package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func addGuess(c *bm.Context) {
	v := new(model.ParamAdd)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.AddGuess(c, v))
}

func delGuess(c *bm.Context) {
	v := new(model.ParamDel)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.DelGuess(c, v))
}

func listGuess(c *bm.Context) {
	v := new(struct {
		Oid int64 `form:"oid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.ListGuess(c, v.Oid))
}

func resultGuess(c *bm.Context) {
	v := new(model.ParamRes)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ResultGuess(c, v))
}
