package service

import (
	"context"

	"go-common/library/log"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	api "go-gateway/app/app-svr/topic/interface/api"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) TopicMergedResource(ctx context.Context, req *api.TopicMergedResourceReq) (resp *api.TopicMergedResourceReply, err error) {
	general := constructGeneralParamFromCtx(ctx)
	general.SetLocalTime(req.LocalTime)
	c := constructPlayerArgs(ctx, general, req.PlayerArgs)
	return s.topicMergedResourceProcess(c, general, req)
}

func (s *Service) topicMergedResourceProcess(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicMergedResourceReq) (*api.TopicMergedResourceReply, error) {
	args := &topicsvc.QueryMergedResourceReq{
		TopicId:   req.TopicId,
		Type:      req.Type,
		Rid:       req.Rid,
		MergeType: req.MergeType,
		Offset:    req.Offset,
	}
	reply, err := s.topicGRPC.QueryMergedResource(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.QueryMergedResource args=%+v, error=%+v", args, err)
		return nil, err
	}
	dynSchemaCtx := initDynSchemaContext(ctx, req.TopicId, req.FromSortBy, req.Offset)
	cardItems, err := s.dynCardProcess(dynSchemaCtx, general, convertMergedResourceToDynMetaCardListItem(reply.Items))
	if err != nil {
		return nil, err
	}
	return &api.TopicMergedResourceReply{
		TopicCardList: &api.TopicCardList{
			TopicCardItems: cardItems,
			Offset:         reply.Offset,
			HasMore:        reply.HasMore,
		},
	}, nil
}

func convertMergedResourceToDynMetaCardListItem(mergedList []*topicsvc.MergedResourceItem) []*topiccardmodel.DynMetaCardListParam {
	var res []*topiccardmodel.DynMetaCardListParam
	for _, v := range mergedList {
		if v.Type != _dynamicCardType {
			continue
		}
		res = append(res, &topiccardmodel.DynMetaCardListParam{
			DynId: v.Rid,
		})
	}
	return res
}
