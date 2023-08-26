package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"
)

func drawImgTagRPCSearchAll(c *bm.Context) {
	p := new(model.SearchAllReq)
	if err := c.Bind(p); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(drawSvc.SearchAll(c, p))
}

func drawImgTagRPCSearchUsers(c *bm.Context) {
	p := new(model.SearchUsersReq)
	if err := c.Bind(p); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(drawSvc.SearchUsers(c, p))
}

func drawImgTagRPCSearchTopics(c *bm.Context) {
	p := new(model.SearchTopicsReq)
	if err := c.Bind(p); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(drawSvc.SearchTopics(c, p))
}

func drawImgTagRPCSearchLocations(c *bm.Context) {
	p := new(model.SearchLocationsReq)
	if err := c.Bind(p); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(drawSvc.SearchLocations(c, p))
}

func drawImgTagRPCSearchItems(c *bm.Context) {
	p := new(model.SearchItemsReq)
	if err := c.Bind(p); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(drawSvc.SearchItems(c, p))
}
