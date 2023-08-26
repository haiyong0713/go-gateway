package common

import (
	"context"
	"encoding/json"
	"regexp"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	sch "go-gateway/app/app-svr/app-car/interface/model/search"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_defaultPn    = 1
	_defaultUpPs  = 3
	_defaultArcPs = 20

	_seasonSuffixSeries  = "追剧"
	_seasonSuffixBangumi = "追番"
	// 时尚分区 - 美妆护肤二级分区
	_regionMakeups = 157
)

func (s *Service) SearchV2(ctx context.Context, param *sch.SearchParamV2) (resp *sch.SearchRespV2, err error) {
	var (
		pageReq   *sch.PageInfo
		arcIdsRes *sch.ArcIdsRes
		matReq    *common.Params
		matResp   *common.CarContext
		arcItems  []*common.Item
		ups       []int64
		upItems   []*sch.UpItemV2
		// 是否命中小鹏近期热门
		hitXiaoPengRegion bool
		channelArcIdsRes  *sch.ChannelArcIdsRes
		// 是否命中小鹏的tab检索
		hitXiaoPengMain bool
		mainArcIdsRes   *sch.MainArcIdsRes
	)
	pageReq, err = extractPageInfo(param.Ps, param.PageNext)
	if err != nil {
		log.Errorc(ctx, "SearchV2 extractSearchPage error:%+v, param:%+v", err, param)
		return nil, err
	}
	eg := errgroup.WithContext(ctx)
	//todo 这没有限制小鹏用户
	if param.Keyword == s.c.CustomModule.XiaoPengKeywordRegion {
		//检索词是：小鹏美妆空间近期热门，用特定分类干预，
		var (
			offset = ""
			arg    = &channelgrpc.ResourceListReq{
				ChannelId: s.c.CustomModule.ChannelMakeups,
				TabType:   channelgrpc.TabType_TAB_TYPE_TOTAL,
				SortType:  channelgrpc.TotalSortType_SORT_BY_HOT,
				Offset:    offset,
				PageSize:  int32(pageReq.Ps)}
		)

		if pageReq.Pn > 1 {
			// search第一页，重置offset
			var regionOffset = s.getRegionOffset(ctx, param.Buvid)
			arg.Offset = regionOffset
		}

		eg.Go(func(ctx context.Context) error {
			var localErr error
			channelArcIdsRes, localErr = s.getChannelSearchRes(ctx, arg, param.Buvid)
			if localErr != nil {
				return localErr
			}
			hitXiaoPengRegion = true
			return nil
		})
	}
	reg := regexp.MustCompile(s.c.CustomModule.XiaoPengKeywordTab)
	matchArr := reg.FindStringSubmatch(param.Keyword)
	if len(matchArr) > 0 {
		//检索词是:小鹏美妆空间XXX教程，用特定检索词，检索分类下
		pageReq.Pn = pageReq.Pn + 1
		realKeyword := matchArr[len(matchArr)-1]
		eg.Go(func(ctx context.Context) error {
			var localErr error
			mainArcIdsRes, localErr = s.getMainSearchRes(ctx, _regionMakeups, *pageReq, realKeyword)
			if localErr != nil {
				return localErr
			}
			hitXiaoPengMain = true
			return nil
		})
	}
	// 1.1 获取ugc+pgc搜索id列表
	eg.Go(func(ctx context.Context) error {
		var localErr error
		arcIdsRes, localErr = s.getArcSearchRes(ctx, param, *pageReq)
		if localErr != nil {
			return localErr
		}
		return nil
	})
	// 1.2 获取up主搜索结果（缺少头像url）
	if pageReq.Pn == 1 {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			upItems, ups, localErr = s.getUpSearchRes(ctx, param)
			if localErr != nil {
				return localErr
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "SearchV2 get ids error:%+v, param:%+v", err, param)
		return nil, err
	}
	if hitXiaoPengRegion {
		//覆盖了getArcSearchRes的返回
		arcIdsRes.Aids = channelArcIdsRes.Aids
	}
	if hitXiaoPengMain {
		arcIdsRes.Sids = mainArcIdsRes.Sids
		arcIdsRes.Aids = mainArcIdsRes.Aids
		arcIdsRes.PageNext = mainArcIdsRes.PageNext
		arcIdsRes.HasNext = mainArcIdsRes.HasNext
	}
	// 2 依据id列表获取原始物料
	matReq = searchMaterialReq(arcIdsRes, ups, param.Mid, param.Buvid)
	matResp, err = s.material(ctx, matReq, param.DeviceInfo)
	if err != nil {
		log.Errorc(ctx, "SearchV2 s.material error:%+v, param:%+v", err, param)
		return nil, err
	}
	// 3 依据原始物料获取item
	arcItems = s.getSearchArcItems(matResp, arcIdsRes, param.DeviceInfo)
	upItems = s.fillSearchUpItems(matResp, upItems)
	return &sch.SearchRespV2{
		ArcItems: arcItems,
		UpItems:  upItems,
		PageNext: arcIdsRes.PageNext,
		HasNext:  arcIdsRes.HasNext,
	}, nil
}

