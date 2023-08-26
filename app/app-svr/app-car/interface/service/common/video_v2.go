package common

import (
	"context"
	"fmt"
	"hash/crc32"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	mediarpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	api "git.bilibili.co/bapis/bapis-go/serial/service"
)

const (
	// tab id
	tabFeedId         = 1
	tabPopularId      = 2
	tabUgcMusicId     = 3
	tabUgcDanceId     = 4
	tabUgcGameId      = 5
	tabUgcKnowledgeId = 6
	tabUgcLifeId      = 7
	tabUgcMiscNewsId  = 8 // 咨询
	tabUgcCarId       = 9
	tabPgcRecId       = 11 // 番剧
	tabPgcCNId        = 12 // 国创
	tabMovieId        = 14 // 电影
	tabDocumentaryId  = 15 // 纪录片

	tab61ChildhoodId   = 16 // 61-童年回来了
	tab61EdenId        = 17 // 61-小朋友乐园
	tabDWRicePuddingId = 18 // 端午-"粽”有陪伴
	tabDWEden          = 19 // 端午-小朋友乐园
	tabXiaoPengLYJId   = 20 // 小鹏727-老友记
	tabXiaoPengZJLId   = 21 // 小鹏727-周杰伦

	// 视频tab类型
	tabTypePgc = 2

	// ugc分区id
	ugcRegionMusic     = 3
	ugcRegionDance     = 129
	ugcRegionGame      = 4
	ugcRegionKnowledge = 36
	ugcRegionLife      = 160
	ugcRegionMiscNews  = 202
	ugcRegionCar       = 223

	// pgc seasonType枚举值
	SeasonTypeSeries      = 1 // 番剧
	SeasonTypeMovie       = 2 // 电影
	SeasonTypeDocumentary = 3 // 纪录片
	SeasonTypeCartoon     = 4 // 国漫

	suffixFollow    = "追番"
	hisProgressOver = -1
	oneThousand     = 1000
	ten             = 10

	build203 = 2030001

	// fm和视频tab位置互换
	_hitExchange = 1

	pgcIndexOrderSynthetical = "8"

	// pgc索引页
	_appSource = 0 // 0=app
	_webSource = 1 // 1=web
)

// RegionMeta 提供给算法caibangyu视频分区tab列表.
func (s *Service) RegionMeta(_ context.Context) ([]*commonmdl.VideoTabV2Item, error) {
	res := make([]*commonmdl.VideoTabV2Item, 0)
	for _, v := range s.c.VideoTabsV2Conf.VideoTabs {
		res = append(res, &commonmdl.VideoTabV2Item{
			Type:      v.Type,
			Id:        v.Id,
			Name:      v.Name,
			IsDefault: v.IsDefault,
		})
	}
	return res, nil
}

// VideoTabs 获取视频下tab列表。
// TODO 开发管理后台时将s.c.VideoTabsV2Conf写入数据库，目前暂时采用配置. 届时type字段区分分区类型，将不同分区的特有属性作为值对象序列化保存.
func (s *Service) VideoTabs(c context.Context, req *commonmdl.VideoTabsReq) (*commonmdl.VideoTabV2Resp, error) {
	if req.Channel == "xiaopeng" && s.c.EnableXP727Tabs {
		return s.xiaopengCustomRegions(c, req)
	}
	regions, err := s.aiSmartRegions(c, req)
	if err != nil || regions == nil || len(regions.Items) == 0 {
		log.Errorc(c, "VideoTabs s.aiSmartRegions err=%+v,regions=%+v", err, regions)
		return s.commonRegions(c, req)
	}
	return regions, nil
}

// aiSmartRegions ai排序tab.
func (s *Service) aiSmartRegions(c context.Context, req *commonmdl.VideoTabsReq) (*commonmdl.VideoTabV2Resp, error) {
	aiRegions, err := s.fmDao.AIRegionList(c, req.Mid, req.Buvid, req.DeviceInfo)
	if err != nil {
		log.Errorc(c, "aiSmartRegions s.fmDao.AIRegionList err=%+v", err)
		return nil, err
	}

	idTabMap := make(map[int64]*commonmdl.VideoTabV2Item)
	var aiTab, popuplar *commonmdl.VideoTabV2Item
	for _, v := range s.c.VideoTabsV2Conf.VideoTabs {
		t := &commonmdl.VideoTabV2Item{
			Type: v.Type,
			Id:   v.Id,
			Name: v.Name,
		}
		idTabMap[v.Id] = t
		if v.Id == tabFeedId {
			aiTab = t
		}
		if v.Id == tabPopularId {
			popuplar = t
		}
	}

	// 推荐、热门位置不受ai影响，固定前两位
	res := make([]*commonmdl.VideoTabV2Item, 0)
	if req.Mid > 0 {
		if aiTab != nil {
			res = append(res, aiTab)
		}
	}
	if popuplar != nil {
		res = append(res, popuplar)
	}

	for _, v := range aiRegions {
		if v == nil {
			continue
		}
		if v.Id == tabFeedId || v.Id == tabPopularId {
			continue
		}
		originalTab := idTabMap[v.Id]
		if originalTab == nil {
			continue
		}
		if originalTab.Type != int64(v.Type) {
			continue
		}
		res = append(res, originalTab)
	}
	if len(res) == 0 {
		return nil, nil
	}
	res[0].IsDefault = true
	return &commonmdl.VideoTabV2Resp{
		Items:    res,
		Exchange: s.TabExchange(req),
	}, nil
}

// commonRegions 正常的tab.
func (s *Service) commonRegions(_ context.Context, req *commonmdl.VideoTabsReq) (*commonmdl.VideoTabV2Resp, error) {
	res := make([]*commonmdl.VideoTabV2Item, 0)
	for _, v := range s.c.VideoTabsV2Conf.VideoTabs {
		if tabFeedId == v.Id && req.Mid <= 0 {
			continue
		}
		res = append(res, &commonmdl.VideoTabV2Item{
			Type: v.Type,
			Id:   v.Id,
			Name: v.Name,
		})
	}
	res[0].IsDefault = true
	return &commonmdl.VideoTabV2Resp{
		Items:    res,
		Exchange: s.TabExchange(req),
	}, nil
}

