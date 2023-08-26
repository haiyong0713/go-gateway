package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	pgrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

func (s *Service) GeneralFeedList(ctx context.Context, req *model.GeneralFeedListReq) (*model.GeneralFeedListRsp, error) {
	const (
		_videoSmallCard  = "video_small_card"
		_videoInlineCard = "video_inline_card"
	)
	general, res := constructGeneralParamFromCtx(ctx), &model.GeneralFeedListRsp{}
	switch req.FeedCardType {
	case _videoSmallCard:
		reply, fanout, err := s.videoStoryFanout(ctx, req, general)
		if err != nil {
			log.Error("GeneralFeedList videoStoryFanout req=%+v, err=%+v", req, err)
			return nil, err
		}
		cards, err := buildTopicAvBasicCards(ctx, reply.Rid, fanout)
		if err != nil {
			log.Error("GeneralFeedList buildTopicAvBasicCards req=%+v, error=%+v", req, err)
			return nil, err
		}
		res.VideoCards = cards
		res.Offset = reply.Offset
		res.HasMore = reply.HasMore
		return res, nil
	case _videoInlineCard:
		reply, fanout, err := s.videoStoryFanout(ctx, req, general)
		if err != nil {
			log.Error("videoStoryFanout req=%+v, err=%+v", req, err)
			return nil, err
		}
		cards, err := buildTopicAvInlineCards(ctx, general, reply.Rid, fanout)
		if err != nil {
			log.Error("GeneralFeedList buildTopicAvInlineCards req=%+v, error=%+v", req, err)
			return nil, err
		}
		res.VideoInlineCards = cards
		res.Offset = reply.Offset
		res.HasMore = reply.HasMore
		return res, nil
	default:
		typesInt32, err := splitInt32s(req.ShowDynamicTypes)
		if err != nil {
			log.Error("splitInt32s parse error req=%+v, error=%+v", req, err)
		}
		args := &topicsvc.GeneralFeedListReq{
			TopicId:  req.TopicId,
			SortBy:   req.SortBy,
			Offset:   req.Offset,
			PageSize: int32(req.PageSize),
			BizTypes: typesInt32,
			Scene:    req.Business,
			Uid:      general.Mid,
		}
		reply, err := s.topicGRPC.GeneralFeedList(ctx, args)
		if err != nil {
			log.Error("GeneralFeedList s.topicGRPC.GeneralFeedList mid=%d, args=%+v, error=%+v", general.Mid, args, err)
			return nil, err
		}
		dynSchemaCtx := initDynSchemaContext(ctx, req.TopicId, req.SortBy, req.Offset)
		cardItems, err := s.dynWebCardProcess(dynSchemaCtx, general, convertGeneralListToDynMetaCardListItem(reply.Items))
		if err != nil {
			log.Error("GeneralFeedList s.dynWebCardProcess args=%+v, error=%+v", args, err)
			return nil, err
		}
		res.TopicCards = makeDynamicCardList(cardItems)
		res.Offset = reply.Offset
		res.HasMore = reply.HasMore
		return res, nil
	}
}

func convertGeneralListToDynMetaCardListItem(items []*topicsvc.GeneralFeedItem) []*topiccardmodel.DynMetaCardListParam {
	var res []*topiccardmodel.DynMetaCardListParam
	for _, v := range items {
		if v.Type != _dynamicCardType {
			continue
		}
		res = append(res, &topiccardmodel.DynMetaCardListParam{
			DynId:          v.Rid,
			ItemFrom:       v.ItemFrom.String(),
			HiddenAttached: v.HiddenAttached,
		})
	}
	return res
}

func (s *Service) videoStoryFanout(ctx context.Context, req *model.GeneralFeedListReq, general *topiccardmodel.GeneralParam) (*topicsvc.VideoStoryRsp, *FanoutResult, error) {
	args := &topicsvc.VideoStoryReq{
		TopicId:    req.TopicId,
		FromSortBy: req.SortBy,
		Offset:     req.Offset,
		PageSize:   req.PageSize,
		Uid:        general.Mid,
	}
	reply, err := s.topicGRPC.VideoStory(ctx, args)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "s.topicGRPC.VideoStory args=%+v", req)
	}
	loader := &TopicFanoutLoader{General: general, Service: s, Archive: loaderArchiveSubset{Aids: reply.Rid}}
	fanout, err := loader.doTopicCardFanoutLoad(ctx)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "loader.doTopicCardFanoutLoad args=%+v", req)
	}
	return reply, fanout, nil
}

func (s *Service) TopicTimeLine(ctx context.Context, req *model.TopicTimeLineReq) (*model.TopicTimeLineRsp, error) {
	upStreamRsp, err := s.topicGRPC.TopConfig(ctx, &topicsvc.TopicInlineResReq{TopicId: req.TopicId})
	if err != nil {
		log.Error("s.topicGRPC.BannerRes req=%+v, error=%+v", req, err)
		return nil, err
	}
	if upStreamRsp == nil || upStreamRsp.ResType != _timeLineResourceType {
		return nil, errors.Wrap(ecode.NothingFound, "未找到资源")
	}
	reply, err := s.managerPopClient.TimeLine(ctx, &pgrpc.TimeLineRequest{LineId: upStreamRsp.ResId, Ps: req.PageSize, Offset: req.Offset})
	if err != nil {
		log.Error("s.managerPopClient.TimeLine upStreamRsp=%+v, err=%+v", upStreamRsp, err)
		return nil, err
	}
	events := constructTimeLineEventsJson(reply.Events, upStreamRsp.TimingAccuracy)
	if len(events) == 0 {
		return nil, errors.Wrap(ecode.NothingFound, "未找到时间轴事件")
	}
	return &model.TopicTimeLineRsp{
		TimeLineId:     upStreamRsp.ResId,
		TimeLineTitle:  upStreamRsp.Title,
		HasMore:        reply.HasMore,
		Offset:         reply.Offset,
		TimeLineEvents: events,
	}, nil
}
