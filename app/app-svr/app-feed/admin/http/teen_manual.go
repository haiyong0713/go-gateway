package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/teen_manual"
)

func searchTeenManual(c *bm.Context) {
	req := new(model.SearchTeenManualReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(teenManualSvc.Search(c, req))
}

func openTeenManual(c *bm.Context) {
	req := new(model.OpenReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	c.JSON(nil, teenManualSvc.Open(c, req, username))
}

func quitTeenManual(c *bm.Context) {
	req := new(model.QuitReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	c.JSON(nil, teenManualSvc.Quit(c, req, username))
}

func teenManualLog(c *bm.Context) {
	req := new(model.TeenManualLogReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(teenManualSvc.Log(c, req))
}
