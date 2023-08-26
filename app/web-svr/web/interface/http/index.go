package http

import (
	bm "go-common/library/net/http/blademaster"
)

func indexSort(c *bm.Context) {
	v := new(struct {
		Version int64 `form:"version" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	c.JSON(webSvc.IndexSort(c, mid, v.Version))
}

func indexSet(c *bm.Context) {
	v := new(struct {
		Settings string `form:"settings" validate:"required"`
		Version  int64  `form:"version" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	c.JSON("", webSvc.IndexSortSet(c, mid, v.Settings, v.Version))
}
