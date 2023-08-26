package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func appLaserList(c *bm.Context) {
	var (
		params                                     = c.Request.Form
		res                                        = map[string]interface{}{}
		appKey, platform, buvid, logDate, operator string
		mid, taskID                                int64
		status, pn, ps                             int
		err                                        error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	platform = params.Get("platform")
	buvid = params.Get("buvid")
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	taskID, _ = strconv.ParseInt(params.Get("task_id"), 10, 64)
	status, _ = strconv.Atoi(params.Get("status"))
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 || ps > 100 {
		ps = 100
	}
	logDate = params.Get("log_date")
	operator = params.Get("operator")
	c.JSON(s.LaserSvr.AppLaserList(c, appKey, platform, buvid, logDate, operator, mid, taskID, status, pn, ps))
}

func appLaserAdd(c *bm.Context) {
	var (
		params                                        = c.Request.Form
		res                                           = map[string]interface{}{}
		appKey, platform, buvid, logDate, description string
		mid                                           int64
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	buvid = params.Get("buvid")
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	if mid == 0 && buvid == "" {
		res["message"] = "mid或buvid至少填一项"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	logDate = params.Get("log_date")
	platform = params.Get("platform")
	description = params.Get("description")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.LaserSvr.AppLaserAdd(c, appKey, platform, buvid, logDate, userName, description, mid))
}

func appLaserDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		taskID int64
		err    error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		res["message"] = "task_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserDel(c, taskID))
}

func appLaserActiveList(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		appKey, buvid string
		mid, laserId  int64
		pn, ps        int
		err           error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	buvid = params.Get("buvid")
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	laserId, _ = strconv.ParseInt(params.Get("laser_id"), 10, 64)
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(s.LaserSvr.AppLaserActiveList(c, appKey, buvid, mid, laserId, pn, ps))
}

func appLaserCmdList(c *bm.Context) {
	var (
		params                                    = c.Request.Form
		res                                       = map[string]interface{}{}
		appKey, platform, buvid, action, operator string
		mid, taskID                               int64
		status, pn, ps                            int
		err                                       error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	platform = params.Get("platform")
	buvid = params.Get("buvid")
	action = params.Get("action")
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	taskID, _ = strconv.ParseInt(params.Get("task_id"), 10, 64)
	status, _ = strconv.Atoi(params.Get("status"))
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 || ps > 100 {
		ps = 100
	}
	operator = params.Get("operator")
	c.JSON(s.LaserSvr.AppLaserCmdList(c, appKey, platform, buvid, action, operator, mid, taskID, status, pn, ps))
}

func appLaserCmdAdd(c *bm.Context) {
	var (
		params                                                  = c.Request.Form
		res                                                     = map[string]interface{}{}
		appKey, buvid, action, description, paramsStr, operator string
		mid                                                     int64
		err                                                     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if action = params.Get("action"); action == "" {
		res["message"] = "action 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	description = params.Get("description")
	buvid = params.Get("buvid")
	midStr := params.Get("mid")
	if buvid == "" && midStr == "" {
		res["message"] = "buvid 或 mid 不能同时为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if midStr != "" {
		if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil {
			res["message"] = "mid 错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	paramsStr = params.Get("params")
	username, _ := c.Get("username")
	operator, ok := username.(string)
	if !ok || operator == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if err = s.LaserSvr.AppLaserCmdAdd(c, appKey, buvid, action, description, paramsStr, operator, mid); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, err)
		return
	}
	c.JSON(nil, nil)
}

func appLaserCmdDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		taskID int64
		err    error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		res["message"] = "task_id 有误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserCmdDel(c, taskID))
}

func appLaserCmdActionAdd(c *bm.Context) {
	var (
		params                                            = c.Request.Form
		res                                               = map[string]interface{}{}
		name, platform, paramsJSON, operator, description string
	)
	if name = params.Get("name"); name == "" {
		res["message"] = "name 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	platform = params.Get("platform")
	paramsJSON = params.Get("params")
	username, _ := c.Get("username")
	operator, ok := username.(string)
	if !ok || operator == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserCmdActionAdd(c, name, platform, paramsJSON, operator, description))
}

func appLaserCmdActionUpdate(c *bm.Context) {
	var (
		params                                            = c.Request.Form
		res                                               = map[string]interface{}{}
		name, platform, paramsJSON, operator, description string
		id                                                int64
		err                                               error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	platform = params.Get("platform")
	paramsJSON = params.Get("params")
	username, _ := c.Get("username")
	operator, ok := username.(string)
	if !ok || operator == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserCmdActionUpdate(c, id, name, platform, paramsJSON, operator, description))
}

func appLaserCmdActionDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		id     int64
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserCmdActionDel(c, id))
}

func appLaserCmdActionList(c *bm.Context) {
	params := c.Request.Form
	name := params.Get("name")
	platform := params.Get("platform")
	c.JSON(s.LaserSvr.LaserCmdActionList(c, name, platform))
}

func laserUser(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, operator string
		mid, startTime   int64
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mid, err = strconv.ParseInt(params.Get("mid"), 10, 64); err != nil {
		res["message"] = "mid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		res["message"] = "start_time异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.LaserSvr.LaserUser(c, appKey, operator, mid, startTime))
}

//
//func appLaserParseStatusUpdate(c *bm.Context) {
//	var (
//		params                      = c.Request.Form
//		res                         = map[string]interface{}{}
//		laserID                     int64
//		status                      int
//		laserType, operator, appKey string
//		err                         error
//	)
//	if laserID, err = strconv.ParseInt(params.Get("laser_id"), 10, 64); err != nil {
//		res["message"] = "laserID异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if status, err = strconv.Atoi(params.Get("status")); err != nil {
//		res["message"] = "status异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if laserType = params.Get("type"); laserType == "" {
//		res["message"] = "type不能为空"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if appKey = params.Get("app_key"); appKey == "" {
//		res["message"] = "appKey不能为空"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	operator = params.Get("operator")
//	c.JSON(nil, s.LaserSvr.AppLaserParseStatusUpdate(c, status, laserID, laserType, operator, appKey))
//}
//
//func appLaserPendingList(c *bm.Context) {
//	c.JSON(s.LaserSvr.AppLaserPendingList(c))
//}
