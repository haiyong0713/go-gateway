package http

import (
	bm "go-common/library/net/http/blademaster"
)

func qq(c *bm.Context) {
	c.JSON(200, nil)
}

func news(c *bm.Context) {
	c.JSON(200, nil)
}
