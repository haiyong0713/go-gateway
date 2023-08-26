package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
)

func familyAggregation(ctx *bm.Context) {
	req := new(model.AggregationReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.Aggregation(ctx, req, mid))
}

func familyTeenGuard(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.TeenGuard(ctx, mid))
}

func familyIdentity(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.Identity(ctx, mid))
}

func createFamilyQrcode(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.CreateQrcode(ctx, mid))
}

func familyQrcodeInfo(ctx *bm.Context) {
	req := new(model.QrcodeInfoReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.QrcodeInfo(ctx, req, mid))
}

func familyQrcodeStatus(ctx *bm.Context) {
	req := new(model.QrcodeStatusReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(teenSvr.QrcodeStatus(ctx, req))
}

func familyParentIndex(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.ParentIndex(ctx, mid))
}

func familyParentUnbind(ctx *bm.Context) {
	req := new(model.ParentUnbindReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.ParentUnbind(ctx, req, mid))
}

func parentUpdateTeenager(ctx *bm.Context) {
	req := new(model.ParentUpdateTeenagerReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.ParentUpdateTeenager(ctx, req, mid))
}

func familyChildIndex(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.ChildIndex(ctx, mid))
}

func familyChildBind(ctx *bm.Context) {
	req := new(model.ChildBindReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.ChildBind(ctx, req, mid))
}

func familyChildUnbind(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.ChildUnbind(ctx, mid))
}

func timelockInfo(ctx *bm.Context) {
	req := new(model.TimelockInfoReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.TimelockInfo(ctx, req, mid))
}

func updateTimelock(ctx *bm.Context) {
	req := new(model.UpdateTimelockReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(nil, teenSvr.UpdateTimelock(ctx, req, mid))
}

func timelockPwd(ctx *bm.Context) {
	req := new(model.TimelockPwdReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.TimelockPwd(ctx, req, mid))
}

func verifyTimelockPwd(ctx *bm.Context) {
	req := new(model.VerifyTimelockPwdReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.VerifyTimelockPwd(ctx, req, mid))
}
