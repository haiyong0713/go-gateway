package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	arcMidV1 "go-gateway/app/app-svr/archive/middleware/v1"

	accSvc "git.bilibili.co/bapis/bapis-go/account/service"
	coinSvc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbupSvc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
	grpcmd "google.golang.org/grpc/metadata"
)

type uri string

func (u uri) Args(args ...interface{}) string {
	return fmt.Sprintf(string(u), args...)
}

func (u uri) StatPath(c *bmClient) string {
	return c.composeURI(strings.Replace(string(u), "%d", ":id", -1))
}

func appParams(ctx context.Context) url.Values {
	ret := url.Values{}
	md, ok := grpcmd.FromIncomingContext(ctx)
	if ok {
		const _maxSplit = 2
		if vals := md.Get("authorization"); len(vals) > 0 {
			if valSplit := strings.SplitN(vals[0], " ", _maxSplit); len(valSplit) == _maxSplit {
				ret.Set("access_key", valSplit[1])
			}
		}
	}

	if d, ok := device.FromContext(ctx); ok {
		ret.Set("build", strconv.Itoa(int(d.Build)))
		ret.Set("device", d.Device)
		ret.Set("mobi_app", d.RawMobiApp)
		ret.Set("platform", d.RawPlatform)
		ret.Set("buvid", d.Buvid)
	}
	return ret
}

func musicF(method string, u string) string {
	return "music." + method + "(" + u + ")"
}

type musicPager struct {
	Ctx           context.Context
	PageSize      int
	PageNum       int
	NoPager       bool
	FetchAll      bool
	SimpleFetchFn fetchSimpleAction
	RawFetchFn    fetchRawAction
}

type fetchSimpleAction func(context.Context, string, url.Values, interface{}, ...string) error
type fetchRawAction func(context.Context, *http.Request, interface{}, ...string) error

func (mp *musicPager) defaultPager() {
	if mp.PageNum <= 0 {
		mp.PageNum = 1
	}
	if mp.PageSize <= 0 {
		mp.PageSize = _defaultMusicPageSize
	}
}

func (mp *musicPager) FetchSimple(uri string, param url.Values, dataTyp interface{}, onReceive func(data interface{}) (lastPage int), v ...string) error {
	if !mp.NoPager {
		mp.defaultPager()
		param.Set("page_size", strconv.Itoa(mp.PageSize))
		param.Set("pageSize", strconv.Itoa(mp.PageSize))
		param.Set("page_index", strconv.Itoa(mp.PageNum))
		param.Set("pageIndex", strconv.Itoa(mp.PageNum))
	}

	var rt reflect.Type
	if dataTyp != nil {
		rt = reflect.TypeOf(dataTyp)
	}

	firstResp, err := mp.simpleFetchAndUnmarshalData(mp.Ctx, uri, param, rt, v...)
	if err != nil {
		return err
	}
	lastPage := onReceive(firstResp)
	if !mp.FetchAll || mp.NoPager {
		return nil
	}
	eg := errgroup.WithCancel(mp.Ctx)
	startCh := make(chan struct{}, 1)
	var tmp chan struct{}
	for i := mp.PageNum + 1; i <= lastPage; i++ {
		paramClone := mp.cloneParam(param)
		paramClone.Set("page_index", strconv.Itoa(i))
		paramClone.Set("pageIndex", strconv.Itoa(i))
		// 保持各个goroutine按调用时的顺序回写数据
		var prev, next chan struct{}
		if i > mp.PageNum+1 {
			prev = tmp
		} else {
			prev = startCh
		}
		next = make(chan struct{}, 1)
		tmp = next
		eg.Go(func(c context.Context) error {
			resp, err := mp.simpleFetchAndUnmarshalData(c, uri, paramClone, rt, v...)
			if err != nil {
				return err
			}
			select {
			case <-c.Done():
				return c.Err()
			case <-prev:
			}
			onReceive(resp)
			next <- struct{}{}
			return nil
		})
	}
	startCh <- struct{}{}
	return eg.Wait()
}

func (mp *musicPager) FetchRaw(req *http.Request, dataTyp interface{}, onReceive func(data interface{}), v ...string) error {
	var rt reflect.Type
	if dataTyp != nil {
		rt = reflect.TypeOf(dataTyp)
	}
	resp := &model.BmGenericResp{}
	err := mp.RawFetchFn(mp.Ctx, req, resp, v...)
	if err != nil {
		return wrapDaoError(err, musicF("FetchRaw", req.URL.String()), req)
	}
	if err = resp.IsNormal(); err != nil {
		return errors.WithMessagef(err, "abnormal music response resp(%+v), req(%+v)", resp, req)
	}
	var data interface{}
	if rt == nil {
		return nil
	} else if rt.Kind() == reflect.String {
		data = resp.Data.String()
	} else {
		data = reflect.New(rt).Interface()
		err = json.Unmarshal(resp.Data.Bytes(), data)
	}
	if err != nil {
		return errors.WithMessagef(err, "failed to unmarshal music data(%s) into structs, req(%+v)", string(resp.Data), req)
	}
	onReceive(data)
	return nil
}

func (mp *musicPager) cloneParam(u url.Values) url.Values {
	ret := url.Values{}
	for k, v := range u {
		for _, vs := range v {
			ret.Add(k, vs)
		}
	}
	return ret
}

func (mp *musicPager) simpleFetchAndUnmarshalData(ctx context.Context, uri string, param url.Values, dataTyp reflect.Type, v ...string) (data interface{}, err error) {
	resp := &model.BmGenericResp{}
	err = mp.SimpleFetchFn(ctx, uri, param, resp, v...)
	if err != nil {
		return nil, wrapDaoError(err, musicF("FetchSimple", uri), param)
	}
	if err = resp.IsNormal(); err != nil {
		return nil, errors.WithMessagef(err, "abnormal music response resp(%+v), req(%s), params(%+v)", resp, uri, param)
	}
	if dataTyp == nil {
		data = nil
	} else if dataTyp.Kind() == reflect.String {
		data = resp.Data.String()
	} else {
		data = reflect.New(dataTyp).Interface()
		err = json.Unmarshal(resp.Data.Bytes(), data)
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to unmarshal music data(%s) into structs, req(%s), params(%+v)", string(resp.Data), uri, param)
	}
	return
}