func (s *Service) getRegionOffset(ctx context.Context, buvid string) string {
	res, err := s.srchDao.GetRegionOffsetCacheById(ctx, buvid)
	if err != nil {
		return ""
	}
	return res
}

func (s *Service) saveRegionOffset(ctx context.Context, buvid string, offset string) error {
	_, err := s.srchDao.SaveRegionOffsetCache(ctx, buvid, offset)
	return err
}

func (s *Service) getArcSearchRes(ctx context.Context, param *sch.SearchParamV2, page sch.PageInfo) (*sch.ArcIdsRes, error) {
	res, err := s.srchDao.Search(ctx, param.Mid, 0, page.Pn, page.Ps, param.Keyword, param.Buvid)
	if err != nil {
		return nil, errors.Wrap(err, "getArcSearchRes error")
	}
	if res == nil || res.Result == nil {
		return nil, ecode.NothingFound
	}
	var (
		aids    = make([]int64, 0)
		sids    = make([]int32, 0)
		hasNext bool
	)
	// ogv
	for _, v := range res.Result.MediaBangumi {
		sids = append(sids, v.SeasonID)
	}
	for _, v := range res.Result.MediaFt {
		sids = append(sids, v.SeasonID)
	}
	// ugc
	for _, v := range res.Result.Video {
		aids = append(aids, v.ID)
	}
	hasNext = page.Pn < res.NumPages // 未到达最后一页，则存在下一页
	page.Pn = page.Pn + 1
	return &sch.ArcIdsRes{
		Aids:     aids,
		Sids:     sids,
		PageNext: &page,
		HasNext:  hasNext,
	}, nil

}

// getChannelSearchRes 获取特定频道下的内容
func (s *Service) getChannelSearchRes(ctx context.Context, req *changrpc.ResourceListReq, buvid string) (*sch.ChannelArcIdsRes, error) {
	var (
		totalId []int64
		aids    []int64
		arcs    map[int64]*archivegrpc.Arc
	)
	reply, err := s.channelDao.ResourceList(ctx, req)
	if err != nil || reply == nil || reply.Cards == nil {
		return nil, errors.Wrap(err, "getRegionSearchRes error")
	}
	for _, v := range reply.Cards {
		//频道下面，只有ugc内容
		if v.CardType != channelgrpc.ChannelCardType_CARD_TYPE_VIDEO_ARCHIVE {
			//只选取普通视频卡片
			continue
		}
		if v.VideoCard.Rid == 0 {
			continue
		}
		totalId = append(totalId, v.VideoCard.Rid)
	}
	if len(totalId) != 0 {
		if arcs, err = s.archiveDao.Archives(ctx, totalId); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
	}
	for _, v := range reply.Cards {
		if v.CardType != channelgrpc.ChannelCardType_CARD_TYPE_VIDEO_ARCHIVE {
			continue
		}
		aid := v.VideoCard.Rid
		if _, ok := arcs[aid]; ok {
			aids = append(aids, aid)
		}
	}
	// 缓存offset
	if reply.NextOffset != "" {
		err := s.saveRegionOffset(ctx, buvid, reply.NextOffset)
		if err != nil {
			log.Errorc(ctx, "save region offset with buvid error, err: %+v | param: %+v |channelArcIdsRes: %+v", err, req, reply)
		}
	}
	return &sch.ChannelArcIdsRes{
		Aids:       aids,
		Arcs:       arcs,
		NextOffset: reply.NextOffset,
	}, nil
}

// getMainSearchRes 获取大搜特定分区下的检索内容
func (s *Service) getMainSearchRes(ctx context.Context, rid int64, page sch.PageInfo, keyword string) (*sch.MainArcIdsRes, error) {
	var (
		cardItem []*ai.Item
		aids     []int64
		ssids    []int32
		hasNext  bool
	)
	all, err := s.srchDao.Search(ctx, 0, rid, page.Pn, page.Ps, keyword, "")
	if err != nil {
		return nil, errors.Wrap(err, "getMainSearchRes error")
	}
	if all == nil || all.Result == nil {
		return nil, errors.Wrap(err, "getMainSearchRes error")
	}
	// 转换成统一结构体
	// pgc
	for _, v := range all.Result.MediaBangumi {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	for _, v := range all.Result.MediaFt {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	// archive
	for _, v := range all.Result.Video {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
	}
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			ssids = append(ssids, int32(v.ID))
		}
	}
	hasNext = page.Pn < all.NumPages // 未到达最后一页，则存在下一页
	page.Pn = page.Pn + 1
	return &sch.MainArcIdsRes{
		Aids:     aids,
		Sids:     ssids,
		HasNext:  hasNext,
		PageNext: &page,
	}, nil
}

