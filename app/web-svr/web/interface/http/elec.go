package http

import (
	bm "go-common/library/net/http/blademaster"
)

func elecShow(c *bm.Context) {
	var (
		aid int64
		err error
		mid int64
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Mid  int64  `form:"mid" validate:"min=1"` // up mid
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.ElecShow(c, v.Mid, aid, mid, nil))
}
