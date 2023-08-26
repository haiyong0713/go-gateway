package vote

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
	riskModel "go-gateway/app/web-svr/activity/interface/model/risk"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"go-gateway/app/web-svr/activity/interface/service/vote/adapters"
)

type Service struct {
	dao *vote.Dao
}

func New(c *conf.Config) *Service {
	ds := make(map[string]vote.DataSource, 0)
	ds["MOCK"] = adapters.MockDS
	ds[model.DSTypeOperVideo] = adapters.OperationVideoDS
	ds[model.DSTypeOperPic] = adapters.OperationPicDS
	ds[model.DSTypeOperUp] = adapters.OperationUpDS
	ds[model.DSTypeUp] = adapters.UpSourceDS
	ds[model.DSTypeVideo] = adapters.VideoSourceDS
	return &Service{
		dao: vote.New(c, ds),
	}
}

func (s *Service) AddVoteActivity(ctx context.Context, req *api.AddVoteActivityReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.AddActivity(ctx, req)
	return
}

func (s *Service) DelVoteActivity(ctx context.Context, req *api.DelVoteActivityReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.DelActivity(ctx, req)
	return
}

func (s *Service) UpdateVoteActivity(ctx context.Context, req *api.UpdateVoteActivityReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.UpdateActivity(ctx, req)
	return
}

func (s *Service) ListVoteActivity(ctx context.Context, req *api.ListVoteActivityReq) (res *api.ListVoteActivityResp, err error) {
	res, err = s.dao.ListActivity(ctx, req)
	return
}

func (s *Service) ListVoteActivityForRefresh(ctx context.Context, req *api.ListVoteActivityForRefreshReq) (res *api.ListVoteActivityForRefreshResp, err error) {
	res, err = s.dao.ListVoteActivityForRefresh(ctx, req)
	return
}

func (s *Service) UpdateVoteActivityRule(ctx context.Context, req *api.UpdateVoteActivityRuleReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.UpdateActivityRule(ctx, req)
	return
}

func (s *Service) AddVoteActivityDataSourceGroup(ctx context.Context, req *api.AddVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.AddActivityDataSourceGroup(ctx, req)
	return
}

func (s *Service) DelVoteActivityDataSourceGroup(ctx context.Context, req *api.DelVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.DelActivityDataSourceGroup(ctx, req)
	return
}

func (s *Service) UpdateActivityDataSourceGroup(ctx context.Context, req *api.UpdateVoteActivityDataSourceGroupReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.UpdateActivityDataSourceGroup(ctx, req)
	return
}

func (s *Service) ListActivityDataSourceGroups(ctx context.Context, req *api.ListVoteActivityDataSourceGroupsReq) (res *api.ListVoteActivityDataSourceGroupsResp, err error) {
	return s.dao.ListActivityDataSourceGroups(ctx, req)
}

func (s *Service) AddVoteActivityBlackList(ctx context.Context, req *api.AddVoteActivityBlackListReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.AddVoteActivityBlackList(ctx, req)
	return
}

func (s *Service) DelVoteActivityBlackList(ctx context.Context, req *api.DelVoteActivityBlackListReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.DelVoteActivityBlackList(ctx, req)
	return
}
func (s *Service) UpdateVoteActivityInterveneVoteCount(ctx context.Context, req *api.UpdateVoteActivityInterveneVoteCountReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.UpdateVoteActivityInterveneVoteCount(ctx, req)
	return
}

func (s *Service) GetVoteActivityRankInternal(ctx context.Context, req *api.GetVoteActivityRankInternalReq) (res *api.GetVoteActivityRankInternalResp, err error) {
	res, err = s.dao.GetDSGRankInternal(ctx, req)
	return
}

func (s *Service) GetVoteActivityRankExternal(ctx context.Context, mid int64, req *model.RankExternalParams) (res *model.RankResultExternal, err error) {
	params := &model.InnerRankParams{
		Mid:               mid,
		ActivityId:        req.ActivityId,
		DataSourceGroupId: req.DataSourceGroupId,
		Version:           req.Version,
		Pn:                req.Pn,
		Ps:                req.Ps,
	}
	switch req.Sort {
	case 1: //票数排序
		res, err = s.dao.GetDSGRankExternal(ctx, params)
	case 2: //随机排序
		res, err = s.dao.GetDSGRankExternalOrder(ctx, true, params)
	case 3: //时间排序
		res, err = s.dao.GetDSGRankExternalOrder(ctx, false, params)
	default:
		res, err = s.dao.GetDSGRankExternal(ctx, params)
	}

	return
}