func (s *Service) getUpSearchRes(ctx context.Context, param *sch.SearchParamV2) (upItems []*sch.UpItemV2, ups []int64, err error) {
	var (
		upReply []*sch.User
	)
	upReply, err = s.srchDao.Upper(ctx, param.Mid, _defaultPn, _defaultUpPs, param.Keyword, param.Buvid)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range upReply {
		upItems = append(upItems, &sch.UpItemV2{
			Mid:        v.Mid,
			Name:       v.Name,
			FansCount:  v.Fans,
			VideoCount: v.Videos,
			Desc:       v.Usign,
		})
		ups = append(ups, v.Mid)
	}
	return upItems, ups, nil
}

func (s *Service) getSearchArcItems(carCtx *common.CarContext, arcs *sch.ArcIdsRes, dev model.DeviceInfo) []*common.Item {
	if len(arcs.Aids) == 0 && len(arcs.Sids) == 0 {
		return make([]*common.Item, 0)
	}
	ogvItems := make([]*common.Item, 0)
	ugcItems := make([]*common.Item, 0)
	for _, sid := range arcs.Sids {
		carCtx.OriginData = &common.OriginData{
			MaterialType: common.MaterialTypeOGVSeaon,
			Oid:          int64(sid),
		}
		item := s.formItem(carCtx, dev)
		if item == nil {
			continue
		}
		fillOgvItem(sid, carCtx, item)
		ogvItems = append(ogvItems, item)
	}
	for _, aid := range arcs.Aids {
		carCtx.OriginData = &common.OriginData{
			MaterialType: common.MaterialTypeUGC,
			Oid:          aid,
		}
		item := s.formItem(carCtx, dev)
		if item == nil {
			continue
		}
		ugcItems = append(ugcItems, item)
	}
	// ogv在前，ugc在后
	return append(ogvItems, ugcItems...)
}

// fillSearchUpItems 填充up主搜索结果中，缺失的信息
func (s *Service) fillSearchUpItems(matResp *common.CarContext, items []*sch.UpItemV2) []*sch.UpItemV2 {
	if len(items) == 0 {
		return make([]*sch.UpItemV2, 0)
	}
	if len(matResp.AccountCardResp) == 0 {
		log.Warn("fillSearchUpItems account card empty, upItems:%+v", items)
		return items
	}
	for _, v := range items {
		var (
			card *accountgrpc.Card
			ok   bool
		)
		if card, ok = matResp.AccountCardResp[v.Mid]; !ok {
			log.Warn("fillSearchUpItems account card miss, mid:%+v", v.Mid)
			continue
		}
		v.Face = card.Face
	}
	return items
}

func extractPageInfo(globalPs int, pageJson string) (*sch.PageInfo, error) {
	res := new(sch.PageInfo)
	if pageJson != "" && pageJson != "null" {
		if err := json.Unmarshal([]byte(pageJson), res); err != nil {
			return nil, errors.Wrap(ecode.RequestErr, err.Error())
		}
	}
	if res.Pn == 0 {
		res.Pn = 1
	}
	if globalPs > 0 {
		res.Ps = globalPs
	} else if res.Ps == 0 {
		res.Ps = _defaultArcPs
	}
	return res, nil
}

func searchMaterialReq(arcs *sch.ArcIdsRes, ups []int64, mid int64, buvid string) *common.Params {
	var (
		arcReq   *common.ArchiveReq
		ssReq    *common.SeasonReq
		accounts *common.AccountCardReq
	)
	if arcs != nil && len(arcs.Aids) > 0 {
		playAvs := make([]*archivegrpc.PlayAv, 0)
		for _, aid := range arcs.Aids {
			playAvs = append(playAvs, &archivegrpc.PlayAv{Aid: aid})
		}
		arcReq = &common.ArchiveReq{PlayAvs: playAvs}
	}
	if arcs != nil && len(arcs.Sids) > 0 {
		ssReq = &common.SeasonReq{Sids: arcs.Sids}
	}
	if len(ups) > 0 {
		accounts = &common.AccountCardReq{Mids: ups}
	}
	return &common.Params{
		ArchiveReq:     arcReq,
		SeasonReq:      ssReq,
		AccountCardReq: accounts,
		Mid:            mid,
		Buvid:          buvid,
	}
}

// fillOgvItem
// 1. ogv填充XX追番/追剧，到sub_title字段
// 2. ogv封面图替换为竖图
func fillOgvItem(sid int32, carCtx *common.CarContext, item *common.Item) {
	if len(carCtx.SeasonResp) == 0 {
		return
	}
	card := carCtx.SeasonResp[sid]
	if card == nil || card.Stat == nil {
		return
	}
	item.Cover = card.Cover
	switch card.SeasonType {
	case 1, 4: // nolint:gomnd // 动漫
		item.SubTitle = model.StatString64(card.Stat.Follow, _seasonSuffixBangumi)
	default: // 影视剧
		item.SubTitle = model.StatString64(card.Stat.Follow, _seasonSuffixSeries)
	}
}
