package http

import (
	bm "go-common/library/net/http/blademaster"
)

// abArcSearch .
func abArcSearch(c *bm.Context) {
	var (
		mid int64
		err error
	)
	v := new(struct {
		VMid int64 `form:"mid" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.AbtestVideoSearch(c, mid, v.VMid), nil)
}
