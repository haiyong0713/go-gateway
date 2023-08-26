package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func acg2020Task(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	var midI int64
	if mid != nil {
		midI = mid.(int64)
	}
	ctx.JSON(service.AcgSvc.Task(ctx, midI))
}