//nolint:deadcode,varcheck,unused
const (
	// 我收藏的歌单/我收藏的合辑列表
	_menuList = uri("/audio/music-service-c/users/%d/menus")
	// 我创建的歌单 mid
	_collectionList = "/audio/music-service-c/collections"

	_collectionNew = "/audio/music-service-c/collections"
	// 编辑我创建的歌单
	_collectionEdit = uri("/audio/music-service-c/collections/%d")
	// 删除我创建的歌单
	_collectionDel = uri("/audio/music-service-c/collections/%d/del")
	// 收藏歌单
	_menuCollAdd = "/audio/music-service-c/menucollect/add"
	// 取消收藏歌单 menuId mid
	_menuCollDel = "/audio/music-service-c/menucollect/del"
	// 计数上报 rid，type，add（+1表示增加，-1表示减少）
	// 单曲播放 type=1 单曲收藏 type=3 单曲分享 type=12
	_clickReport = "/audio/music-service-c/counts/add"

	// 根据menu id获取基本信息
	_menuInfo = uri("/audio/music-service-c/menus/info/%d")
	// 根据menu id查询详情
	_menuDetail = uri("/audio/music-service-c/menus/%d")
	// 获取播放地址 quality|privilege|song_id
	_audioPlayUrl = "/audio/music-service-c/url"
	// 获取单曲播放信息
	_audioPlayingInfo = "/audio/music-service-c/songs/playing"
	// 动态用的根据songid批量获取基本信息 rids
	_audioDynamicByRids = "/x/internal/v1/audio/news/detail"

	// 批量获取单曲基本信息 ids|level(0,1)
	_audioBasicBatch = "/x/internal/v1/audio/songs/batch"
	// 批量获取详细信息 songId
	_audioDetailBatch = "/x/internal/v1/audio/songs/search/baseQuerySongInfo"
	// 根据uid获取用户投稿信息 uid 分页
	_audioListByMid = "/x/internal/v1/audio/songs/create/filter/info"
	// 根据uid获取用户投稿（包括联合投稿）uid
	_audioListByMidV3 = "/audio/music-service-c/songs/internal/upsongslist/v3"

	// 个人歌单的情况 mid
	_personalMenuStatus = "/x/internal/v1/audio/personal/coll"

	_defaultMusicPageSize = 10
)

type MusicMenuCountOpt struct {
	Typ int32
	Mid int64
}

func (d *dao) MusicMenuCount(ctx context.Context, opt MusicMenuCountOpt) (count int64, err error) {
	type menuCountResp struct {
		Total int64 `json:"total"`
	}
	params := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		PageNum: 1, PageSize: 1,
	}
	switch opt.Typ {
	case model.MenuFavored, model.CollectionFavored:
		params.Set("type", strconv.Itoa(int(opt.Typ)))
		err = mp.FetchSimple(_menuList.Args(opt.Mid), params, menuCountResp{}, func(data interface{}) int {
			d := data.(*menuCountResp)
			count = d.Total
			return 0
		}, _menuList.StatPath(d.music))
	case model.MenuCreated:
		params.Set("mid", strconv.Itoa(int(opt.Mid)))
		err = mp.FetchSimple(_collectionList, params, menuCountResp{}, func(data interface{}) int {
			d := data.(*menuCountResp)
			count = d.Total
			return 0
		})
	default:
		return 0, fmt.Errorf("unknown menu typ(%+v)", opt)
	}
	return
}

type MusicMenuListOpt struct {
	Typ      int32
	Mid      int64
	PageNum  int64
	PageSize int64
}

func (d *dao) MusicMenuList(ctx context.Context, opt MusicMenuListOpt) (ret model.MusicMenuList, err error) {
	if opt.PageNum <= 0 {
		opt.PageNum = 1
	}
	type menuListResp struct {
		Total    int64             `json:"total"`
		LastPage int64             `json:"lastPage"`
		List     []model.MusicMenu `json:"list"`
	}
	type collectionListResp struct {
		Total    int64                   `json:"total"`
		LastPage int64                   `json:"lastPage"`
		List     []model.MusicCollection `json:"list"`
	}

	params := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		PageNum: int(opt.PageNum), PageSize: int(opt.PageSize),
	}
	ret.Typ = opt.Typ
	ret.CurrentPageNum = opt.PageNum

	switch opt.Typ {
	case model.MenuFavored, model.CollectionFavored:
		params.Set("type", strconv.Itoa(int(opt.Typ)))
		err = mp.FetchSimple(_menuList.Args(opt.Mid), params, menuListResp{}, func(data interface{}) int {
			d := data.(*menuListResp)
			ret.List = make([]model.MenuItem, 0, len(d.List))
			for _, m := range d.List {
				ret.List = append(ret.List, m.ToMenuItem())
			}
			ret.HasMore = d.LastPage > opt.PageNum
			ret.Total = d.Total
			return int(d.LastPage)
		}, _menuList.StatPath(d.music))
	case model.MenuCreated:
		params.Set("mid", strconv.Itoa(int(opt.Mid)))
		err = mp.FetchSimple(_collectionList, params, collectionListResp{}, func(data interface{}) int {
			d := data.(*collectionListResp)
			// 音频bug， 无数据返回0
			if d.LastPage == 0 {
				d.LastPage = 1
			}
			ret.List = make([]model.MenuItem, 0, len(d.List))
			for _, c := range d.List {
				if c.MenuId <= 0 {
					log.Warnc(ctx, "unexpected MenuId<=0 for music collection(%+v). Discarded", c)
					continue
				}
				ret.List = append(ret.List, c.ToMenuItem())
			}
			ret.HasMore = d.LastPage > opt.PageNum
			ret.Total = d.Total
			return int(d.LastPage)
		})
	default:
		err = fmt.Errorf("unknown menu typ(%+v)", opt)
		return
	}

	return
}

