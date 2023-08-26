package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-view/interface/model"
)

func dots(c *bm.Context) {
	var params = &struct {
		Aid     int64  `form:"aid" validate:"min=1"`
		MobiApp string `form:"mobi_app"`
		Device  string `form:"device"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	plat := model.PlatNew(params.MobiApp, params.Device)
	c.JSON(viewSvr.Dots(c, params.Aid, mid, plat))
}
