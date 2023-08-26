package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

// appMogulLogList
func appMogulLogList(c *bm.Context) {
	param := &manager.AppMogulLogParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(searchSvc.AppMogulLogList(c, param))
}
