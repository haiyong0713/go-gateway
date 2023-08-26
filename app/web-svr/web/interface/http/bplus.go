package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	mdl "go-gateway/app/web-svr/web/interface/model"
)

func materialInfo(c *bm.Context) {
	var (
		request = &mdl.MaterialInfoReq{}
	)
	if err := c.Bind(request); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(webSvc.MaterialInfo(c, request))
}
