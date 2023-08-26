package http

import (
	bm "go-common/library/net/http/blademaster"
	opmdl "go-gateway/app/web-svr/web-show/interface/model/operation"
)

// notice
func notice(c *bm.Context) {
	arg := new(opmdl.ArgOp)
	if err := c.Bind(arg); err != nil {
		return
	}
	notice := opSvc.Notice(c, arg)
	c.JSON(notice, nil)
}

// promote
func promote(c *bm.Context) {
	arg := new(opmdl.ArgPromote)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(opSvc.Promote(c, arg))
}
