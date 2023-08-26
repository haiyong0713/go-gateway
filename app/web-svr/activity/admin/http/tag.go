package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func tagStatus(c *bm.Context) {
	v := new(struct {
		Name string `form:"tag_name" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(actSrv.TagIsActivity(c, v.Name))
}

func tagToActivity(c *bm.Context) {
	v := new(struct {
		Name  string `form:"tag_name" validate:"required"`
		Stime int64  `form:"start_time" validate:"required"`
		Etime int64  `form:"end_time" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(actSrv.TagToActivity(c, v.Name, v.Stime, v.Etime, userName))
}

func tagToNormal(c *bm.Context) {
	v := new(struct {
		Name string `form:"tag_name" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(actSrv.TagToNormal(c, v.Name, userName))
}
