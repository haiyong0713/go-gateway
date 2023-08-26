package grpc

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

func (s *activityService) AddVoteActivity(ctx context.Context, req *api.AddVoteActivityReq) (res *api.NoReply, err error) {
	return service.VoteSvr.AddVoteActivity(ctx, req)
}

func (s *activityService) DelVoteActivity(ctx context.Context, req *api.DelVoteActivityReq) (res *api.NoReply, err error) {
	return service.VoteSvr.DelVoteActivity(ctx, req)
}

func (s *activityService) UpdateVoteActivity(ctx context.Context, req *api.UpdateVoteActivityReq) (res *api.NoReply, err error) {
	return service.VoteSvr.UpdateVoteActivity(ctx, req)
}
func (s *activityService) ListVoteActivity(ctx context.Context, req *api.ListVoteActivityReq) (res *api.ListVoteActivityResp, err error) {
	return service.VoteSvr.ListVoteActivity(ctx, req)
}

func (s *activityService) ListVoteActivityForRefresh(ctx context.Context, req *api.ListVoteActivityForRefreshReq) (res *api.ListVoteActivityForRefreshResp, err error) {
	return service.VoteSvr.ListVoteActivityForRefresh(ctx, req)
}

func (s *activityService) UpdateVoteActivityRule(ctx context.Context, req *api.UpdateVoteActivityRuleReq) (res *api.NoReply, err error) {
	return service.VoteSvr.UpdateVoteActivityRule(ctx, req)
}

func (s *activityService) AddVoteActivityDataSourceGroup(ctx context.Context, req *api.AddVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	return service.VoteSvr.AddVoteActivityDataSourceGroup(ctx, req)
}

func (s *activityService) DelVoteActivityDataSourceGroup(ctx context.Context, req *api.DelVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	return service.VoteSvr.DelVoteActivityDataSourceGroup(ctx, req)
}

func (s *activityService) UpdateVoteActivityDataSourceGroup(ctx context.Context, req *api.UpdateVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	return service.VoteSvr.UpdateActivityDataSourceGroup(ctx, req)
}

func (s *activityService) ListVoteActivityDataSourceGroups(ctx context.Context, req *api.ListVoteActivityDataSourceGroupsReq) (res *api.ListVoteActivityDataSourceGroupsResp, err error) {
	return service.VoteSvr.ListActivityDataSourceGroups(ctx, req)
}

func (s *activityService) AddVoteActivityBlackList(ctx context.Context, req *api.AddVoteActivityBlackListReq) (res *api.NoReply, err error) {
	return service.VoteSvr.AddVoteActivityBlackList(ctx, req)
}

func (s *activityService) DelVoteActivityBlackList(ctx context.Context, req *api.DelVoteActivityBlackListReq) (res *api.NoReply, err error) {
	return service.VoteSvr.DelVoteActivityBlackList(ctx, req)
}

func (s *activityService) UpdateVoteActivityInterveneVoteCount(ctx context.Context, req *api.UpdateVoteActivityInterveneVoteCountReq) (res *api.NoReply, err error) {
	return service.VoteSvr.UpdateVoteActivityInterveneVoteCount(ctx, req)
}
func (s *activityService) GetVoteActivityRankInternal(ctx context.Context, req *api.GetVoteActivityRankInternalReq) (res *api.GetVoteActivityRankInternalResp, err error) {
	return service.VoteSvr.GetVoteActivityRankInternal(ctx, req)
}

func copyTrx(ctx context.Context) context.Context {
	c := metadata.WithContext(ctx)
	tr, ok := trace.FromContext(ctx)
	if ok {
		c = trace.NewContext(c, tr)
	}
	return c
}

func (s *activityService) RefreshVoteActivityDSItems(ctx context.Context, req *api.RefreshVoteActivityDSItemsReq) (res *api.NoReply, err error) {
	ctxWithoutTimeout := copyTrx(ctx)
	res, err = service.VoteSvr.RefreshVoteActivityDSItems(ctxWithoutTimeout, req)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, err.Error())
	}
	return
}

func (s *activityService) RefreshVoteActivityRankExternal(ctx context.Context, req *api.RefreshVoteActivityRankExternalReq) (res *api.NoReply, err error) {
	ctxWithoutTimeout := copyTrx(ctx)
	res, err = service.VoteSvr.RefreshVoteActivityRankExternal(ctxWithoutTimeout, req)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, err.Error())
	}
	return
}

func (s *activityService) RefreshVoteActivityRankZset(ctx context.Context, req *api.RefreshVoteActivityRankZsetReq) (res *api.NoReply, err error) {
	ctxWithoutTimeout := copyTrx(ctx)
	res, err = service.VoteSvr.RefreshVoteActivityRankZset(ctxWithoutTimeout, req)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, err.Error())
	}
	return
}

func (s *activityService) RefreshVoteActivityRankInternal(ctx context.Context, req *api.RefreshVoteActivityRankInternalReq) (res *api.NoReply, err error) {
	ctxWithoutTimeout := copyTrx(ctx)
	res, err = service.VoteSvr.RefreshVoteActivityRankInternal(ctxWithoutTimeout, req)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, err.Error())
	}
	return
}

func (s *activityService) VoteUserDo(ctx context.Context, req *api.VoteUserDoReq) (res *api.VoteUserDoResp, err error) {
	return service.VoteSvr.VoteUserDo(ctx, req)
}

func (s *activityService) VoteUserUndo(ctx context.Context, req *api.VoteUserUndoReq) (res *api.VoteUserUndoResp, err error) {
	return service.VoteSvr.VoteUserUndo(ctx, req)
}

func (s *activityService) GetVoteActivityRank(ctx context.Context, req *api.GetVoteActivityRankReq) (res *api.GetVoteActivityRankResp, err error) {
	return service.VoteSvr.GetVoteActivityRank(ctx, req)
}

func (s *activityService) VoteUserAddTmpTimes(ctx context.Context, req *api.VoteUserAddTmpTimesReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = ecode.RequestErr
	return
}

func (s *activityService) VoteGetItemContributionRank(ctx context.Context, req *api.VoteGetItemContributionRankReq) (res *api.VoteGetItemContributionRankResp, err error) {
	return service.VoteSvr.GetItemContributionRank(ctx, req)
}

func (s *activityService) VoteUserGetTimes(ctx context.Context, req *api.VoteUserGetTimesReq) (res *api.VoteUserGetTimesResp, err error) {
	return service.VoteSvr.UserGetTimes(ctx, req)
}

func (s *activityService) VoteUserAddTimes(ctx context.Context, req *api.VoteUserAddTimesReq) (res *api.NoReply, err error) {
	return service.VoteSvr.UserAddTimes(ctx, req)
}
