package newyear2021

import (
	"context"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/rewards"
)

func (s *Service) Support(ctx context.Context, mid int64) (res *model.UserSupportInfo, err error) {
	res = &model.UserSupportInfo{}
	lt, err := s.GetLevelTaskStatus(ctx, mid)
	if err != nil {
		return
	}
	dt, err := s.GetDailyTaskStatus(ctx, mid)
	if err != nil {
		return
	}
	res.LevelTasks = lt
	res.DailyTasks = dt
	res.Mid = mid
	rr, err := rewards.Client.GetAwardRecordByMidAndActivityId(ctx, mid, []int64{0}, 9999)
	if err != nil {
		return
	}
	res.ReceiveRecords = rr
	isVip, err := s.dao.IsUserPaid(ctx, mid, 33385) //FIXME: make this configurable
	res.IsVip = isVip
	return
}