// xiaopengCustomRegions 小鹏特别tab.
func (s *Service) xiaopengCustomRegions(_ context.Context, req *commonmdl.VideoTabsReq) (*commonmdl.VideoTabV2Resp, error) {
	idTabMap := make(map[int64]*commonmdl.VideoTabV2Item)
	for _, v := range s.c.VideoTabsV2Conf.VideoTabs {
		if tabFeedId == v.Id && req.Mid <= 0 {
			continue
		}
		idTabMap[v.Id] = &commonmdl.VideoTabV2Item{
			Type: v.Type,
			Id:   v.Id,
			Name: v.Name,
		}
	}

	// 小鹏汽车727特别tab
	if s.c.CustomTabLYJ != nil && s.c.CustomTabLYJ.EnableCustomModule {
		idTabMap[tabXiaoPengLYJId] = &commonmdl.VideoTabV2Item{
			Type: tabTypePgc,
			Id:   tabXiaoPengLYJId,
			Name: "老友记",
		}
	}
	if s.c.CustomTabZhouJieLun != nil && s.c.CustomTabZhouJieLun.EnableCustomModule {
		idTabMap[tabXiaoPengZJLId] = &commonmdl.VideoTabV2Item{
			Type: 1,
			Id:   tabXiaoPengZJLId,
			Name: "周杰伦",
		}
	}

	xipengTabIds := []int64{tabFeedId, tabXiaoPengLYJId, tabXiaoPengZJLId, tabPopularId, tabDocumentaryId,
		tabUgcMusicId, tabUgcKnowledgeId, tabPgcCNId, tabPgcRecId, tabMovieId, tabUgcGameId, tabUgcLifeId,
		tabUgcCarId, tabUgcMiscNewsId, tabUgcDanceId}
	res := make([]*commonmdl.VideoTabV2Item, 0)
	for _, v := range xipengTabIds {
		tab := idTabMap[v]
		if tab == nil {
			continue
		}
		res = append(res, tab)
	}
	res[0].IsDefault = true
	return &commonmdl.VideoTabV2Resp{
		Items:    res,
		Exchange: s.TabExchange(req),
	}, nil
}

// VideoTabCards 获取tab下卡片.
func (s *Service) VideoTabCards(c context.Context, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	req.PageNext = s.validatePageNext(req.PageNext)
	switch req.TabId {
	case tabFeedId:
		return s.videoTabFeedCards(c, req)
	case tabPopularId:
		return s.videoTabHotCards(c, req)
	case tabUgcMusicId, tabUgcDanceId, tabUgcGameId, tabUgcKnowledgeId, tabUgcLifeId, tabUgcMiscNewsId, tabUgcCarId:
		typeIdRegionIdMap := map[int64]int64{
			tabUgcMusicId:     ugcRegionMusic,
			tabUgcDanceId:     ugcRegionDance,
			tabUgcGameId:      ugcRegionGame,
			tabUgcKnowledgeId: ugcRegionKnowledge,
			tabUgcLifeId:      ugcRegionLife,
			tabUgcMiscNewsId:  ugcRegionMiscNews,
			tabUgcCarId:       ugcRegionCar,
		}
		return s.videoTabUgcRegionCards(c, typeIdRegionIdMap[req.TabId], req)
	case tabPgcRecId, tabPgcCNId, tabMovieId, tabDocumentaryId:
		typeIdRecTypeMap := map[int64]int{
			tabPgcRecId:      SeasonTypeSeries,
			tabPgcCNId:       SeasonTypeCartoon,
			tabMovieId:       SeasonTypeMovie,
			tabDocumentaryId: SeasonTypeDocumentary,
		}
		return s.videoTabPgcXXRecCards(c, typeIdRecTypeMap[req.TabId], req)
	case tab61ChildhoodId, tab61EdenId, tabDWRicePuddingId, tabDWEden, tabXiaoPengLYJId, tabXiaoPengZJLId:
		return s.videoTabVacationCards(c, req)
	default:
		return nil, xecode.AppMediaNotData
	}
}

// videoTabVacationCards 假期临时tab.
func (s *Service) videoTabVacationCards(c context.Context, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	var cardType = commonmdl.MaterialTypeUGC
	var oids []int64
	switch req.TabId {
	case tab61ChildhoodId:
		if s.c.CustomTab61Childhood != nil {
			oids = s.c.CustomTab61Childhood.ChannelAids[req.Channel]
		}
	case tab61EdenId:
		if s.c.CustomTab61Eden != nil {
			oids = s.c.CustomTab61Eden.ChannelAids[req.Channel]
		}
	case tabDWRicePuddingId:
		if s.c.CustomTabDWRicePudding != nil {
			oids = s.c.CustomTabDWRicePudding.ChannelAids[req.Channel]
		}
	case tabDWEden:
		if s.c.CustomTabDWEden != nil {
			oids = s.c.CustomTabDWEden.ChannelAids[req.Channel]
		}
	case tabXiaoPengLYJId:
		if s.c.CustomTabLYJ != nil {
			oids = s.c.CustomTabLYJ.ChannelAids[req.Channel]
			cardType = commonmdl.MaterialTypeOGVSeaon
		}
	case tabXiaoPengZJLId:
		if s.c.CustomTabZhouJieLun != nil {
			oids = s.c.CustomTabZhouJieLun.ChannelAids[req.Channel]
		}
	default:
		// nop
	}
	if len(oids) == 0 {
		return nil, nil
	}

	sIndex := (req.PageNext.Pn - 1) * req.PageNext.Ps
	if sIndex >= len(oids) {
		return nil, nil
	}
	eIndex := sIndex + req.PageNext.Ps
	if eIndex > len(oids) {
		oids = oids[sIndex:]
	} else {
		oids = oids[sIndex:eIndex]
	}

	var aids, sids []int64
	if cardType == commonmdl.MaterialTypeOGVSeaon {
		sids = oids
	} else if cardType == commonmdl.MaterialTypeUGC {
		aids = oids
	}
	return s.buildTabCardResp(c, cardType, req.DeviceInfo, req.PageNext.Ps, req.PageNext.Pn, req.Mid, req.Buvid, aids, sids, ogvSimpleItemAfterHandler, req.TabId)
}

