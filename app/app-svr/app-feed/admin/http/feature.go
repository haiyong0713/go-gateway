package http

import (
	"encoding/json"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/feature"
)

func appList(c *bm.Context) {
	cookie := c.Request.Header.Get("Cookie")
	c.JSON(featureSvc.AppList(c, cookie))
}

func appPlat(c *bm.Context) {
	req := new(feature.AppPlatReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(featureSvc.AppPlat(c, req))
}

func saveApp(c *bm.Context) {
	req := new(feature.SaveAppReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	c.JSON(nil, featureSvc.SaveApp(c, int(userID), username, req))
}

func buildList(c *bm.Context) {
	req := new(feature.BuildListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(featureSvc.BuildList(c, req))
}

func saveBuild(c *bm.Context) {
	req := new(feature.SaveBuildReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	var res = map[string]interface{}{}
	if !checkKeyName(req.KeyName, req.TreeID) {
		res["message"] = "key名称格式错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	errMsg, err := featureSvc.SaveBuild(c, int(userID), username, req)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func checkKeyName(keyname string, treeID int) bool {
	if treeID == feature.Common {
		if strings.HasPrefix(keyname, "common.") {
			return true
		}
	} else {
		if strings.HasPrefix(keyname, "service.") {
			return true
		}
	}
	return false
}

func handleBuild(c *bm.Context) {
	req := new(feature.HandleBuildReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	c.JSON(nil, featureSvc.HandleBuild(c, int(userID), username, req))
}

func switchTvList(c *bm.Context) {
	req := new(feature.SwitchTvListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(featureSvc.SwitchTV(c, req))
}

func switchTvSave(c *bm.Context) {
	req := new(feature.SwitchTvSaveReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var res = map[string]interface{}{}
	if req.Brand == "" && req.Chid == "" && req.Model == "" {
		res["message"] = "brand/chid/model不能全空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 系统版本校验
	if req.SysVersion != "" {
		var sysVersion *feature.SysVersion
		if err := json.Unmarshal([]byte(req.SysVersion), &sysVersion); err != nil {
			res["message"] = "安卓版本传参错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if sysVersion.Start > sysVersion.End {
			res["message"] = "安卓版本起始版本大于结束版本"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if req.Config == "" {
		res["message"] = "config不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	errMsg, err := featureSvc.SwitchTvSave(c, req)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func switchTvDel(c *bm.Context) {
	req := new(feature.SwitchTvDelReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(nil, featureSvc.SwitchTvDel(c, req))
}

func businessConfigList(c *bm.Context) {
	req := new(feature.BusinessConfigListReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(featureSvc.BusinessConfigList(c, req))
}

func businessConfigSave(c *bm.Context) {
	req := new(feature.BusinessConfigSaveReq)
	var res = map[string]interface{}{}
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	if !checkKeyName(req.KeyName, req.TreeID) {
		res["message"] = "key前缀错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	errMsg, err := featureSvc.BusinessConfigSave(c, req, userID, username)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func businessConfigAct(c *bm.Context) {
	req := new(feature.BusinessConfigActReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	c.JSON(nil, featureSvc.BusinessConfigAct(c, req, userID, username))
}

/*
	分组实验
*/

func abtestList(c *bm.Context) {
	req := new(feature.ABTestReq)
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(featureSvc.ABTestList(c, req))
}

func abtestSave(c *bm.Context) {
	req := new(feature.ABTestSaveReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	var res = map[string]interface{}{}
	if !checkKeyName(req.KeyName, req.TreeID) {
		res["message"] = "key名称格式错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	errMsg, err := featureSvc.ABTestSave(c, int(userID), username, req)
	if err != nil && errMsg != "" {
		c.JSONMap(map[string]interface{}{"message": errMsg}, err)
		return
	}
	c.JSON(nil, err)
}

func abtestHandle(c *bm.Context) {
	req := new(feature.ABTestHandleReq)
	if err := c.Bind(req); err != nil {
		return
	}
	var username string
	if uname, ok := c.Get("username"); ok {
		username = uname.(string)
	}
	var userID int64
	if uid, ok := c.Get("uid"); ok {
		userID = uid.(int64)
	}
	c.JSON(nil, featureSvc.ABTestHandle(c, int(userID), username, req))
}
