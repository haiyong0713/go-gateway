package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	groupv2 "go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-show/interface/dao/resource"
	"go-gateway/app/web-svr/web-show/interface/model"
	rsmdl "go-gateway/app/web-svr/web-show/interface/model/resource"
	xecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"

	seasongrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	vugrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	vumdl "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
)

const (
	_nullImage     = "https://static.hdslb.com/images/transparent.gif"
	_videoPrefix   = "http://www.bilibili.com/video/av"
	_videoPrefixBV = "http://www.bilibili.com/video/"
	_bangumiPrefix = "bilibili://bangumi/season/"
	_GamePrefix    = "bilibili://game/"
	_LivePrefix    = "bilibili://live/"
	_AVprefix      = "bilibili://video/"
	_topicPrefix   = "//www.bilibili.com/tag/"
	_OGVPay        = "https://m.bilibili.com/cheese/play/ep"
	_WebLivePrefix = "https://live.bilibili.com/"
	_forbidReco    = 1
	_videoChannel  = 3
)

var (
	_emptyRelation = []*rsmdl.Relation{}
	_emptyAsgs     = []*rsmdl.Assignment{}
	_contractMap   = map[string]struct{}{
		"banner":     {},
		"focus":      {},
		"promote":    {},
		"app_banner": {},
		"text_link":  {},
		"frontpage":  {},
	}
	_cpmGrayRate   = int64(0)
	_white         = map[int64]struct{}{}
	_cpmOn         = true
	_RelationResID = 162
	// 位置id和禁止项映射
	_locsAdditMap = map[int64]int64{
		2625: 65, // 广告位
		2626: 65, // 广告位
		4330: 65, // 广告位
		3038: 66, // 播放页C位卡
	}
)

// URLMonitor return all urls configured
func (s *Service) URLMonitor(c context.Context, pf int) (urls map[string]string) {
	return s.urlMonitor[pf]
}

// GrayRate return gray  percent
func (s *Service) GrayRate(c context.Context) (r int64, ws []int64, swt bool) {
	r = _cpmGrayRate
	for w := range _white {
		ws = append(ws, w)
	}
	swt = _cpmOn
	return
}

// SetGrayRate set gray percent
func (s *Service) SetGrayRate(c context.Context, swt bool, rate int64, white []int64) {
	_cpmGrayRate = rate
	tmp := map[int64]struct{}{}
	for _, w := range white {
		tmp[w] = struct{}{}
	}
	_cpmOn = swt
	_white = tmp
}