type MusicMenuDetailOpt struct {
	MenuId int64
}

func (d *dao) MusicMenuDetail(ctx context.Context, opt MusicMenuDetailOpt) (ret model.MenuDetail, err error) {
	type menuDetailResp struct {
		Menu  model.MusicMenu    `json:"menusRespones"`
		Songs []model.SongInMenu `json:"songsList"`
	}
	params := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	err = mp.FetchSimple(_menuDetail.Args(opt.MenuId), params, menuDetailResp{}, func(data interface{}) (lastPage int) {
		d := data.(*menuDetailResp)
		if d == nil {
			return
		}
		ret.Menu = d.Menu.ToMenuItem()
		ret.Songs = make([]model.SongItem, 0, len(d.Songs))
		for _, s := range d.Songs {
			ret.Songs = append(ret.Songs, s.ToSongItem())
		}
		return
	}, _menuDetail.StatPath(d.music))
	return
}

type SongPlayingDetailOpt struct {
	Mid        int64
	SongId     int64
	PlayerArgs *arcMidV1.PlayerArgs
	Net        *network.Network
	Dev        *device.Device
}

func (d *dao) SongPlayingDetail(ctx context.Context, opt SongPlayingDetailOpt) (ret model.SongPlayingDetail, err error) {
	type songPlayingResp struct {
		model.SongInPlaying `json:",inline"`
	}
	type songUrlResp struct {
		model.SongUrl `json:",inline"`
	}
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	songid := strconv.Itoa(int(opt.SongId))
	param.Set("song_id", songid)
	param.Set("songid", songid)
	param.Set("mid", strconv.Itoa(int(opt.Mid)))

	// 查询歌曲信息
	err = mp.FetchSimple(_audioPlayingInfo, param, songPlayingResp{}, func(data interface{}) (lastPage int) {
		d := data.(*songPlayingResp)
		ret.Song = d.ToSongItem()
		return
	})
	if err != nil {
		return
	}
	if !ret.Song.IsNormal() {
		err = errors.WithMessagef(ecode.NothingFound, "song is not normal detail(%+v)", ret.Song)
		return
	}

	// 查询播放地址信息
	// TODO: 传递region
	param.Set("quality", strconv.Itoa(int(ret.Song.ChooseQuality(opt.PlayerArgs))))
	param.Set("privilege", "2")
	err = mp.FetchSimple(_audioPlayUrl, param, songUrlResp{}, func(data interface{}) (lastPage int) {
		d := data.(*songUrlResp)
		ret.URL = d.SongUrl
		// 检查mp3问题
		ret.URL.CDNS = ret.URL.CDNS[0:0]
		for _, cdn := range d.SongUrl.CDNS {
			u, e := url.Parse(cdn)
			if e != nil {
				log.Warn("unexpected malform song playURL(%+v). Discarded", d.SongUrl)
				continue
			}
			if !strings.HasSuffix(u.Path, "m4a") {
				log.Warn("unexpected non m4a song(%d) detail(%+v). Discarded", d.SongId, d.SongUrl)
				continue
			}
			ret.URL.CDNS = append(ret.URL.CDNS, cdn)
		}
		return
	})
	if err != nil {
		return
	}

	if len(ret.URL.CDNS) <= 0 {
		err = errors.WithMessagef(ecode.ServerErr, "unexpected zero length song CDNS(%+v)", ret)
	}

	return
}

type SongDetailsOpt struct {
	SongIds interface{}
	Mid     int64
	Net     *network.Network
	Dev     *device.Device
}