// videoTabPgcXXRecCards 查询'PGC-xx推荐'tab下卡片. 不支持分页
func (s *Service) videoTabPgcXXRecCards(c context.Context, pgcRegionId int, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	var tp int32 = _appSource
	if pgcRegionId == SeasonTypeSeries {
		tp = _webSource
	}
	res, err := s.mediaRpc.IndexSearch(c, &mediarpc.IndexSearchReq{
		IndexType: int32(pgcRegionId),
		Type:      tp,
		SubType:   0,
		Order:     pgcIndexOrderSynthetical,
		Sort:      1,
		Pn:        int32(req.PageNext.Pn),
		Ps:        int32(req.PageNext.Ps),
	})
	if err != nil {
		log.Error("videoTabPgcXXRecCards s.mediaRpc.IndexSearch err=%+v, mid=%d, buvid=%s, pgcRegionId=%d,page=%v", err, req.Mid, req.Buvid, pgcRegionId, req.PageNext)
		return nil, ecode.ServerErr
	}
	if res == nil || len(res.Items) == 0 {
		log.Warn("videoTabPgcXXRecCards res has no card. mid=%d, buvid=%s, pgcRegionId=%d,page=%v", req.Mid, req.Buvid, pgcRegionId, req.PageNext)
		return &commonmdl.VideoTabCardResp{
			PageNext: &commonmdl.PageNext{
				Pn: req.PageNext.Pn + 1,
				Ps: req.PageNext.Ps,
			},
		}, nil
	}

	sids := make([]int64, 0)
	for _, v := range res.Items {
		if v == nil || v.SeasonId <= 0 || v.SeasonType != int32(pgcRegionId) {
			continue
		}
		sids = append(sids, int64(v.SeasonId))
	}
	return s.buildTabCardResp(c, commonmdl.MaterialTypeOGVSeaon, req.DeviceInfo, req.PageNext.Ps, req.PageNext.Pn, req.Mid, req.Buvid, nil, sids, ogvSimpleItemAfterHandler, req.TabId)
}

// ogvSimpleItemAfterHandler.
func ogvSimpleItemAfterHandler(item *commonmdl.Item, season *seasongrpc.CardInfoProto) {
	switch season.SeasonType {
	case SeasonTypeSeries, SeasonTypeCartoon:
		item.SubTitle = model.StatString(int32(season.Stat.Follow), suffixFollow)
	case SeasonTypeMovie, SeasonTypeDocumentary:
		item.SubTitle = season.Subtitle
	default:
		// no-op
	}
}

const (
	// banner的卡片形式
	bannerMultiple   = 1
	bannerCollection = 2
)

