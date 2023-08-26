package service

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/topic/interface/internal/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) SearchPubTopics(ctx context.Context, params *model.SearchPubTopicsReq) (*model.SearchPubTopicsRsp, error) {
	args := constructSearchPubTopicsReq(ctx, params)
	rsp, err := s.topicGRPC.SearchPubTopics(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.SearchPubTopics args=%+v, error=%+v", args, err)
		return nil, err
	}
	return convertToSearchPubTopicsRsp(rsp), nil
}

func constructSearchPubTopicsReq(ctx context.Context, params *model.SearchPubTopicsReq) *topicsvc.SearchPubTopicsReq {
	const (
		searchPubSceneryDefault = "dynamic"
	)
	general, from := constructGeneralParamFromCtx(ctx), params.From
	if params.From == "" {
		from = searchPubSceneryDefault
	}
	return &topicsvc.SearchPubTopicsReq{
		Uid:         general.Mid,
		KeyWords:    params.Keywords,
		Content:     params.Content,
		UploadId:    params.UploadId,
		FromTopicId: params.FromTopicId,
		Meta:        constructTopicCommonMetaDataCtrl(general, from),
		PageInfo: &topiccommon.PaginationReq{
			PageSize: params.PageSize,
			PageNum:  params.PageNum,
			Offset:   params.Offset,
		},
	}
}

func convertToSearchPubTopicsRsp(params *topicsvc.SearchPubTopicsRsp) *model.SearchPubTopicsRsp {
	topicItems := make([]*model.TopicItem, 0, len(params.Topics))
	for _, topic := range params.Topics {
		item := convertToTopicInfoJson(topic)
		item.StatDesc = makeTopicItemDesc(topic.View, topic.Discuss)
		item.TopicRcmdType = topic.TopicRcmdType
		topicItems = append(topicItems, item)
	}
	return &model.SearchPubTopicsRsp{
		NewTopic:              makeNewTopic(params.NewTopic, params.HasCreateJurisdiction),
		HasCreateJurisdiction: params.HasCreateJurisdiction,
		TopicItems:            topicItems,
		RequestId:             params.RequestId,
		PageInfo:              params.PageInfo,
	}
}

func makeNewTopic(raw *topicsvc.SearchNewTopic, hasCreateJurisdiction bool) model.SearchNewTopic {
	res := model.SearchNewTopic{
		Name: raw.Name,
	}
	if !hasCreateJurisdiction {
		return res
	}
	res.IsNew = raw.IsNew
	return res
}

func (s *Service) UsrPubTopics(ctx context.Context, mid int64, params *model.UsrPubTopicsReq) (*model.UsrPubTopicsRsp, error) {
	args := constructUsrPubTopicsReq(mid, params)
	rsp, err := s.topicGRPC.UsrPubTopics(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.UsrPubTopics args=%+v, error=%+v", args, err)
		return nil, err
	}
	res := convertToUsrPubTopicsRsp(rsp)
	return res, nil
}

func constructUsrPubTopicsReq(mid int64, params *model.UsrPubTopicsReq) *topicsvc.UsrPubTopicsReq {
	return &topicsvc.UsrPubTopicsReq{
		Uid:   mid,
		State: params.State,
		PageInfo: &topiccommon.PaginationReq{
			PageSize: params.PageSize,
			PageNum:  params.PageNum,
			Offset:   params.Offset,
		},
	}
}

func convertToUsrPubTopicsRsp(params *topicsvc.UsrPubTopicsRsp) *model.UsrPubTopicsRsp {
	topicItems := make([]*model.TopicItem, 0, len(params.Topics))
	for _, topic := range params.Topics {
		item := convertToTopicInfoJson(topic)
		item.State = topic.State
		topicItems = append(topicItems, item)
	}
	return &model.UsrPubTopicsRsp{
		HasCreateJurisdiction: params.HasCreateJurisdiction,
		TopicItems:            topicItems,
		PageInfo:              params.PageInfo,
	}
}

func (s *Service) IsAlreadyExistedTopic(ctx context.Context, params *model.IsAlreadyExistedTopicReq) (*model.IsAlreadyExistedTopicRsp, error) {
	var (
		isExistedTopic bool
		synTopicItems  []*topiccommon.TopicInfo
	)
	general := constructGeneralParamFromCtx(ctx)
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.topicGRPC.IsAreadyExistedTopicV2(ctx, &topicsvc.IsAreadyExistedTopicReq{Name: params.Topic})
		if err != nil {
			log.Error("s.topicGRPC.IsAreadyExistedTopicV2 params=%+v, err=%+v", params, err)
			return nil
		}
		isExistedTopic = reply.AlreadyExisted
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.topicGRPC.TopicSynonym(ctx, &topicsvc.TopicSynonymReq{
			Name:        params.Topic,
			Description: params.Description,
			Uid:         general.Mid,
			Metadata:    constructTopicCommonMetaDataCtrl(general, ""),
		})
		if err != nil {
			log.Error("s.topicGRPC.TopicSynonym params=%+v, err=%+v", params, err)
			return nil
		}
		if len(reply.Topics) > 0 {
			synTopicItems = reply.Topics
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("IsAlreadyExistedTopic eg.Wait() error=%+v", err)
		return &model.IsAlreadyExistedTopicRsp{AlreadyExisted: false}, nil
	}
	return &model.IsAlreadyExistedTopicRsp{
		AlreadyExisted: isExistedTopic,
		SynonymTopic:   &model.SynonymTopic{TopicItems: convertCommonTopicInfoToItems(synTopicItems)},
	}, nil
}