//nolint:gocognit
func (d *dao) SongDetails(ctx context.Context, opt SongDetailsOpt) (ret map[int64]model.SongItem, err error) {
	sidSlc, _ := int64IDM2Slc(opt.SongIds)
	sidStrs := strings.Join(sidSlc, ",")

	type songDetailsResp struct {
		List []model.SongInDetail `json:"list"`
	}
	var songDetailList []model.SongInDetail
	var songDynamicResp map[string]model.SongInDynamic

	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, PageSize: 100, SimpleFetchFn: d.music.Get,
	}
	param2 := mp.cloneParam(param)

	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) error {
		param.Set("rids", sidStrs)
		return mp.FetchSimple(_audioDynamicByRids, param, songDynamicResp, func(data interface{}) (lastPage int) {
			d := data.(*map[string]model.SongInDynamic)
			songDynamicResp = *d
			return
		})
	})

	eg.Go(func(c context.Context) error {
		param2.Set("songId", sidStrs)
		return mp.FetchSimple(_audioDetailBatch, param2, songDetailsResp{}, func(data interface{}) (lastPage int) {
			d := data.(*songDetailsResp)
			songDetailList = d.List
			return
		})
	})

	if err = eg.Wait(); err != nil {
		return
	}

	ret = make(map[int64]model.SongItem)
	// 暂存所有 mid获取关注信息
	midm := make(map[int64]struct{})
	sids := make([]int64, 0, len(sidSlc))

	for _, s := range songDynamicResp {
		ret[s.SongId] = s.ToSongItem()
	}
	for _, s := range songDetailList {
		if sr, ok := ret[s.SongId]; ok {
			s.WriteSongItem(&sr)
			ret[s.SongId] = sr
		} else {
			// 大概率是失效稿件
			ret[s.SongId] = s.ToSongItem()
		}
		// 写入相关id 稍后获取额外信息
		midm[s.Mid] = struct{}{}
		sids = append(sids, s.SongId)
	}

	// level2 填充点赞 收藏等信息
	eg2 := errgroup.WithContext(ctx)
	// 无论登录与否
	// 更新点赞数
	if len(sids) > 0 {
		eg2.Go(func(c context.Context) error {
			req := &thumbupSvc.StatsReq{
				Business: ThumbUpBusinessAudio, MessageIds: sids,
				Mid: opt.Mid, IP: opt.Net.RemoteIP,
			}
			res, err := d.thumbupGRPC.Stats(c, req)
			if err != nil {
				return wrapDaoError(err, "thumbupGRPC.Stats", req)
			}
			if res != nil && res.Stats != nil {
				for i := range ret {
					if stat, ok := res.Stats[i]; ok {
						ret[i].Stat.Like = int32(stat.LikeNumber)
						if opt.Mid > 0 {
							// 避免和buvid like race
							ret[i].Stat.HasLike = stat.LikeState == thumbupSvc.State_STATE_LIKE
						}
					}
				}
			}
			return nil
		})
		// 更新投币数
		eg2.Go(func(c context.Context) error {
			res, err := d.CoinNums(c, CoinNumsOpt{
				Business: CoinBusinessAudio, Oids: sids, Net: opt.Net,
			})
			if err != nil {
				return err
			}
			for i := range ret {
				ret[i].Stat.Coin = int32(res[i])
			}
			return nil
		})
	}

	// 登录态
	if opt.Mid > 0 {
		// 批量获取关注关系
		if len(midm) > 0 {
			eg2.Go(func(c context.Context) error {
				authorIDs := make([]int64, 0, len(midm))
				for k := range midm {
					authorIDs = append(authorIDs, k)
				}
				req := &accSvc.RelationsReq{
					Mid: opt.Mid, Owners: authorIDs, RealIp: opt.Net.RemoteIP,
				}
				rels, err := d.accGRPC.Relations3(c, req)
				if err != nil {
					return wrapDaoError(err, "accGRPC.Relations3", req)
				}
				if rels.GetRelations() != nil {
					for i := range ret {
						if rel, ok := rels.GetRelations()[ret[i].Author.GetMid()]; ok {
							if rel.Following {
								ret[i].Author.Relation.Status = v1.FollowRelation_FOLLOWING
							}
						}
					}
				}
				return nil
			})
		}
		if len(sids) > 0 {
			// 批量更新音频投币状态
			eg2.Go(func(c context.Context) error {
				req := &coinSvc.ItemsUserCoinsReq{
					Mid: opt.Mid, Aids: sids, Business: CoinBusinessAudio,
				}
				coins, err := d.coinGRPC.ItemsUserCoins(c, req)
				if err != nil {
					return wrapDaoError(err, "coinGRPC.ItemsUserCoins", req)
				}
				for s, coin := range coins.GetNumbers() {
					ret[s].Stat.HasCoin = coin > 0
				}
				return nil
			})
			// 批量更新音频收藏状态
			eg2.Go(func(c context.Context) error {
				req := &favSvc.IsFavoredsReq{
					Typ:  model.FavTypeAudio,
					Mid:  opt.Mid,
					Oids: sids,
				}
				favs, err := d.favGRPC.IsFavoreds(c, req)
				if err != nil {
					return wrapDaoError(err, "favGRPC.IsFavoreds", req)
				}
				for s, state := range favs.GetFaveds() {
					ret[s].Stat.HasFav = state
				}
				return nil
			})
		}
	} else {
		// 未登录
		// 获取未登录用户的点赞状态
		if len(sids) > 0 {
			eg2.Go(func(c context.Context) error {
				req := &thumbupSvc.BuvidHasLikeReq{
					Buvid: opt.Dev.Buvid, Business: ThumbUpBusinessAudio, MessageIds: sids, IP: opt.Net.RemoteIP,
				}
				thumbs, err := d.thumbupGRPC.BuvidHasLike(c, req)
				if err != nil {
					return wrapDaoError(err, "thumbupGRPC.BuvidHasLike", req)
				}
				for s, state := range thumbs.States {
					ret[s].Stat.HasLike = state.State == thumbupSvc.State_STATE_LIKE
				}
				return nil
			})
		}
	}

	sideErr := eg2.Wait()
	if sideErr != nil {
		log.Warnc(ctx, "failed to get associated info for SongDetail: %v Discarded", sideErr)
	}

	return
}

type SongInfosOpt struct {
	SongIds  interface{}
	RemoteIP string
}

func (d *dao) SongInfos(ctx context.Context, opt SongInfosOpt) (ret map[int64]model.SongItem, err error) {
	sidSlc, _ := int64IDM2Slc(opt.SongIds)
	sidStrs := strings.Join(sidSlc, ",")

	type songDetailsResp struct {
		List []model.SongInDetail `json:"list"`
	}
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, PageSize: 100, SimpleFetchFn: d.music.Get,
	}
	param.Set("songId", sidStrs)

	var list []model.SongInDetail
	err = mp.FetchSimple(_audioDetailBatch, param, songDetailsResp{}, func(data interface{}) (lastPage int) {
		d := data.(*songDetailsResp)
		list = d.List
		return
	})
	if err != nil {
		return
	}

	ret = make(map[int64]model.SongItem)
	for _, s := range list {
		ret[s.SongId] = s.ToSongItem()
	}
	return
}

type SpaceSongListOpt struct {
	Mid              int64
	WithCollaborator bool // 是否包括联合投稿
}