// buildTabCardResp 负责构建视频tab下的卡片.
// nolint: gocognit
func (s *Service) buildTabCardResp(c context.Context, cardType commonmdl.MaterialType, device model.DeviceInfo, ps, pn int, mid int64,
	buvid string, aids []int64, sids []int64, simpleItemAfterHandler func(*commonmdl.Item, *seasongrpc.CardInfoProto), tabId int64) (*commonmdl.VideoTabCardResp, error) {
	if len(aids) == 0 && len(sids) == 0 {
		return &commonmdl.VideoTabCardResp{
			PageNext: &commonmdl.PageNext{
				Pn: pn + 1,
				Ps: ps,
			},
		}, nil
	}

	// banner
	eg := errgroup.WithContext(c)
	var bannerShow []*commonmdl.Item
	playlists := make([]*conf.BannerPlaylist, 0)
	for _, v := range s.c.BannerPlaylist {
		if v.TabId == tabId && cardType == commonmdl.MaterialType(v.MaterialType) {
			playlists = append(playlists, v)
		}
	}
	if len(playlists) > 0 && pn == 1 {
		// banner的卡片类型与tab的卡片类型一致
		eg.Go(func(ctx context.Context) error {
			for _, playlistInfo := range playlists {
				if playlistInfo.StyleType == bannerCollection && playlistInfo.ShowId > 0 {
					var bAids, bSids []int64
					if commonmdl.MaterialType(playlistInfo.MaterialType) == commonmdl.MaterialTypeUGC {
						bAids = []int64{playlistInfo.ShowId}
					} else {
						bSids = []int64{playlistInfo.ShowId}
					}
					creq := &commonItemsReq{
						Mid:                    mid,
						Buvid:                  buvid,
						Aids:                   bAids,
						Sids:                   bSids,
						SimpleItemAfterHandler: simpleItemAfterHandler,
					}
					banners, err := s.commonItems(ctx, device, creq)
					if err != nil {
						log.Error("buildTabCardResp banner1 s.commonItems err=%+v, cardType=%v, id=%d, buvid=%s.", err, cardType, playlistInfo.ShowId, buvid)
						continue
					}
					if len(banners) == 0 || banners[0] == nil {
						continue
					}
					banners[0].ArcCountShow = fmt.Sprintf("共%d集", len(playlistInfo.PlayList))
					banners[0].Wrapper = &commonmdl.Wrapper{
						Id:   playlistInfo.Id,
						Type: 0,
					}
					if playlistInfo.Title == "" {
						banners[0].Wrapper.Title = banners[0].Title
					} else {
						banners[0].Wrapper.Title = playlistInfo.Title
					}
					if playlistInfo.Cover == "" {
						if cardType == commonmdl.MaterialTypeOGVSeaon {
							banners[0].Wrapper.Cover = banners[0].LandscapeCover
						} else {
							banners[0].Wrapper.Cover = banners[0].Cover
						}
					} else {
						banners[0].Wrapper.Cover = playlistInfo.Cover
					}
					bannerShow = append(bannerShow, banners[0])
					continue
				}
				if playlistInfo.StyleType == bannerMultiple {
					var bAids, bSids []int64
					if commonmdl.MaterialType(playlistInfo.MaterialType) == commonmdl.MaterialTypeUGC {
						bAids = playlistInfo.PlayList
					} else {
						bSids = playlistInfo.PlayList
					}
					var (
						videoSerialIds  []int64
						videoChannelIds []int64
					)
					if s.v23debug(mid, device.Build) {
						videoSerialIds = []int64{111, 222}
						videoChannelIds = []int64{10009, 499}
					}
					creq := &commonItemsReq{
						Mid:                    mid,
						Buvid:                  buvid,
						Aids:                   bAids,
						Sids:                   bSids,
						VideoSerialIds:         videoSerialIds,
						VideoChannelIds:        videoChannelIds,
						SimpleItemAfterHandler: simpleItemAfterHandler,
					}
					banners, err := s.commonItems(ctx, device, creq)
					if err != nil {
						log.Error("buildTabCardResp banner2 s.commonItems err=%+v, cardType=%v, id=%d, buvid=%s.", err, cardType, playlistInfo.ShowId, buvid)
						continue
					}
					if commonmdl.MaterialType(playlistInfo.MaterialType) == commonmdl.MaterialTypeOGVSeaon {
						for _, v := range banners {
							if v != nil && v.LandscapeCover != "" {
								v.Cover = v.LandscapeCover
							}
						}
					}
					bannerShow = append(bannerShow, banners...)
					continue
				}
			}
			return nil
		})
	}
	// cards
	var (
		cards           []*commonmdl.Item
		videoSerialIds  []int64
		videoChannelIds []int64
	)
	eg.Go(func(ctx context.Context) (err error) {
		// todo del mock
		if s.v23debug(mid, device.Build) {
			videoSerialIds = []int64{111, 222}
			videoChannelIds = []int64{10009, 499}
		}

		creq := &commonItemsReq{
			Mid:                    mid,
			Buvid:                  buvid,
			Aids:                   aids,
			Sids:                   sids,
			VideoSerialIds:         videoSerialIds,
			VideoChannelIds:        videoChannelIds,
			SimpleItemAfterHandler: simpleItemAfterHandler,
		}
		cards, err = s.commonItems(ctx, device, creq)

		// todo del mock
		if s.v23debug(mid, device.Build) {
			for _, v := range cards {
				if v == nil {
					continue
				}
				v.Label = &commonmdl.Badge{
					Text:             "有更新",
					TextColorDay:     "#FF6699",
					TextColorNight:   "#FF6699",
					BorderColorDay:   "#FF6699",
					BorderColorNight: "#FF6699",
					BgStyle:          model.BgStyleFill,
				}
			}
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("buildTabCardResp eg.Wait() err=%+v, buvid=%s.", err, buvid)
		return nil, err
	}

	return &commonmdl.VideoTabCardResp{
		Items: &commonmdl.VideoTabCardItems{
			Cards:   cards,
			Banners: bannerShow,
		},
		PageNext: &commonmdl.PageNext{
			Pn: pn + 1,
			Ps: ps,
		},
	}, nil
}

type commonItemsReq struct {
	Mid        int64
	Buvid      string
	Aids, Sids []int64

	// 视频合集ids、视频合集-稿件映射
	VideoSerialIds      []int64
	VideoSerialIdAidMap map[int64]int64

	// FM合集ids、FM合集-稿件映射
	FmSerialIds      []int64
	FmSerialIdAidMap map[int64]int64

	// 视频频道ids、视频频道-稿件映射
	VideoChannelIds      []int64
	VideoChannelIdAidMap map[int64]int64

	// FM频道ids、FM频道-稿件映射
	FmChannelIds      []int64
	FmChannelIdAidMap map[int64]int64

	// 简单item后置处理器，用于处理非常简单的属性干预. 如果逻辑复杂满足不了，在上层方法处理.
	SimpleItemAfterHandler func(*commonmdl.Item, *seasongrpc.CardInfoProto) `json:"-"`
}

// commonItems 组装ugc稿件、ugc单p、ugc多p、pgc稿件、合集、频道卡片.
// 1、ugc稿件、ugc单p、ugc多p、pgc稿件支持历史记录秒开(优先)、首p秒开
// 2、合集支持指定稿件/历史记录/首p外漏和秒开
// 3、频道支持指定稿件/历史记录/首p秒开
func (s *Service) commonItems(c context.Context, device model.DeviceInfo, req *commonItemsReq) ([]*commonmdl.Item, error) { // nolint:gocognit
	// 查询历史记录
	var ugcHisMap map[int64]*hisApi.ModelHistory
	var pgcHisMap map[int64]*hisApi.ModelHistory
	eg := errgroup.WithContext(c)
	if len(req.Aids) > 0 || len(req.Sids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var hisErr error
			ugcHisMap, pgcHisMap, hisErr = s.historyDao.BatchProgress(c, req.Mid, req.Buvid, req.Aids, req.Sids)
			if hisErr != nil {
				log.Error("commonItems s.hisDao.Progress err=%+v, aids=%+v, sids=%+v, mid=%d, buvid=%s", hisErr, req.Aids, req.Sids, req.Mid, req.Buvid)
			}
			return nil
		})
	}
	hisFn := func(ctx context.Context) {
		serialHisReq := make([]*api.BatchSerial, 0)
		if len(req.VideoSerialIds) > 0 {
			for _, x := range req.VideoSerialIds {
				if _, ok := req.VideoSerialIdAidMap[x]; !ok {
					serialHisReq = append(serialHisReq, &api.BatchSerial{SerialId: x, BusinessSerialType: commonmdl.ItemTypeToSerialBusinessType[commonmdl.ItemTypeVideoSerial]})
				}
			}
		}
		if len(req.FmSerialIds) > 0 {
			for _, x := range req.FmSerialIds {
				if _, ok := req.FmSerialIdAidMap[x]; !ok {
					serialHisReq = append(serialHisReq, &api.BatchSerial{SerialId: x, BusinessSerialType: commonmdl.ItemTypeToSerialBusinessType[commonmdl.ItemTypeFmSerial]})
				}
			}
		}
		if len(serialHisReq) == 0 {
			return
		}
		progress, err := s.serialDao.SerialBatchProgress(ctx, req.Mid, req.Buvid, serialHisReq)
		if err != nil {
			log.Errorc(ctx, "commonItems s.serialDao.SerialBatchProgress err=%+v,buvid=%s.", err, req.Buvid)
			return
		}
		for _, v := range progress {
			if v == nil {
				continue
			}
			switch commonmdl.SerialBusinessTypeToItemType[v.BusinessSerialType] {
			case commonmdl.ItemTypeVideoSerial:
				req.VideoSerialIdAidMap[v.SerialId] = v.Episode
			case commonmdl.ItemTypeVideoChannel:
				req.VideoChannelIdAidMap[v.SerialId] = v.Episode
			case commonmdl.ItemTypeFmSerial:
				req.FmSerialIdAidMap[v.SerialId] = v.Episode
			case commonmdl.ItemTypeFmChannel:
				req.FmChannelIdAidMap[v.SerialId] = v.Episode
			default:
				// nop
			}
		}
	}
	firstFn := func(ctx context.Context) {
		serialArcsReq := new(commonmdl.SerialArcsReq)
		channelArcsReq := new(commonmdl.ChannelArcsReq)
		if len(req.VideoSerialIds) > 0 {
			sar := make([]*commonmdl.SerialArcReq, 0)
			for _, x := range req.VideoSerialIds {
				if _, ok := req.VideoSerialIdAidMap[x]; !ok {
					sar = append(sar, &commonmdl.SerialArcReq{
						SerialId:      x,
						SerialPageReq: commonmdl.SerialPageReq{Ps: 1},
					})
				}
			}
			serialArcsReq.Video = sar
		}
		if len(req.FmSerialIds) > 0 {
			sar := make([]*commonmdl.SerialArcReq, 0)
			for _, x := range req.FmSerialIds {
				if _, ok := req.FmSerialIdAidMap[x]; !ok {
					sar = append(sar, &commonmdl.SerialArcReq{
						SerialId:      x,
						SerialPageReq: commonmdl.SerialPageReq{Ps: 1},
					})
				}
			}
			serialArcsReq.FmCommon = sar
		}
		if len(req.VideoChannelIds) > 0 {
			sar := make([]*commonmdl.ChannelArcReq, 0)
			for _, x := range req.VideoChannelIds {
				if _, ok := req.VideoChannelIdAidMap[x]; !ok {
					sar = append(sar, &commonmdl.ChannelArcReq{
						ChanId: x,
						Ps:     1,
					})
				}
			}
			channelArcsReq.Video = sar
		}
		if len(req.FmChannelIds) > 0 {
			sar := make([]*commonmdl.ChannelArcReq, 0)
			for _, x := range req.FmChannelIds {
				if _, ok := req.FmChannelIdAidMap[x]; !ok {
					sar = append(sar, &commonmdl.ChannelArcReq{
						ChanId: x,
						Ps:     1,
					})
				}
			}
			channelArcsReq.Fm = sar
		}
		material, err := s.material(ctx, &commonmdl.Params{
			SerialArcsReq:  serialArcsReq,
			ChannelArcsReq: channelArcsReq,
			Mid:            req.Mid,
			Buvid:          req.Buvid,
		}, device)
		if err != nil || material == nil {
			log.Errorc(ctx, "commonItems s.material firstFn err=%+v, buvid=%s.", err, req.Buvid)
			return
		}
		if material.SerialArcsResp != nil {
			for _, x := range req.VideoSerialIds {
				if y := material.SerialArcsResp.Video[x]; y != nil && len(y.Aids) > 0 {
					req.VideoSerialIdAidMap[x] = y.Aids[0]
				}
			}
			for _, x := range req.FmSerialIds {
				if y := material.SerialArcsResp.FmCommon[x]; y != nil && len(y.Aids) > 0 {
					req.FmSerialIdAidMap[x] = y.Aids[0]
				}
			}
		}
		if material.ChannelArcsResp != nil {
			for _, x := range req.VideoChannelIds {
				if y := material.ChannelArcsResp.Video[x]; y != nil && len(y.Aids) > 0 {
					req.VideoChannelIdAidMap[x] = y.Aids[0]
				}
			}
			for _, x := range req.FmChannelIds {
				if y := material.ChannelArcsResp.Fm[x]; y != nil && len(y.Aids) > 0 {
					req.FmChannelIdAidMap[x] = y.Aids[0]
				}
			}
		}
	}
	eg.Go(func(ctx context.Context) error {
		if len(req.VideoSerialIdAidMap) == 0 {
			req.VideoSerialIdAidMap = make(map[int64]int64)
		}
		if len(req.FmSerialIdAidMap) == 0 {
			req.FmSerialIdAidMap = make(map[int64]int64)
		}
		if len(req.VideoChannelIdAidMap) == 0 {
			req.VideoChannelIdAidMap = make(map[int64]int64)
		}
		if len(req.FmChannelIdAidMap) == 0 {
			req.FmChannelIdAidMap = make(map[int64]int64)
		}
		hisFn(c)
		firstFn(c)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "commonItems eg.Wait() err=%+v, buvid=%s.", req.Mid, req.Buvid)
	}

	// 构建查询参数
	params := &commonmdl.Params{
		Mid:   req.Mid,
		Buvid: req.Buvid,
	}
	if len(req.Aids) > 0 {
		if device.Build < build203 {
			playAvs := make([]*archivegrpc.PlayAv, 0)
			for _, aid := range req.Aids {
				if aid <= 0 {
					continue
				}
				playAv := &archivegrpc.PlayAv{Aid: aid}
				if his := ugcHisMap[aid]; his != nil && his.Pro > -1 && his.Cid > 0 {
					playAv.PlayVideos = []*archivegrpc.PlayVideo{{Cid: his.Cid}}
				}
				playAvs = append(playAvs, playAv)
			}
			if len(playAvs) > 0 {
				params.ArchiveReq = &commonmdl.ArchiveReq{PlayAvs: playAvs}
			}
		} else {
			playAvs := make([]*archivegrpc.PlayAv, 0)
			for _, aid := range req.Aids {
				if aid <= 0 {
					continue
				}
				playAv := &archivegrpc.PlayAv{Aid: aid}
				if his := ugcHisMap[aid]; his != nil && his.Pro > -1 && his.Cid > 0 {
					playAv.PlayVideos = []*archivegrpc.PlayVideo{{Cid: his.Cid}}
				}
				playAvs = append(playAvs, playAv)
			}
			if len(playAvs) > 0 {
				params.ArchivePlusReq = &commonmdl.ArchivePlusReq{PlayAvs: playAvs}
			}
		}
	}
	if len(req.Sids) > 0 {
		sidInt32 := make([]int32, 0)
		epids := make([]int32, 0)
		for _, sid := range req.Sids {
			if sid <= 0 {
				continue
			}
			sidInt32 = append(sidInt32, int32(sid))
			if his := pgcHisMap[sid]; his != nil && his.Business == model.PgcBusinesses && his.Epid > 0 && his.Pro > hisProgressOver {
				epids = append(epids, int32(his.Epid))
			}
		}
		if len(sidInt32) > 0 {
			params.SeasonReq = &commonmdl.SeasonReq{Sids: sidInt32}
		}
		if len(epids) > 0 {
			params.EpisodeReq = &commonmdl.EpisodeReq{Epids: epids}
		}
	}
	if len(req.VideoSerialIds) > 0 && len(req.VideoSerialIdAidMap) > 0 {
		params.SerialInfosReq = &commonmdl.SerialInfosReq{VideoIds: req.VideoSerialIds}
		if params.ArchiveReq == nil {
			params.ArchiveReq = &commonmdl.ArchiveReq{PlayAvs: make([]*archivegrpc.PlayAv, 0)}
		}
		for _, a := range req.VideoSerialIdAidMap {
			params.ArchiveReq.PlayAvs = append(params.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: a})
		}
	}
	if len(req.FmSerialIds) > 0 && len(req.FmSerialIdAidMap) > 0 {
		if params.SerialInfosReq == nil {
			params.SerialInfosReq = new(commonmdl.SerialInfosReq)
		}
		params.SerialInfosReq.FmCommonIds = req.FmSerialIds
		if params.ArchiveReq == nil {
			params.ArchiveReq = &commonmdl.ArchiveReq{PlayAvs: make([]*archivegrpc.PlayAv, 0)}
		}
		for _, a := range req.FmSerialIdAidMap {
			params.ArchiveReq.PlayAvs = append(params.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: a})
		}
	}
	if len(req.VideoChannelIds) > 0 && len(req.VideoChannelIdAidMap) > 0 {
		params.ChannelInfosReq = &commonmdl.ChannelInfosReq{Video: req.VideoChannelIds}
		if params.ArchiveReq == nil {
			params.ArchiveReq = &commonmdl.ArchiveReq{PlayAvs: make([]*archivegrpc.PlayAv, 0)}
		}
		for _, a := range req.VideoChannelIdAidMap {
			params.ArchiveReq.PlayAvs = append(params.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: a})
		}
	}
	if len(req.FmChannelIds) > 0 && len(req.FmChannelIdAidMap) > 0 {
		if params.ChannelInfosReq == nil {
			params.ChannelInfosReq = new(commonmdl.ChannelInfosReq)
		}
		params.ChannelInfosReq.Fm = req.FmChannelIds
		if params.ArchiveReq == nil {
			params.ArchiveReq = &commonmdl.ArchiveReq{PlayAvs: make([]*archivegrpc.PlayAv, 0)}
		}
		for _, a := range req.FmChannelIdAidMap {
			params.ArchiveReq.PlayAvs = append(params.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: a})
		}
	}
	carCxt, err := s.material(c, params, device)
	if err != nil {
		log.Error("commonItems s.material1 err=%+v, mid=%d, buvid=%s", err, req.Mid, req.Buvid)
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, err
		}
		return nil, ecode.ServerErr
	}

	// 组装卡片
	cards := make([]*commonmdl.Item, 0)
	if len(req.Aids) > 0 {
		if device.Build < build203 {
			for _, aid := range req.Aids {
				ugc := carCxt.ArchiveResp[aid]
				if ugc == nil || ugc.Arc == nil {
					continue
				}
				var cid int64
				if his := ugcHisMap[aid]; his != nil && his.Pro > hisProgressOver && his.Cid > 0 {
					cid = his.Cid
				}
				cc := &commonmdl.CarContext{
					OriginData: &commonmdl.OriginData{
						MaterialType: commonmdl.MaterialTypeUGC,
						Oid:          aid,
						Cid:          cid,
					},
					ArchiveResp: carCxt.ArchiveResp,
				}
				if item := s.formItem(cc, device); item != nil {
					cards = append(cards, item)
				}
			}
		} else {
			for _, aid := range req.Aids {
				ugc := carCxt.ArchivePlusResp[aid]
				if ugc == nil || ugc.Player == nil || ugc.View == nil {
					continue
				}
				var cid int64
				if his := ugcHisMap[aid]; his != nil && his.Pro > hisProgressOver && his.Cid > 0 {
					cid = his.Cid
				}
				cc := &commonmdl.CarContext{
					OriginData: &commonmdl.OriginData{
						MaterialType: commonmdl.MaterialTypeUGCPlus,
						Oid:          aid,
						Cid:          cid,
					},
					ArchivePlusResp: carCxt.ArchivePlusResp,
				}
				if item := s.formItem(cc, device); item != nil {
					cards = append(cards, item)
				}
			}
		}
	}
	if len(req.Sids) > 0 {
		for _, sid := range req.Sids {
			season := carCxt.SeasonResp[int32(sid)]
			if season == nil {
				continue
			}
			cc := &commonmdl.CarContext{
				OriginData: &commonmdl.OriginData{
					MaterialType: commonmdl.MaterialTypeOGVSeaon,
					Oid:          sid,
				},
				SeasonResp:        carCxt.SeasonResp,
				EpisodeInlineResp: carCxt.EpisodeInlineResp,
			}
			item := s.formItem(cc, device)
			if item == nil {
				continue
			}

			// 将season默认秒开地址替换为其历史记录的秒开地址
			history := pgcHisMap[sid]
			if history != nil {
				epcc := &commonmdl.CarContext{
					OriginData: &commonmdl.OriginData{
						MaterialType: commonmdl.MaterialTypeOGVEP,
						Oid:          history.Epid,
					},
					EpisodeInlineResp: carCxt.EpisodeInlineResp,
					EpisodeResp:       carCxt.EpisodeResp,
				}
				epItem := s.formItem(epcc, device)
				if epItem != nil && epItem.Url != "" {
					item.Url = epItem.Url
				}
			}

			item.Cover = season.Cover
			item.Badge = pgcBadge(season.BadgeType, season.Badge)
			if req.SimpleItemAfterHandler != nil {
				req.SimpleItemAfterHandler(item, season)
			}
			cards = append(cards, item)
		}
	}
	if len(req.VideoSerialIds) > 0 && len(req.VideoSerialIdAidMap) > 0 {
		for _, sid := range req.VideoSerialIds {
			cc := &commonmdl.CarContext{
				OriginData: &commonmdl.OriginData{
					MaterialType: commonmdl.MaterialTypeVideoSerial,
					Oid:          sid,
					Cid:          req.VideoSerialIdAidMap[sid],
				},
				SerialInfosResp: carCxt.SerialInfosResp,
				ArchiveResp:     carCxt.ArchiveResp,
			}
			if item := s.formItem(cc, device); item != nil {
				cards = append(cards, item)
			}
		}
	}
	if len(req.FmSerialIds) > 0 && len(req.FmSerialIdAidMap) > 0 {
		for _, sid := range req.FmSerialIds {
			cc := &commonmdl.CarContext{
				OriginData: &commonmdl.OriginData{
					MaterialType: commonmdl.MaterialTypeFmSerial,
					Oid:          sid,
					Cid:          req.FmSerialIdAidMap[sid],
				},
				SerialInfosResp: carCxt.SerialInfosResp,
				ArchiveResp:     carCxt.ArchiveResp,
			}
			if item := s.formItem(cc, device); item != nil {
				cards = append(cards, item)
			}
		}
	}
	if len(req.VideoChannelIds) > 0 && len(req.VideoChannelIdAidMap) > 0 {
		for _, sid := range req.VideoChannelIds {
			cc := &commonmdl.CarContext{
				OriginData: &commonmdl.OriginData{
					MaterialType: commonmdl.MaterialTypeVideoChannel,
					Oid:          sid,
					Cid:          req.VideoChannelIdAidMap[sid],
				},
				ChannelInfosResp: carCxt.ChannelInfosResp,
				ArchiveResp:      carCxt.ArchiveResp,
			}
			if item := s.formItem(cc, device); item != nil {
				cards = append(cards, item)
			}
		}
	}
	if len(req.FmChannelIds) > 0 && len(req.FmChannelIdAidMap) > 0 {
		for _, v := range req.FmChannelIds {
			cc := &commonmdl.CarContext{
				OriginData: &commonmdl.OriginData{
					MaterialType: commonmdl.MaterialTypeFmChannel,
					Oid:          v,
					Cid:          req.FmChannelIdAidMap[v],
				},
				ChannelInfosResp: carCxt.ChannelInfosResp,
				ArchiveResp:      carCxt.ArchiveResp,
			}
			if item := s.formItem(cc, device); item != nil {
				cards = append(cards, item)
			}
		}
	}
	if len(cards) == 0 {
		log.Error("commonItems res is nil. mid=%d, buvid=%s, aids=%+v, sids=%+v", req.Mid, req.Buvid, req.Aids, req.Sids)
		return nil, xecode.AppMediaNotData
	}
	return cards, nil
}

