package http

import (
	bm "go-common/library/net/http/blademaster"
)

func sysNotice(c *bm.Context) {
	var (
		err    error
		notice interface{}
	)
	param := struct {
		ID int64 `form:"id" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		c.JSON(nil, err)
		return
	}
	list := spcSvc.SysNotice
	if v, ok := list[param.ID]; ok {
		notice = v
	}
	if notice == nil {
		notice = struct{}{}
	}
	c.JSON(notice, nil)
}
