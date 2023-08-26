package http

import (
	"encoding/json"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	bubblemdl "go-gateway/app/app-svr/app-feed/admin/model/bubble"
)

func bubbleList(c *bm.Context) {
	var (
		params = c.Request.Form
		pn, ps int
	)
	pn, _ = strconv.Atoi(params.Get("pn"))
	if pn < 1 {
		pn = 1
	}
	ps, _ = strconv.Atoi(params.Get("ps"))
	if ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(bubbleSvc.List(c, pn, ps))
}

func bubblePositionList(c *bm.Context) {
	c.JSON(bubbleSvc.PositionList(c))
}

func bubbleAdd(c *bm.Context) {
	var (
		params    = &bubblemdl.Param{}
		res       = map[string]interface{}{}
		uid       int64
		username  string
		positions []*bubblemdl.ParamPostion
		err       error
	)
	if err = c.Bind(params); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(params.Position), &positions); err != nil {
		res["message"] = "位置信息 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.Icon == "" {
		res["message"] = "图标 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.Desc == "" {
		res["message"] = "推送文案 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.WhiteList == "" {
		res["message"] = "推送用户列表 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, username = managerInfo(c)
	c.JSON(nil, bubbleSvc.Add(c, params, uid, username, positions))
}

func bubbleEdit(c *bm.Context) {
	var (
		params    = &bubblemdl.Param{}
		res       = map[string]interface{}{}
		uid       int64
		username  string
		positions []*bubblemdl.ParamPostion
		err       error
	)
	if err = c.Bind(params); err != nil {
		return
	}
	if params.ID == 0 {
		res["message"] = "ID 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = json.Unmarshal([]byte(params.Position), &positions); err != nil {
		res["message"] = "位置信息 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.Icon == "" {
		res["message"] = "图标 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.Desc == "" {
		res["message"] = "推送文案 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.WhiteList == "" {
		res["message"] = "推送用户列表 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, username = managerInfo(c)
	c.JSON(nil, bubbleSvc.Edit(c, params, uid, username, positions))
}

func bubbleState(c *bm.Context) {
	var (
		params   = &bubblemdl.Param{}
		res      = map[string]interface{}{}
		uid      int64
		username string
		err      error
	)
	if err = c.Bind(params); err != nil {
		return
	}
	if params.ID == 0 {
		res["message"] = "ID 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.State != bubblemdl.StateOnline && params.State != bubblemdl.StateOffline {
		res["message"] = "状态 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, username = managerInfo(c)
	c.JSON(nil, bubbleSvc.BuddleState(c, params, uid, username))
}
