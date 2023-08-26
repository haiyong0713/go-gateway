package http

import (
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"strconv"
	"strings"

	"go-common/library/net/http/blademaster/binding"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	confmdl "go-gateway/app/app-svr/fawkes/service/model/config"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func newestVersion(c *bm.Context) {
	c.JSON(s.BusSvr.NewestVersion(c))
}

func versionAll(c *bm.Context) {
	c.JSON(s.BusSvr.VersionAll(c))
}

func upgradeAll(c *bm.Context) {
	c.JSON(s.BusSvr.UpgradeAll(c))
}

func packAll(c *bm.Context) {
	c.JSON(s.BusSvr.PackAll(c))
}

func packLatestStable(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey      string
		versionCode int
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if versionCode, err = strconv.Atoi(params.Get("version_code")); err != nil {
		res["message"] = "version_code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BusSvr.PackLatestStable(c, appKey, versionCode))
}

func filterAll(c *bm.Context) {
	c.JSON(s.BusSvr.FilterAll(c))
}

func patchAll(c *bm.Context) {
	c.JSON(s.BusSvr.PatchAllCache(c))
}

func channelAll(c *bm.Context) {
	c.JSON(s.BusSvr.ChannelAll(c))
}

func flowAll(c *bm.Context) {
	c.JSON(s.BusSvr.FlowAll(c))
}

func hotfixAll(c *bm.Context) {
	c.JSON(s.BusSvr.HotfixAllCache(c))
}

func laserReport(c *bm.Context) {
	var (
		params                                                   = c.Request.Form
		res                                                      = map[string]interface{}{}
		taskID                                                   int64
		status                                                   int
		url, recallMobiApp, build, errorMessage, md5, rawUposUri string
		err                                                      error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		res["message"] = "task_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	url = params.Get("url")
	recallMobiApp = params.Get("recall_mobi_app")
	build = params.Get("build")
	errorMessage = params.Get("error_msg")
	md5 = params.Get("md5")
	rawUposUri = params.Get("raw_upos_uri")
	if status == appmdl.StatusUpSuccess && url == "" {
		res["message"] = "url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserReport(c, taskID, status, url, recallMobiApp, build, errorMessage, md5, rawUposUri))
}

func laserReport2(c *bm.Context) {
	var (
		params                                                   = c.Request.Form
		res                                                      = map[string]interface{}{}
		appKey, buvid                                            string
		mid, taskID                                              int64
		status                                                   int
		url, recallMobiApp, build, errorMessage, md5, rawUposUri string
		err                                                      error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	buvid = params.Get("buvid")
	if mid == 0 && buvid == "" {
		res["message"] = "mid和buvid all err"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		status = 0
	}
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		taskID = 0
	}
	url = params.Get("url")
	recallMobiApp = params.Get("recall_mobi_app")
	build = params.Get("build")
	errorMessage = params.Get("error_msg")
	md5 = params.Get("md5")
	rawUposUri = params.Get("raw_upos_uri")
	c.JSON(s.LaserSvr.AppLaserAdd2(c, appKey, buvid, url, recallMobiApp, build, errorMessage, md5, rawUposUri, status, mid, taskID))
}

func laserCmdReport(c *bm.Context) {
	var (
		params                                                       = c.Request.Form
		taskID                                                       int64
		status                                                       int
		url, result, recallMobiApp, build, errorMsg, md5, rawUposUri string
		err                                                          error
	)
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		status = 0
	}
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		taskID = 0
	}
	recallMobiApp = params.Get("mobi_app")
	build = params.Get("build")
	url = params.Get("url")
	errorMsg = params.Get("error_msg")
	result = params.Get("result")
	md5 = params.Get("md5")
	rawUposUri = params.Get("raw_upos_uri")
	c.JSON(nil, s.AppSvr.AppLaserCmdReport(c, taskID, status, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri))
}

func laser(c *bm.Context) {
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
	c.JSON(s.BusSvr.Laser(c, taskID))
}

func generatesPunlish(c *bm.Context) {
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
	c.JSON(s.BusSvr.GenerateList(c, appKey, buildID))
}

//	func generateTestStateSet(c *bm.Context) {
//		var (
//			params    = c.Request.Form
//			res       = map[string]interface{}{}
//			appKey    string
//			ids       string
//			testState int
//			err       error
//		)
//		if appKey = params.Get("app_key"); appKey == "" {
//			res["message"] = "appkey异常"
//			c.JSONMap(res, ecode.RequestErr)
//			return
//		}
//		if ids = params.Get("ids"); ids == "" {
//			res["message"] = "ids异常"
//			c.JSONMap(res, ecode.RequestErr)
//			return
//		}
//		if testState, err = strconv.Atoi(params.Get("test_state")); err != nil {
//			res["message"] = "test_state异常"
//			c.JSONMap(res, ecode.RequestErr)
//			return
//		}
//		c.JSON(nil, s.CDSvr.AppCDGenerateTestStateSet(c, appKey, ids, "??", testState))
//	}
func generatesUpdate(c *bm.Context) {
	var (
		params                  = c.Request.Form
		res                     = map[string]interface{}{}
		appKey, channelFileInfo string
		jobID                   int64
		channelStatus           int
		err                     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channelStatus, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if jobID, err = strconv.ParseInt(params.Get("job_id"), 10, 64); err != nil {
		res["message"] = "job_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channelFileInfo = params.Get("channel_file_info"); channelFileInfo == "" {
		res["message"] = "channel_file_info异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = s.CDSvr.GeneratesUpdate(c, appKey, channelFileInfo, jobID, channelStatus); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, err)
		return
	}
	c.JSON(nil, nil)
}

func laserReportSilence(c *bm.Context) {
	var (
		params                                  = c.Request.Form
		res                                     = map[string]interface{}{}
		taskID                                  int64
		status                                  int
		url, recallMobiApp, build, errorMessage string
		err                                     error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		res["message"] = "task_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	url = params.Get("url")
	recallMobiApp = params.Get("recall_mobi_app")
	build = params.Get("build")
	errorMessage = params.Get("error_msg")
	if status == appmdl.StatusUpSuccess && url == "" {
		res["message"] = "url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.LaserSvr.AppLaserReportSilence(c, taskID, status, url, recallMobiApp, build, errorMessage))
}

func busAppChannelAdd(c *bm.Context) {
	var (
		params                              = c.Request.Form
		res                                 = map[string]interface{}{}
		channelID, groupID                  int64
		code, name, plate, appKey, operator string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 提供自定义渠道添加服务. 非静态渠道
	chType := 2
	if code = params.Get("code"); code == "" {
		res["message"] = "code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); code == "" || name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if plate = params.Get("plate"); plate == "" {
		res["message"] = "plate异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelAdd(c, chType, channelID, groupID, code, name, plate, operator, appKey))
}

func bizApkListAll(c *bm.Context) {
	c.JSON(s.BusSvr.BizApkListAllCache(c))
}

func tribeListAll(c *bm.Context) {
	c.JSON(s.BusSvr.TribeListAllCache(c))
}
func tribeRelationAll(c *bm.Context) {
	c.JSON(s.BusSvr.TribeRelationAllCache(c))
}

func appUseableTribes(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		buildID     int64
		err         error
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
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BusSvr.AppUseableTribes(c, appKey, env, buildID))
}

func tribeHosts(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		id          int64
		tribeName   string
		err         error
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
	if tribeName = params.Get("tribe_name"); tribeName == "" {
		res["message"] = "tribe_name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BusSvr.TribeHosts(c, appKey, env, tribeName, id))
}

func appCDVersionList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
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
	c.JSON(s.BusSvr.AppCDVersionList(c, appKey, env))
}

//func hawkeyeWebhookCrash(c *bm.Context) {
//	var (
//		bs  []byte
//		err error
//	)
//	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
//		log.Error("ioutil.ReadAll() error(%v)", err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	c.Request.Body.Close()
//	var alterParams *apmmdl.AlertWebhookParams
//	if err = json.Unmarshal(bs, &alterParams); err != nil {
//		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	c.JSON(nil, s.BusSvr.HawkeyeWebhookCrash(c, alterParams))
//}

func modAppKeyList(ctx *bm.Context) {
	ctx.JSON(s.BusSvr.AppkeyList(ctx))
}

func modAppkeyFileList(ctx *bm.Context) {
	param := new(struct {
		AppKey string  `form:"app_key" validate:"required"`
		Env    mod.Env `form:"env" validate:"required"`
		Md5    string  `form:"md5"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Env.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	file, md5Val, err := s.BusSvr.AppKeyFileList(ctx, param.AppKey, param.Env, param.Md5)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["file"] = file
	res["md5"] = md5Val
	ctx.JSON(res, nil)
}

func testflightAll(c *bm.Context) {
	var (
		param = c.Request.Form
		env   string
	)
	if env = param.Get("env"); env != "test" {
		env = "prod"
	}
	c.JSON(s.BusSvr.TestFlightAll(c, env))
}

func patchSetStatus(c *bm.Context) {
	var (
		param       = c.Request.Form
		res         = map[string]interface{}{}
		appKey      string
		status      int
		id, glJobID int64
		err         error
	)
	if id, err = strconv.ParseInt(param.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if glJobID, err = strconv.ParseInt(param.Get("gl_job_id"), 10, 64); err != nil {
		res["message"] = "gl_job_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(param.Get("status")); err != nil {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = param.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BusSvr.PatchSetStatus(c, id, glJobID, status, appKey))
}

func patchUpload(c *bm.Context) {
	var (
		param       = c.Request.Form
		res         = map[string]interface{}{}
		id          int64
		appKey, md5 string
		file        multipart.File
		err         error
	)
	if file, _, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(param.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if md5 = param.Get("md5"); md5 == "" {
		res["message"] = "md5 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = param.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BusSvr.PatchUpload(c, id, file, md5, appKey))
}

func addManagerShareConfig(c *bm.Context) {
	var (
		res                                = map[string]interface{}{}
		appKey, env, description, operator string
		bs                                 []byte
		err                                error
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
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = cs.Env; env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = cs.Operator; operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Desc; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ConfigSvr.AddConfigBusiness(c, appKey, env, cs.Items, confmdl.BusinessShareChannel, cs.Operator, description))
}

func addConfig(c *bm.Context) {
	var (
		res                                           = map[string]interface{}{}
		appKey, env, groupName, description, operator string
		bs                                            []byte
		err                                           error
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
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = cs.Env; env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if groupName = cs.GroupName; groupName == "" {
		res["message"] = "groupName异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = getCurrentUsername(c); operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Desc; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ConfigSvr.AddConfigBusiness(c, appKey, env, cs.Items, groupName, operator, description))
}

func addDefaultConfig(c *bm.Context) {
	var (
		res                                          = map[string]interface{}{}
		appKey, env, description, operator, business string
		bs                                           []byte
		err                                          error
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
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = cs.Env; env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if operator = cs.Operator; operator == "" {
		res["message"] = "operator异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if description = cs.Desc; description == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if business = cs.Business; business == "" {
		res["message"] = "business异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if business != confmdl.BusinessShareChannel && business != confmdl.BusinessActiveListChannel {
		res["message"] = "business异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.ConfigSvr.AddConfigBusinessV2(c, appKey, env, cs.Items, business, cs.Operator, description))
}

func defaultConfigHistory(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		pn, ps      int
		err         error
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
	c.JSON(s.ConfigSvr.DefaultConfigHistory(c, appKey, env, pn, ps))
}

func defaultConfig(c *bm.Context) {
	var (
		params                = c.Request.Form
		res                   = map[string]interface{}{}
		appKey, env, business string
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
	if business = params.Get("business"); business == "" {
		res["message"] = "business异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.ConfigSvr.DefaultConfig(c, appKey, env, business))
}

//
//func addApmEvent(c *bm.Context) {
//	var (
//		res                                                                                                                      = map[string]interface{}{}
//		name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string
//		level, isWideTable                                                                                                       int8
//		shared, sampleRate                                                                                                       int
//		busId, datacenterAppID, datacenterEventID, dataCount                                                                     int64
//		bs                                                                                                                       []byte
//		err                                                                                                                      error
//	)
//	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
//		log.Error("ioutil.ReadAll() error(%v)", err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	c.Request.Body.Close()
//	// params
//	var cs *apm.ParamsEvent
//	if err = json.Unmarshal(bs, &cs); err != nil {
//		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	if logID = cs.LogID; logID == "" {
//		res["message"] = "log_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if name = cs.Name; name == "" {
//		res["message"] = "name异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if description = cs.Description; description == "" {
//		res["message"] = "description异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if owner = cs.Owner; owner == "" {
//		res["message"] = "owner异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if busId = cs.BusID; busId == 0 {
//		res["message"] = "bus_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if datacenterAppID = cs.DatacenterAppID; datacenterAppID == 0 {
//		res["message"] = "datacenter_app_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if datacenterEventID = cs.DatacenterEventID; datacenterEventID == 0 {
//		res["message"] = "datacenter_event_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if sampleRate = cs.SampleRate; sampleRate == 0 {
//		sampleRate = 10000
//	}
//	appKey = cs.AppKey
//	appKeys = cs.AppKeys
//	dbName = cs.Databases
//	tableName = cs.TableName
//	distributedTableName = cs.DistributedTableName
//	shared = cs.Shared
//	topic = cs.Topic
//	level = cs.Level
//	dataCount = cs.DataCount
//	dwdTableName = cs.DatacenterDwdTableName
//	isWideTable = cs.IsWideTable
//	username, _ := c.Get("username")
//	userName, ok := username.(string)
//	if !ok || userName == "" {
//		c.JSON(nil, ecode.NoLogin)
//		return
//	}
//	c.JSON(s.ApmSvr.BusApmEventAdd(c, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName, level, isWideTable, shared, sampleRate, busId, datacenterAppID, datacenterEventID, dataCount))
//}
//
//func updateApmEvent(c *bm.Context) {
//	var (
//		res                                                                                                              = map[string]interface{}{}
//		appKeys, description, name, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string
//		shared, sampleRate                                                                                               int
//		activity, state, level, isWideTable                                                                              int8
//		eventId, datacenterAppID, busId, datacenterEventID, dataCount                                                    int64
//		bs                                                                                                               []byte
//		err                                                                                                              error
//	)
//	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
//		log.Error("ioutil.ReadAll() error(%v)", err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	c.Request.Body.Close()
//	// params
//	var cs *apm.ParamsEvent
//	if err = json.Unmarshal(bs, &cs); err != nil {
//		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
//		c.JSON(nil, ecode.RequestErr)
//		return
//	}
//	if eventId = cs.EventID; eventId == 0 {
//		res["message"] = "event_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if logID = cs.LogID; logID == "" {
//		res["message"] = "log_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if name = cs.Name; name == "" {
//		res["message"] = "name异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if description = cs.Description; description == "" {
//		res["message"] = "description异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if owner = cs.Owner; owner == "" {
//		res["message"] = "owner异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if busId = cs.BusID; busId == 0 {
//		res["message"] = "bus_id异常"
//		c.JSONMap(res, ecode.RequestErr)
//		return
//	}
//	if sampleRate = cs.SampleRate; sampleRate == 0 {
//		sampleRate = 10000
//	}
//
//	appKeys = cs.AppKeys
//	state = cs.State
//	dbName = cs.Databases
//	tableName = cs.TableName
//	distributedTableName = cs.DistributedTableName
//	shared = cs.Shared
//	topic = cs.Topic
//	activity = cs.Activity
//	datacenterAppID = cs.DatacenterAppID
//	datacenterEventID = cs.DatacenterEventID
//	level = cs.Level
//	dataCount = cs.DataCount
//	dwdTableName = cs.DatacenterDwdTableName
//	isWideTable = cs.IsWideTable
//	username, _ := c.Get("username")
//	userName, ok := username.(string)
//	if !ok || userName == "" {
//		c.JSON(nil, ecode.NoLogin)
//		return
//	}
//	c.JSON(nil, s.ApmSvr.BusApmEventUpdate(c, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName, activity, state, level, isWideTable, shared, sampleRate, eventId, datacenterAppID, busId, datacenterEventID, dataCount))
//}

func setApmEventField(c *bm.Context) {
	p := new(apm.EventFieldReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.EventID == 0 && p.CommonFieldsFlag != 1 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	p.Operator = userName
	c.JSON(nil, s.ApmSvr.BusApmEventFieldSet(c, p))
}

func crashIndexListByHashList(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		stackHashList []string
		appKey        string
		eventId       int64
		err           error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if hashStr := params.Get("stack_hash_list"); hashStr != "" {
		stackHashList = append(stackHashList, strings.Split(hashStr, ",")...)
	}
	if len(stackHashList) == 0 {
		res["message"] = "stack_hash_list异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if eventId, err = strconv.ParseInt(params.Get("event_id"), 10, 64); err != nil {
		res["message"] = "event_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.CrashIndexMessageByHashList(c, appKey, eventId, stackHashList))
}

func jankIndexListByHashList(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		stackHashList []string
		appKey        string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if hashStr := params.Get("stack_hash_list"); hashStr != "" {
		stackHashList = append(stackHashList, strings.Split(hashStr, ",")...)
	}
	if len(stackHashList) == 0 {
		res["message"] = "stack_hash_list异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BuglySvr.JankIndexMessageByHashList(c, appKey, stackHashList))
}

func parseIP(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		addr   string
	)
	if addr = params.Get("addr"); addr == "" {
		res["message"] = "addr异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BusSvr.ParseIP(c, addr))
}

func getCurrentUsername(c *bm.Context) string {
	var username string
	if un, ok := c.Get("username"); ok {
		username = un.(string)
	}
	return username
}

func addPcdnFile(c *bm.Context) {
	var (
		params                             = c.Request.Form
		res                                = map[string]interface{}{}
		rid, url, md5, business, versionId string
		size                               int64
		err                                error
	)
	if rid = params.Get("rid"); rid == "" {
		res["message"] = "rid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if url = params.Get("url"); url == "" {
		res["message"] = "url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if md5 = params.Get("md5"); md5 == "" {
		res["message"] = "md5异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if business = params.Get("business"); business == "" {
		res["message"] = "business异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if versionId = params.Get("version_id"); versionId == "" {
		res["message"] = "versionId异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if size, err = strconv.ParseInt(params.Get("size"), 10, 64); err != nil {
		res["message"] = "size异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BusSvr.AddPcdnFile(c, rid, url, md5, business, versionId, size))
}

func pcdnFileList(ctx *bm.Context) {
	var (
		params          = ctx.Request.Form
		res             = map[string]interface{}{}
		versionId, zone string
	)
	versionId = params.Get("version_id")
	if zone = params.Get("zone"); zone == "" {
		res["message"] = "zone异常"
		ctx.JSONMap(res, ecode.RequestErr)
		return
	}
	ctx.JSON(s.BusSvr.PcdnFileList(ctx, versionId, zone))
}
