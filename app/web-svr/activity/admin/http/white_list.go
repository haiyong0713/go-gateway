package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/activity/admin/model"
)

func addWhiteList(c *bm.Context) {
	req := new(model.AddWhiteListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if val, ok := c.Get("username"); ok {
		username = val.(string)
	}
	var uid int64
	if val, ok := c.Get("uid"); ok {
		uid = val.(int64)
	}
	c.JSON(nil, actSrv.AddWhiteList(c, req, uid, username, model.FromActivity))
}

func addWhiteListOuter(c *bm.Context) {
	req := new(model.AddWhiteListOuterReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, actSrv.AddWhiteList(c, &model.AddWhiteListReq{Mid: req.Mid}, req.Uid, req.Uname, req.From))
}

func deleteWhiteList(c *bm.Context) {
	req := new(model.DeleteWhiteListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if val, ok := c.Get("username"); ok {
		username = val.(string)
	}
	var uid int64
	if val, ok := c.Get("uid"); ok {
		uid = val.(int64)
	}
	c.JSON(nil, actSrv.DeleteWhiteList(c, req, uid, username))
}

func whiteList(c *bm.Context) {
	req := new(model.GetWhiteListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(actSrv.WhiteList(c, req))
}
