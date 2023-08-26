package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func nav(c *bm.Context) {
	var (
		rawMid interface{}
		ok     bool
	)

	if rawMid, ok = c.Get("mid"); !ok {
		// NOTE NoLogin here only for web
		c.JSON(model.FailedNavResp{}, ecode.NoLogin)
		return
	}
	mid := rawMid.(int64)
	c.JSON(webSvc.Nav(c, mid, c.Request.Header.Get("Cookie")))
}

func navStat(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(webSvc.NavStat(c, mid), nil)
}
