package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/model"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

func appInfo(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		ID     int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	idStr := params.Get("id")
	if idStr == "" {
		ID = -1
	} else if ID, err = strconv.ParseInt(idStr, 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.AppSvr.AppInfo(c, appKey, ID))
}

func appEdit(c *bm.Context) {
	var (
		params                                                                                           = c.Request.Form
		res                                                                                              = map[string]interface{}{}
		id, datacenterAppID, serverZone, isHost                                                          int64
		name, desc, treePath, appID, mobiApp, platform, gitPath, icon, dsymName, symbolsoName, projectID string
		owners                                                                                           []string
		err                                                                                              error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if projectID = params.Get("git_prj_id"); projectID == "" {
		res["message"] = "git_prj_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appID = params.Get("app_id"); appID == "" {
		res["message"] = "app_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		res["message"] = "mobi_app异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if platform = params.Get("platform"); platform == "" {
		res["message"] = "platform异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitPath = params.Get("git_path"); gitPath == "" {
		res["message"] = "git_path异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if icon = params.Get("icon"); icon == "" {
		res["message"] = "icon异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	name = params.Get("name")
	desc = params.Get("description")
	treePath = params.Get("tree_path")
	dsymName = params.Get("app_dsym_name")
	symbolsoName = params.Get("app_symbolso_name")
	if name == "" && desc == "" && treePath == "" {
		res["message"] = "无效请求"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if os := params.Get("owners"); os != "" {
		owners = strings.Split(params.Get("owners"), ",")
	}
	// 如果上传icon, 需要进行替换
	f, _, _ := c.Request.FormFile("logofile")
	buf := new(bytes.Buffer)
	if f != nil {
		if _, err = io.Copy(buf, f); err != nil {
			res["message"] = "logofile异常 " + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		defer f.Close()
	}
	// 数据平台AppID
	if datacenterAppID, err = strconv.ParseInt(params.Get("datacenter_app_id"), 10, 64); err != nil {
		res["message"] = "datacenter_app_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if serverZoneStr := params.Get("server_zone"); serverZoneStr != "" {
		if serverZone, err = strconv.ParseInt(serverZoneStr, 10, 64); err != nil {
			res["message"] = "server_zone异常"
			c.JSONMap(res, ecode.RequestErr)
		}
	}
	laserWebhook := params.Get("laser_webhook")
	if isHost, err = strconv.ParseInt(params.Get("is_host"), 10, 64); err != nil {
		isHost = 0
	}
	c.JSON(nil, s.AppSvr.AppEdit(c, id, datacenterAppID, serverZone, isHost, appID, mobiApp, platform, gitPath, icon, strings.Join(owners, ","), name, desc, treePath, dsymName, symbolsoName, projectID, userName, laserWebhook, buf.Bytes()))
}

func appUpdateIsHighestPeak(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		appKey        string
		isHighestPeak int64
		err           error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isHighestPeak, err = strconv.ParseInt(params.Get("is_highest_peak"), 10, 64); err != nil {
		res["message"] = "is_highest_peak异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppUpdateIsHighestPeak(c, appKey, isHighestPeak))
}

func appFollowList(c *bm.Context) {
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppFollowList(c, userName))
}

func appList(c *bm.Context) {
	var (
		params          = c.Request.Form
		datacenterAppId int64
		err             error
	)
	if datacenterAppId, err = strconv.ParseInt(params.Get("datacenter_app_id"), 10, 64); err != nil {
		datacenterAppId = 0
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppList(c, userName, datacenterAppId))
}

func appFollowAdd(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppFollowAdd(c, appKey, userName))
}

func appFollowDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppFollowDeL(c, appKey, userName))
}

func appAdd(c *bm.Context) {
	var (
		params                                                          = c.Request.Form
		res                                                             = map[string]interface{}{}
		appID, appKey, mobiApp, platform, name, treePath, desc, gitPath string
		datacenterAppId, isHost                                         int64
		owners                                                          []string
		err                                                             error
	)
	utils.GetUsername(c)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appID = params.Get("app_id"); appID == "" {
		res["message"] = "app_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		res["message"] = "mobi_app异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if platform = params.Get("platform"); platform == "" {
		res["message"] = "platform异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	treePath = params.Get("tree_path")
	// cover
	f, _, err := c.Request.FormFile("icon")
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, f); err != nil {
		res["message"] = "icon异常 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if desc = params.Get("description"); desc == "" {
		res["message"] = "desc异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitPath = params.Get("git_path"); gitPath == "" {
		res["message"] = "gitPath异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if os := params.Get("owners"); os != "" {
		for _, o := range strings.Split(params.Get("owners"), ",") {
			if o != userName {
				owners = append(owners, o)
			}
		}
	}
	owners = append(owners, userName)
	// 数据平台AppId
	if datacenterAppId, err = strconv.ParseInt(params.Get("datacenter_app_id"), 10, 64); err != nil {
		datacenterAppId = 0
	}
	if isHost, err = strconv.ParseInt(params.Get("is_host"), 10, 64); err != nil {
		isHost = 0
	}
	c.JSON(nil, s.AppSvr.AppAdd(c, datacenterAppId, isHost, appID, appKey, mobiApp, platform, name, treePath, desc, gitPath, strings.Join(owners, ","), userName, buf.Bytes()))
}

func appKeys(c *bm.Context) {
	c.JSON(s.AppSvr.AppKeys(c))
}

func appAuditList(c *bm.Context) {
	c.JSON(s.AppSvr.AppAuditList(c))
}

func appAudit(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		id               int64
		appKey, reason   string
		status, isActive int
		err              error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status异常" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 1=通过  -2=驳回  3=注销
	if status != -3 && status != -2 && status != 1 {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status == -2 {
		if reason = params.Get("refusal_reason"); reason == "" {
			res["message"] = "请填写驳回理由"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 是否为恢复应用注销
	if isActive, err = strconv.Atoi(params.Get("is_active")); err != nil || isActive != 1 {
		isActive = 0
	}
	c.JSON(nil, s.AppSvr.AppAudit(c, appKey, reason, userName, status, isActive, id))
}

func appMailtoList(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey, funcModule string
		receiverType       int64
		err                error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if receiverType, err = strconv.ParseInt(params.Get("type"), 10, 64); err != nil {
		receiverType = 1
	}
	funcModule = params.Get("func_module")
	c.JSON(s.AppSvr.AppMailtoList(c, appKey, funcModule, receiverType))
}

func appMailtoUpdate(c *bm.Context) {
	var (
		params                          = c.Request.Form
		res                             = map[string]interface{}{}
		appKey, mailListStr, funcModule string
		receiverType                    int64
		err                             error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mailListStr = params.Get("mail_list"); mailListStr == "" {
		res["message"] = "mail_list 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if receiverType, err = strconv.ParseInt(params.Get("type"), 10, 64); err != nil {
		receiverType = 1
	}
	if funcModule = params.Get("func_module"); funcModule == "" {
		res["message"] = "func_module异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.UpdateAppMailtoList(c, appKey, mailListStr, funcModule, receiverType))
}

func system(c *bm.Context) {
	var (
		params   = c.Request.Form
		res      = map[string]interface{}{}
		platform string
	)
	if platform = params.Get("platform"); platform == "" {
		res["message"] = "platform异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.AppSvr.System(c, platform), nil)
}

func appRobotNotify(c *bm.Context) {
	var (
		params     = c.Request.Form
		res        = map[string]interface{}{}
		botId      int64
		webhookURL string
		msgType    string
		msgBody    string
		err        error
	)
	if botId, err = strconv.ParseInt(params.Get("bot_id"), 10, 64); err != nil {
		botId = 0
	}
	if msgType = params.Get("msg_type"); msgType == "" {
		res["message"] = "msg_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if msgBody = params.Get("msg_body"); msgBody == "" {
		res["message"] = "msg_body 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	webhookURL = params.Get("webhook_url")
	c.JSON(nil, s.AppSvr.RobotNotify(c, webhookURL, msgType, msgBody, botId))
}

func appWXAppNotify(c *bm.Context) {
	var (
		params                                 = c.Request.Form
		res                                    = map[string]interface{}{}
		appKeys, roles, content, assignedUsers string
		isTest                                 int
		err                                    error
	)
	if content = params.Get("content"); content == "" {
		res["message"] = "content 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if isTest, err = strconv.Atoi(params.Get("is_test")); err != nil {
		res["message"] = "is_test 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	appKeys = params.Get("app_keys")
	roles = params.Get("roles")
	assignedUsers = params.Get("assigned_users")
	c.JSON(nil, s.AppSvr.AppWXAppNotify(c, appKeys, roles, content, userName, assignedUsers, isTest))
}

func appWXAppPictureNotify(c *bm.Context) {
	var (
		params                        = c.Request.Form
		res                           = map[string]interface{}{}
		appKeys, roles, assignedUsers string
		isTest                        int
		err                           error
	)
	pic, picHeader, err := c.Request.FormFile("picture")
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if isTest, err = strconv.Atoi(params.Get("is_test")); err != nil {
		res["message"] = "is_test 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	appKeys = params.Get("app_keys")
	roles = params.Get("roles")
	assignedUsers = params.Get("assigned_users")
	c.JSON(nil, s.AppSvr.AppWXAppPictureNotify(c, appKeys, roles, userName, assignedUsers, pic, picHeader, isTest))
}

func appRobotSet(c *bm.Context) {
	var (
		params          = c.Request.Form
		res             = map[string]interface{}{}
		appKey          string
		robotName       string
		robotWebhookURL string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	robotName = params.Get("robot_name")
	robotWebhookURL = params.Get("robot_webhook_url")
	// 企业微信机器人名称存在的情况下， webhookUrl不可为空
	if robotName != "" && robotWebhookURL == "" {
		res["message"] = "robot_webhook_url异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppRobotSet(c, appKey, robotName, robotWebhookURL, userName))
}

func appRobotUpload(c *bm.Context) {
	var (
		res    = map[string]interface{}{}
		err    error
		file   multipart.File
		header *multipart.FileHeader
	)
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.AppSvr.UploadFile(c, "upload", file, header))
}

func appRobotList(c *bm.Context) {
	var (
		params                      = c.Request.Form
		res                         = map[string]interface{}{}
		botName, appKey, funcModule string
		state                       int
		err                         error
	)
	botName = params.Get("bot_name")
	appKey = params.Get("app_key")
	funcModule = params.Get("func_module")
	stateStr := params.Get("state")
	if stateStr == "" {
		state = -1
	} else if state, err = strconv.Atoi(stateStr); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppRobotList(c, appKey, funcModule, botName, userName, state))
}

func appRobotAdd(c *bm.Context) {
	var (
		params                                                    = c.Request.Form
		res                                                       = map[string]interface{}{}
		botName, appKeys, webhook, users, description, funcModule string
		state, isGlobal, isDefault                                int
		err                                                       error
	)
	if botName = params.Get("bot_name"); botName == "" {
		res["message"] = "bot_name 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if webhook = params.Get("webhook"); webhook == "" {
		res["message"] = "webhook 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if state, err = strconv.Atoi(params.Get("state")); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isGlobal, err = strconv.Atoi(params.Get("is_global")); err != nil {
		res["message"] = "is_global异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isDefault, err = strconv.Atoi(params.Get("is_default")); err != nil {
		res["message"] = "is_default异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	appKeys = params.Get("app_keys")
	funcModule = params.Get("func_module")
	users = params.Get("users")
	description = params.Get("description")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppRobotAdd(c, botName, webhook, appKeys, funcModule, users, description, userName, state, isGlobal, isDefault))
}

func appRobotUpdate(c *bm.Context) {
	var (
		params                                                    = c.Request.Form
		res                                                       = map[string]interface{}{}
		botName, appKeys, webhook, users, description, funcModule string
		state, isGlobal, isDefault                                int
		ID                                                        int64
		err                                                       error
	)
	if botName = params.Get("bot_name"); botName == "" {
		res["message"] = "bot_name 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if webhook = params.Get("webhook"); webhook == "" {
		res["message"] = "webhook 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if state, err = strconv.Atoi(params.Get("state")); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isGlobal, err = strconv.Atoi(params.Get("is_global")); err != nil {
		res["message"] = "is_global异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isDefault, err = strconv.Atoi(params.Get("is_default")); err != nil {
		res["message"] = "is_default异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	appKeys = params.Get("app_keys")
	funcModule = params.Get("func_module")
	users = params.Get("users")
	description = params.Get("description")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppRobotUpdate(c, botName, webhook, appKeys, funcModule, users, description, userName, state, isGlobal, isDefault, ID))
}

func appRobotDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		ID     int64
		err    error
	)
	if ID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
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
	c.JSON(nil, s.AppSvr.AppRobotDel(c, ID))
}

func appNotificationList(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, platform string
		state            int64
		err              error
	)
	appKey = params.Get("app_key")
	platform = params.Get("platform")
	if stateStr := params.Get("state"); stateStr == "" {
		state = -1
	} else if state, err = strconv.ParseInt(params.Get("state"), 10, 64); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.AppSvr.AppNotificationList(c, appKey, platform, state))
}

func appNotificationUpdate(c *bm.Context) {
	var (
		params                                                                    = c.Request.Form
		res                                                                       = map[string]interface{}{}
		appKeys, platform, routePath, title, content, url, effectTime, expireTime string
		id, state, isGlobal, closeable, showType                                  int64
		err                                                                       error
	)
	if isGlobal, err = strconv.ParseInt(params.Get("is_global"), 10, 64); err != nil {
		res["message"] = "is_global异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKeys = params.Get("app_keys"); isGlobal != 1 && appKeys == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if platform = params.Get("platform"); platform == "" {
		res["message"] = "platform异常"
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
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil || id == 0 {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if state, err = strconv.ParseInt(params.Get("state"), 10, 64); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if closeable, err = strconv.ParseInt(params.Get("closeable"), 10, 64); err != nil {
		res["message"] = "closeable异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if showType, err = strconv.ParseInt(params.Get("type"), 10, 64); err != nil {
		res["message"] = "type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routePath = params.Get("route_path"); routePath == "" {
		res["message"] = "route_path异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	effectTime = params.Get("effect_time")
	expireTime = params.Get("expire_time")
	url = params.Get("url")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppNotificationUpdate(c, id, appKeys, platform, routePath, title, content, url, state, isGlobal, showType, closeable, effectTime, expireTime, userName))
}

func appNotificationAdd(c *bm.Context) {
	var (
		params                                                                    = c.Request.Form
		res                                                                       = map[string]interface{}{}
		appKeys, platform, routePath, title, content, url, effectTime, expireTime string
		state, isGlobal, showType, closeable                                      int64
		err                                                                       error
	)
	if isGlobal, err = strconv.ParseInt(params.Get("is_global"), 10, 64); err != nil {
		res["message"] = "is_global异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKeys = params.Get("app_keys"); isGlobal != 1 && appKeys == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if platform = params.Get("platform"); platform == "" {
		res["message"] = "platform异常"
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
	if state, err = strconv.ParseInt(params.Get("state"), 10, 64); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if closeable, err = strconv.ParseInt(params.Get("closeable"), 10, 64); err != nil {
		res["message"] = "closeable异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if showType, err = strconv.ParseInt(params.Get("type"), 10, 64); err != nil {
		res["message"] = "type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if routePath = params.Get("route_path"); routePath == "" {
		res["message"] = "route_path异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	effectTime = params.Get("effect_time")
	expireTime = params.Get("expire_time")
	url = params.Get("url")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppNotificationAdd(c, appKeys, platform, routePath, title, content, url, state, isGlobal, showType, closeable, effectTime, expireTime, userName))
}

func appFileUpload(c *bm.Context) {
	var (
		err     error
		res     = map[string]interface{}{}
		fileURL string
		data    = new(struct {
			FileURL string `json:"file_url"`
		})
	)
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		res["message"] = "file异常" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, f); err != nil {
		res["message"] = "file异常：io.Copy error" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fileURL, err = s.AppSvr.Upload(c, model.BFSBucket, "", "", buf.Bytes()); err != nil {
		res["message"] = "文件上传失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	data.FileURL = fileURL
	c.JSON(data, nil)
}

func appMailConfigAdd(c *bm.Context) {
	var (
		params                                       = c.Request.Form
		res                                          = map[string]interface{}{}
		appKey, funcModule, host, address, pwd, name string
		port                                         int
		err                                          error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if funcModule = params.Get("func_module"); funcModule == "" {
		res["message"] = "func_module异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if host = params.Get("host"); host == "" {
		res["message"] = "host异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if port, err = strconv.Atoi(params.Get("port")); err != nil {
		res["message"] = "port异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if address = params.Get("address"); address == "" {
		res["message"] = "address异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pwd = params.Get("pwd"); pwd == "" {
		res["message"] = "pwd异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["name"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppMailConfigAdd(c, appKey, funcModule, host, address, pwd, name, userName, port))
}

func appMailConfigDel(c *bm.Context) {
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
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppMailConfigDel(c, id))
}

func appMailConfigUpdate(c *bm.Context) {
	var (
		params                                       = c.Request.Form
		res                                          = map[string]interface{}{}
		appKey, funcModule, host, address, pwd, name string
		id                                           int64
		port                                         int
		err                                          error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if funcModule = params.Get("func_module"); funcModule == "" {
		res["message"] = "func_module异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if host = params.Get("host"); host == "" {
		res["message"] = "host异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if port, err = strconv.Atoi(params.Get("port")); err != nil {
		res["message"] = "port异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if address = params.Get("address"); address == "" {
		res["message"] = "address异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pwd = params.Get("pwd"); pwd == "" {
		res["message"] = "pwd异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppMailConfigUpdate(c, appKey, funcModule, host, address, pwd, name, userName, port, id))
}

func appMailConfigList(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey, funcModule string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	funcModule = params.Get("func_module")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppMailConfigList(c, appKey, funcModule, userName))
}

func appMailList(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey, funcModule string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	funcModule = params.Get("func_module")
	c.JSON(s.AppSvr.AppMailList(c, appKey, funcModule))
}

func appTriggerPipeline(c *bm.Context) {
	var (
		params                      = c.Request.Form
		res                         = map[string]interface{}{}
		appKey, envVars, buildIDStr string
		buildID                     int64
		err                         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildIDStr = params.Get("build_id"); buildIDStr == "" {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(buildIDStr, 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	envVars = params.Get("env_var")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.AppSvr.AppTriggerPipeline(c, appKey, buildID, envVars, userName))
}

func appServicePing(c *bm.Context) {
	var (
		params     = c.Request.Form
		requestUrl string
		res        = map[string]interface{}{}
	)
	if requestUrl = params.Get("request_url"); requestUrl == "" {
		res["message"] = "request_url 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.ServicePing(c, requestUrl))
}
