package service

import (
	"context"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/app/web-svr/web/interface/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	dynccommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

func (s *Service) ActivityMovieList(ctx context.Context, mid int64, req *model.ActivityMovieListReq) (*model.ActivityMovieListRsp, error) {
	// 话题信息获取基础物料
	interRaw, err := s.resolveInterRawFromReqAndFetch(ctx, mid, req)
	if err != nil {
		log.Error("s.fetchDynTopicMeta err=%+v", err)
		return nil, err
	}
	var (
		accCardMap   = make(map[int64]*accgrpc.Card, len(interRaw.Mids))
		dynCommonMap = make(map[int64]*dynmdlV2.DynamicCommonCard, len(interRaw.DynamicIdMap))
	)
	// 调各业务方拿数据
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		dynCommonMap, err = s.dao.DynamicCommonInfos(ctx, interRaw.Rids)
		if err != nil {
			log.Error("s.dao.CommonInfos rids=%+v, err=%+v", interRaw.Rids, err)
			return ecode.NothingFound
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		accCardMap, err = s.dao.Cards3(ctx, interRaw.Mids)
		if err != nil {
			log.Error("s.dao.Cards3 mids=%+v, err=%+v", interRaw.Mids, err)
			return ecode.NothingFound
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return nil, errors.WithMessagef(err, "ActivityMovieList eg.Wait() req=%+v", req)
	}
	// 整合数据
	var res []*model.MovieReviewMeta
	for _, v := range interRaw.Rids {
		if meta, ok := makeMovieReviewList(dynCommonMap[v], accCardMap, interRaw.DynamicIdMap); ok {
			res = append(res, meta)
		}
	}
	return &model.ActivityMovieListRsp{
		List:           res,
		HasMore:        interRaw.HasMore,
		Offset:         interRaw.Offset,
		NewTopicOffset: interRaw.NewTopicOffset,
	}, nil
}

func (s *Service) resolveInterRawFromReqAndFetch(ctx context.Context, mid int64, req *model.ActivityMovieListReq) (*model.MovieReviewIntermediate, error) {
	if req.NewTopicId > 0 {
		// 话题id存在时优先拿新话题数据
		return s.fetchNewTopicMeta(ctx, mid, req)
	}
	if req.TopicName != "" {
		// 话题名称存在拿老话题数据
		return s.fetchDynTopicMeta(ctx, mid, req)
	}
	return nil, ecode.RequestErr
}

func (s *Service) fetchNewTopicMeta(ctx context.Context, mid int64, req *model.ActivityMovieListReq) (*model.MovieReviewIntermediate, error) {
	// 新话题取动态列表
	args := &topicsvc.GeneralFeedListReq{
		TopicId:  req.NewTopicId,
		SortBy:   2, // 热度序
		Offset:   req.NewTopicOffset,
		PageSize: req.PageSize,
		Scene:    "/x/web-interface/activity/movie/review/list",
		BizTypes: []int32{dynmdlV2.DynTypeCommonSquare, dynmdlV2.DynTypeCommonVertical},
		Uid:      mid,
	}
	reply, err := s.dao.TopicGeneralFeedList(ctx, args)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get s.dao.TopicSortList args=%+v", args)
	}
	// 填入数据
	res := &model.MovieReviewIntermediate{
		HasMore:        reply.HasMore,
		NewTopicOffset: reply.Offset,
		DynamicIdMap:   make(map[int64]int64, len(reply.Items)),
	}
	var dynIds []int64
	for _, v := range reply.Items {
		if v.Type == 0 {
			dynIds = append(dynIds, v.Rid)
		}
	}
	// simpleInfo接口获取动态信息
	simpleInfo, err := s.dao.DynSimpleInfo(ctx, &dyngrpc.DynSimpleInfosReq{
		DynIds: dynIds,
		Uid:    mid,
	})
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get s.dao.DynSimpleInfo dynIds=%+v", dynIds)
	}
	if simpleInfo == nil {
		return nil, errors.Errorf("simpleInfo is nil dynIds=%+v", dynIds)
	}
	for _, v := range simpleInfo.DynSimpleInfos {
		res.Rids = append(res.Rids, v.Rid)
		res.Mids = append(res.Mids, v.Uid)
		res.DynamicIdMap[v.Rid] = v.DynId
	}
	return res, nil
}

