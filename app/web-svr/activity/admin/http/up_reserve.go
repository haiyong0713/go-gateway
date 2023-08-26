package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/up_reserve"
)

func upReserveList(c *bm.Context) {
	arg := new(up_reserve.ParamList)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.UpReserveList(c, arg))
}

func upReserveHang(c *bm.Context) {
	arg := new(up_reserve.ParamHang)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, actSrv.UpReserveHang(c, arg))
}

func upReserveHangLogList(c *bm.Context) {
	arg := new(up_reserve.HangLogListParams)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(actSrv.UpReserveHangLogList(c, arg))
}
