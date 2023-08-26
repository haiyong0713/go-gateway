package http

import (
	bm "go-common/library/net/http/blademaster"
	kfcmdl "go-gateway/app/web-svr/activity/admin/model/kfc"
)

func kfcList(c *bm.Context) {
	arg := new(kfcmdl.ListParams)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(kfcSrv.List(c, arg))
}
