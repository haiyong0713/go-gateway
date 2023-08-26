package http

import (
	bm "go-common/library/net/http/blademaster"
	ac "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/actionlog"
)

func actionLog(c *bm.Context) {
	params := &ac.Log{}

	if err := c.Bind(params); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.LogAction(c, params))
}
