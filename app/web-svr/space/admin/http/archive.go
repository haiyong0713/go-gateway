package http

import bm "go-common/library/net/http/blademaster"

func topArc(c *bm.Context) {
	arg := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(spcSvc.TopArc(c, arg.Mid))
}

func masterpiece(c *bm.Context) {
	arg := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(spcSvc.Masterpiece(c, arg.Mid))
}
