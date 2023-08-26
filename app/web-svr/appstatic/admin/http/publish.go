package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func publish(c *bm.Context) {
	arg := new(struct {
		ResID int `form:"res_id" validate:"required"`
	})
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(apsSvc.Publish(c, arg.ResID))
}

func push(c *bm.Context) {
	var (
		err  error
		name string
	)
	res := map[string]interface{}{}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if err = apsSvc.Push(c, name); err != nil {
		res["message"] = "推送失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
