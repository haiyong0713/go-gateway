package http

import (
	bm "go-common/library/net/http/blademaster"
)

func onlineInfo(c *bm.Context) {
	c.JSON(webSvc.OnlineArchiveCount(c), nil)
}

func onlineList(c *bm.Context) {
	c.JSON(webSvc.OnlineList(c))
}

func onlineTotal(c *bm.Context) {
	v := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	total, err := webSvc.OnlineTotal(c, v.Token)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"online_total": total}, nil)
}
