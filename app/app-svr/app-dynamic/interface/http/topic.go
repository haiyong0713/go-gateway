package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	topicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/topic"
)

func square(c *bm.Context) {
	var req = new(topicmdl.SquareReq)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var (
		mid   int64
		buvid string
	)
	buvid = c.Request.Header.Get("Buvid")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(topicSvc.Square(c, mid, buvid, req))
}

func hotList(c *bm.Context) {
	var req = new(topicmdl.HotListReq)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var (
		mid   int64
		buvid string
	)
	buvid = c.Request.Header.Get("Buvid")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(topicSvc.HotList(c, mid, buvid, req))
}

func subscribeSave(c *bm.Context) {
	var req = new(topicmdl.SubscribeSaveReq)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(nil, topicSvc.SubscribeSave(c, mid, req))
}
