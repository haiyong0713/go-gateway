package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) HotWordVideos(ctx context.Context, req *model.HotWordVideosReq) (*model.HotWordVideosRsp, error) {
	general := constructGeneralParamFromCtx(ctx)
	args := &topicsvc.VideoStoryReq{
		TopicId:    req.TopicId,
		FromSortBy: _topicSortByHot, // 热度序
		Offset:     req.Offset,
		PageSize:   req.PageSize,
		Uid:        general.Mid,
	}
	reply, err := s.topicGRPC.VideoStory(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.VideoStory mid=%d, args=%+v, error=%+v", general.Mid, args, err)
		return nil, err
	}
	loader := &TopicFanoutLoader{General: general, Service: s, Archive: loaderArchiveSubset{Aids: reply.Rid}}
	fanout, err := loader.doTopicCardFanoutLoad(ctx)
	if err != nil {
		log.Error("HotWordVideos doTopicCardFanoutLoad Aids=%+v, error=%+v", reply.Rid, err)
		return nil, err
	}
	cards, err := buildTopicAvBasicCards(ctx, reply.Rid, fanout)
	if err != nil {
		log.Error("HotWordVideos buildTopicAvBasicCards Aids=%+v, error=%+v", reply.Rid, err)
		return nil, err
	}
	return &model.HotWordVideosRsp{VideoCards: cards, HasMore: reply.HasMore, Offset: reply.Offset}, nil
}

func (s *Service) HotWordDynamics(ctx context.Context, req *model.HotWordDynamicReq) (*model.HotWordDynamicRsp, error) {
	general := constructGeneralParamFromCtx(ctx)
	args := &topicsvc.TopicDynListReq{
		TopicId:  req.TopicId,
		SortBy:   _topicSortByNew, // 最新序
		Offset:   req.Offset,
		PageSize: req.PageSize,
	}
	reply, err := s.topicGRPC.TopicDynList(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.TopicDynList mid=%d, args=%+v, error=%+v", general.Mid, args, err)
		return nil, err
	}
	dynSchemaCtx := initDynSchemaContext(ctx, req.TopicId, args.SortBy, req.Offset)
	cardItems, err := s.dynWebCardProcess(dynSchemaCtx, general, convertSortListToDynMetaCardListItem(reply.Items))
	if err != nil {
		log.Error("HotWordDynamics s.dynWebCardProcess args=%+v, error=%+v", args, err)
		return nil, err
	}
	return &model.HotWordDynamicRsp{Items: makeDynamicCardList(cardItems), HasMore: reply.HasMore, Offset: reply.Offset}, nil
}
