package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func indexPopUp(c *bm.Context) {
	v := new(struct {
		MobiApp string `form:"mobi_app" validate:"required"`
		Build   int32  `form:"build" validate:"min=1"`
		Device  string `form:"device"`
		Buvid   string `form:"-"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	v.Buvid = c.Request.Header.Get("Buvid")
	if v.Buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	teenagersMode, _ := strconv.Atoi(c.Request.Form.Get("teenagers_mode"))
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(showSvc.IndexPopUp(c, mid, v.Buvid, v.MobiApp, v.Device, v.Build, teenagersMode), nil)
}
