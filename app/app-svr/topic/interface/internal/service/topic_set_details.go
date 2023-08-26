package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	api "go-gateway/app/app-svr/topic/interface/api"

	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

func (s *Service) TopicSetDetails(ctx context.Context, req *api.TopicSetDetailsReq) (resp *api.TopicSetDetailsReply, err error) {
	general := constructGeneralParamFromCtx(ctx)
	return s.topicSetDetailsProcess(ctx, general, req)
}

func (s *Service) topicSetDetailsProcess(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicSetDetailsReq) (*api.TopicSetDetailsReply, error) {
	res := &api.TopicSetDetailsReply{}
	// 修复ios phone672之前版本bug:调整pagesize
	if isNeedFixPageProblemVersion(ctx) {
		req.PageSize = 7
	}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.topicGRPC.TopicSetInfo(ctx, &topicsvc.TopicSetInfoReq{
			SetId:    req.SetId,
			Uid:      general.Mid,
			Metadata: constructTopicCommonMetaDataCtrl(general, "/bilibili.app.topic.v1.Topic/TopicSetDetails"),
		})
		if err != nil {
			log.Error("s.topicGRPC.TopicSetInfo req=%+v, err=%+v", req, err)
			return err
		}
		res.TopicSetHeadInfo = constructTopicSetHeadInfo(reply)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.topicGRPC.TopicListOfTopicSet(ctx, &topicsvc.TopicListOfTopicSetReq{
			SetId:    req.SetId,
			SortBy:   int32(req.SortBy),
			Offset:   req.Offset,
			PageSize: req.PageSize,
		})
		if err != nil {
			log.Error("s.topicGRPC.TopicListOfTopicSet req=%+v, err=%+v", req, err)
			return nil
		}
		res.TopicInfo = convertToTopicInfoList(reply.Topics)
		if reply.SortCfg != nil {
			res.SortCfg = &api.TopicSetSortCfg{
				DefaultSortBy: int64(reply.SortCfg.SortBy),
				AllSortBy:     convertToAllSortBy(reply.SortCfg.AllSortBy),
			}
		}
		if isNeedFixPageProblemVersion(ctx) {
			// 不给翻页的返回数据
			return nil
		}
		res.HasMore = reply.HasMore
		res.Offset = reply.Offset
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func isNeedFixPageProblemVersion(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build("<", 67200000)
	}).MustFinish()
}

func convertToAllSortBy(contents []*topicsvc.SortContent) []*api.SortContent {
	var res []*api.SortContent
	for _, v := range contents {
		res = append(res, &api.SortContent{
			SortBy:   v.SortBy,
			SortName: v.SortName,
		})
	}
	return res
}

func convertToTopicInfoList(topics []*topiccommon.TopicInfo) []*api.TopicInfo {
	const (
		topicIcon = "https://i0.hdslb.com/bfs/feed-admin/716042de651fca2d23b63be46d6291f7196682df.png"
	)
	var res []*api.TopicInfo
	for _, v := range topics {
		res = append(res, &api.TopicInfo{
			Id:             v.Id,
			Name:           v.Name,
			View:           v.View,
			Discuss:        v.Discuss,
			Description:    v.Description,
			JumpUrl:        v.JumpUrl,
			StatsDesc:      makeTopicItemDesc(v.View, v.Discuss),
			FixedTopicIcon: topicIcon,
		})
	}
	return res
}

func constructTopicSetHeadInfo(reply *topicsvc.TopicSetInfoRsp) *api.TopicSetHeadInfo {
	if reply.BasicInfo == nil {
		return nil
	}
	res := &api.TopicSetHeadInfo{
		TopicSet:     constructTopicSet(reply.BasicInfo),
		TopicCntText: fmt.Sprintf("收录%d个话题", reply.TopicCnt),
		HeadImgUrl:   reply.HeadImgUrl,
		IconUrl:      reply.IconUrl,
		IsFav:        reply.IsFav,
		IsFirstTime:  reply.IsFirstTime,
	}
	if reply.MissionUrl != "" {
		res.MissionUrl = reply.MissionUrl
		res.MissionText = reply.MissionText
	}
	return res
}
