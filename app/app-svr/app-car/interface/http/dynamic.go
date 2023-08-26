package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
)

func video(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &dynamic.DynamicParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Offset != "" {
		param.RefreshType = 1
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, page, err := showSvc.DynVideo(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func videoWeb(c *bm.Context) {
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
	param := &dynamic.DynamicParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	if param.Offset != "" {
		param.RefreshType = 1
	}
	data, page, err := showSvc.DynVideoWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func dynamicV2(c *bm.Context) {
	var (
		req = new(commonmdl.DynamicReq)
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
	c.JSON(commonSvc.Dynamic(c, req, mid, buvid))
}
