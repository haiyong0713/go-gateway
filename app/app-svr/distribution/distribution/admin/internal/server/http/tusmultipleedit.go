package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func Overview(c *bm.Context) {
	c.JSON(svc.Overview(c))
}

func Performance(c *bm.Context) {
	params := &struct {
		Mid       int64  `form:"mid" validate:"required"`
		FieldName string `form:"field_name" validate:"required"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(svc.Performance(c, params.FieldName, params.Mid))
}

func PerformanceSave(c *bm.Context) {
	params := &struct {
		Mids      []int64 `form:"mids" validate:"required"`
		FieldName string  `form:"field_name" validate:"required"`
		TusValue  string  `form:"tus_value" validate:"required"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	const maxMidsLen = 50
	if len(params.Mids) > maxMidsLen {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "mid长度不要超过50~"))
		return
	}
	c.JSON(nil, svc.PerformanceSave(c, params.FieldName, params.TusValue, params.Mids))
}