// Resources get resource info by pf,ids
// nolint: gocognit
func (s *Service) Resources(c context.Context, arg *rsmdl.ArgRess) (mres map[string][]*rsmdl.Assignment, adsControl json.RawMessage, count int, err error) {
	var (
		aids                    []int64
		arcs                    map[int64]*api.Arc
		epids                   []int32
		seasons                 map[int32]*seasongrpc.SeasonCard
		country, province, city string
		info                    *locgrpc.InfoComplete
		roomIDs                 []int64
		roomList                map[int64]*rsmdl.LiveRoomInfo
		arc                     *api.SimpleArcReply
		viewAddit               *vugrpc.ArcViewAdditReply
	)
	arg.IP = metadata.String(c, metadata.RemoteIP)
	if info, err = s.LocalInfo(c, arg.IP); err != nil {
		log.Error("%+v", err)
		err = nil
	}
	if info != nil {
		country = info.Country
		province = info.Province
		city = info.City
	}
	area := checkAera(country)
	var cpmInfos map[int64]*rsmdl.Assignment
	if !arg.IsNotAD {
		var upID int64
		aid := arg.Aid
		if aid <= 0 && arg.BVID != "" {
			aid, _ = bvid.BvToAv(arg.BVID)
		}
		if aid > 0 {
			group := groupv2.WithContext(c)
			group.Go(func(ctx context.Context) error {
				var err error
				in := &arcgrpc.SimpleArcRequest{Aid: aid}
				if arc, err = s.arcGRPC.SimpleArc(c, in); err != nil {
					log.Error("SimpleArc aid:%d, err:%+v", aid, err)
					return err
				}
				upID = arc.GetArc().Mid
				return nil
			})
			group.Go(func(ctx context.Context) error {
				var err error
				if viewAddit, err = s.vuGRPC.ArcViewAddit(ctx, &vumdl.ArcViewAdditReq{Aid: aid}); err != nil {
					log.Error("ArcViewAddit aid:%d, error:%+v", arg.Aid, err)
					return err
				}
				return nil
			})
			if err = group.Wait(); err != nil {
				log.Error("s.Resources err:%+v", err)
				return
			}
		}
		val := &rsmdl.CpmsRequestParam{
			Mid:       arg.Mid,
			Ids:       arg.Ids,
			Sid:       arg.Sid,
			IP:        arg.IP,
			Country:   country,
			Province:  province,
			City:      city,
			Buvid:     arg.Buvid,
			Aid:       aid,
			UpID:      upID,
			UserAgent: arg.UserAgent,
			FromSpmID: arg.FromSpmID,
		}
		if _cpmOn {
			cpmInfos, adsControl = s.cpms(c, val)
		} else if _, ok := _white[arg.Mid]; ok || (arg.Mid%100 < _cpmGrayRate && arg.Mid != 0) {
			cpmInfos, adsControl = s.cpms(c, val)
		}
	}
	mres = make(map[string][]*rsmdl.Assignment)
	for _, id := range arg.Ids {
		if additID, ok := _locsAdditMap[id]; ok && viewAddit != nil {
			if state, ok := viewAddit.GetForbidReco().GetGroupState()[additID]; ok && state == 1 {
				// 命中禁止项
				continue
			}
		}
		pts := s.posCache[posKey(arg.Pf, int(id))]
		if pts == nil {
			continue
		}
		count = pts.Counter
		// add ads if exists
		res, as, es, rooms := s.res(c, cpmInfos, int(id), area, pts, arg.Mid)
		mres[strconv.FormatInt(id, 10)] = res
		for _, zoneID := range info.GetZoneId() {
			if zoneID == 4308992 && id == 142 && (time.Now().Unix() <= s.c.FrontPage.ETime) {
				mres["142"] = []*rsmdl.Assignment{
					{
						ID:         534907,
						ContractID: "frontpage",
						PosNum:     1,
						Name:       s.c.FrontPage.Name,
						Pic:        s.c.FrontPage.Pic,
						LitPic:     s.c.FrontPage.LitPic,
						URL:        s.c.FrontPage.URL,
						SrcID:      141,
						Area:       1,
						RequestID:  strconv.FormatInt(time.Now().Unix(), 10),
					},
				}
				break
			}
		}
		aids = append(aids, as...)
		roomIDs = append(roomIDs, rooms...)
		// epid去重
		for _, e := range es {
			var isRepeat bool
			for _, epid := range epids {
				if epid == e {
					isRepeat = true
					break
				}
			}
			if !isRepeat {
				epids = append(epids, e)
			}
		}
	}
	// fill archive if has video ad
	var mutex sync.Mutex
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if len(aids) == 0 {
			return nil
		}
		if arcs, err = s.Arcs(ctx, aids); err != nil {
			log.Error("Resources arg=%+v,aids=%+v,error=%+v", arg, aids, err)
			resource.PromError("arcGRPC.Arcs", "s.Arcs(arcAids:(%v), arcs), err(%v)", aids, err)
			return
		}
		mutex.Lock()
		for _, tres := range mres {
			for _, rs := range tres {
				if arc, ok := arcs[rs.Aid]; ok {
					// rs.Archive = arc
					bvID, _ := bvid.AvToBv(rs.Aid)
					rs.URL = _videoPrefixBV + bvID
					if rs.Name == "" {
						rs.Name = arc.Title
					}
					if rs.Pic == "" {
						rs.Pic = arc.Pic
					}
					model.ClearAttrAndAccess(arc)
					rs.Archive = &rsmdl.ArchiveBV{
						Arc:  arc,
						BVID: bvID,
					}
				}
			}
		}
		mutex.Unlock()
		return
	})
	if len(epids) > 0 {
		g.Go(func() (err error) {
			if seasons, err = s.Seasons(ctx, epids); err != nil {
				log.Error("Resources arg=%+v,epids=%+v,error=%+v", arg, epids, err)
				return
			}
			mutex.Lock()
			for _, tres := range mres {
				for _, rs := range tres {
					if season, ok := seasons[rs.EpID]; ok {
						rs.Season = season
						if season != nil {
							rs.URL = season.Url
						}
					}
				}
			}
			mutex.Unlock()
			return
		})
	}
	if len(roomIDs) > 0 {
		g.Go(func() (err error) {
			if roomList, err = s.liveDao.RoomByIds(ctx, roomIDs); err != nil {
				log.Error("Resources arg=%+v,roomIDs=%+v,error=%+v", arg, roomIDs, err)
				resource.PromError("s.liveDao.RoomByIds", "s.liveDao.RoomByIds(roomIDs:(%v)), err(%v)", roomIDs, err)
				return
			}
			mutex.Lock()
			for _, tres := range mres {
				for _, rs := range tres {
					if room, ok := roomList[rs.RoomID]; ok {
						rs.Room = room
						if room.Show != nil {
							if rs.Name == "" {
								rs.Name = room.Show.Title
							}
							if rs.Pic == "" {
								rs.Pic = room.Show.Cover
							}
						}
					}
				}
			}
			mutex.Unlock()
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	// if id is banner and not content add defult
	for i, rs := range mres {
		if len(rs) == 0 {
			id, _ := strconv.ParseInt(i, 10, 64)
			for _, val := range s.c.BannerID {
				if id != val {
					continue
				}
				mres[i] = append(mres[i], s.defBannerCache)
			}
		}
	}
	return
}

func (s *Service) getTagIds(c context.Context, aid, mid int64) (tagIds []int64, err error) {
	resourceChannels, err := s.ResourceChannels(c, mid, aid, _videoChannel)
	if err != nil {
		log.Error("【@getTagIds】s.ResourceChannels(%d,%d) (%+v)", mid, aid, err)
		return nil, err
	}
	for _, channel := range resourceChannels.GetChannels() {
		if channel != nil && channel.ID != 0 {
			tagIds = append(tagIds, channel.ID)
		}
	}
	return
}

// cpmBanners
func (s *Service) cpms(c context.Context, val *rsmdl.CpmsRequestParam) (res map[int64]*rsmdl.Assignment, adsControl json.RawMessage) {
	if val.Aid != 0 {
		tagIds, err := s.getTagIds(c, val.Aid, val.Mid)
		if err == nil && len(tagIds) > 0 {
			val.TagIds = tagIds
		}
	}
	cpmInfos, err := s.adDao.Cpms(c, val)
	if err != nil {
		log.Error("s.adDao.Cpms error(%v)", err)
		return
	}
	res = make(map[int64]*rsmdl.Assignment, len(cpmInfos.AdsInfo))
	adsControl = cpmInfos.AdsControl
	for _, id := range val.Ids {
		idStr := strconv.FormatInt(id, 10)
		if adsInfos := cpmInfos.AdsInfo[idStr]; len(adsInfos) > 0 {
			for srcStr, adsInfo := range adsInfos {
				// var url string
				srcIDInt, _ := strconv.ParseInt(srcStr, 10, 64)
				if adInfo := adsInfo.AdInfo; adInfo != nil {
					//switch adInfo.CreativeType {
					// case 0:
					// 	url = adInfo.CreativeContent.URL
					// case 1:
					// 	url = "www.bilibili.com/video/av" + adInfo.CreativeContent.VideoID
					// }
					ad := &rsmdl.Assignment{
						CreativeType: adInfo.CreativeType,
						Aid:          adInfo.CreativeContent.VideoID,
						RequestID:    cpmInfos.RequestID,
						SrcID:        srcIDInt,
						IsAdLoc:      true,
						IsAd:         adsInfo.IsAd,
						CmMark:       adsInfo.CmMark,
						CreativeID:   adInfo.CreativeID,
						AdCb:         adInfo.AdCb,
						ShowURL:      adInfo.CreativeContent.ShowURL,
						ClickURL:     adInfo.CreativeContent.ClickURL,
						Name:         adInfo.CreativeContent.Title,
						AdDesc:       adInfo.CreativeContent.Desc,
						Pic:          adInfo.CreativeContent.ImageURL,
						LitPic:       adInfo.CreativeContent.ThumbnailURL,
						URL:          adInfo.CreativeContent.URL,
						PosNum:       int(adsInfo.Index),
						Title:        adInfo.CreativeContent.Title,
						ServerType:   rsmdl.FromCpm,
						IsCpm:        true,
						AdverName:    adInfo.Extra.Card.AdverName,
						CardType:     adInfo.CardType,
						BusinessMark: adInfo.Extra.Card.BusinessMark,
					}
					res[srcIDInt] = ad
				} else {
					ad := &rsmdl.Assignment{
						IsAdLoc:   true,
						RequestID: cpmInfos.RequestID,
						IsAd:      false,
						SrcID:     srcIDInt,
						ResID:     int(id),
						CmMark:    adsInfo.CmMark,
					}
					res[srcIDInt] = ad
				}
			}
		}
	}
	return
}

func checkAera(country string) (area int8) {
	switch country {
	case "中国":
		area = 1
	case "香港", "台湾", "澳门":
		area = 2
	case "日本":
		area = 3
	case "美国":
		area = 4
	default:
		if _, ok := rsmdl.OverSeasCountry[country]; ok {
			area = 5
		} else {
			area = 0
		}
	}
	return
}

// Relation get relation archives by aid
// nolint:gomnd
func (s *Service) Relation(c context.Context, arg *rsmdl.ArgAid) (rls []*rsmdl.Relation, err error) {
	var (
		aids   []int64
		relErr error
		forbid int64
	)
	rls = _emptyRelation
	arg.IP = metadata.String(c, metadata.RemoteIP)
	group := groupv2.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if aids, relErr = s.dataDao.Related(ctx, arg.Aid, arg.IP); relErr != nil {
			log.Error("s.dataDao.Related aid(%v) error(%v)", arg.Aid, relErr)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if viewAddit, e := s.vuGRPC.ArcViewAddit(ctx, &vumdl.ArcViewAdditReq{Aid: arg.Aid}); e != nil {
			log.Error("s.videoUpGRPC.ArcViewAddit aid(%d) error(%v)", arg.Aid, e)
		} else if viewAddit != nil && viewAddit.ForbidReco != nil {
			forbid = viewAddit.ForbidReco.State
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if len(aids) == 0 || forbid == _forbidReco {
		log.Warn("zero_relates %d forbid(%d)", arg.Aid, forbid)
		return
	}
	arcs, err := s.Arcs(c, aids)
	if err != nil {
		log.Info("s.Arcs error(%v)", err)
		return
	}
	var res []*rsmdl.Relation
	for _, arc := range arcs {
		model.ClearAttrAndAccess(arc)
		res = append(res, &rsmdl.Relation{Arc: arc})
	}
	rls = res
	if len(res) < 3 {
		return
	}
	var (
		country, province, city string
		info                    *locgrpc.InfoComplete
	)
	if info, err = s.LocalInfo(c, arg.IP); err != nil {
		log.Error("%+v", err)
		err = nil
	}
	if info != nil {
		country = info.Country
		province = info.Province
		city = info.City
	}
	area := checkAera(country)
	//pts := s.posCache[posKey(0, _RelationResID)]
	val := &rsmdl.CpmsRequestParam{
		Mid:       arg.Mid,
		Ids:       []int64{int64(_RelationResID)},
		Sid:       arg.Sid,
		IP:        arg.IP,
		Country:   country,
		Province:  province,
		City:      city,
		Buvid:     arg.Buvid,
		Aid:       0,
		UpID:      0,
		UserAgent: arg.UserAgent,
	}
	cpmInfos, _ := s.cpms(c, val)
	for _, rs := range cpmInfos {
		// just fet one ad
		if rs.IsAd {
			var arc *api.Arc
			arc, err = s.Arc(c, rs.Aid)
			if err != nil {
				resource.PromError("arcGRPC.Arc", "s.Arc(arcAid:(%v), arcs), err(%v)", rs.Aid, err)
				err = nil
				rls = res
				return
			}
			model.ClearAttrAndAccess(arc)
			rl := &rsmdl.Relation{
				Arc:        arc,
				Area:       area,
				RequestID:  rs.RequestID,
				CreativeID: rs.CreativeID,
				AdCb:       rs.AdCb,
				SrcID:      rs.SrcID,
				ShowURL:    rs.ShowURL,
				ClickURL:   rs.ClickURL,
				IsAdLoc:    rs.IsAdLoc,
				ResID:      _RelationResID,
				IsAd:       true,
			}
			if rs.Pic != "" {
				rl.Pic = rs.Pic
			}
			if rs.Title != "" {
				rl.Title = rs.Title
			}
			rls = append(res[:2], append([]*rsmdl.Relation{rl}, res[2:]...)...)
			return
		}
		res[2].AdCb = rs.AdCb
		res[2].SrcID = rs.SrcID
		res[2].ShowURL = rs.ShowURL
		res[2].ClickURL = rs.ClickURL
		res[2].IsAdLoc = rs.IsAdLoc
		res[2].RequestID = rs.RequestID
		res[2].CreativeID = rs.CreativeID
		res[2].ResID = _RelationResID
		return
	}
	return
}

// Resource get resource info by pf,id
// nolint: gocognit
func (s *Service) Resource(c context.Context, arg *rsmdl.ArgRes) (res []*rsmdl.Assignment, count int, err error) {
	var (
		aids                    []int64
		arcs                    map[int64]*api.Arc
		epids                   []int32
		seasons                 map[int32]*seasongrpc.SeasonCard
		country, province, city string
		info                    *locgrpc.InfoComplete
		roomIDs                 []int64
		roomList                map[int64]*rsmdl.LiveRoomInfo
	)
	arg.IP = metadata.String(c, metadata.RemoteIP)
	res = _emptyAsgs
	pts := s.posCache[posKey(arg.Pf, int(arg.ID))]
	if pts == nil {
		return
	}
	count = pts.Counter
	if info, err = s.LocalInfo(c, arg.IP); err != nil {
		log.Error("%+v", err)
		err = nil
	}
	if info != nil {
		country = info.Country
		province = info.Province
		city = info.City
	}
	area := checkAera(country)
	var cpmInfos map[int64]*rsmdl.Assignment
	if !arg.IsNotAD {
		val := &rsmdl.CpmsRequestParam{
			Mid:       arg.Mid,
			Ids:       []int64{arg.ID},
			Sid:       arg.Sid,
			IP:        arg.IP,
			Country:   country,
			Province:  province,
			City:      city,
			Buvid:     arg.Buvid,
			Aid:       0,
			UpID:      0,
			UserAgent: arg.UserAgent,
			FromSpmID: arg.FromSpmID,
		}
		if _cpmOn {
			cpmInfos, _ = s.cpms(c, val)
		} else if _, ok := _white[arg.Mid]; ok || (arg.Mid%100 < _cpmGrayRate && arg.Mid != 0) {
			cpmInfos, _ = s.cpms(c, val)
		}
	}
	res, aids, epids, roomIDs = s.res(c, cpmInfos, int(arg.ID), area, pts, arg.Mid)
	// fill archive if has video ad
	var mutex sync.Mutex
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if arcs, err = s.Arcs(ctx, aids); err != nil {
				log.Error("Resource arg=%+v,aids=%+v,error=%+v", arg, aids, err)
				resource.PromError("arcGRPC.Arcs", "s.Arcs(arcAid:(%v), arcs), err(%v)", aids, err)
				return
			}
			mutex.Lock()
			for _, rs := range res {
				if arc, ok := arcs[rs.Aid]; ok {
					bvID, _ := bvid.AvToBv(rs.Aid)
					rs.URL = _videoPrefixBV + bvID
					if rs.Name == "" {
						rs.Name = arc.Title
					}
					if rs.Pic == "" {
						rs.Pic = arc.Pic
					}
					model.ClearAttrAndAccess(arc)
					rs.Archive = &rsmdl.ArchiveBV{
						Arc:  arc,
						BVID: bvID,
					}
				}
			}
			mutex.Unlock()
			return
		})
	}
	if len(epids) > 0 {
		g.Go(func() (err error) {
			if seasons, err = s.Seasons(ctx, epids); err != nil {
				log.Error("Resource arg=%+v,epids=%+v,error=%+v", arg, epids, err)
				return
			}
			mutex.Lock()
			for _, rs := range res {
				if season, ok := seasons[rs.EpID]; ok {
					rs.Season = season
					if season != nil {
						rs.URL = season.Url
					}
				}
			}
			mutex.Unlock()
			return
		})
	}
	if len(roomIDs) > 0 {
		g.Go(func() (err error) {
			if roomList, err = s.liveDao.RoomByIds(ctx, roomIDs); err != nil {
				log.Error("Resource arg=%+v,roomIDs=%+v,error=%+v", arg, roomIDs, err)
				resource.PromError("s.liveDao.RoomByIds", "s.liveDao.RoomByIds(roomIDs:(%v)), err(%v)", roomIDs, err)
				return
			}
			mutex.Lock()
			for _, rs := range res {
				if room, ok := roomList[rs.RoomID]; ok {
					rs.Room = room
					if room.Show != nil {
						if rs.Name == "" {
							rs.Name = room.Show.Title
						}
						if rs.Pic == "" {
							rs.Pic = room.Show.Cover
						}
					}
				}
			}
			mutex.Unlock()
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	// add defBanner if contnent not exits
	if len(res) == 0 {
		for _, val := range s.c.BannerID {
			if arg.ID != val {
				continue
			}
			// 142 is index banner
			if rs := s.resByID(142); rs != nil {
				res = append(res, rs)
			} else {
				res = append(res, s.defBannerCache)
			}
		}
	}
	return
}

// nolint: gocognit
func (s *Service) res(_ context.Context, cpmInfos map[int64]*rsmdl.Assignment, id int, area int8, pts *rsmdl.Position, mid int64) (res []*rsmdl.Assignment, aids []int64, epids []int32, roomIDs []int64) {
	// add ads if exists
	var (
		reqID     string
		index     int
		resIndex  int
		bossIndex int
		ts        = strconv.FormatInt(time.Now().Unix(), 10)
		resBs     []*rsmdl.Assignment
	)
	for _, pt := range pts.Pos {
		var (
			isAdLoc bool
			rs      *rsmdl.Assignment
			cpm     *rsmdl.Assignment
			ok      bool
			// rs指针传递临时变量
			rstmp *rsmdl.Assignment
		)
		// 优先级 强运营帧 > 广告 > 固定投放 > 推荐池
		// 目前只有一个强运营帧，并且出现在首帧
		if _, ok = s.bossAsgCache[id]; ok && bossIndex == 0 {
			if bossIndex < len(s.bossAsgCache[id]) {
				rstmp = s.bossAsgCache[id][bossIndex]
				bossIndex++
				// 固定投放缓冲队列中补充当前位置应有的固定投放
				if rsb := s.resByID(pt.ID); rsb != nil {
					// url mean aid in Asgtypevideo
					resBs = append(resBs, rsb)
				}
				goto RESOURCE
			}
		}
		// 广告优先
		cpm = cpmInfos[int64(pt.ID)]
		if cpm != nil {
			isAdLoc = true
			reqID = cpm.RequestID
			if cpm.IsCpm {
				cpm.Area = area
				if mid != 0 {
					cpm.Mid = strconv.FormatInt(mid, 10)
				}
				if cpm.CreativeType == rsmdl.CreativeVideo {
					aids = append(aids, cpm.Aid)
				}
				cpm.ServerType = rsmdl.FromCpm
				res = append(res, cpm)
				// 固定投放缓冲队列中补充当前位置应有的固定投放
				if rsb := s.resByID(pt.ID); rsb != nil {
					// url mean aid in Asgtypevideo
					resBs = append(resBs, rsb)
				}
				continue
			}
		}
		// 运营后台内容
		if rstmp = s.resByID(pt.ID); rstmp != nil {
			// 当前位置应有的固定投放优先
		} else if resIndex < len(resBs) {
			// 固定投放缓冲队列
			rstmp = resBs[resIndex]
			resIndex++
		} else {
			// 推荐池
			rstmp = s.resByIndex(id, index)
			index++
		}
	RESOURCE:
		if rstmp != nil {
			rs = new(rsmdl.Assignment)
			*rs = *rstmp
			if rs.Atype == rsmdl.AsgTypeVideo || rs.Atype == rsmdl.AsgTypeAv {
				aids = append(aids, rs.Aid)
			}
			if rs.Atype == rsmdl.AsgTypeOGVPay {
				// epids去重
				var isRepeat bool
				for _, epid := range epids {
					if epid == rs.EpID {
						isRepeat = true
						break
					}
				}
				if !isRepeat {
					epids = append(epids, rs.EpID)
				}
			}
			if rs.Atype == rsmdl.AsgTypeWebLive || rs.Atype == rsmdl.AsgTypeLive {
				roomIDs = append(roomIDs, rs.RoomID)
			}
			rs.PosNum = pt.PosNum
			rs.SrcID = int64(pt.ID)
			rs.Area = area
			rs.IsAdLoc = isAdLoc
			if isAdLoc {
				rs.RequestID = reqID
			} else {
				rs.RequestID = ts
			}
			if mid != 0 {
				rs.Mid = strconv.FormatInt(mid, 10)
			}
			res = append(res, rs)
		} else if isAdLoc {
			rs = &rsmdl.Assignment{
				PosNum:    pt.PosNum,
				SrcID:     int64(pt.ID),
				IsAdLoc:   isAdLoc,
				RequestID: reqID,
				Area:      area,
				Pic:       _nullImage,
				NullFrame: true,
			}
			if mid != 0 {
				rs.Mid = strconv.FormatInt(mid, 10)
			}
			res = append(res, rs)
		}
	}
	return
}

// resByIndex return res of index
func (s *Service) resByIndex(id, index int) (res *rsmdl.Assignment) {
	ss := s.asgCache[id]
	if index >= len(ss) {
		return
	}
	res = new(rsmdl.Assignment)
	*res = *(ss[index])
	return
}

// resByID return res of id
func (s *Service) resByID(id int) (res *rsmdl.Assignment) {
	ss := s.asgCache[id]
	l := len(ss)
	if l == 0 {
		return
	}
	res = ss[0]
	for _, s := range ss {
		// ContractId not in contractMap ,it is ad and ad first
		if _, ok := _contractMap[s.ContractID]; !ok {
			res = s
			return
		}
	}
	return
}

// rpc resourcesALL
func (s *Service) resourcesALL() (rscs []*rsmdl.Res, err error) {
	resourcesRPC, err := s.recrpc.ResourceAll(context.Background())
	if err != nil {
		resource.PromError("recRPC.ResourcesALL", "s.recrpc.resourcesRPC error(%v)", err)
		return
	}
	rscs = make([]*rsmdl.Res, 0)
	for _, res := range resourcesRPC {
		if res == nil {
			continue
		}
		rsc := &rsmdl.Res{
			ID:       res.ID,
			Platform: res.Platform,
			Name:     res.Name,
			Parent:   res.Parent,
			Counter:  res.Counter,
			Position: res.Position,
		}
		rscs = append(rscs, rsc)
	}
	return
}

// rpc assignmentAll
// nolint:gomnd
func (s *Service) assignmentAll() (asgs, bossAsgs []*rsmdl.Assignment, err error) {
	assignRPC, err := s.recrpc.AssignmentAll(context.Background())
	if err != nil {
		resource.PromError("recRPC.AssignmentAll", "s.recrpc.assignRPC error(%v)", err)
		return
	}
	asgs = make([]*rsmdl.Assignment, 0)
	for _, asgr := range assignRPC {
		if asgr == nil {
			continue
		}
		asg := &rsmdl.Assignment{
			ID:           asgr.ID,
			Name:         asgr.Name,
			ContractID:   asgr.ContractID,
			ResID:        asgr.ResID,
			AsgID:        int64(asgr.AsgID),
			Pic:          asgr.Pic,
			LitPic:       asgr.LitPic,
			URL:          asgr.URL,
			Atype:        asgr.Atype,
			Weight:       asgr.Weight,
			Rule:         asgr.Rule,
			Agency:       asgr.Agency,
			STime:        asgr.STime,
			SubTitle:     asgr.SubTitle,
			PicMainColor: asgr.PicMainColor,
			Inline: rsmdl.Inline{
				InlineUseSame:       asgr.Inline.InlineUseSame,
				InlineType:          asgr.Inline.InlineType,
				InlineUrl:           asgr.Inline.InlineUrl,
				InlineBarrageSwitch: asgr.Inline.InlineBarrageSwitch,
			},
			Operater: asgr.Operater,
		}
		if asgr.ActivityID > 0 {
			if xtime.Time(time.Now().Unix()) < asgr.ActivitySTime {
				asg.ActivityType = rsmdl.BeforActivity
			} else if xtime.Time(time.Now().Unix()) > asgr.ActivityETime {
				asg.ActivityType = rsmdl.OverActivity
			} else {
				asg.ActivityType = rsmdl.InActivity
			}
		}
		// 投放类型 0固定投放 1推荐池 2强运营帧
		if asgr.Category == 2 {
			bossAsgs = append(bossAsgs, asg)
			continue
		}
		asgs = append(asgs, asg)
	}
	return
}

// default banner
func (s *Service) defBanner() (asg *rsmdl.Assignment, err error) {
	bannerRPC, err := s.recrpc.DefBanner(context.Background())
	if err != nil {
		resource.PromError("recRPC.defBanner", "s.recrpc.defBanner error(%v)", err)
		return
	}
	if bannerRPC != nil {
		asg = &rsmdl.Assignment{
			ID:         bannerRPC.ID,
			Name:       bannerRPC.Name,
			ContractID: bannerRPC.ContractID,
			ResID:      bannerRPC.ResID,
			Pic:        bannerRPC.Pic,
			LitPic:     bannerRPC.LitPic,
			URL:        bannerRPC.URL,
			Atype:      bannerRPC.Atype,
			Weight:     bannerRPC.Weight,
			Rule:       bannerRPC.Rule,
			Agency:     bannerRPC.Agency,
		}
	}
	return
}

// LoadRes load Res info to cache
// nolint: gocognit
func (s *Service) loadRes() {
	if s.resRunning {
		return
	}
	s.resRunning = true
	defer func() {
		s.resRunning = false
	}()
	assign, bossAssign, err := s.assignmentAll()
	if err != nil {
		log.Error("loadRes assignmentAll error(%v)", err)
		return
	}
	resources, err := s.resourcesALL()
	if err != nil {
		log.Error("loadRes resourcesALL error(%v)", err)
		return
	}
	resMap := make(map[int]*rsmdl.Res)
	posMap := make(map[string]*rsmdl.Position)
	for _, res := range resources {
		resMap[res.ID] = res
		if res.Counter > 0 {
			key := posKey(res.Platform, res.ID)
			pos := &rsmdl.Position{
				Pos:     make([]*rsmdl.Loc, 0),
				Counter: res.Counter,
			}
			posMap[key] = pos
		} else {
			key := posKey(res.Platform, res.Parent)
			if pos, ok := posMap[key]; ok {
				loc := &rsmdl.Loc{
					ID:     res.ID,
					PosNum: res.Position,
				}
				pos.Pos = append(pos.Pos, loc)
			}
		}
	}
	// 强运营帧
	tmpBoss := make(map[int][]*rsmdl.Assignment)
	for _, a := range bossAssign {
		if res, ok := resMap[a.ResID]; ok {
			if err = s.convertURL(a); err != nil {
				log.Error("loadRes convertURL error(%v)", err)
				return
			}
			var data struct {
				Cover        int32  `json:"is_cover"`
				Style        int32  `json:"style"`
				Label        string `json:"label"`
				Intro        string `json:"intro"`
				CreativeType int8   `json:"creative_type"`
			}
			//  unmarshal rule for frontpage style
			if a.Rule != "" {
				e := json.Unmarshal([]byte(a.Rule), &data)
				if e != nil {
					log.Error("json.Unmarshal data:%+v,error:%+v", a.Rule, e)
				} else {
					a.Style = data.Style
					a.CreativeType = data.CreativeType
				}
			}
			tmpBoss[res.ID] = append(tmpBoss[res.ID], a)
		}
	}
	for _, a := range assign {
		if res, ok := resMap[a.ResID]; ok {
			if err = s.convertURL(a); err != nil {
				return
			}
			var data struct {
				Cover        int32  `json:"is_cover"`
				Style        int32  `json:"style"`
				Label        string `json:"label"`
				Intro        string `json:"intro"`
				CreativeType int8   `json:"creative_type"`
			}
			//  unmarshal rule for frontpage style
			if a.Rule != "" {
				e := json.Unmarshal([]byte(a.Rule), &data)
				if e != nil {
					log.Error("json.Unmarshal data:%+v,error:%+v", a.Rule, e)
				} else {
					a.Style = data.Style
					a.CreativeType = data.CreativeType
					if a.ContractID == "rec_video" {
						a.Label = data.Label
						a.Intro = data.Intro
					}
				}
			}
			res.Assignments = append(res.Assignments, a)
		}

	}
	urlMonitor := make(map[int]map[string]string)
	tmp := make(map[int][]*rsmdl.Assignment, len(resMap))
	for _, res := range resMap {
		tmp[res.ID] = res.Assignments
		urlMap, ok := urlMonitor[res.Platform]
		if !ok {
			urlMap = make(map[string]string)
			urlMonitor[res.Platform] = urlMap
		}
		for _, a := range res.Assignments {
			urlMap[fmt.Sprintf("%d_%s", a.ResID, a.Name)] = a.URL
		}
	}
	s.asgCache = tmp
	s.bossAsgCache = tmpBoss
	s.posCache = posMap
	s.urlMonitor = urlMonitor
	// load default banner
	banner, err := s.defBanner()
	if err != nil {
		log.Error("loadRes defBanner error(%v)", err)
		return
	} else if banner != nil {
		var data struct {
			Cover int32 `json:"is_cover"`
			Style int32 `json:"style"`
		}
		err := json.Unmarshal([]byte(banner.Rule), &data)
		if err != nil {
			log.Error("json.Unmarshal (%s) error(%v)", banner.Rule, err)
		} else {
			banner.Style = data.Style
		}
		s.defBannerCache = banner
	}
}

func posKey(pf, id int) string {
	return fmt.Sprintf("%d_%d", pf, id)
}

func (s *Service) convertURL(a *rsmdl.Assignment) (err error) {
	switch a.Atype {
	case rsmdl.AsgTypeVideo:
		var aid int64
		if aid, err = strconv.ParseInt(a.URL, 10, 64); err != nil {
			log.Error("strconv.ParseInt(%s) err(%v)", a.URL, err)
			return
		}
		a.Aid = aid
		a.URL = _videoPrefix + a.URL
	case rsmdl.AsgTypeURL:
	case rsmdl.AsgTypeBangumi:
		a.URL = _bangumiPrefix + a.URL
	case rsmdl.AsgTypeLive:
		var roomID int64
		if roomID, err = strconv.ParseInt(a.URL, 10, 64); err != nil {
			log.Error("roomID strconv.ParseInt(%s) err(%v)", a.URL, err)
			return
		}
		a.RoomID = roomID
		a.URL = _LivePrefix + a.URL
	case rsmdl.AsgTypeGame:
		a.URL = _GamePrefix + a.URL
	case rsmdl.AsgTypeAv:
		var aid int64
		if aid, err = strconv.ParseInt(a.URL, 10, 64); err != nil {
			log.Error("strconv.ParseInt(%s) err(%v)", a.URL, err)
			return
		}
		a.Aid = aid
		a.URL = _AVprefix + a.URL
	case rsmdl.AsgTypeTopic:
		a.URL = _topicPrefix + a.URL
	case rsmdl.AsgTypeOGVPay:
		var epid int64
		if epid, err = strconv.ParseInt(a.URL, 10, 32); err != nil {
			log.Error("strconv.Atoi(%s) err(%v)", a.URL, err)
			return
		}
		a.EpID = int32(epid)
		a.URL = _OGVPay + a.URL
	case rsmdl.AsgTypeWebLive:
		var roomID int64
		if roomID, err = strconv.ParseInt(a.URL, 10, 64); err != nil {
			log.Error("roomID strconv.ParseInt(%s) err(%v)", a.URL, err)
			return
		}
		a.RoomID = roomID
		a.URL = _WebLivePrefix + a.URL
	}
	return
}

func (s *Service) Seasons(c context.Context, epids []int32) (res map[int32]*seasongrpc.SeasonCard, err error) {
	var cards *seasongrpc.CardsByEpIdsReply
	if cards, err = s.seasonDao.CardsByEpIds(c, epids); err != nil {
		log.Error("%v", err)
		return
	}
	res = cards.GetCards()
	return
}

func (s *Service) LocalInfo(c context.Context, ip string) (res *locgrpc.InfoComplete, err error) {
	var (
		args = &locgrpc.InfoCompleteReq{Addr: ip}
		info *locgrpc.InfoCompleteReply
	)
	if info, err = s.locGRPC.InfoComplete(c, args); err != nil {
		log.Error("%+v", err)
		return
	}
	res = info.GetInfo()
	return
}

func (s *Service) Arc(c context.Context, aid int64) (res *arcgrpc.Arc, err error) {
	var (
		args    = &arcgrpc.ArcRequest{Aid: aid}
		arcsTmp *arcgrpc.ArcReply
	)
	if arcsTmp, err = s.arcGRPC.Arc(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = arcsTmp.GetArc()
	return
}

func (s *Service) Arcs(c context.Context, aids []int64) (res map[int64]*arcgrpc.Arc, err error) {
	var (
		args    = &arcgrpc.ArcsRequest{Aids: aids}
		arcsTmp *arcgrpc.ArcsReply
	)
	res = make(map[int64]*arcgrpc.Arc)
	if arcsTmp, err = s.arcGRPC.Arcs(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = arcsTmp.GetArcs()
	return
}

func (s *Service) FrontPage(c context.Context, resid int64) (*rsmdl.FrontPage, error) {
	resTmp, err := s.resdao.FrontPage(c, resid)
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	if resTmp == nil {
		return nil, nil
	}
	var res *rsmdl.FrontPage
	for _, online := range resTmp.Online {
		if online == nil {
			continue
		}
		res = new(rsmdl.FrontPage)
		res.FormFrontPage(online)
		return res, nil
	}
	if resTmp.Default != nil {
		res = new(rsmdl.FrontPage)
		res.FormFrontPage(resTmp.Default)
	}
	return res, nil
}

func (s *Service) PageHeader(ctx context.Context, resourceID int64, ip string) (*rsmdl.PageHeader, error) {
	reply, err := s.resdao.PageHeader(ctx, resourceID, ip)
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	return &rsmdl.PageHeader{
		Name:         reply.GetName(),
		Pic:          reply.GetPic(),
		Litpic:       reply.GetLitpic(),
		Url:          reply.GetUrl(),
		IsSplitLayer: reply.GetIsSplitLayer(),
		SplitLayer:   reply.GetSplitLayer(),
		RequestId:    strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}

func (s *Service) slbRetryCode(originErr error) error {
	retryCode := []int{-500, -502, -504}
	for _, val := range retryCode {
		if ecode.EqualError(ecode.Int(val), originErr) {
			return errors.Wrapf(xecode.WebSLBRetry, "%v", originErr)
		}
	}
	return originErr
}

func (s *Service) SLBRetry(err error) bool {
	return ecode.EqualError(xecode.WebSLBRetry, err)
}
