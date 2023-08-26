package http

import (
	bm "go-common/library/net/http/blademaster"
)

func exportStatistics(ctx *bm.Context) {
	params := &struct {
		Select string `form:"select" default:"1d"`
	}{}
	if err := ctx.Bind(params); err != nil {
		return
	}
	midV, _ := ctx.Get("mid")
	mid := midV.(int64)
	ctx.JSON(accSvr.ExportStatistics(ctx, mid, params.Select))
}
