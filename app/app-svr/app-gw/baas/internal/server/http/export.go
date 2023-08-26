package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/baas/api"
)

func exportList(ctx *bm.Context) {
	req := &api.ExportListRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.ExportList(ctx, req))
}

func addExport(ctx *bm.Context) {
	req := &api.AddExportRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddExport(ctx, req))
}

func updateExport(ctx *bm.Context) {
	req := &api.UpdateExportRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UpdateExport(ctx, req))
}