// pgcBadge 获取pgc角标.
func pgcBadge(style int32, text string) *commonmdl.Badge {
	if text == "" {
		return nil
	}
	res := &commonmdl.Badge{
		Text: text,
	}
	switch model.PGCBageType[style] {
	case model.BgColorRed:
		res.TextColorDay = "#FFFFFF"
		res.TextColorNight = "#FFFFFF"
		res.BgColorDay = "#FF5377"
		res.BgColorNight = "#FF5377"
		res.BorderColorDay = "#FF5377"
		res.BorderColorNight = "#FF5377"
		res.BgStyle = model.BgStyleFill
	case model.BgColorBlue:
		res.TextColorDay = "#FFFFFF"
		res.TextColorNight = "#FFFFFF"
		res.BgColorDay = "#20AAE2"
		res.BgColorNight = "#20AAE2"
		res.BorderColorDay = "#20AAE2"
		res.BorderColorNight = "#20AAE2"
		res.BgStyle = model.BgStyleFill
	case model.BgColorYellow:
		res.TextColorDay = "#7E2D11"
		res.TextColorNight = "#7E2D11"
		res.BgColorDay = "#FFB112"
		res.BgColorNight = "#FFB112"
		res.BorderColorDay = "#FFB112"
		res.BorderColorNight = "#FFB112"
		res.BgStyle = model.BgStyleFill
	default:
		return nil
	}
	return res
}