func (d *dao) SpaceSongList(ctx context.Context, opt SpaceSongListOpt) (ret []model.SongItem, err error) {
	type spaceSongListResp struct {
		List     []model.SongInSpace `json:"data"`
		Total    int64               `json:"total"`
		PageSize int64               `json:"pageSize"`
	}
	type spaceSongListV3Resp struct {
		List     []model.SongInSpaceV3 `json:"list"`
		Total    int64                 `json:"total"`
		PageSize int64                 `json:"pageSize"`
	}
	ret = make([]model.SongItem, 0, 30)
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		PageSize: 9999, FetchAll: true,
	}
	param.Set("uid", strconv.FormatInt(opt.Mid, 10))
	if opt.WithCollaborator {
		// 这个接口有性能问题，按小数据量分页来
		// 端上使用的页面大小是20，所以这边用同样的pageSize理论上可以复用端上打开该页面时产生的缓存
		mp.PageSize = 20
		err = mp.FetchSimple(_audioListByMidV3, param, spaceSongListV3Resp{}, func(data interface{}) (lastPage int) {
			d := data.(*spaceSongListV3Resp)
			for _, s := range d.List {
				ret = append(ret, s.ToSongItem())
			}
			return int((d.Total / d.PageSize) + 1)
		})
	} else {
		// 这个接口直接大批量拉数据即可
		err = mp.FetchSimple(_audioListByMid, param, spaceSongListResp{}, func(data interface{}) (lastPage int) {
			d := data.(*spaceSongListResp)
			for _, s := range d.List {
				ret = append(ret, s.ToSongItem())
			}
			return int((d.Total / d.PageSize) + 1)
		})
	}
	return
}

type PersonalMenuStatusOpt struct {
	Mid int64
}

func (d *dao) PersonalMenuStatus(ctx context.Context, opt PersonalMenuStatusOpt) (ret model.PersonalMenuStatus, err error) {
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	param.Set("mid", strconv.FormatInt(opt.Mid, 10))
	err = mp.FetchSimple(_personalMenuStatus, param, model.PersonalMenuStatus{}, func(data interface{}) (lastPage int) {
		d := data.(*model.PersonalMenuStatus)
		ret = *d
		return
	})
	return
}

type MenuEditOpt struct {
	MenuId      int64
	Mid         int64
	Title, Desc string
	IsOpen      int32
}

func (d *dao) MenuEdit(ctx context.Context, opt MenuEditOpt) (err error) {
	collectionId, err := d.menuId2CollectionId(ctx, opt.MenuId)
	if err != nil {
		return
	}
	type menuEditResp struct {
		CollectionId int64 `json:"collection_id"`
		MenuId       int64 `json:"menu_id"`
		Mid          int64 `json:"mid"`
	}
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Post,
		NoPager: true,
	}
	param.Set("collection_id", strconv.FormatInt(collectionId, 10))
	param.Set("is_open", strconv.Itoa(int(opt.IsOpen)))
	param.Set("title", opt.Title)
	param.Set("desc", opt.Desc)
	param.Set("mid", strconv.FormatInt(opt.Mid, 10))
	err = mp.FetchSimple(_collectionEdit.Args(collectionId), param, menuEditResp{}, func(data interface{}) (lastPage int) {
		// no need to read the resp data
		return
	}, _collectionEdit.StatPath(d.music))
	return
}

type MenuDelOpt struct {
	MenuId int64
	Mid    int64
}

func (d *dao) MenuDel(ctx context.Context, opt MenuDelOpt) (err error) {
	collectionId, err := d.menuId2CollectionId(ctx, opt.MenuId)
	if err != nil {
		return
	}
	type menuDelResp struct {
		CollectionId int64 `json:"id"`
	}
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Post,
		NoPager: true,
	}
	param.Set("mid", strconv.FormatInt(opt.Mid, 10))
	param.Set("collection_id", strconv.FormatInt(collectionId, 10))
	err = mp.FetchSimple(_collectionDel.Args(collectionId), param, menuDelResp{}, func(data interface{}) (lastPage int) {
		return
	}, _collectionDel.StatPath(d.music))
	return
}

type MenuAddOpt struct {
	Title, Desc string
	IsOpen      int32
}

func (d *dao) MenuAdd(ctx context.Context, opt MenuAddOpt) (menuId int64, err error) {
	type menuAddResp struct {
		CollectionId int64 `json:"collection_id"`
		MenuId       int64 `json:"menu_id"`
		Mid          int64 `json:"mid"`
	}
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Post,
		NoPager: true,
	}
	param.Set("is_open", strconv.Itoa(int(opt.IsOpen)))
	param.Set("title", opt.Title)
	param.Set("desc", opt.Desc)
	err = mp.FetchSimple(_collectionNew, param, menuAddResp{}, func(data interface{}) (lastPage int) {
		d := data.(*menuAddResp)
		menuId = d.MenuId
		return
	})
	return
}

func (d *dao) menuId2CollectionId(ctx context.Context, menuId int64) (collectionId int64, err error) {
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	param.Set("menusId", strconv.FormatInt(menuId, 10))
	err = mp.FetchSimple(_menuInfo.Args(menuId), param, model.MusicMenu{}, func(data interface{}) (lastPage int) {
		d := data.(*model.MusicMenu)
		collectionId = d.CollectionId
		return
	}, _menuInfo.StatPath(d.music))
	if collectionId == 0 {
		return 0, errors.WithMessagef(ecode.ServerErr, "no collection id found for menu(%d)", menuId)
	}
	return
}

type MenuCollAddOpt struct {
	MenuId int64
	Mid    int64
}

func (d *dao) MenuCollAdd(ctx context.Context, opt MenuCollAddOpt) (err error) {
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	param.Set("menuId", strconv.FormatInt(opt.MenuId, 10))
	param.Set("mid", strconv.FormatInt(opt.Mid, 10))
	err = mp.FetchSimple(_menuCollAdd, param, nil, func(data interface{}) (lastPage int) {
		return
	})
	return
}

type MenuCollDelOpt struct {
	MenuId int64
	Mid    int64
}

