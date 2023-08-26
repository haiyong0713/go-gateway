package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/model/reply"
)

func replyList(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &reply.ReplyParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := replySvc.Replys(c, mid, param)
	c.JSON(data, err)
}

func replyChild(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	param := &reply.ReplyParam{}
	// get params
	if err := c.Bind(param); err != nil {
		return
	}
	data, err := replySvc.ReplyChild(c, mid, param)
	c.JSON(data, err)
}
