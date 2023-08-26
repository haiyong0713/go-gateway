package service

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/topic/ecode"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) HasCreateJurisdiction(ctx context.Context, mid int64) (*topicsvc.HasCreatJurisdictionRsp, error) {
	return s.topicGRPC.HasCreateJurisdiction(ctx, &topicsvc.HasCreateJurisdictionReq{Uid: mid})
}

func (s *Service) CreateTopic(ctx context.Context, mid int64, params *model.CreateTopicReq) (*model.CreateTopicRsp, error) {
	rsp, err := s.topicGRPC.CreateTopic(ctx, &topicsvc.CreateTopicReq{
		Uid:         mid,
		Name:        params.TopicName,
		Description: params.Description,
		Type:        params.SubmitTopicType,
		FromSource:  params.Scene,
	})
	if err != nil {
		log.Error("s.topicGRPC.CreateTopic mid=%d, params=%+v, error=%+v", mid, params, err)
		return nil, ecode.HandlePubEcodeToastErr(err)
	}
	reply := constructCreateTopicRsp(params.Scene, rsp)
	return reply, nil
}

func constructCreateTopicRsp(scene string, rsp *topicsvc.CreateTopicRsp) *model.CreateTopicRsp {
	var desc string
	switch scene {
	case _topicCreateSceneDynamic:
		desc = "审核通过后，你的动态将会自动关联上该话题。\n我们将通过私信通知您审核结果，请注意关注。\n您可以在「动态-话题广场-我的话题」内查看审核进度。"
	case _topicCreateSceneView:
		desc = "审核通过后，你的视频将会自动关联上该话题。\n我们将通过私信通知您审核结果，请注意关注。\n您可以在「动态-话题广场-我的话题」内查看审核进度。"
	default:
		desc = "审核通过后，我们将通过私信通知您审核结果，请注意关注。\n您可以在「动态-话题广场-我的话题」内查看审核进度。"
	}
	return &model.CreateTopicRsp{
		TopicId:     rsp.Id,
		TopicName:   rsp.Name,
		SuccessDesc: desc,
	}
}

func (s *Service) WebCreateTopic(ctx context.Context, mid int64, params *model.CreateTopicReq) (*model.CreateTopicRsp, error) {
	rsp, err := s.topicGRPC.CreateTopic(ctx, &topicsvc.CreateTopicReq{
		Uid:         mid,
		Name:        params.TopicName,
		Description: params.Description,
		Type:        params.SubmitTopicType,
		FromSource:  params.Scene,
	})
	if err != nil {
		log.Error("s.topicGRPC.WebCreateTopic mid=%d, params=%+v, error=%+v", mid, params, err)
		return nil, ecode.HandlePubEcodeToastErr(err)
	}
	reply := constructWebCreateTopicRsp(params.Scene, rsp)
	return reply, nil
}

func constructWebCreateTopicRsp(scene string, rsp *topicsvc.CreateTopicRsp) *model.CreateTopicRsp {
	var desc string
	switch scene {
	case _topicCreateSceneDynamic:
		desc = "审核通过后，你的动态将会自动关联上该话题。\n我们将通过私信通知您审核结果，请注意关注。\n您可以在移动端App「动态-话题广场-我的话题」内查看审核进度。"
	case _topicCreateSceneView:
		desc = "审核通过后，你的视频将会自动关联上该话题。\n我们将通过私信通知您审核结果，请注意关注。\n您可以在移动端App「动态-话题广场-我的话题」内查看审核进度。"
	default:
		if scene != _topicCreateSceneTopic {
			log.Error("constructCreateTopicRsp Unrecognized rsp=%+v, scene: %s", rsp, scene)
		}
		desc = "审核通过后，我们将通过私信通知您审核结果，请注意关注。\n您可以在「动态-话题广场-我的话题」内查看审核进度。"
	}
	return &model.CreateTopicRsp{
		TopicId:     rsp.Id,
		TopicName:   rsp.Name,
		SuccessDesc: desc,
	}
}
