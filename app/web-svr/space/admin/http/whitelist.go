package http

import (
	"fmt"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/model"
	"go-gateway/app/web-svr/space/admin/util"
)

func whitelistIndex(c *bm.Context) {
	param := &struct {
		Mid    int64 `form:"mid"`
		Status int   `json:"state" form:"state"`
		Pn     int   `form:"pn" default:"1"`
		Ps     int   `form:"ps" default:"20"`
	}{}
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(spcSvc.WhitelistIndex(param.Mid, param.Pn, param.Ps, param.Status))
}

func whitelistAdd(c *bm.Context) {
	var (
		err        error
		failedList []int64
	)
	res := map[string]interface{}{}
	param := &model.WhitelistReq{}
	if err = c.Bind(param); err != nil {
		return
	}
	param.Username, _ = util.UserInfo(c)
	if param.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}

	if failedList, err = spcSvc.WhitelistAdd(c, param); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if len(failedList) > 0 {
		res["message"] = "无法添加的Uid: " + strings.Replace(strings.Trim(fmt.Sprint(failedList), "[]"), " ", ",", -1)
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
}

func whitelistUp(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := &model.WhitelistAdd{}
	if err = c.Bind(param); err != nil {
		return
	}
	param.Username, _ = util.UserInfo(c)
	if param.Username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if err = spcSvc.WhitelistUp(param); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func whitelistDel(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID   int64 `json:"id" form:"id"`
		Type int   `json:"type" form:"type" default:"0"`
	}{}
	if err = c.Bind(req); err != nil {
		res["message"] = "传参错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	username, _ := util.UserInfo(c)
	if username == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登陆"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID < 0 {
		res["message"] = "id无效"
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = spcSvc.WhitelistDel(req.ID, req.Type); err != nil {
		res["message"] = "删除错误: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		c.Abort()
		return
	}
	c.JSON(nil, nil)
}
