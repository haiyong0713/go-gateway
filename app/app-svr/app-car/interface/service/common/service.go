package common

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"

	"go-gateway/app/app-svr/app-car/interface/conf"
	accountdao "go-gateway/app/app-svr/app-car/interface/dao/account"
	archivedao "go-gateway/app/app-svr/app-car/interface/dao/archive"
	bangumidao "go-gateway/app/app-svr/app-car/interface/dao/bangumi"
	channeldao "go-gateway/app/app-svr/app-car/interface/dao/channel"
	dyndao "go-gateway/app/app-svr/app-car/interface/dao/dynamic"
	"go-gateway/app/app-svr/app-car/interface/dao/exp"
	favdao "go-gateway/app/app-svr/app-car/interface/dao/favorite"
	"go-gateway/app/app-svr/app-car/interface/dao/fm"
	historydao "go-gateway/app/app-svr/app-car/interface/dao/history"
	rcmddao "go-gateway/app/app-svr/app-car/interface/dao/recommend"
	regdao "go-gateway/app/app-svr/app-car/interface/dao/region"
	relationdao "go-gateway/app/app-svr/app-car/interface/dao/relation"
	srchdao "go-gateway/app/app-svr/app-car/interface/dao/search"
	serialdao "go-gateway/app/app-svr/app-car/interface/dao/serial"
	showdao "go-gateway/app/app-svr/app-car/interface/dao/show"
	sbdao "go-gateway/app/app-svr/app-car/interface/dao/silverbullet"
	thumbupdao "go-gateway/app/app-svr/app-car/interface/dao/thumbup"
	updao "go-gateway/app/app-svr/app-car/interface/dao/up"
	"go-gateway/app/app-svr/app-car/interface/model"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	mediarpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

type Service struct {
	c                  *conf.Config
	historyDao         *historydao.Dao
	archiveDao         *archivedao.Dao
	bangumiDao         *bangumidao.Dao
	accountDao         *accountdao.Dao
	relationDao        *relationdao.Dao
	dynDao             *dyndao.Dao
	upDao              *updao.Dao
	favDao             *favdao.Dao
	srchDao            *srchdao.Dao
	fmDao              *fm.Dao
	rcmdDao            *rcmddao.Dao
	thumbupDao         *thumbupdao.Dao
	showDao            *showdao.Dao
	regionDao          *regdao.Dao
	channelDao         *channeldao.Dao
	sbDao              *sbdao.Dao
	serialDao          *serialdao.Dao
	expDao             *exp.Dao
	grpcClientShortURL grpcShortURL.ShortUrlClient
	mediaRpc           mediarpc.IndexClient
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		historyDao:  historydao.New(c),
		archiveDao:  archivedao.New(c),
		bangumiDao:  bangumidao.New(c),
		accountDao:  accountdao.New(c),
		relationDao: relationdao.New(c),
		upDao:       updao.New(c),
		dynDao:      dyndao.New(c),
		favDao:      favdao.New(c),
		rcmdDao:     rcmddao.New(c),
		srchDao:     srchdao.New(c),
		fmDao:       fm.New(c),
		thumbupDao:  thumbupdao.New(c),
		showDao:     showdao.New(c),
		regionDao:   regdao.New(c),
		channelDao:  channeldao.New(c),
		sbDao:       sbdao.New(c),
		serialDao:   serialdao.New(c),
		expDao:      exp.New(c),
	}
	initFmListHandler(s)
	initFmTabHandler(s)
	var err error
	if s.grpcClientShortURL, err = grpcShortURL.NewClient(nil); err != nil {
		panic(err)
	}
	if s.mediaRpc, err = mediarpc.NewClientIndex(nil); err != nil {
		panic(err)
	}
	return
}

