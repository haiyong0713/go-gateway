package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-car/interface/model"

	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
)

func payInfo(c *bm.Context) {
	var (
		req = new(commonmdl.PayInfoReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
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
	c.JSON(commonSvc.PayInfo(c, req, mid, buvid), nil)
}

func payResult(c *bm.Context) {
	var (
		req = new(commonmdl.PayStateReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
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
	cookie := c.Request.Header.Get(_headerCookie)
	referer := c.Request.Referer()
	if req.Ptype == 2 && req.SeasonId == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if req.DeviceInfo.MobiApp == "" {
		req.DeviceInfo.MobiApp = model.AndroidBilithings
	}
	if req.DeviceInfo.Platform == "" {
		req.DeviceInfo.Platform = "android"
	}
	c.JSON(commonSvc.PayState(c, req, mid, buvid, cookie, referer))
}
