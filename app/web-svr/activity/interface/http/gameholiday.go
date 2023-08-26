package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func addGhLotteryTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.GameHolidaySvc.AddLotteryTimes(ctx, midI))
}

func ghLikes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.GameHolidaySvc.Likes(ctx, midI))
}