func (d *dao) MenuCollDel(ctx context.Context, opt MenuCollDelOpt) (err error) {
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, SimpleFetchFn: d.music.Get,
		NoPager: true,
	}
	param.Set("menuId", strconv.FormatInt(opt.MenuId, 10))
	param.Set("mid", strconv.FormatInt(opt.Mid, 10))
	err = mp.FetchSimple(_menuCollDel, param, nil, func(data interface{}) (lastPage int) {
		return
	})
	return
}

const (
	MusicClickShare = 12
	MusicClickPlay  = 1
	MusicClickFav   = 3
)

type MusicClickReportOpt struct {
	ClickTyp  int
	SongId    int64
	AddMetric bool
}

func (d *dao) MusicClickReport(ctx context.Context, opt MusicClickReportOpt) (err error) {
	param := appParams(ctx)
	mp := &musicPager{
		Ctx: ctx, RawFetchFn: d.music.RawJSON,
	}
	type reportData struct {
		Rid   int64 `json:"rid"`
		Typ   int   `json:"type"`
		Count int   `json:"count"`
	}
	body := reportData{
		Rid: opt.SongId,
		Typ: opt.ClickTyp,
	}
	if opt.AddMetric {
		body.Count = 1
	} else {
		body.Count = -1
	}
	bodyData, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, d.music.composeURI(_clickReport), bytes.NewReader(bodyData))
	if err != nil {
		return fmt.Errorf("failed to build report request(%s), body(%s)", _clickReport, string(bodyData))
	}
	req.URL.RawQuery = param.Encode()
	req.Header.Set("Content-Type", "application/json")
	err = mp.FetchRaw(req, nil, func(data interface{}) {
		// nothing to do
	})
	return
}

// PersonalMenuStatusV1 done checked
func (d *dao) PersonalMenuStatusV1(ctx context.Context, opt PersonalMenuStatusOpt) (ret model.PersonalMenuStatus, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.PersonalMenuStatus(ctx, opt)
	}
	const (
		MenuV1        = "menu"
		MenuCreatedV1 = "menu_created"
		CollectionV1  = "collection"
	)

	req := &listenerSvc.GetLikeMenusStatusReq{
		Mid: opt.Mid,
	}
	res, err := d.musicGRPC.GetLikeMenusStatus(ctx, req)
	if err != nil {
		return model.PersonalMenuStatus{}, wrapDaoError(err, "musicGRPC.GetLikeMenusStatus", req)
	}
	if res.GetStatus() == nil {
		return model.PersonalMenuStatus{}, fmt.Errorf("listener svc failed to GetLikeMenusStatus")
	}
	ret = model.PersonalMenuStatus{
		HasMenu:        res.Status[MenuV1],
		HasMenuCreated: res.Status[MenuCreatedV1],
		HasCollection:  res.Status[CollectionV1],
	}
	return
}

// MusicMenuListV1 done  miss IsOff
func (d *dao) MusicMenuListV1(ctx context.Context, opt MusicMenuListOpt) (ret model.MusicMenuList, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.MusicMenuList(ctx, opt)
	}
	const (
		DefaultPageNumV1  = 1
		DefaultPageSizeV1 = 10
	)

	if opt.PageNum <= 0 {
		opt.PageNum = DefaultPageNumV1
	}
	if opt.PageSize <= 0 {
		opt.PageSize = DefaultPageSizeV1
	}

	req := &listenerSvc.GetLikeMenusReq{
		Mid:      opt.Mid,
		MenuType: int64(opt.Typ),
		PageNum:  opt.PageNum,
		PageSize: opt.PageSize,
	}
	res, err := d.musicGRPC.GetLikeMenus(ctx, req)
	if err != nil {
		return model.MusicMenuList{}, wrapDaoError(err, "musicGRPC.GetLikeMenusReq", req)
	}
	if res.GetData() == nil {
		return model.MusicMenuList{}, fmt.Errorf("listener svc failed to GetLikeMenusReq")
	}

	ret.List = make([]model.MenuItem, 0, len(res.Data))
	for _, m := range res.Data {
		if m.MenuId <= 0 {
			log.Warnc(ctx, "unexpected MenuId<=0 for music collection(%+v). Discarded", m)
			continue
		}

		ms := model.MenuSongItem{
			Menu: m,
		}
		menuItem := ms.ToMenuItem(opt.Typ)
		ret.List = append(ret.List, menuItem)
	}

	lastPage := res.Total / res.PageSize
	if res.Total%res.PageSize != 0 {
		lastPage = lastPage + 1
	}
	ret.HasMore = lastPage > opt.PageNum
	ret.Typ = opt.Typ
	ret.CurrentPageNum = opt.PageNum
	ret.Total = res.Total
	return
}

// MusicMenuDetailV1 done checked
func (d *dao) MusicMenuDetailV1(ctx context.Context, opt MusicMenuDetailOpt) (ret model.MenuDetail, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.MusicMenuDetail(ctx, opt)
	}
	req := &listenerSvc.GetMenuInfoReq{
		MenuId: opt.MenuId,
	}
	res, err := d.musicGRPC.GetMenuInfo(ctx, req)
	if err != nil {
		return model.MenuDetail{}, wrapDaoError(err, "musicGRPC.GetMenuInfo", req)
	}
	if res.GetData() == nil {
		return model.MenuDetail{}, fmt.Errorf("listener svc failed to GetMenuInfo")
	}
	data := res.Data
	ms := model.MenuSongItem{
		Menu: res.Data,
	}
	ret.Menu = ms.ToMenuItem(0)

	ret.Songs = make([]model.SongItem, 0, len(data.Songs))
	for _, s := range data.Songs {
		ms := model.MenuSongItem{
			Song: s,
		}
		songItem := ms.ToSongItem()
		//miss Author MaxQuality PlayNum , don't need here , so cut
		ret.Songs = append(ret.Songs, songItem)
	}
	return
}