func (s *Service) PubEvents(ctx context.Context, params *model.TopicPubEventsReq, timeStamp int64) (*model.TopicPubEventsRsp, error) {
	mids, err := xstr.SplitInts(params.ShowMids)
	if err != nil {
		return nil, ecode.RequestErr
	}
	reply, err := s.accGRPC.Cards3(ctx, &accountgrpc.MidsReq{Mids: mids, RealIp: metadata.String(ctx, metadata.RemoteIP)})
	if err != nil {
		log.Error("s.accGRPC.Cards3 error=%+v", err)
		return nil, nil
	}
	var showMembers []*model.ShowMember
	for _, mid := range mids {
		if v, ok := reply.Cards[mid]; ok {
			showMembers = append(showMembers, &model.ShowMember{Mid: mid, Avatar: v.Face})
		}
	}
	return &model.TopicPubEventsRsp{
		ShowText:     makePubEventShowText(params.PubNum),
		ShowMembers:  showMembers,
		ReqTimestamp: timeStamp,
	}, nil
}

func makePubEventShowText(pubNum int64) string {
	const (
		maxShowTextNum = 99
	)
	if pubNum > maxShowTextNum {
		return "99+条新增内容"
	}
	return fmt.Sprintf("%d条新增内容", pubNum)
}

func (s *Service) SearchRcmdPubTopics(ctx context.Context, req *model.SearchRcmdPubTopicsReq) (*model.SearchRcmdPubTopicsRsp, error) {
	res := &model.SearchRcmdPubTopicsRsp{TopicItems: []*model.SeachRcmdTopicItem{}}
	if req.Keywords == "" {
		return res, nil
	}
	general := constructGeneralParamFromCtx(ctx)
	reply, err := s.topicGRPC.SearchRcmdPubTopics(ctx, &topicsvc.SearchRcmdPubTopicsReq{
		Mid:         general.Mid,
		Content:     req.Keywords,
		UploadId:    req.UploadId,
		Platform:    general.GetPlatform(),
		Build:       general.GetBuildStr(),
		FromTopicId: req.FromTopicId,
		Ip:          general.IP,
	})
	if err != nil {
		log.Info("s.topicGRPC.SearchRcmdPubTopics err=%+v", err)
		return res, nil
	}
	for _, v := range reply.Topics {
		res.TopicItems = append(res.TopicItems, &model.SeachRcmdTopicItem{TopicId: v.Id, TopicName: v.Name})
	}
	res.RequestId = reply.RequestId
	return res, nil
}

func (s *Service) PubTopicEndpoint(ctx context.Context, mid int64) (*model.PubTopicEndpointRsp, error) {
	res := &model.PubTopicEndpointRsp{}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		reply, err := s.topicGRPC.TopicUsrPubCnt(ctx, &topicsvc.TopicUsrPubCntReq{Uid: mid})
		if err != nil {
			log.Error("s.topicGRPC.TopicUsrPubCnt mid=%d, err=%+v", mid, err)
			return nil
		}
		res.MaxCnt = reply.MaxCnt
		res.RemainCnt = reply.RemainCnt
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		JurisdictionReply, err := s.topicGRPC.HasCreateJurisdiction(ctx, &topicsvc.HasCreateJurisdictionReq{
			Uid: mid,
		})
		if err != nil {
			log.Error("s.topicGRPC.HasCreateJurisdiction mid=%d, err=%+v", mid, err)
			return nil
		}
		res.HasCreateJurisdiction = JurisdictionReply.HasJurisdiction
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("PubTopicEndpoint eg.Wait() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func (s *Service) PubTopicUpload(ctx context.Context, mid int64, params *model.PubTopicUploadReq) {
	payload := infocv2.NewLogStreamV(s.customConfig.PubInfoc.LogID, log.String(strconv.FormatInt(mid, 10)), log.String(params.TopicId), log.String(params.RequestId),
		log.String(params.UploadId))
	if err := s.pubInfocv2.Info(ctx, payload); err != nil {
		log.Warn("infocV2SendTopicPubData() s.pubInfocv2.Info() mid(%d) params(%+v) error(%+v)", mid, params, err)
	}
}
