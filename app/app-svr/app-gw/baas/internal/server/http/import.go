package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/baas/api"
)

func addImport(ctx *bm.Context) {
	req := &api.AddImportRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddImport(ctx, req))
}

func updateImport(ctx *bm.Context) {
	req := &api.UpdateImportRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UpdateImport(ctx, req))
}
