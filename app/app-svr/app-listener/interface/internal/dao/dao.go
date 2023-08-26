package dao

import (
	"context"
	"net/http"
	"net/url"
	"reflect"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/silverbullet/gaia"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	purlSvcV2 "go-gateway/app/app-svr/playurl/service/api/v2"

	arcSvc "go-gateway/app/app-svr/archive/service/api"

	accSvc "git.bilibili.co/bapis/bapis-go/account/service"
	hisSvc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	coinSvc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbupSvc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	copyrightSvc "git.bilibili.co/bapis/bapis-go/copyright-manage/interface"
	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	epCardSvc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	ogvEpisodeSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	ogvSeasonSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	ugcSeasonSvc "git.bilibili.co/bapis/bapis-go/ugc-season/service"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Dao dao interface
type Dao interface {
	// acc
	UpInfoByMids(ctx context.Context, mids interface{}, ip string) (ret map[int64]model.MemberInfo, err error)
	UpInfoStatByMid(ctx context.Context, mid int64, ip string) (ret *model.MemberInfo, err error)

	// UGC 稿件服务
	ArchiveDetails(ctx context.Context, opt ArcDetailsOpt) (map[int64]model.ArchiveDetail, error)
	ArcPlayUrl(ctx context.Context, opt ArcPlayUrlOpt) (map[int64]model.PlayUrlInfo, error)
	ArchiveInfos(ctx context.Context, opt ArchiveInfoOpt) (map[int64]model.ArchiveInfo, error)
	FilterArchives(ctx context.Context, aids []int64) (map[int64]string, error)

	// OGV 相关服务
	SeasonDetails(ctx context.Context, opt SeasonDetailsOpt) (map[int32]model.SeasonDetail, error)
	Epids2Aids(ctx context.Context, epids []int32) (map[int32]int64, error)
	EpisodeDetails(ctx context.Context, opt EpisodeDetailsOpt) (map[int32]model.EpisodeDetail, error)
	OGVEpCards(ctx context.Context, opt OGVEpCardOpt) (map[int32]model.EpCard, error)

	// 播放历史
	PlayHistory(ctx context.Context, mid int64, dev *device.Device) ([]model.PlayHistory, error)
	PlayHistoryAdd(ctx context.Context, opt PlayHistoryAddOpt) error
	PlayHistoryDelete(ctx context.Context, opt PlayHistoryDeleteOpt) error
	PlayHistoryTruncate(ctx context.Context, mid int64, buvid string) error
	PlayHisoryByItemID(ctx context.Context, mid int64, buvid string, item *v1.PlayItem) (subid int64, progress int64, err error)
	PlayHisoryByItemIDs(ctx context.Context, mid int64, buvid string, items ...*v1.PlayItem) (ret map[string]PlayHistoryResult, err error)
	PlayActionReport(ctx context.Context, opt PlayActionReportOpt) error

	// 播放列表
	Playlist(ctx context.Context, mid int64, buvid string) (model.Playlist, error)
	PlaylistAdd(ctx context.Context, opt PlaylistAddOpt) error
	PlaylistDelete(ctx context.Context, opt PlaylistDeleteOpt) error
	PlaylistTruncate(ctx context.Context, mid int64, buvid string) error
	PlaylistReplace(ctx context.Context, opt PlaylistReplaceOpt) (rets []*v1.PlayItem, err error)

	// 交互
	CoinAdd(ctx context.Context, opt CoinAddOpt) error
	ThumbAction(ctx context.Context, opt ThumbActionOpt) error

	// 收藏
	FavFolderList(ctx context.Context, opt FavFolderListOpt) (ret []model.FavFolder, err error)
	FavFoldersInfo(ctx context.Context, opt FavFoldersInfoOpt) (ret map[string]model.FavFolder, err error)
	FavFoldersDetail(ctx context.Context, opt FavFolderDetailsOpt) (ret map[string][]model.FavItemDetail, err error)
	FavFolderDetailPaged(ctx context.Context, opt FavFolderDetailPagedOpt) (ret *FavFolderDetailPagedResp, err error)
	FavFolderDetail(ctx context.Context, opt FavFolderDetailOpt) (ret []model.FavItemDetail, err error)
	UgcSeasonsInfo(ctx context.Context, ss []int64) (ret map[int64]*ugcSeasonSvc.Season, err error)
	UgcSeasonDetail(ctx context.Context, ss int64) (model.UgcSeasonDetail, error)

	FavFolderCreate(ctx context.Context, opt FavFolderCreateOpt) (model.FavFolderMeta, error)
	FavFolderDelete(ctx context.Context, opt FavFolderDeleteOpt) error

	FavItemAdd(ctx context.Context, opt FavItemAddOpt) (err error)
	FavItemDelete(ctx context.Context, opt FavItemDeleteOpt) (err error)
	FavoredInFolders(ctx context.Context, opt FavoredInFoldersOpt) ([]model.FavFolderMeta, error)

	RecommendArchives(ctx context.Context, opt RecommendArchivesOpt) (RecommendArchivesRes, error)
	RcmdTopCards(ctx context.Context, opt RcmdTopCardsOpt) ([]model.RcmdTopCard, error)

	// 发现
	PickCards(ctx context.Context, opt PickCardsOpt) ([]model.SinglePick, int64, error)
	CardDetail(ctx context.Context, opt CardDetailsOpt) (model.SingleCollection, error)

	// 版权
	CopyrightBans(ctx context.Context, opt CopyrightBansOpt) (map[int64]bool, error)

	// 老音频
	MusicMenuCount(ctx context.Context, opt MusicMenuCountOpt) (count int64, err error)
	MusicMenuList(ctx context.Context, opt MusicMenuListOpt) (ret model.MusicMenuList, err error)
	SongPlayingDetail(ctx context.Context, opt SongPlayingDetailOpt) (ret model.SongPlayingDetail, err error)
	MusicMenuDetail(ctx context.Context, opt MusicMenuDetailOpt) (ret model.MenuDetail, err error)
	SongDetails(ctx context.Context, opt SongDetailsOpt) (map[int64]model.SongItem, error)
	SongInfos(ctx context.Context, opt SongInfosOpt) (ret map[int64]model.SongItem, err error) // 简单信息
	SpaceSongList(ctx context.Context, opt SpaceSongListOpt) (ret []model.SongItem, err error)
	PersonalMenuStatus(ctx context.Context, opt PersonalMenuStatusOpt) (ret model.PersonalMenuStatus, err error)
	MenuEdit(ctx context.Context, opt MenuEditOpt) (err error)
	MenuDel(ctx context.Context, opt MenuDelOpt) (err error)
	MenuCollAdd(ctx context.Context, opt MenuCollAddOpt) (err error)
	MenuCollDel(ctx context.Context, opt MenuCollDelOpt) (err error)
	MusicClickReport(ctx context.Context, opt MusicClickReportOpt) error
	GuideBarShowReport(ctx context.Context, opt GuideBarShowReportOpt) (success bool, err error)

	// 音频，rpc
	PersonalMenuStatusV1(ctx context.Context, opt PersonalMenuStatusOpt) (ret model.PersonalMenuStatus, err error)
	MusicMenuListV1(ctx context.Context, opt MusicMenuListOpt) (ret model.MusicMenuList, err error)
	MusicMenuDetailV1(ctx context.Context, opt MusicMenuDetailOpt) (ret model.MenuDetail, err error)

	SongPlayingDetailV1(ctx context.Context, opt SongPlayingDetailOpt) (ret model.SongPlayingDetail, err error)
	SongDetailsV1(ctx context.Context, opt SongDetailsOpt) (map[int64]model.SongItem, error)
	SongInfosV1(ctx context.Context, opt SongInfosOpt) (ret map[int64]model.SongItem, err error) // 简单信息

	MenuEditV1(ctx context.Context, opt MenuEditOpt) (err error)
	MenuDelV1(ctx context.Context, opt MenuDelOpt) (err error)
	MenuCollDelV1(ctx context.Context, opt MenuCollDelOpt) (err error)

	// 播单
	MediaListDetail(ctx context.Context, opt MediaListDetailOpt) ([]model.MediaListItem, error)
	MediaListPaged(ctx context.Context, opt MediaListPagedOpt) (resp *MediaListPagedResp, err error)

	Close()
	Ping(ctx context.Context) (err error)
}

// dao dao.
type dao struct {
	// 播客专用后端服务
	listenerGRPC listenerSvc.ListenerSvrClient

	// 音频服务
	musicGRPC listenerSvc.MusicSvrClient
	// 旧音频
	music *bmClient

	// 风控
	silverBullet gaia.EngineInterface

	// 稿件服务
	arcGRPC arcSvc.ArchiveClient
	// 硬币服务 GRPC
	coinGRPC coinSvc.CoinClient
	// 硬币网关 HTTP
	coinHTTP *bmClient
	// 点赞服务
	thumbupGRPC thumbupSvc.ThumbupClient
	// 播放地址解析服务
	playUrlV2GRPC purlSvcV2.PlayURLClient
	// 账号服务
	accGRPC accSvc.AccountClient
	// 社区（主站）播放历史服务
	hisGRPC hisSvc.HistoryClient
	// ogv season 服务
	ogvSeasonGRPC ogvSeasonSvc.SeasonClient
	// ogv episode 服务
	ogvEpisodeGRPC ogvEpisodeSvc.EpisodeClient
	// 主站收藏服务
	favGRPC favSvc.FavoriteClient
	// UGC合集服务
	ugcSeasonGRPC ugcSeasonSvc.UGCSeasonClient
	// 版权中台
	copyrightGRPC copyrightSvc.CopyrightManageClient
	// 播单
	mediaListHTTP *bmClient
	// ep详情
	epCardGRPC epCardSvc.CardClient
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	d = &dao{}
	appConf := &paladin.TOML{}
	if err = paladin.Get("db.toml").Unmarshal(appConf); err != nil {
		return
	}
	daoConf := new(paladin.TOML)
	if err = appConf.Get("Dao").Unmarshal(daoConf); err != nil {
		if err != paladin.ErrNotExist {
			return
		}
		// silence the error
		err = nil
		daoConf = nil
	}
	// 播客专用后端 GRPC
	d.listenerGRPC = newDaoClient(d, "listenerGRPC", listenerSvc.NewClient, daoConf).(listenerSvc.ListenerSvrClient)

	// 风控
	silver := new(struct {
		Enable bool
		Config *gaia.Config
	})
	if err = appConf.Get("SilverBullet").UnmarshalTOML(silver); err != nil {
		if err != paladin.ErrNotExist {
			return
		}
		err = nil
	} else {
		if silver.Enable {
			d.silverBullet, err = gaia.New(silver.Config)
			if err != nil {
				return
			}
		}
	}

	// 其他GRPC服务
	d.musicGRPC = newDaoClient(d, "musicGRPC", listenerSvc.NewClientMusicSvr, daoConf).(listenerSvc.MusicSvrClient)
	d.arcGRPC = newDaoClient(d, "arcGRPC", arcSvc.NewClient, daoConf).(arcSvc.ArchiveClient)
	d.coinGRPC = newDaoClient(d, "coinGRPC", coinSvc.NewClient, daoConf).(coinSvc.CoinClient)
	d.thumbupGRPC = newDaoClient(d, "thumbupGRPC", thumbupSvc.NewClient, daoConf).(thumbupSvc.ThumbupClient)
	d.playUrlV2GRPC = newDaoClient(d, "playUrlV2GRPC", purlSvcV2.NewClient, daoConf).(purlSvcV2.PlayURLClient)
	d.accGRPC = newDaoClient(d, "accGRPC", accSvc.NewClient, daoConf).(accSvc.AccountClient)
	d.hisGRPC = newDaoClient(d, "hisGRPC", hisSvc.NewClient, daoConf).(hisSvc.HistoryClient)
	d.ogvSeasonGRPC = newDaoClient(d, "ogvSeasonGRPC", ogvSeasonSvc.NewClient, daoConf).(ogvSeasonSvc.SeasonClient)
	d.ogvEpisodeGRPC = newDaoClient(d, "ogvEpisodeGRPC", ogvEpisodeSvc.NewClient, daoConf).(ogvEpisodeSvc.EpisodeClient)
	d.favGRPC = newDaoClient(d, "favGRPC", favSvc.NewClient, daoConf).(favSvc.FavoriteClient)
	d.ugcSeasonGRPC = newDaoClient(d, "ugcSeasonGRPC", ugcSeasonSvc.NewClient, daoConf).(ugcSeasonSvc.UGCSeasonClient)
	d.copyrightGRPC = newDaoClient(d, "copyrightGRPC", copyrightSvc.NewClient, daoConf).(copyrightSvc.CopyrightManageClient)
	d.epCardGRPC = newDaoClient(d, "epCardGRPC", epCardSvc.NewClient, daoConf).(epCardSvc.CardClient)

	bmConf := new(paladin.TOML)
	if err = appConf.Get("BmClient").Unmarshal(bmConf); err != nil {
		if err != paladin.ErrNotExist {
			return
		}
		err, bmConf = nil, nil
	}
	// http服务
	d.music = newBmClient(d, "music", bmConf)
	d.coinHTTP = newBmClient(d, "coinHTTP", bmConf)
	d.mediaListHTTP = newBmClient(d, "mediaListHTTP", bmConf)

	cf = d.Close
	return
}

func newBmClient(d *dao, fieldName string, bmConf *paladin.TOML) *bmClient {
	rt := reflect.TypeOf(d).Elem()
	_, found := rt.FieldByName(fieldName)
	if !found {
		panic("bm field " + fieldName + " not found")
	}
	cfg := new(struct {
		Host   string
		Config *bm.ClientConfig
	})
	if bmConf == nil {
		panic("unexpected nil bmConf")
	}
	if err := bmConf.Get(fieldName).UnmarshalTOML(cfg); err != nil {
		panic(err)
	}
	if len(cfg.Host) == 0 {
		panic("unexpected empty Host config for " + fieldName + " bm client")
	}
	uri, err := url.Parse(cfg.Host)
	if err != nil {
		panic(err)
	}

	return &bmClient{Host: cfg.Host, hostURL: uri, Client: bm.NewClient(cfg.Config)}
}

func newDaoClient(d *dao, fieldName string, constructor interface{}, daoConf *paladin.TOML) interface{} {
	rt := reflect.TypeOf(d).Elem()
	fd, found := rt.FieldByName(fieldName)
	if !found {
		panic("warden field " + fieldName + " not found")
	}
	cfg := new(warden.ClientConfig)
	if daoConf != nil {
		if err := daoConf.Get(fd.Type.Name()).UnmarshalTOML(cfg); err != nil {
			if err != paladin.ErrNotExist {
				panic(err)
			}
			cfg = nil
		}
	} else {
		cfg = nil
	}
	ret := reflect.ValueOf(constructor).Call([]reflect.Value{reflect.ValueOf(cfg)})
	if err, ok := ret[1].Interface().(error); ok && err != nil {
		panic(err)
	}
	return ret[0].Interface()
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(_ context.Context) (err error) {
	return nil
}

type bmClient struct {
	*bm.Client
	Host    string
	hostURL *url.URL
}

func (c *bmClient) composeURI(uri string) string {
	out, err := c.hostURL.Parse(uri)
	if err != nil {
		panic("error compose uri " + uri + err.Error())
	}
	return out.String()
}

func (c *bmClient) Get(ctx context.Context, uri string, params url.Values, res interface{}, v ...string) error {
	net, _ := network.FromContext(ctx)
	req, err := c.Client.NewRequest(http.MethodGet, c.composeURI(uri), net.RemoteIP, params)
	if err != nil {
		return err
	}
	return c.Client.Do(ctx, req, res, v...)
}

func (c *bmClient) Post(ctx context.Context, uri string, params url.Values, res interface{}, v ...string) error {
	net, _ := network.FromContext(ctx)
	req, err := c.Client.NewRequest(http.MethodPost, c.composeURI(uri), net.RemoteIP, params)
	if err != nil {
		return err
	}
	return c.Client.Do(ctx, req, res, v...)
}

func (c *bmClient) NewRequest(ctx context.Context, method, uri string, params url.Values) (*http.Request, error) {
	net, _ := network.FromContext(ctx)
	return c.Client.NewRequest(method, c.composeURI(uri), net.RemoteIP, params)
}

func (c *bmClient) Do(ctx context.Context, req *http.Request, res interface{}, v ...string) error {
	return c.Client.Do(ctx, req, res, v...)
}

func (c *bmClient) RawJSON(ctx context.Context, req *http.Request, res interface{}, v ...string) error {
	return c.Client.JSON(ctx, req, res, v...)
}
