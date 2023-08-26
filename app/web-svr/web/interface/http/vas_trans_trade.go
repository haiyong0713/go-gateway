package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func tradeCreate(c *bm.Context) {
	v := &model.TradeCreateReq{}
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		v.Mid = midInter.(int64)
	}
	c.JSON(webSvc.TradeCreate(c, v))
}
