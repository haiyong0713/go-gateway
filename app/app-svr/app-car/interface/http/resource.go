package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model/banner"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	"go-gateway/app/app-svr/app-car/interface/model/tab"
)

func showTab(c *bm.Context) {
	param := &show.ShowParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(struct {
		Item []*tab.Tab `json:"items"`
	}{Item: resSvc.ShowTab(c, param)}, nil)
}

func showBanner(c *bm.Context) {
	param := &show.ShowParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(struct {
		Item []*banner.Banner `json:"items"`
	}{Item: resSvc.Banner(c, param)}, nil)
}

func showTabWeb(c *bm.Context) {
	param := &show.ShowParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(struct {
		Item []*tab.TabWeb `json:"items"`
	}{Item: resSvc.ShowTabWeb(c, param)}, nil)
}