// SongPlayingDetailV1 done checked
func (d *dao) SongPlayingDetailV1(ctx context.Context, opt SongPlayingDetailOpt) (ret model.SongPlayingDetail, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.SongPlayingDetail(ctx, opt)
	}
	req := &listenerSvc.GetPlayingSongReq{
		Mid:    opt.Mid,
		SongId: opt.SongId,
	}
	res, err := d.musicGRPC.GetPlayingSong(ctx, req)
	if err != nil {
		return model.SongPlayingDetail{}, wrapDaoError(err, "musicGRPC.GetPlayingSong", req)
	}
	if res.GetSong() == nil {
		return model.SongPlayingDetail{}, fmt.Errorf("listener svc failed to GetPlayingSong")
	}
	ms := model.MenuSongItem{
		Song: res.Song,
	}
	songItem := ms.ToSongItem()
	// miss Author, judge whether need, seems like not, so cut

	// convert to Qn
	songItem.Qn = make([]model.SongQn, 0, len(res.Qualities))
	qualities := res.Qualities
	for _, qn := range qualities {
		songItem.Qn = append(songItem.Qn, model.SongQn{
			Typ:  int32(qn.Type),
			Desc: qn.Desc,
			Bps:  qn.Bps,
			Size: qn.Size_,
		})
		if qn.Type > int64(songItem.MaxQuality) {
			songItem.MaxQuality = int32(qn.Type)
		}
	}
	ret.Song = songItem

	// get play url
	urlReq := &listenerSvc.GetPlayUrlReq{
		SongId:   opt.SongId,
		Platform: opt.Dev.RawPlatform,
		Ip:       opt.Net.RemoteIP,
		Type:     int64(ret.Song.ChooseQuality(opt.PlayerArgs)),
	}
	urlRes, urlErr := d.musicGRPC.GetPlayUrl(ctx, urlReq)
	if urlErr != nil {
		return model.SongPlayingDetail{}, wrapDaoError(urlErr, "musicGRPC.GetPlayUrl", urlReq)
	}
	if !urlRes.GetSuccess() {
		return model.SongPlayingDetail{}, fmt.Errorf("listener svc failed to GetPlayUrl")
	}
	urlTemp := model.SongUrl{
		SongId:  urlRes.SongId,
		Size:    urlRes.Size_,
		Timeout: urlRes.Timestamp,
	}
	for _, value := range urlRes.Url {
		u, e := url.Parse(value)
		if e != nil {
			log.Warnc(ctx, "unexpected malform song playURL(%+v). Discarded", urlTemp)
			continue
		}
		if !strings.HasSuffix(u.Path, "m4a") {
			log.Warnc(ctx, "unexpected non m4a song(%d) detail(%+v). Discarded", urlTemp.SongId, urlTemp)
			continue
		}
		urlTemp.CDNS = append(urlTemp.CDNS, value)
	}
	if len(urlTemp.CDNS) <= 0 {
		err = errors.WithMessagef(ecode.ServerErr, "unexpected zero length song CDNS(%+v)", ret)
	}
	ret.URL = urlTemp
	return
}

// SongDetailsV1 done
//
//nolint:gocognit
func (d *dao) SongDetailsV1(ctx context.Context, opt SongDetailsOpt) (ret map[int64]model.SongItem, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.SongDetails(ctx, opt)
	}
	_, sidSlc := int64IDM2Slc(opt.SongIds)
	req := &listenerSvc.GetSongsReq{
		SongIds: sidSlc,
	}
	res, err := d.musicGRPC.GetSongs(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "musicGRPC.GetSongs", req)
	}
	if res.GetData() == nil {
		return make(map[int64]model.SongItem), nil
	}

	data := res.Data
	ret = make(map[int64]model.SongItem)
	// 暂存所有 mid获取关注信息
	midm := make(map[int64]struct{})
	for id, s := range data {
		ms := model.MenuSongItem{
			Song: s,
		}
		songItem := ms.ToSongItem()
		// 写入相关id 稍后获取额外信息
		midm[s.Mid] = struct{}{}
		ret[id] = songItem
	}

	//填充点赞 收藏等信息
	eg := errgroup.WithContext(ctx)
	sids := sidSlc

	// 登录态
	if opt.Mid > 0 {
		if len(midm) > 0 {
			// 批量获取up主详细信息
			eg.Go(func(ctx context.Context) error {
				upInfos, err := d.UpInfoByMids(ctx, midm, opt.Net.RemoteIP)
				if err != nil {
					return err
				}
				for i := range ret {
					if upInfo, ok := upInfos[ret[i].Author.GetMid()]; ok {
						au := ret[i].Author
						au.Name, au.Avatar = upInfo.GetName(), upInfo.GetFace()
					}
				}
				return nil
			})
			// 批量获取关注关系
			eg.Go(func(c context.Context) error {
				authorIDs := make([]int64, 0, len(midm))
				for k := range midm {
					authorIDs = append(authorIDs, k)
				}
				req := &accSvc.RelationsReq{
					Mid: opt.Mid, Owners: authorIDs, RealIp: opt.Net.RemoteIP,
				}
				rels, err := d.accGRPC.Relations3(c, req)
				if err != nil {
					return wrapDaoError(err, "accGRPC.Relations3", req)
				}
				if rels.GetRelations() != nil {
					for i := range ret {
						if rel, ok := rels.GetRelations()[ret[i].Author.GetMid()]; ok {
							if rel.Following {
								ret[i].Author.Relation.Status = v1.FollowRelation_FOLLOWING
							}
						}
					}
				}
				return nil
			})
		}
		if len(sids) > 0 {
			// 批量更新音频点赞状态
			eg.Go(func(c context.Context) error {
				req := &thumbupSvc.HasLikeReq{
					Mid: opt.Mid, Business: ThumbUpBusinessAudio, MessageIds: sids, IP: opt.Net.RemoteIP,
				}
				thumbs, err := d.thumbupGRPC.HasLike(c, req)
				if err != nil {
					return wrapDaoError(err, "thumbupGRPC.HasLike", req)
				}
				if thumbs.States == nil {
					return nil
				}
				for i := range ret {
					if state, ok := thumbs.States[i]; ok {
						ret[i].Stat.HasLike = state.State == thumbupSvc.State_STATE_LIKE
					}
				}
				return nil
			})
			// 批量更新音频投币状态
			eg.Go(func(c context.Context) error {
				req := &coinSvc.ItemsUserCoinsReq{
					Mid: opt.Mid, Aids: sids, Business: CoinBusinessAudio,
				}
				coins, err := d.coinGRPC.ItemsUserCoins(c, req)
				if err != nil {
					return wrapDaoError(err, "coinGRPC.ItemsUserCoins", req)
				}
				if coins.Numbers == nil {
					return nil
				}
				for i := range ret {
					if coin, ok := coins.Numbers[i]; ok {
						ret[i].Stat.HasCoin = coin > 0
					}
				}
				return nil
			})
			// 批量更新音频收藏状态
			eg.Go(func(c context.Context) error {
				req := &favSvc.IsFavoredsReq{
					Typ:  model.FavTypeAudio,
					Mid:  opt.Mid,
					Oids: sids,
				}
				favs, err := d.favGRPC.IsFavoreds(c, req)
				if err != nil {
					return wrapDaoError(err, "favGRPC.IsFavoreds", req)
				}
				if favs.Faveds == nil {
					return nil
				}
				for i := range ret {
					if fav, ok := favs.Faveds[i]; ok {
						ret[i].Stat.HasFav = fav
					}
				}
				return nil
			})
		}
	} else {
		// 未登录
		// 获取未登录用户的点赞状态
		if len(sids) > 0 {
			eg.Go(func(c context.Context) error {
				req := &thumbupSvc.BuvidHasLikeReq{
					Buvid: opt.Dev.Buvid, Business: ThumbUpBusinessAudio, MessageIds: sids, IP: opt.Net.RemoteIP,
				}
				thumbs, err := d.thumbupGRPC.BuvidHasLike(c, req)
				if err != nil {
					return wrapDaoError(err, "thumbupGRPC.BuvidHasLike", req)
				}
				if thumbs.States == nil {
					return nil
				}
				for i := range ret {
					if state, ok := thumbs.States[i]; ok {
						ret[i].Stat.HasLike = state.State == thumbupSvc.State_STATE_LIKE
					}
				}
				return nil
			})
		}
	}

	sideErr := eg.Wait()
	if sideErr != nil {
		log.Warnc(ctx, "failed to get associated info for SongDetail: %v Discarded", sideErr)
	}

	return
}

