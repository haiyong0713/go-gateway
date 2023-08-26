package http

import bm "go-common/library/net/http/blademaster"

func channel(c *bm.Context) {
	arg := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(spcSvc.Channel(c, arg.Mid))
}
