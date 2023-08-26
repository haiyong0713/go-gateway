package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func archiveInfo(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.LikeSvc.ArcInfo(ctx, mid))
}

func contriLikes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.LikeSvc.Likes(ctx, midI))
}

func addContriTimes(ctx *bm.Context) {
	v := new(struct {
		ActionType int `form:"action_type" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.LikeSvc.AddContriLotteryTimes(ctx, midI, v.ActionType))
}

func lightBcut(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.LikeSvc.LightBcutInfo(ctx, mid))
}

func totalRank(ctx *bm.Context) {
	v := new(struct {
		Datetime string `form:"dt"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.TotalRank(ctx, v.Datetime))
}

func haveMoney(ctx *bm.Context) {
	go func() {
		service.LikeSvc.HaveMoney(ctx)
	}()
	ctx.JSON("ok", nil)
}
