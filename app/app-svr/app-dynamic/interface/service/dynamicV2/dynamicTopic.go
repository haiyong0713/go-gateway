package dynamicV2

import (
	"context"
	"encoding/base64"

	"go-common/library/ecode"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	"github.com/pkg/errors"
)

const (
	_sortDefault    = 5 // 话题配置决定
	_sortHeat       = 0 // 热度
	_sortTime       = 1 // 时间
	_sortMix        = 2 // 综合
	_sortHeatRandom = 3 // 热度随机
	_sortRcmd       = 4 // 推荐

	_cardAll   = 0 // 全部卡片
	_cardAlbum = 1 // 图文
	_cardVideo = 2 // 视频
)

var (
	sortType2Name = map[int32]string{
		_sortDefault:    "默认",
		_sortHeat:       "最热",
		_sortTime:       "最新",
		_sortMix:        "综合",
		_sortHeatRandom: "热度",
		_sortRcmd:       "推荐",
	}
	defaultCardFilters = []*api.SortType{
		{SortType: _cardAll, SortTypeName: "全部"},
		{SortType: _cardVideo, SortTypeName: "视频"},
		{SortType: _cardAlbum, SortTypeName: "图文"},
	}
	cardType2DynTypes = map[int32][]int64{
		_cardAll:   {mdlv2.DynTypeVideo, mdlv2.DynTypeDraw, mdlv2.DynTypeForward, mdlv2.DynTypeWord, mdlv2.DynTypeArticle},
		_cardAlbum: {mdlv2.DynTypeDraw},
		_cardVideo: {mdlv2.DynTypeVideo},
	}
)

func (s *Service) LegacyTopicFeed(ctx context.Context, general *mdlv2.GeneralParam, req *api.LegacyTopicFeedReq) (*api.LegacyTopicFeedReply, error) {
	const (
		_defaultPageSize = 10
	)
	feedReq := &dyntopicgrpc.ListDynsV2Req{
		TopicName: req.TopicName,
		TopicId:   req.TopicId,
		Uid:       general.Mid, PageSize: _defaultPageSize,
	}
	// 处理offset
	if len(req.Offset) > 0 {
		offsetData, err := base64.StdEncoding.DecodeString(req.Offset)
		if err != nil {
			return nil, errors.WithMessagef(ecode.RequestErr, "invalid base64 offset encoding: %v", err)
		}
		offset := new(dyntopicgrpc.FeedOffset)
		if err = offset.Unmarshal(offsetData); err != nil {
			return nil, errors.WithMessagef(ecode.RequestErr, "invalid protobuf offset: %v", err)
		}
		feedReq.Offset = offset
	}
	// 处理sort by
	if req.SortType != nil {
		feedReq.SortBy = req.SortType.GetSortType()
	} else {
		feedReq.SortBy = _sortDefault
	}
	// 处理卡片过滤
	feedReq.Types = cardType2DynTypes[req.CardFilter.GetSortType()]

	feedReply, err := s.topDao.LegacyTopicFeed(ctx, feedReq)
	if err != nil {
		return nil, err
	}
	resp := &api.LegacyTopicFeedReply{
		HasMore: feedReply.HasMore,
	}
	// 处理翻页参数
	if feedReply.HasMore && feedReply.Offset != nil {
		offsetData, _ := feedReply.Offset.Marshal()
		resp.Offset = base64.StdEncoding.EncodeToString(offsetData)
	}
	// 首页的情况下给出排序信息
	if len(req.Offset) <= 0 {
		resp.FeedCardFilters = defaultCardFilters
		resp.SupportedSortTypes = make([]*api.SortType, 0, len(feedReply.AllSortBy))
		defaultAdded := false
		for _, sort := range feedReply.AllSortBy {
			s := &api.SortType{SortType: int32(sort), SortTypeName: sortType2Name[int32(sort)]}
			if sort == _sortDefault {
				defaultAdded = true
				copy(resp.SupportedSortTypes[1:], resp.SupportedSortTypes[0:])
				resp.SupportedSortTypes[0] = s
			} else {
				resp.SupportedSortTypes = append(resp.SupportedSortTypes, s)
			}
		}
		if !defaultAdded {
			resp.SupportedSortTypes = append([]*api.SortType{{SortType: _sortDefault, SortTypeName: sortType2Name[_sortDefault]}}, resp.SupportedSortTypes...)
		}
	}
	if len(feedReply.List) <= 0 {
		return resp, nil
	}

	// 处理feed流
	dynList := make([]*mdlv2.Dynamic, 0, len(feedReply.List))
	for _, d := range feedReply.List {
		dyn := new(mdlv2.Dynamic)
		dyn.FromDynamic(d)
		dynList = append(dynList, dyn)
	}

	dynCtx, err := s.getMaterial(ctx, getMaterialOption{general: general, dynamics: dynList})
	if err != nil {
		return nil, err
	}
	foldList := s.procListReply(ctx, dynList, dynCtx, general, _handleTypeLegacyTopic)
	s.procBackfill(ctx, dynCtx, general, foldList)
	retDynList := s.procFold(foldList, dynCtx, general)
	resp.List = retDynList

	return resp, nil
}
