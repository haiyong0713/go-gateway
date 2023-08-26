package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/region"
)

func regionList(c *bm.Context) {
	param := &region.RegionParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, err := showSvc.RegionList(c, plat, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func regionListWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &region.RegionParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.RegionListWeb(c, model.PlatH5, mid, param)
	c.JSON(struct {
		Item []card.Handler `json:"items"`
		Page *card.Page     `json:"page"`
	}{Item: data,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}
