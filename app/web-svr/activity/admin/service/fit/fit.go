package fit

import (
	"context"
	fit "go-gateway/app/web-svr/activity/admin/model/fit"
)

// AddOnePlan service层添加一条系列计划
func (s *Service) AddOnePlan(ctx context.Context, req *fit.PlanRecord) (lastID int64, err error) {
	return s.dao.AddOnePlan(ctx, req)
}