// SongInfosV1  done checked
func (d *dao) SongInfosV1(ctx context.Context, opt SongInfosOpt) (ret map[int64]model.SongItem, err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.SongInfos(ctx, opt)
	}
	_, sidSlc := int64IDM2Slc(opt.SongIds)
	req := &listenerSvc.GetSongsReq{
		SongIds: sidSlc,
	}
	res, err := d.musicGRPC.GetSongs(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "musicGRPC.GetSongs", req)
	}
	if res.GetData() == nil {
		return make(map[int64]model.SongItem), nil
	}

	data := res.Data
	midm := make(map[int64]struct{})
	ret = make(map[int64]model.SongItem)
	for id, s := range data {
		ms := model.MenuSongItem{
			Song: s,
		}
		songItem := ms.ToSongItem()
		ret[id] = songItem
		midm[s.Mid] = struct{}{}
	}
	if len(midm) > 0 {
		upInfos, err := d.UpInfoByMids(ctx, midm, opt.RemoteIP)
		if err != nil {
			return nil, err
		}
		for _, s := range ret {
			if up, ok := upInfos[s.Author.GetMid()]; ok {
				s.Author.Name = up.Name
				s.Author.Avatar = up.Face
			}
		}
	}

	return
}

// MenuEditV1 done checked
func (d *dao) MenuEditV1(ctx context.Context, opt MenuEditOpt) (err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.MenuEdit(ctx, opt)
	}
	req := &listenerSvc.UpdateUserMenuReq{
		MenuId: opt.MenuId,
		Mid:    opt.Mid,
		Title:  opt.Title,
		Open:   opt.IsOpen == 1,
		Desc:   opt.Desc,
	}

	res, err := d.musicGRPC.UpdateUserMenu(ctx, req)
	if err != nil {
		return wrapDaoError(err, "musicGRPC.UpdateUserMenu", req)
	}
	if !res.GetSuccess() {
		return fmt.Errorf("listener svc failed to UpdateUserMenu")
	}
	return
}

// MenuDelV1 done checked
func (d *dao) MenuDelV1(ctx context.Context, opt MenuDelOpt) (err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.MenuDel(ctx, opt)
	}
	req := &listenerSvc.DeleteUserMenuReq{
		MenuId: opt.MenuId,
		Mid:    opt.Mid,
	}
	res, err := d.musicGRPC.DeleteUserMenu(ctx, req)
	if err != nil {
		return wrapDaoError(err, "musicGRPC.DeleteUserMenu", req)
	}
	if !res.GetSuccess() {
		return fmt.Errorf("listener svc failed to DeleteUserMenu")
	}
	return
}

// MenuCollDelV1 done checked
func (d *dao) MenuCollDelV1(ctx context.Context, opt MenuCollDelOpt) (err error) {
	if conf.C.Switch.LegacyMusicAPI {
		return d.MenuCollDel(ctx, opt)
	}
	req := &listenerSvc.DelLikeMenuReq{
		MenuId: opt.MenuId,
		Mid:    opt.Mid,
	}
	res, err := d.musicGRPC.DelLikeMenu(ctx, req)
	if err != nil {
		return wrapDaoError(err, "musicGRPC.DelLikeMenu", req)
	}
	if !res.GetSuccess() {
		return fmt.Errorf("listener svc failed to DelLikeMenu")
	}
	return
}
