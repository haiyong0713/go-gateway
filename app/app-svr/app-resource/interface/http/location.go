package http

import bm "go-common/library/net/http/blademaster"

func IpInfo(c *bm.Context) {
	c.JSON(locationSvc.Info(c))
}
