package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func kfcInfo(c *bm.Context) {
	p := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.KfcSvc.KfcInfo(c, p.ID, mid))
}

func kfcUse(c *bm.Context) {
	c.JSON(200, nil)
}

func deliverKfc(c *bm.Context) {
	p := new(struct {
		ID  int64 `form:"id" validate:"min=1"`
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, service.KfcSvc.DeliverKfc(c, p.ID, p.Mid))
}
