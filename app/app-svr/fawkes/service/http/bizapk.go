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

	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"

	goGitlab "github.com/xanzy/go-gitlab"
)

func bizapkBuildsList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		packBuildID int64
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
	if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
		res["message"] = "pack_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BizapkSvr.BizApkBuildsList(c, appKey, env, packBuildID))
}

func bizApkList(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		packBuildID int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
		res["message"] = "pack_build_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); appKey == "" {
		res["message"] = "env 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BizapkSvr.BizApkList(c, appKey, packBuildID, env))
}

func bizApkSet(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		active, priority int
		settingsID       int64
		err              error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if settingsID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if activeStr := params.Get("active"); activeStr != "" {
		if active, err = strconv.Atoi(activeStr); err != nil {
			res["message"] = "active 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		err = s.BizapkSvr.SetBizApkActive(context.Background(), active, userName, settingsID)
	}
	if priorityStr := params.Get("priority"); priorityStr != "" {
		if priority, err = strconv.Atoi(priorityStr); err != nil {
			res["message"] = "priority 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		err = s.BizapkSvr.SetBizapkPriority(context.Background(), priority, userName, settingsID)
	}
	c.JSON(nil, err)
}

func bizApkAdd(c *bm.Context) {
	var (
		params                          = c.Request.Form
		res                             = map[string]interface{}{}
		appKey, name, gitName, buildEnv string
		packBuildID, buildID            int64
		gitType                         int
		pipeline                        *goGitlab.Pipeline
		err                             error
		packURLRes                      *bizapkmdl.OrgPackURLResp
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitTypeStr := params.Get("git_type"); gitTypeStr != "" {
		if gitType, err = strconv.Atoi(gitTypeStr); err != nil {
			res["message"] = "git_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "git_type 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitName = params.Get("git_name"); gitName == "" {
		res["message"] = "git_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
		res["message"] = "pack_build_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var variables = map[string]string{
		"APP_KEY":      appKey,
		"FAWKES":       "1",
		"FAWKES_USER":  userName,
		"TASK":         "biz_build",
		"BIZ_APK_NAME": name,
	}
	if ciInfo, err := s.CiSvr.PackInfo(c, appKey, 0, packBuildID); ciInfo != nil && err == nil {
		variables["PKG_TYPE"] = strconv.FormatInt(int64(ciInfo.PkgType), 10)
	} else {
		res["message"] = "构建号不存在"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var envVarMap = make(map[string]string)
	if buildEnv = params.Get("build_env"); buildEnv != "" {
		if err = json.Unmarshal([]byte(buildEnv), &envVarMap); err != nil {
			res["message"] = fmt.Sprintf("bizApkAdd json.Unmarshal(%s) error(%v)", string(buildEnv), err)
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		for key, value := range envVarMap {
			variables[key] = value
		}
	}
	if buildID, err = s.BizapkSvr.BizApkBuildCreate(c, appKey, name, packBuildID, gitType, gitName, userName); err != nil {
		res["message"] = "构建创建失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packURLRes, err = s.BizapkSvr.OrgPackBuildByJobID(c, appKey, packBuildID); err != nil {
		res["message"] = "未找到原始包"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	variables["APK_BUILD_ID"] = strconv.FormatInt(buildID, 10)
	variables["ORG_PACK_APK"] = packURLRes.PkgURL
	variables["ORG_PACK_MAPPING"] = packURLRes.MappingURL
	if pipeline, err = s.GitSvr.TriggerPipeline(context.Background(), appKey, gitType, gitName, variables); err != nil {
		res["message"] = "pipeline trigger 运行失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BizapkSvr.BizApkBuildUpdatePpl(c, pipeline.ID, pipeline.Sha, buildID))
}

func bizApkUpload(c *bm.Context) {
	var (
		params                                 = c.Request.Form
		fmd5, apk, mapping, meta, appKey, name string
		packBuildID, bizapkBuildID             int64
		res                                    = map[string]interface{}{}
		file                                   multipart.File
		header                                 *multipart.FileHeader
		priority, builtIn, needUploadToCDN     int
		err                                    error
	)
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
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
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		c.JSON(nil, err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	if fmd5 != hex.EncodeToString(md5Bs[:]) {
		res["message"] = "md5 校验错误"
		c.JSON(res, ecode.RequestErr)
		return
	}
	if apk = c.Request.FormValue("apk"); apk == "" {
		res["message"] = "需指定 apk 文件相对路径"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if builtIn, _ = strconv.Atoi(params.Get("built_in")); builtIn != 1 {
		builtIn = 0
	}
	mapping = c.Request.FormValue("mapping")
	meta = c.Request.FormValue("meta")
	bizapkBuildID, _ = strconv.ParseInt(params.Get("id"), 10, 64)
	if needUploadToCDN, err = strconv.Atoi(params.Get("need_upload_to_cdn")); err != nil {
		needUploadToCDN = bizapkmdl.NEED_UPLOAD
	}
	// apk build id 无值则不是从 web 前端触发，而是直接从 CI 上传的第一版，需要补全 name，构建号，版本等信息
	if bizapkBuildID == 0 {
		if appKey = c.Request.FormValue("app_key"); appKey == "" {
			res["message"] = "app_key 为空"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if name = c.Request.FormValue("name"); name == "" {
			res["message"] = "name 为空"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
			res["message"] = "pack_build_id 异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if priority, err = strconv.Atoi(params.Get("priority")); err != nil {
			res["message"] = "priority 异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		c.JSON(s.BizapkSvr.UploadBizApkBuildFromCI(context.Background(), file, header, apk, mapping, meta, appKey, name, priority, builtIn, packBuildID))
		return
	}
	c.JSON(s.BizapkSvr.UploadBizApkBuildFromCD(context.Background(), file, header, apk, mapping, meta, bizapkBuildID, builtIn, needUploadToCDN))
}

func bizApkUpdate(c *bm.Context) {
	var (
		params                     = c.Request.Form
		res                        = map[string]interface{}{}
		gitlabJobID, bizapkBuildID int64
		err                        error
	)
	if gitlabJobID, err = strconv.ParseInt(params.Get("gl_job_id"), 10, 64); err != nil {
		res["message"] = "gl_job_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if bizapkBuildID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BizapkSvr.UpdateBizApkBuildInfo(c, gitlabJobID, bizapkBuildID))
}

func bizApkEvolution(c *bm.Context) {
	var (
		params        = c.Request.Form
		bizapkBuildID int64
		res           = map[string]interface{}{}
		err           error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if bizapkBuildID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.BizapkSvr.BizApkBuildEvolution(c, bizapkBuildID, userName))
}

func bizApkCancel(c *bm.Context) {
	var (
		params        = c.Request.Form
		bizapkBuildID int64
		res           = map[string]interface{}{}
		err           error
	)
	if bizapkBuildID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelBizApkJob(context.Background(), bizapkBuildID)
	}()
	c.JSON(nil, s.BizapkSvr.BizApkBuildCancel(c, bizapkBuildID))
}

func bizApkDelete(c *bm.Context) {
	var (
		params        = c.Request.Form
		bizapkBuildID int64
		res           = map[string]interface{}{}
		err           error
	)
	if bizapkBuildID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelBizApkJob(context.Background(), bizapkBuildID)
	}()
	c.JSON(nil, s.BizapkSvr.BizApkBuildDelete(c, bizapkBuildID))
}

func bizApkFilterConfigSet(c *bm.Context) {
	var (
		params                                                           = c.Request.Form
		res                                                              = map[string]interface{}{}
		appKey, env, network, isp, channel, city, device, excludesSystem string
		percent, status                                                  int
		buildID                                                          int64
		err                                                              error
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
	if buildID, err = strconv.ParseInt(params.Get("apk_build_id"), 10, 64); err != nil {
		res["message"] = "apk_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	network = params.Get("network")
	isp = params.Get("isp")
	channel = params.Get("channel")
	excludesSystem = params.Get("excludes_system")
	if city = params.Get("city"); city == "" {
		city = "0"
	}
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
	c.JSON(nil, s.BizapkSvr.FilterConfigSet(c, appKey, env, buildID, network, isp, channel, city, excludesSystem, percent, device, userName, status))
}

func bizApkFilterConfig(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		env     string
		buildID int64
		err     error
	)
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("apk_build_id"), 10, 64); err != nil {
		res["message"] = "apk_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BizapkSvr.FilterConfig(c, env, buildID))
}

func bizApkFlowConfigSet(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		appKey, env, flow  string
		packBuildID, apkID int64
		err                error
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
	if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
		res["message"] = "pack_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if apkID, err = strconv.ParseInt(params.Get("apk_id"), 10, 64); err != nil {
		res["message"] = "apk_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var fs map[int64]string
	if err = json.Unmarshal([]byte(flow), &fs); err != nil {
		res["message"] = "flow异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.BizapkSvr.FlowConfigSet(c, appKey, env, userName, packBuildID, apkID, fs))
}

func bizApkFlowConfig(c *bm.Context) {
	var (
		params             = c.Request.Form
		res                = map[string]interface{}{}
		env                string
		packBuildID, apkID int64
		err                error
	)
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if packBuildID, err = strconv.ParseInt(params.Get("pack_build_id"), 10, 64); err != nil {
		res["message"] = "pack_build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if apkID, err = strconv.ParseInt(params.Get("apk_id"), 10, 64); err != nil {
		res["message"] = "apk_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.BizapkSvr.FlowConfig(c, env, packBuildID, apkID))
}