func (s *Service) RefreshVoteActivityDSItems(ctx context.Context, req *api.RefreshVoteActivityDSItemsReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.RefreshVoteActivityDSItems(ctx, req.ActivityId)
	return
}

func (s *Service) RefreshVoteActivityRankExternal(ctx context.Context, req *api.RefreshVoteActivityRankExternalReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.RefreshVoteActivityRankExternal(ctx, req.ActivityId)
	return
}

func (s *Service) RefreshVoteActivityRankInternal(ctx context.Context, req *api.RefreshVoteActivityRankInternalReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.RefreshVoteActivityRankInternal(ctx, req.ActivityId)
	return
}

func (s *Service) RefreshVoteActivityRankZset(ctx context.Context, req *api.RefreshVoteActivityRankZsetReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	err = s.dao.RefreshVoteActivityRankZset(ctx, req)
	return
}

func (s *Service) DoVote(ctx context.Context, mid int64, risk *riskModel.Base, req *model.DoVoteParams) (resp *api.VoteUserDoResp, err error) {
	return s.dao.DoVote(ctx, mid, risk, req)
}

func (s *Service) UndoVote(ctx context.Context, mid int64, req *model.UndoVoteParams) (resp *api.VoteUserUndoResp, err error) {
	return s.dao.UndoVote(ctx, mid, req)
}

func (s *Service) VoteUserDo(ctx context.Context, req *api.VoteUserDoReq) (res *api.VoteUserDoResp, err error) {
	var risk *riskModel.Base
	res = &api.VoteUserDoResp{}
	if req.Risk != nil {
		risk = req.Risk.ToBase(req.Mid, riskModel.ActionVoteNew)
	}
	res, err = s.dao.DoVote(ctx, req.Mid, risk, &model.DoVoteParams{
		ActivityId:        req.ActivityId,
		DataSourceGroupId: req.SourceGroupId,
		DataSourceItemId:  req.SourceItemId,
		Vote:              req.VoteCount,
	})
	return
}

func (s *Service) VoteUserUndo(ctx context.Context, req *api.VoteUserUndoReq) (res *api.VoteUserUndoResp, err error) {
	return s.dao.UndoVote(ctx, req.Mid, &model.UndoVoteParams{
		ActivityId:        req.ActivityId,
		DataSourceGroupId: req.SourceGroupId,
		DataSourceItemId:  req.SourceItemId,
	})
}

func (s *Service) GetVoteActivityRank(ctx context.Context, req *api.GetVoteActivityRankReq) (res *api.GetVoteActivityRankResp, err error) {
	tmpRes, err := s.GetVoteActivityRankExternal(ctx, req.Mid, &model.RankExternalParams{
		ActivityId:        req.ActivityId,
		DataSourceGroupId: req.SourceGroupId,
		Version:           0,
		Pn:                req.Pn,
		Ps:                req.Ps,
		Sort:              req.Sort,
	})
	if err != nil {
		return
	}
	res = tmpRes.ToPB()
	return
}

func (s *Service) Search(ctx context.Context, mid int64, req *model.RankSearchParams) (*model.RankSearchResultExternal, error) {
	return s.dao.Search(ctx, mid, req)
}

func (s *Service) GetItemContributionRank(ctx context.Context, req *api.VoteGetItemContributionRankReq) (res *api.VoteGetItemContributionRankResp, err error) {
	return s.dao.GetItemContributionRank(ctx, req)
}

func (s *Service) UserAddTimes(ctx context.Context, req *api.VoteUserAddTimesReq) (res *api.NoReply, err error) {
	return s.dao.AddUserTimes(ctx, req)
}

func (s *Service) UserGetTimes(ctx context.Context, req *api.VoteUserGetTimesReq) (res *api.VoteUserGetTimesResp, err error) {
	return s.dao.UserGetTimes(ctx, req)
}
