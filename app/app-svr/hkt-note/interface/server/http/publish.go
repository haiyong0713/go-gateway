package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/hkt-note/interface/model/article"
)

func publishNoteInfo(c *bm.Context) {
	req := new(article.PubNoteInfoReq)
	err := c.Bind(req)
	if err != nil {
		return
	}
	c.JSON(artSvr.PublishNoteInfo(c, req))
}

func publicListInArc(c *bm.Context) {
	req := new(article.PubListInArcReq)
	err := c.Bind(req)
	if err != nil {
		return
	}
	var midInt int64
	mid, ok := c.Get("mid")
	if ok {
		midInt = mid.(int64)
	}
	c.JSON(artSvr.PubListInArc(c, req, midInt))
}

func publicListInUser(c *bm.Context) {
	req := new(struct {
		Pn int64 `form:"pn" validate:"required"`
		Ps int64 `form:"ps" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(artSvr.PubListInUser(c, req.Pn, req.Ps, mid.(int64)))
}

func publishDel(c *bm.Context) {
	req := new(struct {
		Cvids []int64 `form:"cvids,split" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, artSvr.PublishDel(c, req.Cvids, mid.(int64)))
}
