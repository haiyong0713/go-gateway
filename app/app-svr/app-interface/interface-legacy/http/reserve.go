package http

import (
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
)

func reserve(c *bm.Context) {
	arg := &space.AddReserveReq{}
	if err := c.Bind(arg); err != nil {
		return
	}
	arg.Buvid = c.Request.Header.Get(_headerBuvid)
	midInter, _ := c.Get("mid")
	arg.Mid = midInter.(int64)
	c.JSON(spaceSvr.Reserve(c, arg))
}

func reserveCancel(c *bm.Context) {
	var arg = new(struct {
		Sid          int64 `form:"sid" validate:"min=1"`
		ReserveTotal int64 `form:"reserve_total"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(spaceSvr.ReserveCancel(c, mid, arg.Sid, arg.ReserveTotal))
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
	c.JSON(nil, spaceSvr.UpReserveCancel(c, mid, arg.Sid))
}

func reserveShareInfo(c *bm.Context) {
	args := &space.ReserveShareInfoReq{}
	if err := c.Bind(args); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	dev, ok := device.FromContext(c)
	if !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(spaceSvr.GetReserveDynShareContent(c, mid, args, dev))
}
