package http

import (
	bm "go-common/library/net/http/blademaster"

	match "go-gateway/app/web-svr/native-page/interface/model/like"
)

func clearCache(c *bm.Context) {
	p := new(match.ParamMsg)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(nil, matchSvc.ClearCache(c, p.Msg))
}
