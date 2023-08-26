package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
)

func crashIndexList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		matchOption *apm.MatchOption
		column      string
		eventId     int64
		err         error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.CrashIndexList(c, column, eventId, matchOption))
}

func crashInfoList(c *bm.Context) {
	var (
		params           = c.Request.Form
		isLaser, eventId int64
		matchOption      *apm.MatchOption
		res              = map[string]interface{}{}
		err              error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.ErrorStackHashWithoutUseless == "" {
		res["message"] = "error_stack_hash_without_useless不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	isLaserConv := params.Get("is_laser")
	if isLaserConv != "" {
		if isLaser, err = strconv.ParseInt(isLaserConv, 10, 64); err != nil {
			res["message"] = "is_laser异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.CrashInfoList(c, isLaser, eventId, matchOption))
}

func jankInfoList(c *bm.Context) {
	var (
		matchOption *apm.MatchOption
		res         = map[string]interface{}{}
		err         error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.AnalyseJankStackHash == "" {
		res["message"] = "analyse_jank_stack_hash不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.JankInfoList(c, matchOption))
}

func jankIndexList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		matchOption *apm.MatchOption
		column      string
		eventId     int64
		err         error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.OrderBy == "" {
		matchOption.OrderBy = "count() DESC"
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	column = params.Get("column")
	c.JSON(s.BuglySvr.JankIndexList(c, column, eventId, matchOption))
}

func updateCrashIndex(c *bm.Context) {
	var (
		params                                                  = c.Request.Form
		res                                                     = map[string]interface{}{}
		solveStatus, solveVersionCode                           int
		eventId                                                 int64
		solveOperator, errorStackHash, solveDescription, appKey string
		err                                                     error
	)
	if errorStackHash = params.Get("error_stack_hash_without_useless"); errorStackHash == "" {
		res["message"] = "error_stack_hash_without_useless异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveOperator = params.Get("solve_operator"); solveOperator == "" {
		res["message"] = "solve_operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveVersionCode, err = strconv.Atoi(params.Get("solve_version_code")); err != nil {
		res["message"] = "solve_version_code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveStatus, err = strconv.Atoi(params.Get("solve_status")); err != nil {
		res["message"] = "solve_status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	solveDescription = params.Get("solve_description")
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BuglySvr.UpdateCrashIndex(c, errorStackHash, appKey, solveOperator, solveDescription, solveVersionCode, solveStatus, eventId))
}

func updateJankIndex(c *bm.Context) {
	var (
		params                                                        = c.Request.Form
		res                                                           = map[string]interface{}{}
		solveStatus, solveVersionCode                                 int
		solveOperator, analyseJankStackHash, solveDescription, appKey string
		err                                                           error
	)
	if analyseJankStackHash = params.Get("analyse_jank_stack_hash"); analyseJankStackHash == "" {
		res["message"] = "analyse_jank_stack_hash异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveOperator = params.Get("solve_operator"); solveOperator == "" {
		res["message"] = "solve_operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveVersionCode, err = strconv.Atoi(params.Get("solve_version_code")); err != nil {
		res["message"] = "solve_version_code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveStatus, err = strconv.Atoi(params.Get("solve_status")); err != nil {
		res["message"] = "solve_status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	solveDescription = params.Get("solve_description")
	c.JSON(nil, s.BuglySvr.UpdateJankIndex(c, analyseJankStackHash, appKey, solveOperator, solveDescription, solveVersionCode, solveStatus))
}

func updateOOMIndex(c *bm.Context) {
	var (
		params                                        = c.Request.Form
		res                                           = map[string]interface{}{}
		solveStatus, solveVersionCode                 int
		solveOperator, hash, solveDescription, appKey string
		err                                           error
	)
	if hash = params.Get("hash"); hash == "" {
		res["message"] = "hash异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveOperator = params.Get("solve_operator"); solveOperator == "" {
		res["message"] = "solve_operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveVersionCode, err = strconv.Atoi(params.Get("solve_version_code")); err != nil {
		res["message"] = "solve_version_code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveStatus, err = strconv.Atoi(params.Get("solve_status")); err != nil {
		res["message"] = "solve_status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	solveDescription = params.Get("solve_description")
	c.JSON(nil, s.BuglySvr.UpdateOOMIndex(c, hash, appKey, solveOperator, solveDescription, solveVersionCode, solveStatus))
}

func crashLaserRelationAdd(c *bm.Context) {
	var (
		params                       = c.Request.Form
		res                          = map[string]interface{}{}
		laserId                      int64
		errorStackHashWithoutUseless string
		err                          error
	)
	if laserId, err = strconv.ParseInt(params.Get("laser_id"), 10, 64); err != nil {
		res["message"] = "laser_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if errorStackHashWithoutUseless = params.Get("error_stack_hash_without_useless"); errorStackHashWithoutUseless == "" {
		res["message"] = "error_stack_hash_without_useless异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.BuglySvr.CrashLaserRelationAdd(c, laserId, errorStackHashWithoutUseless, userName))
}

func oomIndexList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		matchOption *apm.MatchOption
		column      string
		eventId     int64
		err         error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.OrderBy == "" {
		matchOption.OrderBy = "count() DESC"
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	column = params.Get("column")
	c.JSON(s.BuglySvr.OOMIndexList(c, column, eventId, matchOption))
}

func oomInfoList(c *bm.Context) {
	var (
		matchOption *apm.MatchOption
		res         = map[string]interface{}{}
		err         error
	)
	matchOption = &apm.MatchOption{}
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.Hash == "" {
		res["message"] = "hash不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.OOMInfoList(c, matchOption))
}

func solveStatus(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		hash, appKey string
		eventId      int64
		err          error
	)
	hash = params.Get("hash")
	if hash == "" {
		res["message"] = "hash异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.SolveStatus(c, eventId, hash, appKey))
}

func updateIndex(c *bm.Context) {
	var (
		params                                                                            = c.Request.Form
		res                                                                               = map[string]interface{}{}
		solveStatus, solveVersionCode, eventId                                            int64
		assignOperator, solveOperator, operator, errorStackHash, solveDescription, appKey string
		wxNotify                                                                          bool
		err                                                                               error
	)
	if errorStackHash = params.Get("hash"); errorStackHash == "" {
		res["message"] = "hash异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	assignOperator = params.Get("assign_operator")
	if solveOperator = params.Get("solve_operator"); solveOperator == "" {
		res["message"] = "solve_operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveVersionCode, err = strconv.ParseInt(params.Get("solve_version_code"), 10, 64); err != nil {
		res["message"] = "solve_version_code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if solveStatus, err = strconv.ParseInt(params.Get("solve_status"), 10, 64); err != nil {
		res["message"] = "solve_status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	solveDescription = params.Get("solve_description")
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.Get("wx_notify") != "" {
		wxNotify, _ = strconv.ParseBool(params.Get("wx_notify"))
	}
	c.JSON(nil, s.BuglySvr.UpdateIndex(c, errorStackHash, appKey, assignOperator, solveOperator, operator, solveDescription, solveVersionCode, solveStatus, eventId, wxNotify))
}

func crashLogList(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		hash, appKey string
	)
	if hash = params.Get("hash"); hash == "" {
		res["message"] = "hash异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.LogList(c, hash, appKey))
}
