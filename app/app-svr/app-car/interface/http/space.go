package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/space"
)

func spaceArchive(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &space.SpaceParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Vmid == 0 && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if param.Vmid == 0 {
		param.Vmid = mid
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.Space(c, mid, plat, buvid, param)
	data.Page = &card.Page{
		Pn: pagePn(data.Items, param.Pn),
		Ps: param.Ps,
	}
	c.JSON(data, err)
}

func spaceArchiveWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	param := &space.SpaceParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Vmid == 0 && mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if param.Vmid == 0 {
		param.Vmid = mid
	}
	data, err := showSvc.SpaceWeb(c, mid, model.PlatH5, buvid, param)
	data.Page = &card.Page{
		Pn: pagePn(data.Items, param.Pn),
		Ps: param.Ps,
	}
	c.JSON(data, err)
}

// spaceV2 up主空间页 V2
func spaceV2(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)

	param := &space.SpaceParamV2{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	c.JSON(commonSvc.SpaceV2(c, param))
}
