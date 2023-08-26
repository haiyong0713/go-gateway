package http

import (
	"net/http"

	"go-common/library/log"

	bm "go-common/library/net/http/blademaster"
)

func baasImpl(ctx *bm.Context) {
	exportApi := ctx.Params.ByName("path")
	out, err := svc.Common.BaasImpl(ctx, exportApi)
	if err != nil {
		log.Error("Failed to impl baas: %+v", err)
		ctx.String(http.StatusNotFound, "")
		return
	}
	ctx.String(200, "%s", out)
}
