package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func newstarJoin(ctx *bm.Context) {
	arg := new(struct {
		ActivityUID string `form:"activity_uid" validate:"required"`
		InviterMid  int64  `form:"inviter_mid" json:"inviter_mid"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.NewstarSvc.JoinNewstar(ctx, arg.ActivityUID, mid, arg.InviterMid))
}

func newstarCreation(ctx *bm.Context) {
	arg := new(struct {
		ActivityUID string `form:"activity_uid" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.NewstarSvc.NewstarCreation(ctx, arg.ActivityUID, mid))
}

func newstarInvite(ctx *bm.Context) {
	arg := new(struct {
		ActivityUID string `form:"activity_uid" validate:"required"`
		Pn          int    `form:"pn" default:"1" validate:"min=1"`
		Ps          int    `form:"ps" default:"2" validate:"min=1,max=5"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(service.NewstarSvc.NewstarInvite(ctx, arg.ActivityUID, mid, arg.Pn, arg.Ps))
}
