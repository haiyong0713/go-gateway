package http

import bm "go-common/library/net/http/blademaster"

func cusCenter(c *bm.Context) {
	c.JSON(srvWeb.CusCenter(c))
}
