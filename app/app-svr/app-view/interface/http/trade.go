package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-view/interface/model/trade"
)

func productInfo(c *bm.Context) {
	req := &trade.ProductInfoReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(viewSvr.TradeProductInfo(c, req))
}

func orderState(c *bm.Context) {
	req := &trade.OrderStateReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(viewSvr.TradeOrderState(c, mid, req))
}

func orderCreate(c *bm.Context) {
	req := &trade.OrderCreateReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	mid := midInter.(int64)
	c.JSON(viewSvr.TradeOrderCreate(c, mid, req))
}
