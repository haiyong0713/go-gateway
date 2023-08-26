package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func esportAdd(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		matchID, mid int64
		err          error
	)
	if matchID, err = strconv.ParseInt(params.Get("match_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err = relSvr.EsportAdd(c, mid, matchID); err != nil {
		c.JSON(nil, err)
		return
	}
	res["message"] = "订阅成功，直播开始时通知你"
	c.JSONMap(res, ecode.OK)
}

func esportCancel(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		matchID, mid int64
		err          error
	)
	if matchID, err = strconv.ParseInt(params.Get("match_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err = relSvr.EsportCancel(c, mid, matchID); err != nil {
		c.JSON(nil, err)
		return
	}
	res["message"] = "订阅已取消"
	c.JSONMap(res, ecode.OK)
}
