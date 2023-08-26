package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

func checkExpressionWithDevice(ctx *bm.Context) {
	req := new(model.CheckExpressionReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.CheckExpressionWithDevice(ctx, req))
}

func checkExpressionWithContext(ctx *bm.Context) {
	req := new(model.CheckExpressionReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.CheckExpressionWithContext(ctx, req))
}
