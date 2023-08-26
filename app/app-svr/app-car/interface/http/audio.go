package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/audio"
	"go-gateway/app/app-svr/app-car/interface/model/card"
)

func audioShow(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &audio.ShowAudioParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.AudioShow(c, mid, plat, buvid, param)
	c.JSON(data, err)
}

func audioFeed(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &audio.ShowAudioParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.AudioFeed(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
	}{Item: data}, err)
}

func audioChannel(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &audio.ChannelAudioParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.AudioChannel(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{
		Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
		},
	}, err)
}

func reportPlayAction(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &audio.ReportPlayParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	err := showSvc.ReportPlayAction(c, mid, buvid, param)
	c.JSON(nil, err)
}
