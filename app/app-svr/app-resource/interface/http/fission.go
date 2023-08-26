package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	fissiMdl "go-gateway/app/app-svr/app-resource/interface/model/fission"
)

func checkNew(c *bm.Context) {
	var header = c.Request.Header
	buvid := header.Get("Buvid")
	if buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	p := new(fissiMdl.ParamCheck)
	if err := c.Bind(p); err != nil {
		return
	}
	midI, _ := c.Get("mid")
	p.Mid = midI.(int64)
	p.Buvid = buvid
	c.JSON(fissionSvc.CheckNew(c, p))
}

func checkDevice(c *bm.Context) {
	var header = c.Request.Header
	buvid := header.Get("Buvid")
	if buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	p := new(fissiMdl.ParamCheck)
	if err := c.Bind(p); err != nil {
		return
	}
	p.Buvid = buvid
	c.JSON(fissionSvc.CheckDevice(c, p))
}
