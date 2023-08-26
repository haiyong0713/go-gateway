package http

import (
	bm "go-common/library/net/http/blademaster"
	bl "go-gateway/app/app-svr/app-interface/interface-legacy/model/bili_link"
)

func biliLinkReport(c *bm.Context) {
	params := new(bl.BiliLinkReport)
	if err := c.Bind(params); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		params.Mid = midInter.(int64)
	}
	c.JSON(nil, accSvr.BiliLinkReport(c, params))
}
