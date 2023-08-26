package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/fawkes/service/api/app/webcontainer"
)

func AddWhiteList(c *bm.Context) {
	p := new(webcontainer.AddWhiteListReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := s.WebContainerSvr.AddWhiteList(c, p)
	c.JSON(resp, err)
}
