package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
)

const (
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
)

func popularIndex(c *bm.Context) {
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
	data, page, err := showSvc.Index(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}

func mediaPopular(c *bm.Context) {
	param := &popular.MediaPopularParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.MediaPopular(c, param)
	c.JSON(struct {
		Item []*card.MediaItem `json:"items"`
	}{Item: data}, err)
}

func popularIndexWeb(c *bm.Context) {
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
	param := &popular.PopularParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, page, err := showSvc.PopularListWeb(c, mid, model.PlatH5, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data, Page: page}, err)
}
