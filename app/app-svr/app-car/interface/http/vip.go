package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/vip"
)

func addVip(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &vip.VipParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if mid == 0 || buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	path := c.Request.URL.Path
	ua := c.Request.UserAgent()
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.AddVip(c, plat, mid, buvid, path, ua, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{Item: data}, err)
}

func codeOpen(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &vip.CodeOpenParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if mid == 0 || buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	path := c.Request.URL.Path
	ua := c.Request.UserAgent()
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.CodeOpen(c, plat, mid, buvid, path, ua, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{Item: data}, err)
}