// videoTabFeedCards 获取'推荐'tab下卡片.
func (s *Service) videoTabFeedCards(c context.Context, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	if req.Mid <= 0 {
		return nil, xecode.AppMediaNotData
	}
	group := feedGroup(req.Mid, req.Buvid)
	list, err := s.rcmdDao.FeedRecommend(c, 0, "", req.Buvid, req.Mid, req.DeviceInfo.Build, req.LoginEvent, group, req.PageNext.Ps, req.Mode)
	if err != nil {
		log.Error("videoTabFeedCards s.rcmd.FeedRecommend err=%+v, mid=%d, buvid=%s, group=%d", err, req.Mid, req.Buvid, group)
		return nil, ecode.ServerErr
	}
	if len(list) == 0 {
		log.Warn("videoTabFeedCards has no card. mid=%d, buvid=%s, group=%d", req.Mid, req.Buvid, group)
	}

	aids := make([]int64, 0)
	for i, v := range list {
		if v == nil || v.Goto != model.GotoAv || v.ID <= 0 {
			continue
		}
		if i >= req.PageNext.Ps {
			break
		}
		aids = append(aids, v.ID)
	}
	resp, err := s.buildTabCardResp(c, commonmdl.MaterialTypeUGC, req.DeviceInfo, req.PageNext.Ps, req.PageNext.Pn, req.Mid, req.Buvid, aids, nil, nil, req.TabId)

	mock := s.mockOgvEpHorizontalCard(c, req.Mid, req.DeviceInfo)
	if mock != nil && resp != nil && resp.Items != nil {
		resp.Items.Cards = append(resp.Items.Cards, mock)
		resp.Items.Banners = append(resp.Items.Banners, mock)
	}

	return resp, err
}

