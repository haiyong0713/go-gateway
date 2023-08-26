package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-admin/admin/model/aids"
)

// saveAids save
func saveAids(c *bm.Context) {
	v := &aids.Param{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, aidsSvc.Save(c, v.Aids))
}
