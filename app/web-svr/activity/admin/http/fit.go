package http

import (
	bm "go-common/library/net/http/blademaster"
	fit "go-gateway/app/web-svr/activity/admin/model/fit"
)

// addOnePlan 后台添加一条系列计划
func addOnePlan(ctx *bm.Context) {
	v := new(fit.PlanRecord)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(fitSrv.AddOnePlan(ctx, v))
}

// updatePlanById 根据id更新计划
func updatePlanById(ctx *bm.Context) {
	v := new(fit.UpdatePlanRecord)
	if err := ctx.Bind(v); err != nil {
		return
	}

}
