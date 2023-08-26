package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/pkg/idsafe/bvid"
)

func bvToAv(c *bm.Context) {
	var (
		err error
		id  int64
	)
	param := &struct {
		BVID string `form:"bvid" validate:"required"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	res := map[string]interface{}{}
	if id, err = bvid.BvToAv(param.BVID); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(id, nil)
}

func avToBv(c *bm.Context) {
	var (
		err error
		id  string
	)
	param := &struct {
		AVID int64 `form:"avid" validate:"required"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	res := map[string]interface{}{}
	if id, err = bvid.AvToBv(param.AVID); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(id, nil)
}
