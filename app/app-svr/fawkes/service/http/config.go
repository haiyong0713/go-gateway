package http

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	apmmdl "go-gateway/app/app-svr/fawkes/service/model/apm"
	confmdl "go-gateway/app/app-svr/fawkes/service/model/config"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func appConfigVersionAdd(c *bm.Context) {
	var (
		params               = c.Request.Form
		res                  = map[string]interface{}{}
		appKey, env, version string
		versionCode          int64
		err                  error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		res["message"] = "version异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version != "default" {
		if versionCode, err = strconv.ParseInt(params.Get("version_code"), 10, 64); err != nil {
			res["message"] = "version_code异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigVersionAdd(c, appKey, env, version, versionCode, userName))
}

func appConfigVersionList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		pn, ps      int
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
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
	c.JSON(s.ConfigSvr.ConfigVersionList(c, appKey, env, pn, ps))
}

func appConfigVersionHistory(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid        int64
		pn, ps      int
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
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
	c.JSON(s.ConfigSvr.ConfigVersionHistory(c, appKey, env, cvid, pn, ps))
}

func appConfigVersionHistoryByID(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		cid    int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cid, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.AppConfigVersionHistoryByID(c, appKey, cid))
}

func appConfigVersionHistorys(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		pn, ps      int
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
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
	c.JSON(s.ConfigSvr.ConfigVersionHistorys(c, appKey, env, pn, ps))
}

func appConfigVersionDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		cvid   int64
		err    error
	)
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigVersionDel(c, cvid))
}

func appConfigFastAdd(c *bm.Context) {
	var (
		params                                                = c.Request.Form
		res                                                   = map[string]interface{}{}
		appKey, env, group, key, value, userName, description string
		cvid                                                  int64
		err                                                   error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if group = params.Get("group"); group == "" {
		res["message"] = "group异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if key = params.Get("key"); key == "" {
		res["message"] = "key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if value = params.Get("value"); value == "" {
		res["message"] = "value异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigFastAdd(c, appKey, env, cvid, group, key, value, userName, description))
}

func appConfigPublishView(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid        int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.ConfigPublishView(c, appKey, env, cvid))
}

func appConfigSave(c *bm.Context) {
	var (
		res                      = map[string]interface{}{}
		appKey, env, description string
		bs                       []byte
		cvid                     int64
		err                      error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	var cs *confmdl.ParamsConfig
	if err = json.Unmarshal(bs, &cs); err != nil {
		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if appKey = cs.AppKey; appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = cs.Env; env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid = cs.CVID; cvid == 0 {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Desc; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigAdd(c, appKey, env, cvid, cs.Items, userName, description))
}

func appConfigDiff(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid, cv    int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cv, err = strconv.ParseInt(params.Get("cv"), 10, 64); err != nil {
		res["message"] = "cv异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.ConfigDiff(c, appKey, env, cvid, cv))
}

func appConfigPublish(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid        int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigPublish(c, appKey, env, cvid, userName))
}

func appConfigPublishMultiple(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid        int64
		pubConf     []*confmdl.PubConfig
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	config := params.Get("config")
	if err = json.Unmarshal([]byte(config), &pubConf); err != nil {
		log.Error("appConfigPublish json.Unmarshal(%s) error(%v)", string(config), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigPublishMultiple(c, appKey, env, cvid, userName, pubConf))
}

func appConfig(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid        int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.Config(c, appKey, env, cvid))
}

func appConfigFile(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		cvid, cv    int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cv, err = strconv.ParseInt(params.Get("cv"), 10, 64); err != nil {
		res["message"] = "cv异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.AppConfigFile(c, appKey, env, cvid, cv))
}

func appConfigKeyPublishHistory(c *bm.Context) {
	var (
		params                    = c.Request.Form
		res                       = map[string]interface{}{}
		appKey, env, ckey, cgroup string
		cvid                      int64
		err                       error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cvid, err = strconv.ParseInt(params.Get("cvid"), 10, 64); err != nil {
		res["message"] = "cvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ckey = params.Get("ckey"); ckey == "" {
		res["message"] = "ckey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cgroup = params.Get("cgroup"); cgroup == "" {
		res["message"] = "cgroup异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.AppConfigKeyPublishHistory(c, appKey, env, ckey, cgroup, cvid))
}

func appConfigModifyCount(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.AppConfigModifyCount(c, appKey))
}

func appConfigSetKV(c *bm.Context) {
	var (
		params                           = c.Request.Form
		res                              = map[string]interface{}{}
		appKey, env, ckey, cgroup, value string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cgroup = params.Get("cgroup"); cgroup == "" {
		res["message"] = "cgroup异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ckey = params.Get("ckey"); ckey == "" {
		res["message"] = "ckey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if value = params.Get("value"); value == "" {
		res["message"] = "value异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.AppConfigSetKV(c, appKey, env, cgroup, ckey, value, userName))
}

func appConfigSreFallbackSet(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		ckey, cgroup string
	)
	if cgroup = params.Get("cgroup"); cgroup == "" {
		res["message"] = "cgroup异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ckey = params.Get("ckey"); ckey == "" {
		res["message"] = "ckey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cgroup != "grpc" {
		res["message"] = "您没有权限编辑该cgroup"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ckey != "fallback_list" && ckey != "fallback_list_v2" {
		res["message"] = "您没有权限编辑该ckey"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	appConfigSetKV(c)
}

func appConfigDCSampleRateSync(c *bm.Context) {
	var (
		params               = c.Request.Form
		res                  = map[string]interface{}{}
		appKey, env          string
		sampleRateConfigResp *apmmdl.EventSampleRateConfigResp
		err                  error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sampleRateConfigResp, err = s.ApmSvr.ApmEventSampleRateConfig(c, &apmmdl.EventSampleRateConfigReq{AppKey: appKey})
	if err != nil {
		res["message"] = "埋点采样率 - 配置异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.AppConfigSetKV(c, appKey, env, "neuron", "event_rates", sampleRateConfigResp.EventRates, userName))
}

func appConfigPublishDefault(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.ConfigSvr.ConfigPublishDefault(c, appKey, env, userName))
}
