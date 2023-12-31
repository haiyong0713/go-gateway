package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func addTree(c *bm.Context) {
	v := new(model.Tree)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Pid > 0 && v.RootID == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(esSvc.AddTree(c, v))
}

func editTree(c *bm.Context) {
	v := new(model.TreeEditParam)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.EditTree(c, v))
}

func delTree(c *bm.Context) {
	v := new(model.TreeDelParam)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.DelTree(c, v))
}

func listTree(c *bm.Context) {
	v := new(model.TreeListParam)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.TreeList(c, v))
}
