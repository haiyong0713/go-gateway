package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/search"
)

func searchAll(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &search.SearchParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, upItems, args, err := showSvc.Search(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item    []card.Handler     `json:"items"`
		Args    *search.SearchArgs `json:"args"`
		Page    *card.Page         `json:"page"`
		UpItems []*search.UpItem   `json:"up_items,omitempty"`
	}{
		Item: data,
		Args: args,
		Page: &card.Page{
			Pn: pagePn(data, param.Pn),
			Ps: param.Ps,
		},
		UpItems: upItems,
	}, err)
}

func searchWebAll(c *bm.Context) {
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
	param := &search.SearchParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	pgcItem, item, err := showSvc.SearchWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item     []card.Handler `json:"items"`
		PgcItems []card.Handler `json:"pgc_items"`
		Page     *card.Page     `json:"page"`
	}{
		Item:     item,
		PgcItems: pgcItem,
		Page: &card.Page{
			Pn: pagePn(item, param.Pn),
			Ps: param.Ps,
		},
	}, err)
}

func suggest(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &search.SearchSuggestParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, args, err := showSvc.Suggest(c, plat, mid, buvid, param)
	c.JSON(struct {
		Item []*search.SuggestItem `json:"items"`
		Args *search.SearchArgs    `json:"args"`
	}{Item: data, Args: args}, err)
}

func suggestWeb(c *bm.Context) {
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
	param := &search.SearchSuggestParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := showSvc.SuggestWeb(c, model.PlatH5, mid, buvid, param)
	c.JSON(struct {
		Item []*search.SuggestItem `json:"items"`
	}{Item: data}, err)
}

// searchV2 综合搜索 V2
func searchV2(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)

	param := &search.SearchParamV2{}
	if err := c.Bind(param); err != nil {
		return
	}
	param.Mid = mid
	param.Buvid = buvid
	c.JSON(commonSvc.SearchV2(c, param))
}
