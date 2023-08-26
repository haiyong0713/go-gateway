package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/banner"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	"go-gateway/app/app-svr/app-car/interface/model/show"
)

func showIndex(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &show.ShowParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, config, err := showSvc.Show(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item   []*show.Item `json:"items"`
		Config *show.Config `json:"config"`
	}{Item: data, Config: config}, err)
}

func showWebIndex(c *bm.Context) {
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
	param := &show.ShowParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.ShowWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []*show.ItemWeb `json:"items"`
	}{Item: data}, err)
}

func fmList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &popular.PopularParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, page, err := showSvc.FmList(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func feedBanner(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &banner.ShowBannerParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data := showSvc.Banner(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{Item: data}, nil)
}

func feedList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &banner.ShowBannerParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data := showSvc.Feed(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{Item: data}, nil)
}