// mockOgvEpHorizontalCard 为推荐feed、fm feed、banner位置mock ogv ep横卡.
func (s *Service) mockOgvEpHorizontalCard(c context.Context, mid int64, device model.DeviceInfo) *commonmdl.Item {
	if !s.v23debug(mid, device.Build) {
		return nil
	}

	var sid int32 = 39733
	var epid int32 = 446571
	epParam := &commonmdl.Params{
		EpisodeReq: &commonmdl.EpisodeReq{
			Epids: []int32{epid},
		},
	}
	material, err := s.material(c, epParam, device)
	if err != nil {
		return nil
	}
	material.OriginData = &commonmdl.OriginData{
		MaterialType: commonmdl.MaterialTypeOGVEP,
		Oid:          int64(epid),
	}
	item := s.formItem(material, device)
	if item == nil {
		return nil
	}
	cards, err := s.bangumiDao.Cards(c, []int32{sid})
	if err != nil || cards[sid] == nil {
		return nil
	}
	season := cards[sid]
	item.Catalog = &commonmdl.Catalog{
		CatalogId:   int64(season.SeasonType),
		CatalogName: season.SeasonTypeName,
	}
	if season.Rating != nil && season.Rating.Score > 0 {
		item.Score = fmt.Sprintf("%v分", season.Rating.Score)
	}
	item.SubTitle = season.Subtitle
	return item
}

