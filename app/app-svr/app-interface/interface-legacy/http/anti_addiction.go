package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
)

func antiAddictionRule(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.AntiAddictionRule(ctx, mid))
}

func aggregationStatus(ctx *bm.Context) {
	req := &model.AggregationStatusReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.AggregationStatus(ctx, req, mid))
}

func setSleepRemind(ctx *bm.Context) {
	req := &model.SetSleepRemindReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.SetSleepRemind(ctx, req, mid))
}
