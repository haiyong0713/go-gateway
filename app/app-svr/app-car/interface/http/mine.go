package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	minemodel "go-gateway/app/app-svr/app-car/interface/model/mine"
	"go-gateway/app/app-svr/app-car/interface/model/show"
)

func mine(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	buvid := header.Get(_headerBuvid)
	param := &minemodel.MineParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	plat := model.Plat(param.MobiApp, param.Device)
	data, acc, err := showSvc.Mine(c, mid, plat, buvid, param)
	c.JSON(struct {
		Item    []*show.Item    `json:"items"`
		Account *minemodel.Mine `json:"account"`
	}{Item: data, Account: acc}, err)
}

func mineWeb(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := c.Request.Header
	referer := c.Request.Referer()
	buvid := header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	param := &minemodel.MineParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	cookie := c.Request.Header.Get(_headerCookie)
	data, acc, err := showSvc.MineWeb(c, mid, model.PlatH5, buvid, cookie, referer, param)
	c.JSON(struct {
		Item    []*show.ItemWeb    `json:"items"`
		Account *minemodel.MineWeb `json:"account"`
	}{Item: data, Account: acc}, err)
}

func mineV2Tabs(c *bm.Context) {
	var (
		req = new(commonmdl.MineTabsReq)
		err error
	)
	if err = c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(commonSvc.MineTabs(c, req), nil)
}
