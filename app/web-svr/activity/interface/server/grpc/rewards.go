package grpc

import (
	"context"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/rewards"
)

func (s *activityService) RewardsSendAward(ctx context.Context, req *v1.RewardsSendAwardReq) (reply *v1.RewardsSendAwardReply, err error) {
	reply = &v1.RewardsSendAwardReply{
		ExtraInfo: make(map[string]string, 0),
	}
	ctxForever := copyTrx(ctx)
	if req.Sync {
		reply, err = rewards.Client.SendAwardById(ctxForever, req.Mid, req.UniqueId, req.Business, req.AwardId, req.UpdateCache)
	} else {
		reply, err = rewards.Client.SendAwardByIdAsync(ctxForever, req.Mid, req.UniqueId, req.Business, req.AwardId, req.UpdateCache, req.UpdateDb)
	}

	return
}

func (s *activityService) RetryRewardsSendAward(ctx context.Context, req *v1.RetryRewardsSendAwardReq) (reply *v1.NoReply, err error) {
	ctxForever := copyTrx(ctx)
	reply = &v1.NoReply{}
	err = rewards.Client.RetrySendAwardById(ctxForever, req.Mid, req.UniqueId, req.Business, req.AwardId)
	return
}

func (s *activityService) RewardsAddAward(ctx context.Context, req *v1.RewardsAddAwardReq) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.AddAward(ctx, req)
	return
}

func (s *activityService) RewardsDelAward(ctx context.Context, req *v1.RewardsDelAwardReq) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.DelAward(ctx, req.Id)
	return
}

func (s *activityService) RewardsUpdateAward(ctx context.Context, req *v1.RewardsAwardInfo) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.UpdateAward(ctx, req)
	return
}

func (s *activityService) RewardsListAward(ctx context.Context, req *v1.RewardsListAwardReq) (reply *v1.RewardsListAwardReply, err error) {
	reply = &v1.RewardsListAwardReply{}
	res, err := rewards.Client.GetAwards(ctx, req)
	reply.List = res
	return
}

func (s *activityService) RewardsAddActivity(ctx context.Context, req *v1.RewardsAddActivityReq) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.AddActivity(ctx, req)
	return
}

func (s *activityService) RewardsDelActivity(ctx context.Context, req *v1.RewardsDelActivityReq) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.DelActivity(ctx, req)
	return
}

func (s *activityService) RewardsUpdateActivity(ctx context.Context, req *v1.RewardsUpdateActivityReq) (reply *v1.NoReply, err error) {
	reply = &v1.NoReply{}
	err = rewards.Client.UpdateActivity(ctx, req)
	return
}

func (s *activityService) RewardsListActivity(ctx context.Context, req *v1.RewardsListActivityReq) (reply *v1.RewardsListActivityReply, err error) {
	return rewards.Client.ListActivity(ctx, req)
}

func (s *activityService) RewardsGetActivityDetail(ctx context.Context, req *v1.RewardsGetActivityDetailReq) (reply *v1.RewardsGetActivityDetailReply, err error) {
	return rewards.Client.GetActivityDetail(ctx, req)
}

func (s *activityService) RewardsListAwardType(ctx context.Context, req *v1.RewardsListAwardTypeReq) (reply *v1.RewardsListAwardTypeReply, err error) {
	return rewards.Client.ListAwardType(ctx, req)
}

func (s *activityService) RewardsCheckSentStatus(ctx context.Context, req *v1.RewardsCheckSentStatusReq) (reply *v1.RewardsCheckSentStatusResp, err error) {
	return rewards.Client.RewardsCheckSentStatusReq(ctx, req)
}

func (s *activityService) RewardsGetAwardConfigById(ctx context.Context, req *v1.RewardsGetAwardConfigByIdReq) (reply *v1.RewardsAwardInfo, err error) {
	return rewards.Client.GetAwardConfigById(ctx, req.Id)
}

func (s *activityService) RewardsSendAwardV2(ctx context.Context, req *v1.RewardsSendAwardV2Req) (reply *v1.RewardsSendAwardReply, err error) {
	reply = &v1.RewardsSendAwardReply{
		ExtraInfo: make(map[string]string, 0),
	}
	ctxForever := copyTrx(ctx)
	reply, err = rewards.Client.SendAwardByIdAsync(ctxForever, req.Mid, req.UniqueId, req.Business, req.AwardId, true, true)

	return
}
