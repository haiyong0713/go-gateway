package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func articleDayJoin(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.LikeSvc.JoinArticleDay(ctx, mid))
}

func articleDayInfo(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.LikeSvc.ArticleDayInfo(ctx, mid))
}
