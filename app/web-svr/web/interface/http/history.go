package http

import bm "go-common/library/net/http/blademaster"

func historyCursor(c *bm.Context) {
	v := new(struct {
		Max      int64  `form:"max"`
		ViewAt   int64  `form:"view_at"`
		Business string `form:"business"`
		Type     string `form:"type"`
		Ps       int32  `form:"ps" default:"20" validate:"min=1,max=30"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(webSvc.HistoryCursor(c, mid, v.Max, v.ViewAt, v.Business, v.Type, v.Ps))
}

func wxHistoryCursor(c *bm.Context) {
	v := new(struct {
		Max      int64  `form:"max"`
		ViewAt   int64  `form:"view_at"`
		Business string `form:"business"`
		Ps       int32  `form:"ps" default:"20" validate:"min=1,max=30"`
		Platform string `form:"platform"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(webSvc.WxHistoryCursor(c, mid, v.Max, v.ViewAt, v.Business, v.Ps, v.Platform))
}
