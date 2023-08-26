package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/show"
)

func myanime(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &bangumi.MyAnimeParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.MyAnime(c, plat, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{
		Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func bangumiList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &bangumi.ListParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	data, err := showSvc.List(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{
		Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func showPGC(c *bm.Context) {
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
	data, err := showSvc.PGCShow(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item []*show.Item `json:"items"`
	}{Item: data}, err)
}

func myanimeWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &bangumi.MyAnimeParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.MyAnimeWeb(c, model.PlatH5, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{
		Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func bangumiListWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &bangumi.ListParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	data, err := showSvc.BangumiListWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{
		Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func showPGCWeb(c *bm.Context) {
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
	data, err := showSvc.PGCShowWeb(c, mid, model.PlatH5, buvid, param)
	c.JSON(struct {
		Item []*show.ItemWeb `json:"items"`
	}{Item: data}, err)
}
