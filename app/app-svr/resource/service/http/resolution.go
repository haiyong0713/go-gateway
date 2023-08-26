package http

import (
	bm "go-common/library/net/http/blademaster"
)

func limitFree(c *bm.Context) {
	c.JSON(resSvc.FetchLimitFreeOnline())
}