// material 集中获取物料
func (s *Service) material(c context.Context, params *commonmdl.Params, deviceInfo model.DeviceInfo) (res *commonmdl.CarContext, err error) { // nolint:gocognit
	res = new(commonmdl.CarContext)
	muEpInline := sync.Mutex{}
	eg := errgroup.WithContext(c)
	if params.ArchiveReq != nil {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.ArchiveResp, errTmp = s.archiveDao.ArcsPlayerV2(ctx, params.ArchiveReq.PlayAvs, true, ""); errTmp != nil {
				b, _ := json.Marshal(params.ArchiveReq.PlayAvs)
				log.Errorc(c, "material ArcsPlayerV2(%+v) error(%v)", string(b), errTmp)
				return errTmp
			}
			return nil
		})
	}
	if params.ArchivePlusReq != nil {
		eg.Go(func(ctx context.Context) error {
			var aidPlayerMap map[int64]*archivegrpc.ArcPlayer
			var aidViewMap map[int64]*archivegrpc.ViewReply
			eg1 := errgroup.WithContext(c)
			eg1.Go(func(c1 context.Context) error {
				var pErr error
				aidPlayerMap, pErr = s.archiveDao.ArcsPlayerV2(c1, params.ArchivePlusReq.PlayAvs, true, "")
				if pErr != nil {
					log.Errorc(c, "material ArchivePlusReq s.archiveDao.ArcsPlayerV2 error(%v)", pErr)
				}
				return pErr
			})
			eg1.Go(func(c1 context.Context) error {
				aids := make([]int64, 0)
				for _, v := range params.ArchivePlusReq.PlayAvs {
					aids = append(aids, v.GetAid())
				}
				var sErr error
				aidViewMap, sErr = s.archiveDao.ViewsAll(c1, aids)
				if sErr != nil {
					log.Errorc(c, "material ArchivePlusReq s.archiveDao.SimpleArcs aids=%v,error(%v)", aids, sErr)
				}
				return sErr
			})
			if err1 := eg1.Wait(); err1 != nil {
				return err1
			}
			res.ArchivePlusResp = make(map[int64]*commonmdl.ArchivePlusResp)
			for aid, v := range aidPlayerMap {
				if v == nil {
					continue
				}
				view := aidViewMap[aid]
				if view == nil {
					continue
				}
				res.ArchivePlusResp[aid] = &commonmdl.ArchivePlusResp{
					Player: v,
					View:   view,
				}
			}
			return nil
		})
	}
	if params.EpisodeReq != nil && len(params.EpisodeReq.Epids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.EpisodeResp, errTmp = s.bangumiDao.EpCards(ctx, params.EpisodeReq.Epids); errTmp != nil {
				log.Error("material EpCards(%+v) error(%v)", params.EpisodeReq.Epids, errTmp)
				return errTmp
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			var (
				errTmp       error
				tmpEpInlines map[int32]*pgcinline.EpisodeCard
			)
			if tmpEpInlines, errTmp = s.bangumiDao.InlineCardsAll(ctx, params.EpisodeReq.Epids, deviceInfo.MobiApp, deviceInfo.Platform, deviceInfo.Device, deviceInfo.Build); errTmp != nil {
				log.Error("material ep InlineCardsAll(%+v, %v, %v, %v, %v) error(%v)", params.EpisodeReq.Epids, deviceInfo.MobiApp, deviceInfo.Platform, deviceInfo.Device, deviceInfo.Build, errTmp)
				return nil
			}
			muEpInline.Lock()
			for epid, tmpEpInline := range tmpEpInlines {
				if res.EpisodeInlineResp == nil {
					res.EpisodeInlineResp = make(map[int32]*pgcinline.EpisodeCard)
				}
				res.EpisodeInlineResp[epid] = tmpEpInline
			}
			muEpInline.Unlock()
			return nil
		})
	}
	if params.SeasonReq != nil && len(params.SeasonReq.Sids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.SeasonResp, errTmp = s.bangumiDao.Cards(ctx, params.SeasonReq.Sids); errTmp != nil {
				log.Error("material SeasonCards(%+v) error(%v)", params.SeasonReq.Sids, errTmp)
				return errTmp
			}
			var epids []int32
			for _, tmpSeason := range res.SeasonResp {
				if tmpSeason == nil || tmpSeason.FirstEpInfo == nil || tmpSeason.FirstEpInfo.Id == 0 {
					continue
				}
				epids = append(epids, tmpSeason.FirstEpInfo.Id)
			}
			// 获取ep信息(秒开)
			var tmpEpInlines map[int32]*pgcinline.EpisodeCard
			if tmpEpInlines, errTmp = s.bangumiDao.InlineCardsAll(ctx, epids, deviceInfo.MobiApp, deviceInfo.Platform, deviceInfo.Device, deviceInfo.Build); errTmp != nil {
				log.Error("material season InlineCardsAll(%+v, %v, %v, %v, %v) error(%v)", epids, deviceInfo.MobiApp, deviceInfo.Platform, deviceInfo.Device, deviceInfo.Build, errTmp)
				return nil
			}
			muEpInline.Lock()
			for epid, tmpEpInline := range tmpEpInlines {
				if res.EpisodeInlineResp == nil {
					res.EpisodeInlineResp = make(map[int32]*pgcinline.EpisodeCard)
				}
				res.EpisodeInlineResp[epid] = tmpEpInline
			}
			muEpInline.Unlock()
			return nil
		})
	}
	if params.AccountCardReq != nil && len(params.AccountCardReq.Mids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.AccountCardResp, errTmp = s.accountDao.Cards3All(ctx, params.AccountCardReq.Mids); err != nil {
				log.Error("material Account Cards3All(%+v) error(%v)", params.AccountCardReq.Mids, errTmp)
				return errTmp
			}
			return nil
		})
	}
	if params.UGCViewReq != nil && len(params.UGCViewReq.Aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.UGCViewResp, errTmp = s.archiveDao.ViewsAll(ctx, params.UGCViewReq.Aids); errTmp != nil {
				log.Error("material ViewsAll(%+v) error(%v)", params.UGCViewReq.Aids, errTmp)
				return errTmp
			}
			return nil
		})
	}
	if params.OGVViewReq != nil && params.OGVViewReq.Sid != 0 {
		eg.Go(func(ctx context.Context) error {
			var errTmp error
			if res.OGVViewResp, errTmp = s.bangumiDao.View(ctx, params.Mid, params.OGVViewReq.Sid, params.OGVViewReq.AccessKey, params.OGVViewReq.Cookie, deviceInfo.MobiApp, deviceInfo.Platform, params.Buvid, params.OGVViewReq.Referer, deviceInfo.Build); errTmp != nil {
				log.Error("material View(%+v) error(%v)", params.OGVViewReq.Sid, errTmp)
				return errTmp
			}
			return nil
		})
	}
	if params.SerialInfosReq != nil {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			if res.SerialInfosResp, localErr = s.serialInfoIntegrate(c, *params.SerialInfosReq); localErr != nil {
				log.Error("material s.serialInfoIntegrate error(%+v), req(%+v)", localErr, *params.SerialInfosReq)
				return localErr
			}
			return nil
		})
	}
	if params.SerialArcsReq != nil {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			if res.SerialArcsResp, localErr = s.serialOidsByPageIntegrate(c, *params.SerialArcsReq); localErr != nil {
				log.Error("material s.serialOidsByPageIntegrate error(%+v), req(%+v)", localErr, *params.SerialArcsReq)
				return localErr
			}
			return nil
		})
	}
	if params.ChannelInfosReq != nil {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			if res.ChannelInfosResp, localErr = s.channelInfoIntegrate(ctx, params.ChannelInfosReq, params.Mid, params.Buvid, deviceInfo); localErr != nil {
				log.Error("material s.channelInfoIntegrate error(%+v), req(%+v), mid(%d), buvid(%s), dev(%+v)",
					localErr, *params.ChannelInfosReq, params.Mid, params.Buvid, deviceInfo)
				return localErr
			}
			return nil
		})
	}
	if params.ChannelArcsReq != nil {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			if res.ChannelArcsResp, localErr = s.channelArcIntegrate(ctx, params.ChannelArcsReq, params.Mid, params.Buvid, deviceInfo); localErr != nil {
				log.Error("material s.channelArcIntegrate error(%+v), req(%+v), mid(%d), buvid(%s), dev(%+v)",
					localErr, *params.ChannelArcsReq, params.Mid, params.Buvid, deviceInfo)
				return localErr
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		b, _ := json.Marshal(params)
		log.Error("material eg.Wait(%v) error(%v)", string(b), err)
	}
	return
}