func (s *Service) fetchDynTopicMeta(ctx context.Context, mid int64, req *model.ActivityMovieListReq) (*model.MovieReviewIntermediate, error) {
	// 老话题取动态列表
	args := &dyntopicgrpc.ListDynsReq{
		TopicName:   req.TopicName,
		Uid:         mid,
		SortBy:      2,
		PageSize:    req.PageSize,
		Offset:      req.Offset,
		WithTop:     true,
		Types:       []int64{dynmdlV2.DynTypeCommonSquare, dynmdlV2.DynTypeCommonVertical},
		VersionCtrl: &dynccommon.MetaDataCtrl{Platform: "Web", From: "/x/web-interface/activity/movie/review/list"},
	}
	reply, err := s.dao.ListDyns(ctx, args)
	if err != nil {
		return nil, errors.WithMessagef(err, "s.dao.ListDyns args=%+v", args)
	}
	// 填入数据
	dynList := append(reply.TopList, reply.HotList...)
	dynList = append(dynList, reply.FeedList...)
	res := &model.MovieReviewIntermediate{
		HasMore:      reply.HasMore,
		Offset:       reply.Offset,
		DynamicIdMap: make(map[int64]int64, len(dynList)),
	}
	for _, v := range dynList {
		res.Rids = append(res.Rids, v.Rid)
		res.Mids = append(res.Mids, v.Uid)
		res.DynamicIdMap[v.Rid] = v.DynId
	}
	return res, nil
}

func makeMovieReviewList(commonCard *dynmdlV2.DynamicCommonCard, cardMap map[int64]*accgrpc.Card, dynamicIdMap map[int64]int64) (*model.MovieReviewMeta, bool) {
	if commonCard == nil || commonCard.User == nil {
		return nil, false
	}
	card, ok := cardMap[commonCard.User.UID]
	if !ok {
		return nil, false
	}
	score := parseScoreFromDynamicContent(commonCard.Vest.Content)
	if score <= 0 || score > 10 {
		// 不展示评分异常动态
		return nil, false
	}
	return &model.MovieReviewMeta{
		Author: &model.MovieReviewAuthor{
			Avatar: card.Face,
			Mid:    card.Mid,
			Uname:  card.Name,
			Level:  card.Level,
			Vip: &model.MovieReviewVip{
				ThemeType: card.Vip.ThemeType,
				VipStatus: card.Vip.Status,
				Type:      card.Vip.Type,
			},
			VipLabel: &model.MovieReviewVipLabel{
				LabelTheme: card.Vip.Label.LabelTheme,
				Path:       card.Vip.Label.Path,
				Text:       card.Vip.Label.Text,
			},
		},
		Content:        commonCard.Vest.Content,
		DynamicId:      dynamicIdMap[commonCard.RID],
		Score:          score,
		PtimeLabelText: parsePTimeLabelFromDynamicId(dynamicIdMap[commonCard.RID]),
	}, true
}

// 从动态id解析出动态发布时间
// nolint:gomnd
func parsePTimeLabelFromDynamicId(dynId int64) string {
	const (
		_k20170701 = 1498838400
	)
	return cardmdl.PubDataString(time.Unix((dynId>>32)+_k20170701, 0))
}

// 从动态文本解析出评分信息
// nolint:gomnd
func parseScoreFromDynamicContent(content string) int {
	const (
		_parseStarCharacter = "[星]"
	)
	return strings.Count(content, _parseStarCharacter) * 2
}
