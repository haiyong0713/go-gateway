package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"
)

func predictionAdd(c *bm.Context) {
	args := make([]*premdl.BatchAdd, 0)
	if err := c.BindWith(&args, binding.JSON); err != nil {
		c.JSON(nil, err)
		return
	}
	if len(args) == 0 || len(args) > 50 {
		return
	}
	c.JSON(nil, preSrv.BatchAdd(c, args))
}

func predSearch(c *bm.Context) {
	arg := new(premdl.PredSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(preSrv.PredSearch(c, arg))
}

func predUp(c *bm.Context) {
	arg := &premdl.PresUp{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, preSrv.PresUp(c, arg))
}

func itemAdd(c *bm.Context) {
	args := make([]*premdl.ItemAdd, 0)
	if err := c.BindWith(&args, binding.JSON); err != nil {
		c.JSON(nil, err)
		return
	}
	if len(args) == 0 || len(args) > 60 {
		return
	}
	c.JSON(nil, preSrv.ItemAdd(c, args))
}

func itemUp(c *bm.Context) {
	args := &premdl.ItemUp{}
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(nil, preSrv.ItemUp(c, args))
}

func itemSearch(c *bm.Context) {
	args := &premdl.ItemSearch{}
	if err := c.Bind(args); err != nil {
		return
	}
	c.JSON(preSrv.ItemSearch(c, args))
}