func feedGroup(mid int64, buvid string) int {
	if mid <= 0 && buvid == "" {
		return hisProgressOver
	}
	if mid != 0 {
		return int(mid % 20)
	}
	return int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
}

// videoTabUgcRegionCards 根据typeId获取ugc不同分区tab卡片.
func (s *Service) videoTabUgcRegionCards(c context.Context, typeId int64, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	dynamic, err := s.regionDao.RegionDynamic(c, typeId, req.PageNext.Pn, req.PageNext.Ps)
	if err != nil {
		log.Error("videoTabUgcRegionCards s.regionDao.RegionDynamic err=%+v, mid=%d, buvid=%s", err, req.Mid, req.Buvid)
		return nil, ecode.ServerErr
	}
	if len(dynamic) == 0 {
		log.Warn("videoTabUgcRegionCards ugc region has no card. typeId=%d, mid=%d, buvid=%s", typeId, req.Mid, req.Buvid)
	}

	aids := make([]int64, 0)
	for i, v := range dynamic {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if i >= req.PageNext.Ps {
			break
		}
		aids = append(aids, v.Aid)
	}
	return s.buildTabCardResp(c, commonmdl.MaterialTypeUGC, req.DeviceInfo, req.PageNext.Ps, req.PageNext.Pn, req.Mid, req.Buvid, aids, nil, nil, req.TabId)
}

// videoTabHotCards 获取热门tab下卡片.
func (s *Service) videoTabHotCards(c context.Context, req *commonmdl.VideoTabCardReq) (*commonmdl.VideoTabCardResp, error) {
	sIndex := (req.PageNext.Pn - 1) * req.PageNext.Ps
	key := popularGroup(req.Mid, req.Buvid)
	popularCards, err := s.showDao.PopularCardTenCache(c, key, sIndex, req.PageNext.Ps)
	if err != nil {
		log.Error("videoTabHotCards s.showDao.PopularCardTenCache err=%+v, mid=%d, buvid=%s", err, req.Mid, req.Buvid)
		return nil, ecode.ServerErr
	}
	if len(popularCards) == 0 {
		log.Warn("videoTabHotCards has no card. mid=%d, buvid=%s, sIndex=%d", req.Mid, req.Buvid, sIndex)
	}

	aids := make([]int64, 0)
	for _, v := range popularCards {
		if v == nil || v.Value <= 0 {
			continue
		}
		aids = append(aids, v.Value)
	}
	return s.buildTabCardResp(c, commonmdl.MaterialTypeUGC, req.DeviceInfo, req.PageNext.Ps, req.PageNext.Pn, req.Mid, req.Buvid, aids, nil, nil, req.TabId)
}

// popularGroup 根据mid、buvid计算出对应热门的index.
func popularGroup(mid int64, buvid string) int {
	if mid > 0 {
		return int((mid / oneThousand) % ten)
	}
	if buvid != "" {
		return int((crc32.ChecksumIEEE([]byte(buvid)) / oneThousand) % ten)
	}
	log.Error("popularGroup mid and buvid is nil.")
	return 0
}

// validatePageNext 校验pageNext入参，并返回合法的pageNext.
func (s *Service) validatePageNext(pageNext *commonmdl.PageNext) *commonmdl.PageNext {
	if pageNext == nil {
		return &commonmdl.PageNext{
			Ps: s.c.VideoTabsV2Conf.DefaultPs,
			Pn: 1,
		}
	}
	if pageNext.Pn <= 0 {
		pageNext.Pn = 1
	}
	pageNext.Ps = s.c.VideoTabsV2Conf.DefaultPs
	return pageNext
}

// CardPlaylist 查询banner合集.
func (s *Service) CardPlaylist(c context.Context, req *commonmdl.CardPlaylistReq) (*commonmdl.CardPlaylistResp, error) {
	var playlistInfo *conf.BannerPlaylist
	for _, v := range s.c.BannerPlaylist {
		if v.Id == req.Id {
			playlistInfo = v
			break
		}
	}
	if playlistInfo == nil || len(playlistInfo.PlayList) == 0 || playlistInfo.StyleType != bannerCollection {
		log.Errorc(c, "serialCustom res is nil. id=%d, buvid=%s.", req.Id, req.Buvid)
		return &commonmdl.CardPlaylistResp{}, nil
	}
	var aids, sids []int64
	if commonmdl.MaterialType(playlistInfo.MaterialType) == commonmdl.MaterialTypeUGC {
		aids = playlistInfo.PlayList
	} else {
		sids = playlistInfo.PlayList
	}
	creq := &commonItemsReq{
		Mid:                    req.Mid,
		Buvid:                  req.Buvid,
		Aids:                   aids,
		Sids:                   sids,
		SimpleItemAfterHandler: ogvSimpleItemAfterHandler,
	}
	cards, err := s.commonItems(c, req.DeviceInfo, creq)
	if err != nil {
		log.Errorc(c, "CardPlaylist s.commonItems err=%+v, id=%d, buvid=%s.", err, req.Id, req.Buvid)
		return nil, err
	}
	return &commonmdl.CardPlaylistResp{
		Cards: cards,
	}, nil
}

func (s *Service) TabExchange(req *commonmdl.VideoTabsReq) int {
	if s.c.TabExchange == nil || len(s.c.TabExchange.Channels) == 0 {
		return 0
	}
	if inArray(req.Channel, s.c.TabExchange.Channels) {
		return _hitExchange
	}
	return 0
}

func inArray(input string, arr []string) bool {
	for _, v := range arr {
		if input == v {
			return true
		}
	}
	return false
}
