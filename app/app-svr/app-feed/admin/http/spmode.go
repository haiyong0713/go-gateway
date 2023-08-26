package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func searchSpmode(c *bm.Context) {
	req := new(model.SearchReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(spmodeSvc.Search(req))
}

func relieveSpmode(c *bm.Context) {
	req := new(model.RelieveReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userid int64
	if uid, ok := c.Get("uid"); ok {
		userid = uid.(int64)
	}
	c.JSON(nil, spmodeSvc.Relieve(c, req, userid, username))
}

func spmodeLog(c *bm.Context) {
	req := new(model.LogReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(spmodeSvc.Log(req))
}
