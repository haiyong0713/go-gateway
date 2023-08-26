package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
)

func winterCourse(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.LikeSvc.WinterCourse(ctx, mid))
}

func winterJoin(ctx *bm.Context) {
	arg := new(like.ParamWinterJoin)
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	arg.IP = metadata.String(ctx, metadata.RemoteIP)
	ctx.JSON(nil, service.LikeSvc.WinterJoin(ctx, mid, arg))
}

func winterProgress(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.LikeSvc.WinterProgress(ctx, mid))
}

func winterInnerProgress(ctx *bm.Context) {
	p := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(service.LikeSvc.WinterProgress(ctx, p.Mid))
}

func upProgress(ctx *bm.Context) {
	go func() {
		service.LikeSvc.UpWinterProgress(ctx)
	}()
	ctx.JSON("ok", nil)
}
