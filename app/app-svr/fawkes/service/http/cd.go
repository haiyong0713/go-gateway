package http

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func appPortalTest(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		appKey, desc string
		buildID      int64
		err          error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if desc = params.Get("description"); desc == "" {
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
	c.JSON(nil, s.CDSvr.AppPortalTest(c, appKey, desc, userName, buildID))
}

func appCDVersions(c *bm.Context) {
	var (
		params                 = c.Request.Form
		res                    = map[string]interface{}{}
		appKey, env, filterKey string
		pn, ps                 int
		err                    error
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
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 100
	}
	if ps < 1 || ps > 100 {
		ps = 100
	}
	filterKey = params.Get("filter_key")
	c.JSON(s.CDSvr.PackVersionByAppKey(c, appKey, env, filterKey, ps, pn))
}

func appCDBuilds(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		versionID   int64
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
	if versionID, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.PackBuilds(c, appKey, env, versionID))
}

func appCDHotfixVersions(c *bm.Context) {
	var (
		params              = c.Request.Form
		res                 = map[string]interface{}{}
		appKey, env, filter string
		pn, ps              int
		err                 error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
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
	if ps < 1 || ps > 100 {
		ps = 100
	}
	filter = params.Get("filter")
	env = params.Get("env")
	c.JSON(s.CDSvr.PackVersionForHotfix(c, appKey, env, filter, pn, ps))
}

func appCDList(c *bm.Context) {
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
	c.JSON(s.CDSvr.AppCDList(c, appKey, env, pn, ps))
}

func appCDListFilter(c *bm.Context) {
	var (
		params                 = c.Request.Form
		res                    = map[string]interface{}{}
		appKey, env, filterKey string
		steadyState, pn, ps    int
		err                    error
		hasBbrUrl              bool
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
	if steadyState, err = strconv.Atoi(params.Get("steady_state")); err != nil {
		steadyState = 0
	}
	filterKey = params.Get("filter_key")
	if params.Get("bbr_url") != "" {
		hasBbrUrl, _ = strconv.ParseBool(params.Get("bbr_url"))
	}
	c.JSON(s.CDSvr.AppCDListFilter(c, appKey, env, filterKey, steadyState, hasBbrUrl, pn, ps))
}

func appCDConfigSwitchSet(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		isUpgrade   bool
		versionID   int64
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
	if versionID, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isUpgrade, err = strconv.ParseBool(params.Get("is_upgrade")); err != nil {
		res["message"] = "is_upgrade异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDConfigSwitchSet(c, appKey, env, versionID, isUpgrade, userName))
}

func appCDUpgradConfigSet(c *bm.Context) {
	var (
		params                                                                                                                             = c.Request.Form
		res                                                                                                                                = map[string]interface{}{}
		appKey, env, normal, exnormal, force, exforce, system, exSystem, title, content, policyURL, confirmBtnText, cancelBtnText, iconURL string
		cycle, silent, policy                                                                                                              int
		versionID                                                                                                                          int64
		err                                                                                                                                error
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
	if versionID, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if normal = params.Get("normal"); normal == "" {
		res["message"] = "normal异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if normal != "" {
		var n *cdmdl.UpgradeVersion
		if err = json.Unmarshal([]byte(normal), &n); err != nil {
			res["message"] = "normal值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	exnormal = params.Get("exclude_normal")
	if exnormal != "" {
		var exu []*cdmdl.ExcludeUpgradeVersion
		if err = json.Unmarshal([]byte(exnormal), &exu); err != nil {
			res["message"] = "exclude_normal值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	force = params.Get("force")
	if force != "" {
		var f *cdmdl.UpgradeVersion
		if err = json.Unmarshal([]byte(force), &f); err != nil {
			res["message"] = "force值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	exforce = params.Get("exclude_force")
	if exnormal != "" {
		var exf []*cdmdl.ExcludeUpgradeVersion
		if err = json.Unmarshal([]byte(exforce), &exf); err != nil {
			res["message"] = "exclude_force值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	system = params.Get("system")
	exSystem = params.Get("exclude_system")
	if cycle, err = strconv.Atoi(params.Get("cycle")); err != nil {
		res["message"] = "cycle异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "title异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if content = params.Get("content"); content == "" {
		res["message"] = "content异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if silent, err = strconv.Atoi(params.Get("is_silent")); err != nil {
		res["message"] = "is_silent异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if policy, err = strconv.Atoi(params.Get("policy")); err != nil {
		policy = cdmdl.UpdateByDefault
	}
	policyURL = params.Get("policy_url")
	if policy != cdmdl.UpdateByDefault && policyURL == "" {
		res["message"] = "policy_url 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	confirmBtnText = params.Get("confirm_btn_text")
	cancelBtnText = params.Get("cancel_btn_text")
	iconURL = params.Get("icon_url")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.UpgradConfigSet(c, appKey, env, versionID, normal, exnormal, force, exforce, system, exSystem, cycle, title, content, userName, policyURL, iconURL, confirmBtnText, cancelBtnText, policy, silent))
}

func appCDUpgradConfig(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		versionID   int64
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
	if versionID, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.UpgradConfig(c, appKey, env, versionID))
}

func appCDFilterConfigSet(c *bm.Context) {
	var (
		params                                                              = c.Request.Form
		res                                                                 = map[string]interface{}{}
		appKey, env, network, isp, channel, city, device, phoneModel, brand string
		percent, status                                                     int
		buildID                                                             int64
		err                                                                 error
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
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	network = params.Get("network")
	isp = params.Get("isp")
	channel = params.Get("channel")
	brand = params.Get("brand")
	if city = params.Get("city"); city == "" {
		city = "0"
	}
	phoneModel = params.Get("phone_model")
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	percent, _ = strconv.Atoi(params.Get("percent"))
	device = params.Get("device")
	if status == 0 {
		if percent == 0 && device == "" {
			res["message"] = "自定义模式下，升级比例和设备需至少配置一项"
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
	c.JSON(nil, s.CDSvr.FilterConfigSet(c, appKey, env, buildID, network, isp, channel, city, percent, device, userName, phoneModel, brand, status))
}

func appCDFilterConfig(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		buildID     int64
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
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.FilterConfig(c, appKey, env, buildID))
}

func appCDFlowConfigSet(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, env, flow string
		versionId         int64
		err               error
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
	if flow = params.Get("flow"); flow == "" {
		res["message"] = "flow异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var fs map[int64]string
	if err = json.Unmarshal([]byte(flow), &fs); err != nil {
		res["message"] = "flow异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if versionId, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.FlowConfigSet(c, appKey, env, userName, fs, versionId))
}

func appCDFlowConfig(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		versionID   int64
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
	if versionID, err = strconv.ParseInt(params.Get("version_id"), 10, 64); err != nil {
		res["message"] = "version_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.FlowConfig(c, appKey, env, versionID))
}

func appCDEvolution(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, env       string
		disPermil         int
		buildID, disLimit int64
		err               error
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
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	disPermil, _ = strconv.Atoi(params.Get("dis_permil"))
	if disPermil > cfg.AppstoreConnect.DisPermilLimit {
		res["message"] = fmt.Sprintf("testflight 分发不能大于千分之%d", cfg.AppstoreConnect.DisPermilLimit)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	disLimit, _ = strconv.ParseInt(params.Get("dis_limit"), 10, 64)
	_, _, err = s.CDSvr.CDEvolution(c, appKey, env, userName, disPermil, disLimit, buildID)
	c.JSON(nil, err)
}

func appCDGenerate(c *bm.Context) {
	var (
		params                           = c.Request.Form
		res                              = map[string]interface{}{}
		appKey, env, filter, order, sort string
		buildID, groupID                 int64
		pn, ps                           int
		err                              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	order = params.Get("order")
	sort = params.Get("sort")
	if sort != "desc" && sort != "asc" {
		sort = "desc"
	}
	env = params.Get("env")
	filter = params.Get("filter")
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	groupIDStr := params.Get("group_id")
	if groupIDStr == "" {
		groupID = -1
	} else if groupID, err = strconv.ParseInt(groupIDStr, 10, 64); err != nil {
		res["message"] = "group_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.GenerateList(c, appKey, env, filter, order, sort, buildID, groupID, pn, ps))
}

func appCDGenerateAdd(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey             string
		buildID, channleID int64
		err                error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channleID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		res["message"] = "channel_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDGenerateAdd(context.Background(), appKey, userName, buildID, channleID))
}

func appCDGenerateAddGit(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, channels string
		buildID          int64
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channels = params.Get("channels"); channels == "" {
		res["message"] = "channels异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if err = s.CDSvr.AppCDGenerateAddGit(c, appKey, channels, userName, buildID); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, err)
		return
	}
	c.JSON(nil, nil)
}

func appCDGenerateAdds(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDGenerateAdds(c, appKey, userName, buildID))
}

func appCDGenerateStatus(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		id     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDGenerateStatus(c, appKey, userName, id))
}

func appCDGenerateUpload(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		id     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDGenerateUpload(c, appKey, userName, id))
}

func appCDGeneratePublish(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		id     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDGeneratePublish(c, appKey, userName, id))
}

func appCDGeneratePublishList(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		cdnUrl string
	)
	if cdnUrl = params.Get("cdn_url"); cdnUrl == "" {
		res["message"] = "cdn_url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.AppCDGeneratePublishList(c, cdnUrl))
}

func appCDPatchList(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		pn, ps  int
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
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
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.PatchList(c, appKey, buildID, pn, ps))
}

func appCDPatchBuild(c *bm.Context) {
	var (
		params                                                   = c.Request.Form
		res                                                      = map[string]interface{}{}
		appKey, dstBuildID, dstCdnURL, dstLocalURL, srcBuildJSON string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if dstBuildID = params.Get("dst_build_id"); dstBuildID == "" {
		res["message"] = "dst_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if dstCdnURL = params.Get("dst_cdn_url"); dstCdnURL == "" {
		res["message"] = "dst_cdn_url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if dstLocalURL = params.Get("dst_local_url"); dstLocalURL == "" {
		res["message"] = "dst_local_url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if srcBuildJSON = params.Get("src_build_json"); srcBuildJSON == "" {
		res["message"] = "src_build_json异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var variables = map[string]string{
		"APP_KEY":        appKey,
		"TASK":           "PATCH",
		"DST_BUILD_ID":   dstBuildID,
		"DST_CDN_URL":    dstCdnURL,
		"DST_LOCAL_URL":  dstLocalURL,
		"SRC_BUILD_JSON": srcBuildJSON,
	}
	c.JSON(s.GitSvr.TriggerPipeline(c, "android", 0, cdmdl.PatchGitName, variables))
}

func appCDSyncMacross(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		isGray  int
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isGray, err = strconv.Atoi(params.Get("is_gray")); err != nil {
		res["message"] = "is_gray异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.SyncMacross(c, appKey, buildID, isGray))
}

func appCDSyncManager(c *bm.Context) {
	var (
		params               = c.Request.Form
		res                  = map[string]interface{}{}
		appKey, md5, channel string
		buildID              int64
		isGray, isPush       int
		err                  error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if md5 = params.Get("md5"); md5 == "" {
		res["message"] = "md5异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channel = params.Get("channel"); channel == "" {
		res["message"] = "channel异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isGray, err = strconv.Atoi(params.Get("is_gray")); err != nil {
		res["message"] = "is_gray异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isPush, err = strconv.Atoi(params.Get("is_push")); err != nil {
		res["message"] = "is_push异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.SyncManager(c, appKey, md5, channel, userName, buildID, isGray, isPush))
}

//func appCDGenerateTestStateSet(c *bm.Context) {
//	var (
//		params    = c.Request.Form
//		res       = map[string]interface{}{}
//		appKey    string
//		ids       string
//		testState int
//		err       error
//	)
//	if appKey = params.Get("app_key"); appKey == "" {
//		res["message"] = "appkey异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if ids = params.Get("ids"); ids == "" {
//		res["message"] = "ids异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if testState, err = strconv.Atoi(params.Get("test_state")); err != nil {
//		res["message"] = "test_state异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	username, _ := c.Get("username")
//	userName, ok := username.(string)
//	if !ok || userName == "" {
//		c.JSON(nil, ecode.NoLogin)
//		return
//	}
//	c.JSON(nil, s.CDSvr.AppCDGenerateTestStateSet(c, appKey, ids, userName, testState))
//}

func appCDPackSteadyStateSet(c *bm.Context) {
	var (
		params                      = c.Request.Form
		res                         = map[string]interface{}{}
		appKey                      string
		description                 string
		buildID, autoGenChannelPack int64
		steadyState                 int
		err                         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = params.Get("description"); description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if steadyState, err = strconv.Atoi(params.Get("steady_state")); err != nil {
		res["message"] = "steady_state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	autoGenStr := params.Get("auto_gen_channel_pack")
	if len(autoGenStr) != 0 {
		if autoGenChannelPack, err = strconv.ParseInt(params.Get("auto_gen_channel_pack"), 10, 64); err != nil {
			res["message"] = "auto_gen_channel_pack异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	c.JSON(nil, s.CDSvr.AppCDPackSteadyStateSet(c, appKey, description, buildID, steadyState, autoGenChannelPack))
}

func appCDRefreshCDN(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		cdnUrls string
	)
	if cdnUrls = params.Get("cdn_urls"); cdnUrls == "" {
		res["message"] = "cdn_urls异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDRefreshCDN(c, cdnUrls))
}

func appCDCustomChannelList(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.AppCDCustomChannelList(c, appKey, buildID))
}

func appCDCustomChannelAdd(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDCustomChannelAdd(c, appKey, userName, buildID))
}

func appCDCustomChannelUpload(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		id     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CDSvr.AppCDCustomChannelUpload(c, appKey, userName, id))
}

func testflightAppSet(c *bm.Context) {
	var (
		res                                                                     = map[string]interface{}{}
		appKey, storeAppID, issuerID, keyID, tagPrefix, buglyAppID, buglyAppKey string
		file                                                                    multipart.File
		header                                                                  *multipart.FileHeader
		err                                                                     error
	)
	if file, header, err = c.Request.FormFile("pk"); err != nil {
		res["message"] = "pk 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = c.Request.FormValue("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if storeAppID = c.Request.FormValue("store_app_id"); storeAppID == "" {
		res["message"] = "store_app_id 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if issuerID = c.Request.FormValue("issuer_id"); issuerID == "" {
		res["message"] = "issuer_id 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if keyID = c.Request.FormValue("key_id"); keyID == "" {
		res["message"] = "key_id 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	tagPrefix = c.Request.FormValue("tag_prefix")
	buglyAppID = c.Request.FormValue("bugly_app_id")
	buglyAppKey = c.Request.FormValue("bugly_app_key")
	c.JSON(nil, s.CDSvr.TestFlightAppInfoSet(c, appKey, storeAppID, issuerID, keyID, tagPrefix, buglyAppID, buglyAppKey, file, header))
}

func testflightAppInfo(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.TestFlightAppInfo(c, appKey))
}

func testflightBetaReview(c *bm.Context) {
	var (
		params              = c.Request.Form
		res                 = map[string]interface{}{}
		appKey, betaBuildID string
		packTFID            int64
		err                 error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if betaBuildID = params.Get("beta_build_id"); betaBuildID == "" {
		res["message"] = "beta_build_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.SubmitBetaReview(appKey, packTFID, betaBuildID))
}

func testflightDistribute(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey             string
		packTFID, disLimit int64
		disPermil          int
		err                error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if disPermilStr := params.Get("dis_permil"); disPermilStr != "" {
		if disPermil, err = strconv.Atoi(disPermilStr); err != nil {
			res["message"] = "dis_permil 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if disPermil > cfg.AppstoreConnect.DisPermilLimit {
			res["message"] = fmt.Sprintf("testflight 分发不能大于千分之%d", cfg.AppstoreConnect.DisPermilLimit)
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "dis_permil 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if disLimit, err = strconv.ParseInt(params.Get("dis_limit"), 10, 64); err != nil {
		res["message"] = "dis_limit 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.DistributeTestFlight(appKey, packTFID, disPermil, disLimit))
}

func testflightStop(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		appKey   string
		packTFID int64
		err      error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.StopTestFlight(appKey, packTFID))
}

func testflightRemindUpdate(c *bm.Context) {
	var (
		params                  = c.Request.Form
		res                     = map[string]interface{}{}
		packTFID, remindUpdTime int64
		err                     error
	)
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if remindUpdTime, err = strconv.ParseInt(params.Get("remind_upd_time"), 10, 64); err != nil {
		res["message"] = "remind_upd_time 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// warning 增加鉴权
	c.JSON(nil, s.CDSvr.UpdateRemindUpdTime(packTFID, remindUpdTime))
}

func testflightForceUpdate(c *bm.Context) {
	var (
		params                 = c.Request.Form
		res                    = map[string]interface{}{}
		packTFID, forceUpdTime int64
		err                    error
	)
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if forceUpdTime, err = strconv.ParseInt(params.Get("force_upd_time"), 10, 64); err != nil {
		res["message"] = "force_upd_time 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// warning 增加鉴权
	c.JSON(nil, s.CDSvr.UpdateForceUpdTime(packTFID, forceUpdTime))
}

func testflightBetagroupSet(c *bm.Context) {
	var (
		params                             = c.Request.Form
		res                                = map[string]interface{}{}
		appKey, publicLink, publicLinkTest string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if publicLink = params.Get("public_link"); publicLink == "" {
		res["message"] = "public_link 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if publicLinkTest = params.Get("public_link_test"); publicLinkTest == "" {
		res["message"] = "public_link_test 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.SetBetaGroups(appKey, publicLink, publicLinkTest))
}

func testflightSetUpdTxt(c *bm.Context) {
	var (
		params                                = c.Request.Form
		res                                   = map[string]interface{}{}
		packTFID                              int64
		guideTFTxt, remindUpdTxt, forceUpdTxt string
		err                                   error
	)
	if packTFID, err = strconv.ParseInt(params.Get("pack_tf_id"), 10, 64); err != nil {
		res["message"] = "pack_tf_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	guideTFTxt = params.Get("guide_tf_txt")
	if remindUpdTxt = params.Get("remind_upd_txt"); remindUpdTxt == "" {
		res["message"] = "remind_upd_txt 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if forceUpdTxt = params.Get("force_upd_txt"); forceUpdTxt == "" {
		res["message"] = "force_upd_txt 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// warning 增加鉴权
	c.JSON(nil, s.CDSvr.TFSetUpdTxt(packTFID, guideTFTxt, remindUpdTxt, forceUpdTxt))
}

func testflightUploadBugly(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey      string
		buildPackID int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildPackID, err = strconv.ParseInt(params.Get("build_pack_id"), 10, 64); err != nil {
		res["message"] = "build_pack_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.TFUploadBugly(c, appKey, buildPackID))
}

func testflightBWAdd(c *bm.Context) {
	var (
		params                      = c.Request.Form
		res                         = map[string]interface{}{}
		appKey, env, nick, listType string
		mid                         int64
		err                         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mid, err = strconv.ParseInt(params.Get("mid"), 10, 64); err != nil {
		res["message"] = "mid 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if nick = params.Get("nick"); nick == "" {
		res["message"] = "nick 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if listType = params.Get("list_type"); listType == "" {
		res["message"] = "list_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.TestflightBWAdd(appKey, env, mid, nick, userName, listType))
}

func testflightBWList(c *bm.Context) {
	var (
		params                = c.Request.Form
		res                   = map[string]interface{}{}
		appKey, env, listType string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if listType = params.Get("list_type"); listType == "" {
		res["message"] = "list_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.TestflightBWList(c, appKey, listType, env))
}

func testflightBWDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		id     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.TestflightBWDel(id))
}

func testflightTestingPack(c *bm.Context) {
	c.JSON(s.CDSvr.BetaPacks(c))
}

func releaseNotify(c *bm.Context) {
	var (
		params            = c.Request.Form
		buildID           int64
		appKey, env, bots string
		notyfyGroup       bool
		err               error
		res               = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	bots = params.Get("bots")
	notyfyGroup, _ = strconv.ParseBool(c.Request.FormValue("notify_group"))
	c.JSON(nil, s.CDSvr.ReleaseNotify(c, buildID, appKey, env, bots, notyfyGroup))
}

func windowsAppinstallerUpload(c *bm.Context) {
	var (
		params  = c.Request.Form
		buildID int64
		appKey  string
		err     error
		res     = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.WindowsAppinstallerUpload(c, appKey, buildID))
}

func windowsAppinstallerPublish(c *bm.Context) {
	var (
		params  = c.Request.Form
		buildID int64
		appKey  string
		err     error
		res     = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CDSvr.WindowsAppinstallerPublish(c, appKey, buildID))
}

// 资源包推送到cd的正式环境
func assetsEvolution(c *bm.Context) {
	var (
		params                = c.Request.Form
		buildID               int64
		err                   error
		unzip                 bool
		res                   = map[string]interface{}{}
		file                  multipart.File
		header                *multipart.FileHeader
		appKey, fmd5, pkgName string
	)

	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(c.Request.FormValue("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fmd5 = c.Request.FormValue("md5"); fmd5 == "" {
		res["message"] = "md5 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, file); err != nil {
		c.JSON(nil, err)
		log.Errorc(c, "%v", err)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		c.JSON(nil, err)
		log.Errorc(c, "%v", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	if fmd5 != hex.EncodeToString(md5Bs[:]) {
		res["message"] = "md5 校验错误"
		c.JSON(res, ecode.RequestErr)
		return
	}
	if unzip, err = strconv.ParseBool(c.Request.FormValue("unzip")); err != nil {
		unzip = false
	}
	if unzip {
		if pkgName = c.Request.FormValue("pkg_name"); pkgName == "" {
			res["message"] = "pkg_name 为空"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		pkgName = header.Filename
	}
	c.JSON(nil, s.CDSvr.AssetsEvolution(c, appKey, buildID, file, header, unzip, pkgName))
}

func packGreyList(c *bm.Context) {
	var (
		params                                   = c.Request.Form
		appKey, version                          string
		versionCode, glJobId, startTime, endTime int64
		pn, ps                                   int
		res                                      = map[string]interface{}{}
		err                                      error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	versionCodeStr := params.Get("version_code")
	if versionCodeStr != "" {
		if versionCode, err = strconv.ParseInt(versionCodeStr, 10, 64); err != nil {
			res["message"] = "version_code 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	glJobIdStr := params.Get("gl_job_id")
	if glJobIdStr != "" {
		if glJobId, err = strconv.ParseInt(glJobIdStr, 10, 64); err != nil {
			res["message"] = "gl_job_id 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 100
	}
	if startTime, err = strconv.ParseInt(params.Get("start_time"), 10, 64); err != nil {
		startTime = 0
	}
	if endTime, err = strconv.ParseInt(params.Get("end_time"), 10, 64); err != nil {
		endTime = 0
	}
	version = params.Get("version")
	c.JSON(s.CDSvr.PackGreyList(c, appKey, version, versionCode, glJobId, startTime, endTime, pn, ps))
}

// appCDCDNPublish 将指定cd产物发布到CDN的固定地址 {{cdnhost}}/{{originpath}}/fixed/{{app_key}}/filename
func appCDCDNPublish(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, filename string
		buildId          int64
		err              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildId, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if filename = params.Get("filename"); filename == "" {
		res["message"] = "filename异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CDSvr.AppCDCDNPublish(c, appKey, buildId, filename))
}

func appCDGeneratePackUpload(c *bm.Context) {
	var (
		params         = c.Request.Form
		res            = map[string]interface{}{}
		appKey, md5Str string
		buildID        int64
		err            error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		err = ecode.Error(ecode.RequestErr, "app_key异常")
		c.JSON(nil, err)
		return
	}
	if md5Str = params.Get("md5"); md5Str == "" {
		err = ecode.Error(ecode.RequestErr, "md5异常")
		c.JSON(nil, err)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		err = ecode.Error(ecode.RequestErr, "file异常")
		c.JSON(nil, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, file); err != nil {
		c.JSON(nil, err)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		c.JSON(nil, err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	log.Infoc(c, "md5 %v", hex.EncodeToString(md5Bs[:]))
	if md5Str != hex.EncodeToString(md5Bs[:]) {
		err = ecode.Error(ecode.RequestErr, "md5 校验错误")
		c.JSON(nil, err)
		return
	}
	c.JSON(s.CDSvr.AppCDGeneratePackUpload(context.Background(), appKey, userName, buildID, file, header))
}
