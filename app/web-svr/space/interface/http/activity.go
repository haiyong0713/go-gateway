package http

import (
	bm "go-common/library/net/http/blademaster"
	v1 "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
)

func activityTab(ctx *bm.Context) {
	req := &v1.UserTabReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(spcSvc.UserTab(ctx, req))
}

func reservation(ctx *bm.Context) {
	req := &struct {
		Vmid int64 `form:"vmid" default:"1" validate:"min=1"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(spcSvc.Reservation(ctx, mid, req.Vmid))
}

func reserve(c *bm.Context) {
	arg := &model.AddReserveReq{}
	if err := c.Bind(arg); err != nil {
		return
	}
	arg.Buvid = c.Request.Header.Get(_headerBuvid)
	midInter, _ := c.Get("mid")
	arg.Mid = midInter.(int64)
	c.JSON(nil, spcSvc.Reserve(c, arg))
}

func reserveCancel(c *bm.Context) {
	var arg = new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(nil, spcSvc.ReserveCancel(c, mid, arg.Sid))
}

func upReserveCancel(c *bm.Context) {
	var arg = new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(nil, spcSvc.UpReserveCancel(c, mid, arg.Sid))
}