// formItem 统一item处理
func (s *Service) formItem(carContext *commonmdl.CarContext, deviceInfo model.DeviceInfo) (res *commonmdl.Item) { // nolint:gocognit
	if carContext.OriginData == nil {
		return
	}
	res = new(commonmdl.Item)
	switch carContext.OriginData.MaterialType {
	case commonmdl.MaterialTypeUGC:
		if carContext.ArchiveResp == nil {
			return nil
		}
		arc, ok := carContext.ArchiveResp[carContext.OriginData.Oid]
		if !ok {
			return nil
		}
		// 过滤互动视频
		if arc.Arc.AttrVal(archivegrpc.AttrBitSteinsGate) == archivegrpc.AttrYes {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeUGC
		res.Otype = commonmdl.OtypeUGC
		res.Oid = carContext.OriginData.Oid
		res.Cid = arc.Arc.FirstCid
		res.Title = arc.Arc.Title
		res.Cover = arc.Arc.Pic
		var cid = arc.Arc.FirstCid
		if carContext.OriginData.Cid != 0 {
			cid = carContext.OriginData.Cid
		}
		res.Url = model.FillURI(model.GotoAv, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(arc.Arc.Aid, 10), model.AvPlayHandlerGRPC(arc, cid, true))
		res.Author = &commonmdl.Author{
			Mid:  arc.Arc.Author.Mid,
			Name: arc.Arc.Author.Name,
			Face: arc.Arc.Author.Face,
		}
		res.PlayCount = int(arc.Arc.Stat.View)
		res.DanmakuCount = int(arc.Arc.Stat.Danmaku)
		res.Duration = arc.Arc.Duration
		res.FavCount = int(arc.Arc.Stat.Share)
		res.ReplyCount = int(arc.Arc.Stat.Reply)
		res.Pubtime = arc.Arc.PubDate
		res.Desc = arc.Arc.Desc
	case commonmdl.MaterialTypeUGCPlus:
		if carContext.ArchivePlusResp == nil {
			return nil
		}
		arc, ok := carContext.ArchivePlusResp[carContext.OriginData.Oid]
		if !ok {
			return nil
		}
		// 过滤互动视频
		if arc.Player.Arc.AttrVal(archivegrpc.AttrBitSteinsGate) == archivegrpc.AttrYes {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeUGC
		res.Otype = commonmdl.OtypeUGC
		res.Oid = carContext.OriginData.Oid
		res.Cid = arc.Player.Arc.FirstCid
		res.Title = arc.Player.Arc.Title
		res.Cover = arc.Player.Arc.Pic
		var cid = arc.Player.Arc.FirstCid
		if carContext.OriginData.Cid != 0 {
			cid = carContext.OriginData.Cid
		}
		res.Url = model.FillURI(model.GotoAv, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(arc.Player.Arc.Aid, 10), model.AvPlayHandlerGRPC(arc.Player, cid, true))
		res.Author = &commonmdl.Author{
			Mid:  arc.Player.Arc.Author.Mid,
			Name: arc.Player.Arc.Author.Name,
			Face: arc.Player.Arc.Author.Face,
		}
		res.PlayCount = int(arc.Player.Arc.Stat.View)
		res.DanmakuCount = int(arc.Player.Arc.Stat.Danmaku)
		res.Duration = arc.Player.Arc.Duration
		res.FavCount = int(arc.Player.Arc.Stat.Share)
		res.ReplyCount = int(arc.Player.Arc.Stat.Reply)
		res.Pubtime = arc.Player.Arc.PubDate
		res.Desc = arc.Player.Arc.Desc
		if arc.View.Videos > 1 {
			res.ItemType = commonmdl.ItemTypeUGCMulti
			res.ArcCountShow = fmt.Sprintf("共%d集", arc.View.Videos)
			for _, p := range arc.View.Pages {
				if p == nil || p.Cid != cid {
					continue
				}
				res.Duration = p.Duration
			}
		} else {
			res.ItemType = commonmdl.ItemTypeUGCSingle
		}
	case commonmdl.MaterialTypeOGVEP:
		if carContext.EpisodeResp == nil {
			return nil
		}
		ep, ok := carContext.EpisodeResp[int32(carContext.OriginData.Oid)]
		if !ok {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeOGV
		res.Otype = commonmdl.OtypePGC
		res.Oid = int64(ep.Season.SeasonId)
		res.Cid = int64(ep.EpisodeId)
		res.Title = ep.Season.Title
		res.Cover = ep.Cover
		res.LandscapeCover = ep.Cover
		res.Url = model.FillURI(model.GotoPGC, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(int64(ep.Season.SeasonId), 10), model.ParamHandler(nil, int64(ep.EpisodeId), 0, "", "", ""))
		if epInline, ok := carContext.EpisodeInlineResp[int32(carContext.OriginData.Oid)]; ok {
			res.Url = model.FillURI("", model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, res.Url, model.PGCPlayHandler(epInline))
		}
		if ep.Season != nil {
			if ep.Season.Stat != nil {
				res.PlayCount = int(ep.Season.Stat.View)
				res.DanmakuCount = int(ep.Season.Stat.Danmaku)
				res.FavCount = int(ep.Season.Stat.Follow)
			}
			if ep.Season.BadgeInfo != nil {
				res.Badge = &commonmdl.Badge{
					Text:         ep.Season.BadgeInfo.Text,
					BgColorDay:   ep.Season.BadgeInfo.BgColor,
					BgColorNight: ep.Season.BadgeInfo.BgColorNight,
					BgStyle:      "fill",
				}
			}
		}
		res.Desc = ep.Season.NewEpShow
		res.Duration = int64(math.Ceil(float64(ep.Duration) / 1000))
	case commonmdl.MaterialTypeOGVSeaon:
		if len(carContext.SeasonResp) == 0 {
			return nil
		}
		season := carContext.SeasonResp[int32(carContext.OriginData.Oid)]
		if season == nil || season.FirstEpInfo == nil {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeOGV
		res.Otype = commonmdl.OtypePGC
		res.Oid = int64(season.SeasonId)
		res.Cid = int64(season.FirstEpInfo.Id)
		res.Title = season.Title
		res.Cover = season.FirstEpInfo.Cover
		res.LandscapeCover = season.FirstEpInfo.Cover
		res.Url = model.FillURI(model.GotoPGC, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(int64(season.SeasonId), 10), model.ParamHandler(nil, int64(season.FirstEpInfo.Id), 0, "", "", ""))
		if epInline, ok := carContext.EpisodeInlineResp[season.FirstEpInfo.Id]; ok {
			res.Url = model.FillURI("", model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, res.Url, model.PGCPlayHandler(epInline))
		}
		if season.Stat != nil {
			res.PlayCount = int(season.Stat.View)
			res.DanmakuCount = int(season.Stat.Danmaku)
			res.FavCount = int(season.Stat.Follow)
			res.ReplyCount = int(season.Stat.Reply)
		}
		res.Desc = season.NewEp.IndexShow
		res.Badge = &commonmdl.Badge{
			Text:         season.BadgeInfo.Text,
			BgColorDay:   season.BadgeInfo.BgColor,
			BgColorNight: season.BadgeInfo.BgColorNight,
			BgStyle:      "fill",
		}
	case commonmdl.MaterialTypeUGCView:
		if carContext.UGCViewResp == nil {
			return nil
		}
		p, ok := carContext.UGCViewResp[carContext.OriginData.Oid]
		if !ok {
			return nil
		}
		if len(p.Pages) == 0 || !p.Arc.IsNormal() || (p.Arc.AttrVal(archivegrpc.AttrBitIsPUGVPay) == archivegrpc.AttrYes) {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeUGC
		res.Otype = commonmdl.OtypeUGC
		res.Oid = carContext.OriginData.Oid
		res.Cid = p.Arc.FirstCid
		res.Title = p.Arc.Title
		res.Cover = p.Arc.Pic
		res.Cid = p.Arc.FirstCid
		res.Url = model.FillURI(model.GotoAv, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(p.Arc.Aid, 10), nil)
		res.Author = &commonmdl.Author{
			Mid:  p.Arc.Author.Mid,
			Name: p.Arc.Author.Name,
			Face: p.Arc.Author.Face,
		}
		res.PlayCount = int(p.Arc.Stat.View)
		res.DanmakuCount = int(p.Arc.Stat.Danmaku)
		res.Duration = p.Arc.Duration
		res.FavCount = int(p.Arc.Stat.Share)
		res.Pubtime = p.Arc.PubDate
		res.Desc = p.Arc.Desc
		res.ReplyCount = int(p.Arc.Stat.Reply)
		if carContext.UGCViewResp != nil {
			for _, page := range carContext.UGCViewResp[carContext.OriginData.Oid].Pages {
				var tmpPlayList = &commonmdl.Playlist{
					Title:     page.Part,
					Aid:       carContext.OriginData.Oid,
					Cid:       page.Cid,
					Duration:  page.Duration,
					LongTitle: page.Part,
					Dimension: &commonmdl.Dimension{
						Height: page.Dimension.Height,
						Width:  page.Dimension.Width,
						Rotate: page.Dimension.Rotate,
					},
				}
				if bvid, err := model.GetBvID(carContext.OriginData.Oid); err == nil {
					tmpPlayList.ShareURL = model.FillURI(model.GotoWebBV, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, bvid, model.SuffixHandler(fmt.Sprintf("p=%d", page.Page)))
				}
				res.Playlist = append(res.Playlist, tmpPlayList)
			}
		}
		// 详情页专属字段
		res.View = &commonmdl.View{
			PagesStyle: "horizontal",
			Introduction: &commonmdl.Introduction{
				Title: res.Title,
				Desc:  p.Arc.Desc,
			},
		}
		var splice = " "
		res.Introduction.Info = model.ViewInfo(res.Introduction.Info, model.StatString(int32(res.PlayCount), "播放"), splice)
		res.Introduction.Info = model.ViewInfo(res.Introduction.Info, model.StatString(int32(res.DanmakuCount), "弹幕"), splice)
		res.Introduction.Info = model.ViewInfo(res.Introduction.Info, model.PubDataString(res.Pubtime.Time()), splice)
		if bvid, errTmp := model.GetBvID(res.Oid); errTmp == nil {
			res.Introduction.Info = model.ViewInfo(res.Introduction.Info, bvid, splice)
		}
	case commonmdl.MaterialTypeOGVView:
		if carContext.OGVViewResp == nil {
			return nil
		}
		res.ItemType = commonmdl.ItemTypeOGV
		res.Otype = commonmdl.OtypePGC
		res.Oid = carContext.OGVViewResp.SeasonID
		res.Title = carContext.OGVViewResp.Title
		res.Cover = carContext.OGVViewResp.Cover
		if carContext.OGVViewResp.Stat != nil {
			res.PlayCount = int(carContext.OGVViewResp.Stat.Views)
			res.DanmakuCount = int(carContext.OGVViewResp.Stat.Danmakus)
			res.FavCount = int(carContext.OGVViewResp.Stat.Favorites)
			res.ReplyCount = int(carContext.OGVViewResp.Stat.Reply)
		}
		res.View = &commonmdl.View{
			SeasonType: carContext.OGVViewResp.Type,
			PagesStyle: "horizontal",
		}
		// 简介
		res.Introduction = new(commonmdl.Introduction)
		res.Introduction.FromOGVIntroduction(carContext.OGVViewResp)
		// 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
		if carContext.OGVViewResp.Type == 1 || carContext.OGVViewResp.Type == 4 {
			res.View.PagesStyle = "grid"
		}
		if carContext.OGVViewResp.BadgeInfo != nil && carContext.OGVViewResp.BadgeInfo.BgColor != "" && carContext.OGVViewResp.Badge != "" {
			// 因ogv pgc/view/v2/app/season接口不再下发badgeType字段，导致这里的角标逻辑与外侧卡片的逻辑不一致，存在内外角标不一致问题
			res.Badge = &commonmdl.Badge{
				Text:             carContext.OGVViewResp.Badge,
				TextColorDay:     "#FFFFFF",
				TextColorNight:   "#FFFFFF",
				BgColorDay:       carContext.OGVViewResp.BadgeInfo.BgColor,
				BgColorNight:     carContext.OGVViewResp.BadgeInfo.BgColor,
				BorderColorDay:   carContext.OGVViewResp.BadgeInfo.BgColor,
				BorderColorNight: carContext.OGVViewResp.BadgeInfo.BgColor,
				BgStyle:          model.BgStyleFill,
			}
		}
		if carContext.OGVViewResp.UserStatus != nil {
			if carContext.OGVViewResp.UserStatus.Follow == 1 {
				res.IsFollow = true
			}
			if carContext.OGVViewResp.UserStatus.Progress != nil {
				res.View.History = &commonmdl.History{
					Epid:     carContext.OGVViewResp.UserStatus.Progress.LastEpID,
					Progress: carContext.OGVViewResp.UserStatus.Progress.LastTime,
				}
			}
		}
		var positive, section []*commonmdl.Playlist
		for _, page := range carContext.OGVViewResp.Modules {
			for _, ep := range page.Data.Episodes {
				// 互动视频不展示
				if ep.Interaction != nil {
					continue
				}
				var tmpPlayList = &commonmdl.Playlist{
					Title:      ep.Title,
					Aid:        ep.Aid,
					Cid:        ep.Cid,
					Epid:       ep.ID,
					LongTitle:  ep.LongTitle,
					ShareURL:   model.FillURI(model.GotoWebPGC, 0, 0, strconv.FormatInt(ep.ID, 10), nil),
					ReplyCount: int(ep.Stat.Reply),
					Dimension: &commonmdl.Dimension{
						Height: ep.Dimension.Height,
						Width:  ep.Dimension.Width,
						Rotate: ep.Dimension.Rotate,
					},
					Duration: int64(math.Ceil(float64(ep.Duration) / 1000)),
					Cover:    ep.Cover,
				}
				if ep.LongTitle == "" {
					tmpPlayList.Title = ep.Title
				}
				if ep.BadgeInfo != nil && ep.BadgeInfo.Text != "" {
					tmpPlayList.Badge = reasonStyleFrom(model.PGCBageType[ep.BadgeType], ep.Badge)
				}
				switch page.Style {
				case "positive":
					positive = append(positive, tmpPlayList)
				case "section":
					section = append(section, tmpPlayList)
				default:
					continue
				}
			}
		}
		res.Playlist = section
		if len(positive) > 0 {
			res.Playlist = positive
		}
		if len(res.Playlist) == 0 {
			return nil
		}
	case commonmdl.MaterialTypeVideoSerial, commonmdl.MaterialTypeFmSerial:
		if carContext.ArchiveResp == nil || carContext.SerialInfosResp == nil {
			return nil
		}
		var data map[int64]*commonmdl.SerialInfo
		var itemType commonmdl.ItemType
		if carContext.OriginData.MaterialType == commonmdl.MaterialTypeVideoSerial {
			data = carContext.SerialInfosResp.Video
			itemType = commonmdl.ItemTypeVideoSerial
		} else {
			data = carContext.SerialInfosResp.FmCommon
			itemType = commonmdl.ItemTypeFmSerial
		}
		arc := carContext.ArchiveResp[carContext.OriginData.Cid]
		serial := data[carContext.OriginData.Oid]
		if serial == nil || arc == nil {
			return nil
		}
		// 过滤互动视频
		if arc.Arc.AttrVal(archivegrpc.AttrBitSteinsGate) == archivegrpc.AttrYes {
			return nil
		}
		res.ItemType = itemType
		res.ItemId = carContext.OriginData.Oid
		res.Otype = commonmdl.OtypeUGC // 合集目前只出ugc
		res.Oid = arc.Arc.Aid
		res.Cid = arc.Arc.FirstCid
		res.Title = arc.Arc.Title
		res.Cover = arc.Arc.Pic
		var cid = arc.Arc.FirstCid
		if carContext.OriginData.Cid != 0 {
			cid = carContext.OriginData.Cid
		}
		res.Url = model.FillURI(model.GotoAv, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(arc.Arc.Aid, 10), model.AvPlayHandlerGRPC(arc, cid, true))
		res.Author = &commonmdl.Author{
			Mid:  arc.Arc.Author.Mid,
			Name: arc.Arc.Author.Name,
			Face: arc.Arc.Author.Face,
		}
		res.PlayCount = int(arc.Arc.Stat.View)
		res.DanmakuCount = int(arc.Arc.Stat.Danmaku)
		res.Duration = arc.Arc.Duration
		res.FavCount = int(arc.Arc.Stat.Share)
		res.ReplyCount = int(arc.Arc.Stat.Reply)
		res.Pubtime = arc.Arc.PubDate
		res.Desc = arc.Arc.Desc
		if serial.Count > 0 {
			res.ArcCountShow = fmt.Sprintf("共%d集", serial.Count)
		}
	case commonmdl.MaterialTypeVideoChannel, commonmdl.MaterialTypeFmChannel:
		if carContext.ArchiveResp == nil || carContext.ChannelInfosResp == nil {
			return nil
		}
		var channelMap map[int64]*commonmdl.ChannelInfo
		var itemType commonmdl.ItemType
		if carContext.OriginData.MaterialType == commonmdl.MaterialTypeVideoChannel {
			channelMap = carContext.ChannelInfosResp.Video
			itemType = commonmdl.ItemTypeVideoChannel
		} else {
			channelMap = carContext.ChannelInfosResp.Fm
			itemType = commonmdl.ItemTypeFmChannel
		}
		arc := carContext.ArchiveResp[carContext.OriginData.Cid]
		channel := channelMap[carContext.OriginData.Oid]
		if channel == nil {
			return nil
		}
		f := func() {
			res.ItemType = itemType
			res.ItemId = carContext.OriginData.Oid
			res.Title = channel.Title
			res.SubTitle = channel.SubTitle
			res.Cover = channel.Cover
			res.HotRate = channel.HotRate
			//if channel.Count > 0 {
			/// 本期暂时不下发
			//res.ArcCountShow = fmt.Sprintf("共%d集", channel.Count)
			//}
		}
		// 过滤互动视频
		if arc == nil || arc.Arc.AttrVal(archivegrpc.AttrBitSteinsGate) == archivegrpc.AttrYes {
			f()
			return
		}
		f()
		res.Otype = commonmdl.OtypeUGC
		res.Oid = arc.Arc.GetAid()
		res.Cid = arc.Arc.GetFirstCid()
		if arc.Arc.GetFirstCid() > 0 {
			res.Url = model.FillURI(model.GotoAv, model.Plat(deviceInfo.MobiApp, deviceInfo.Device), deviceInfo.Build, strconv.FormatInt(arc.Arc.GetAid(), 10), model.AvPlayHandlerGRPC(arc, arc.Arc.GetFirstCid(), true))
		}
	default:
		return nil
	}
	// 后置校验 待完善
	if res.Url == "" && carContext.OriginData.MaterialType != commonmdl.MaterialTypeOGVView {
		log.Warn("formItem res.Url invalid (%+v)", res)
	}
	return
}

func reasonStyleFrom(style string, text string) *commonmdl.Badge {
	if text == "" {
		return nil
	}
	res := &commonmdl.Badge{
		Text:    text,
		BgStyle: "fill",
	}
	switch style {
	case model.BgColorRed:
		res.TextColorDay = "#FFFFFF"
		res.BgColorDay = "#FF5377"
	case model.BgColorBlue:
		res.TextColorDay = "#FFFFFF"
		res.BgColorDay = "#20AAE2"
	case model.BgColorYellow:
		res.TextColorDay = "#7E2D11"
		res.BgColorDay = "#FFB112"
	default:
		return nil
	}
	return res
}

func (s *Service) v23debug(mid int64, build int) bool {
	if build < build203 {
		return false
	}
	if s.c.V23Debug == nil || !s.c.V23Debug.Switch {
		return false
	}
	for _, v := range s.c.V23Debug.Mids {
		if v == mid {
			return true
		}
	}
	return false
}
