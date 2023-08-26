package http

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	sagamdl "go-gateway/app/app-svr/fawkes/service/model/saga"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const _resignTag = "resign"

func appCIPackInfo(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		appKey  string
		buildID int64
		glJobID int64
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	buildID, _ = strconv.ParseInt(params.Get("build_id"), 10, 64)
	glJobID, _ = strconv.ParseInt(params.Get("gl_job_id"), 10, 64)
	if buildID == 0 && glJobID == 0 {
		res["message"] = "参数异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CiSvr.PackInfo(c, appKey, buildID, glJobID))
}

func buildPackList(c *bm.Context) {
	var (
		params                                    = c.Request.Form
		appKey, order, sort, gitKeyword, operator string
		pn, ps, pkgType, status, gitType          int
		res                                       = map[string]interface{}{}
		err                                       error
		gitlabJobID, ID                           int64
		hasBbrUrl                                 bool
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	if ps < 0 || ps > 20 {
		ps = 20
	}
	order = params.Get("order")
	if order != "version_code" && order != "version" && order != "mtime" {
		order = "id"
	}
	sort = params.Get("sort")
	if sort != "desc" && sort != "asc" {
		sort = "desc"
	}
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if statusStr := params.Get("status"); statusStr != "" {
		if status, err = strconv.Atoi(statusStr); err != nil {
			res["message"] = "status 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if gitTypeStr := params.Get("git_type"); gitTypeStr != "" {
		if gitType, err = strconv.Atoi(gitTypeStr); err != nil {
			res["message"] = "git_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	gitKeyword = params.Get("git_keyword")
	didPushCD := params.Get("push_cd")
	operator = params.Get("operator")
	if params.Get("bbr_url") != "" {
		hasBbrUrl, _ = strconv.ParseBool(params.Get("bbr_url"))
	}
	if gitlabJobIDStr := params.Get("gl_job_id"); gitlabJobIDStr != "" {
		if gitlabJobID, err = strconv.ParseInt(params.Get("gl_job_id"), 10, 64); err != nil {
			res["message"] = "gl_job_id 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if idStr := params.Get("id"); idStr != "" {
		if ID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
			res["message"] = "id 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	c.JSON(s.CiSvr.BuildPackList(c, appKey, pn, ps, pkgType, status, gitType, gitKeyword, operator, order, sort, gitlabJobID, ID, didPushCD, hasBbrUrl))
}

func recordBuildPack(c *bm.Context) {
	var (
		params                                                 = c.Request.Form
		appKey, gitName, commit, version, operator             string
		pkgType, gitType                                       int
		versionCode, gitlabJobID, buildID, internalVersionCode int64
		res                                                    = map[string]interface{}{}
		err                                                    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitlabJobID, err = strconv.ParseInt(params.Get("gl_job_id"), 10, 64); err != nil {
		res["message"] = "gl_job_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "pkg_type 为空"
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
	if commit = params.Get("commit"); commit == "" {
		res["message"] = "commit 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		res["message"] = "version 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if versionCode, err = strconv.ParseInt(params.Get("version_code"), 10, 64); err != nil {
		res["message"] = "version_code 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 不传内部版本号，则默认和版本号相同
	if internalVersionCode, err = strconv.ParseInt(params.Get("internal_version_code"), 10, 64); err != nil {
		internalVersionCode = versionCode
	}
	if operator = params.Get("operator"); operator == "" {
		res["message"] = "operator 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	operator = strings.TrimSuffix(operator, "@bilibili.com")
	if buildID, err = s.CiSvr.RecordBuildPack(c, appKey, gitlabJobID, pkgType, gitType, gitName, commit, version,
		versionCode, internalVersionCode, operator); err != nil {
		c.JSON(nil, err)
	} else {
		res["id"] = strconv.FormatInt(buildID, 10)
		c.JSON(res, nil)
	}
}

func createBuildPack(c *bm.Context) {
	var (
		params                                         = c.Request.Form
		appKey, gitName, description, webhookURL, send string
		pkgType, resignPkgType, gitType                int
		res                                            = map[string]interface{}{}
		err                                            error
		buildID, resignBuildID, depGitlabJobId         int64
		shouldNotify                                   bool
		ciEnvVar                                       string
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
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "pkg_type 为空"
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
	if get := params.Get("dep_gitlab_job_id"); get != "" {
		if depGitlabJobId, err = strconv.ParseInt(get, 10, 64); err != nil {
			log.Error("parse dep_gitlab_job_id[%s] error:%v", get, err)
			return
		}
	}
	description = params.Get("description")
	webhookURL = params.Get("webhook_url")
	shouldNotify, _ = strconv.ParseBool(c.Request.FormValue("notify_group"))
	var variables = map[string]string{
		"APP_KEY":               appKey,
		"PKG_TYPE":              strconv.Itoa(pkgType),
		"FAWKES":                "1",
		"FAWKES_USER":           userName,
		"TASK":                  "pack",
		"TRIBE_HOST_BBR_JOB_ID": "",
		"TRIBE_PRE_BBR_JOB_ID":  params.Get("dep_gitlab_job_id"),
	}
	var envVarMap = make(map[string]string)
	if ciEnvVar = params.Get("env_var"); ciEnvVar != "" {
		if err = json.Unmarshal([]byte(ciEnvVar), &envVarMap); err != nil {
			res["message"] = fmt.Sprintf("createBuildPack json.Unmarshal(%s) error(%v)", ciEnvVar, err)
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		for key, value := range envVarMap {
			variables[key] = value
		}
	}
	send = params.Get("send")
	if buildID, err = s.CiSvr.CreateBuildPack(c, appKey, send, pkgType, gitType, gitName, userName, ciEnvVar, description, webhookURL, shouldNotify, depGitlabJobId); err != nil {
		res["message"] = "CI 创建失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	variables["BUILD_ID"] = strconv.FormatInt(buildID, 10)
	if envVarMap["RESIGN_TASK"] == _resignTag {
		// 需重签
		if v, ok := variables["IPA_PATH"]; !ok || v == "" {
			// 新包
			resignPkgType, _ = strconv.Atoi(envVarMap["RESIGN_PKG_TYPE"])
			if resignBuildID, err = s.CiSvr.CreateBuildPack(c, appKey, send, resignPkgType, gitType, gitName, userName, ciEnvVar, fmt.Sprintf("【重签 FROM ID: %v】:%v", buildID, description), webhookURL, shouldNotify, depGitlabJobId); err != nil {
				res["message"] = "CI 创建失败"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			variables["RESIGN_BUILD_ID"] = strconv.FormatInt(resignBuildID, 10)
		} else {
			// 老包重签
			variables["RESIGN_BUILD_ID"] = strconv.FormatInt(buildID, 10)
			delete(variables, "TASK")
			// 旧包重签 因为原来的分支可能被删除 所以pipeline统一走master分支， 但是数据库存入的还是原来的分支
			gitName = "master"
		}
	}
	c.JSON(s.GitSvr.TriggerPipeline(context.Background(), appKey, gitType, gitName, variables))
}

func createBuildPackCommon(c *bm.Context) {
	var (
		params                                         = c.Request.Form
		appKey, gitName, description, webhookURL, send string
		pkgType, gitType                               int
		buildId                                        int64
		res                                            = map[string]interface{}{}
		err                                            error
		shouldNotify                                   bool
		ciEnvVar                                       string
		depGitlabJobId                                 int64
		triggerTribeIds                                []int64
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
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "pkg_type 为空"
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
	description = params.Get("description")
	webhookURL = params.Get("webhook_url")
	shouldNotify, _ = strconv.ParseBool(c.Request.FormValue("notify_group"))
	ciEnvVar = params.Get("env_var")
	if get := params.Get("dep_gitlab_job_id"); get != "" {
		if depGitlabJobId, err = strconv.ParseInt(get, 10, 64); err != nil {
			log.Error("parse dep_gitlab_job_id[%s] error:%v", get, err)
			return
		}
	}
	if tribeIds := params.Get("trigger_tribe_id"); tribeIds != "" {
		for _, v := range strings.Split(tribeIds, cimdl.Comma) {
			id, _ := strconv.ParseInt(v, 10, 64)
			triggerTribeIds = append(triggerTribeIds, id)
		}
	}
	send = params.Get("send")
	buildId, _, err = s.CiSvr.CreateBuildPackCommon(c, appKey, send, pkgType, gitType, gitName, userName, ciEnvVar, description, webhookURL, shouldNotify, depGitlabJobId, triggerTribeIds)
	if err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, err)
		return
	}
	c.JSON(struct {
		BuildID int64 `json:"id"`
	}{buildId}, nil)
}

func updateBuildPackInfo(c *bm.Context) {
	var (
		params                                                 = c.Request.Form
		buildID, gitlabJobID, versionCode, internalVersionCode int64
		commit, version                                        string
		res                                                    = map[string]interface{}{}
		err                                                    error
	)
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gitlabJobID, err = strconv.ParseInt(params.Get("gl_job_id"), 10, 64); err != nil {
		res["message"] = "gl_job_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if commit = params.Get("commit"); commit == "" {
		res["message"] = "commit 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	version = params.Get("version")
	if versionCode, err = strconv.ParseInt(params.Get("version_code"), 10, 64); err != nil {
		versionCode = 0
	}
	if internalVersionCode, err = strconv.ParseInt(params.Get("internal_version_code"), 10, 64); err != nil {
		internalVersionCode = versionCode
	}
	c.JSON(nil, s.CiSvr.UpdateBuildPackBase(c, buildID, gitlabJobID, commit, version, versionCode, internalVersionCode))
}

func updateBuildPackStatus(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		buildID int64
		status  int
		err     error
	)
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		res["message"] = "status 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.UpdateBuildPackStatus(c, buildID, status))
}

func uploadBuildPack(c *bm.Context) {
	var (
		fmd5, pkgName, mappingName, rName, rMappingName, bbrName string
		buildID                                                  int64
		unzip                                                    bool
		res                                                      = map[string]interface{}{}
		file                                                     multipart.File
		header                                                   *multipart.FileHeader
		err                                                      error
	)
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
	changeLog := c.Request.FormValue("change_log")
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
		mappingName = c.Request.FormValue("mapping_name")
		rName = c.Request.FormValue("r_name")
		rMappingName = c.Request.FormValue("r_mapping_name")
	} else {
		pkgName = header.Filename
		mappingName = ""
		rName = ""
		rMappingName = ""
	}
	bbrName = c.Request.FormValue("bbr_name")
	// 获取子仓Commit
	var subRepoCommitList []*cimdl.BuildPackSubRepo
	_ = json.Unmarshal([]byte(c.Request.FormValue("subrepo_commits")), &subRepoCommitList)
	c.JSON(nil, s.CiSvr.UploadBuildPack(c, buildID, file, header, unzip, pkgName, mappingName, bbrName, rName, rMappingName, changeLog, subRepoCommitList))
}

func uploadBuildFile(c *bm.Context) {
	var (
		appKey, fmd5 string
		jobID        int64
		unzip        bool
		res          = map[string]interface{}{}
		file         multipart.File
		header       *multipart.FileHeader
		err          error
	)
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if jobID, err = strconv.ParseInt(c.Request.FormValue("job_id"), 10, 64); err != nil {
		res["message"] = "job_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = c.Request.FormValue("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
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
	if unzip, err = strconv.ParseBool(c.Request.FormValue("unzip")); err != nil {
		unzip = false
	}
	c.JSON(nil, s.CiSvr.UploadBuildFile(c, appKey, jobID, file, header, unzip))
}

func uploadMobileEPBusiness(c *bm.Context) {
	var (
		business, appKey, fmd5, dirname string
		unzip                           bool
		res                             = map[string]interface{}{}
		file                            multipart.File
		header                          *multipart.FileHeader
		err                             error
	)
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if business = c.Request.FormValue("business"); business == "" {
		res["message"] = "business 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = c.Request.FormValue("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fmd5 = c.Request.FormValue("md5"); fmd5 == "" {
		res["message"] = "md5 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if dirname = c.Request.FormValue("dirname"); dirname == "" {
		res["message"] = "dirname 为空"
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
	if unzip, err = strconv.ParseBool(c.Request.FormValue("unzip")); err != nil {
		unzip = false
	}
	c.JSON(nil, s.CiSvr.UploadMobileEPBusiness(c, appKey, business, dirname, fmd5, file, header, unzip))
}

func cancelBuildPack(c *bm.Context) {
	var (
		params  = c.Request.Form
		buildID int64
		res     = map[string]interface{}{}
		err     error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelJob(context.Background(), buildID)
	}()
	c.JSON(nil, s.CiSvr.CancelBuildPack(c, buildID))
}

func deleteBuildPack(c *bm.Context) {
	var (
		params  = c.Request.Form
		buildID int64
		res     = map[string]interface{}{}
		err     error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelJob(context.Background(), buildID)
	}()
	c.JSON(nil, s.CiSvr.DeleteBuildPack(c, buildID))
}

func notifyGroup(c *bm.Context) {
	var (
		params        = c.Request.Form
		buildID       int64
		notifyCCGroup bool
		bots          string
		err           error
		res           = map[string]interface{}{}
	)
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	bots = params.Get("bots")
	notifyCCGroup, _ = strconv.ParseBool(c.Request.FormValue("notify_group"))
	c.JSON(nil, s.CiSvr.NotifyGroup(c, buildID, notifyCCGroup, bots))
}

// sendMail send mail
func sendmail(c *bm.Context) {
	req := c.Request
	res := map[string]interface{}{}
	res["message"] = "success"
	var attach *mailmdl.Attach
	// 附件
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		attach = &mailmdl.Attach{}
		attach.Name = header.Filename
		attach.File = file
		unzip := c.Request.Form.Get("unzip")
		if unzip != "" && unzip != "0" {
			attach.ShouldUnzip = true
		} else {
			attach.ShouldUnzip = false
		}
	}
	var bs []byte
	if attach == nil {
		bs, err = ioutil.ReadAll(req.Body)
	} else {
		// 使用 multipart 上传附件时，body 并不是 json，因此原来的 json 放在 form 的 json_body 中
		jsonBody := c.Request.Form.Get("json_body")
		bs = []byte(jsonBody)
	}
	if err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		res["message"] = fmt.Sprintf("%v", err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	req.Body.Close()
	// params
	var m = &mailmdl.Mail{}
	if err = json.Unmarshal(bs, m); err != nil {
		log.Error("http sendmail() json.Unmarshal(%s) error(%v)", string(bs), err)
		res["message"] = fmt.Sprintf("%v", err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if m.Subject == "" {
		res["message"] = "mail title can not be empty!"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if m.Body == "" {
		res["message"] = "mail content can not be empty!"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if len(m.ToAddresses) == 0 {
		res["message"] = "mail address list can not be empty!"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var attribution = &mailmdl.Attribution{}
	if err = json.Unmarshal(bs, attribution); err != nil {
		res["message"] = "参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = s.CiSvr.SendMail(c, attribution.AppKey, attribution.FuncModule, m, attach); err != nil {
		res["message"] = fmt.Sprintf("%v", err)
		c.JSONMap(res, err)
		return
	}
	c.JSONMap(res, nil)
}

func updateTestStatus(c *bm.Context) {
	pack := cimdl.BuildPack{}
	if err := c.BindWith(&pack, binding.JSON); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, s.CiSvr.UpdateTestStatus(c, &pack))
}

func repoPushHook(c *bm.Context) {
	var (
		err                error
		appKey, branchName string
		hookPush           = &sagamdl.HookPush{}
		params             = c.Request.Form
		res                = map[string]interface{}{}
	)
	if err = c.BindWith(hookPush, binding.JSON); err != nil {
		log.Error("Push hook error: %v", err)
		return
	}
	// 忽略由 fawkes 账号联动 Merge 产生的子仓 Push events，防止一个联动 Merge 使得多个子仓 Push events 触发多次主仓 pipeline
	if hookPush.UserUserName == "fawkes1" {
		return
	}
	// 忽略删除远程分支引起的 push events，防止子仓 Merge 后自动删除分支的 events 触发主仓 pipeline
	if hookPush.CheckoutSHA == "" {
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	branchName = strings.TrimPrefix(hookPush.Ref, "refs/heads/")
	c.JSON(nil, s.GitSvr.TriggerBuild(c, appKey, branchName, hookPush.UserUserName, hookPush.Project.GitSSHURL, hookPush.CheckoutSHA, "", ""))
}

func subRepoMRHook(c *bm.Context) {
	var (
		err    error
		appKey string
		hookMR = &sagamdl.HookMR{}
		params = c.Request.Form
		res    = map[string]interface{}{}
	)
	if err = c.BindWith(hookMR, binding.JSON); err != nil {
		log.Error("Sub repo MR hook error")
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.SubRepoMRHook(c, appKey, hookMR))
}

func mainRepoMRHook(c *bm.Context) {
	var (
		err    error
		appKey string
		hookMR = &sagamdl.HookMR{}
		params = c.Request.Form
		res    = map[string]interface{}{}
	)
	if err = c.BindWith(hookMR, binding.JSON); err != nil {
		log.Error("Main repo MR hook error")
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.MainRepoMRHook(c, appKey, hookMR))
}

func mainRepoCommentHook(c *bm.Context) {
	var (
		err         error
		appKey      string
		HookComment = &sagamdl.HookComment{}
		params      = c.Request.Form
		res         = map[string]interface{}{}
	)
	if err = c.BindWith(HookComment, binding.JSON); err != nil {
		log.Error("Main repo Comment hook error")
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.MainRepoCommentHook(c, appKey, HookComment))
}

func releaseBranchHook(c *bm.Context) {
	var (
		err      error
		appKey   string
		hookPush = &sagamdl.HookPush{}
		params   = c.Request.Form
		res      = map[string]interface{}{}
	)
	if err = c.BindWith(hookPush, binding.JSON); err != nil {
		log.Error("Main repo Comment hook error")
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.MainRepoReleaseBranchHook(c, appKey, hookPush))
}

func mainRepoRebuild(c *bm.Context) {
	var (
		appKey, arch, operator, commit string
		params                         = c.Request.Form
		res                            = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if commit = params.Get("commit"); commit == "" {
		res["message"] = "commit 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	operator = params.Get("operator")
	arch = params.Get("arch")
	var variables = map[string]string{
		"APP_KEY":          appKey,
		"FAWKES":           "1",
		"FAWKES_USER":      operator,
		"TASK":             "build",
		"MAIN_REPO_COMMIT": commit,
		"PUT_CACHE_ALL":    "1",
	}
	if arch != "" {
		variables["ARCH"] = arch
	}
	c.JSON(s.GitSvr.TriggerPipeline(c, appKey, 2, commit, variables))
}

func mainRepoBuild(c *bm.Context) {
	var (
		appKey, operator, branch string
		params                   = c.Request.Form
		res                      = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if branch = params.Get("branch"); branch == "" {
		res["message"] = "branch 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	operator = params.Get("operator")
	var variables = map[string]string{
		"APP_KEY":     appKey,
		"FAWKES":      "1",
		"FAWKES_USER": operator,
		"TASK":        "build",
	}
	c.JSON(s.GitSvr.TriggerPipeline(c, appKey, 0, branch, variables))
}

func pipelineStatus(c *bm.Context) {
	var (
		err        error
		appKey     string
		pipelineID int
		params     = c.Request.Form
		res        = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pipelineID, err = strconv.Atoi(params.Get("pipeline_id")); err != nil {
		res["message"] = "pipeline_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.GitSvr.PipelineStatus(c, appKey, pipelineID))
}

func checkoutBranch(c *bm.Context) {
	var (
		err       error
		repoID    int
		srcBranch string
		tgtBranch string
		params    = c.Request.Form
		res       = map[string]interface{}{}
	)
	if repoID, err = strconv.Atoi(params.Get("repo_id")); err != nil {
		res["message"] = "repo_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if srcBranch = params.Get("src_branch"); srcBranch == "" {
		srcBranch = "master"
	}
	if tgtBranch = params.Get("tgt_branch"); tgtBranch == "" {
		res["message"] = "tgt_branch 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.GitSvr.CheckoutBranch(c, repoID, srcBranch, tgtBranch))
}

func lintMRHook(c *bm.Context) {
	var (
		err    error
		appKey string
		hookMR = &sagamdl.HookMR{}
		params = c.Request.Form
		res    = map[string]interface{}{}
	)
	if err = c.BindWith(hookMR, binding.JSON); err != nil {
		log.Error("Main repo MR hook error")
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 只有创建 MR 并且不为 WIP 状态才触发
	if hookMR.ObjectAttributes.Action != "open" || hookMR.ObjectAttributes.WorkInProgress {
		return
	}
	var commit = hookMR.ObjectAttributes.LastCommit
	var variables = map[string]string{
		"APP_KEY":               appKey,
		"FAWKES":                "1",
		"TASK":                  "lint",
		"MR_OPERATOR":           hookMR.User.UserName,
		"MR_LAST_COMMIT_ID":     commit.ID,
		"MR_LAST_COMMIT_MSG":    commit.Message,
		"MR_LAST_COMMIT_TS":     commit.Timestamp,
		"MR_LAST_COMMIT_AUTHOR": commit.Author.Name,
		"MR_LAST_COMMIT_EMAIL":  commit.Author.Email,
		"MR_SOURCE_BRANCH":      hookMR.ObjectAttributes.SourceBranch,
		"MR_TARGET_BRANCH":      hookMR.ObjectAttributes.TargetBranch,
		"MR_STATE":              hookMR.ObjectAttributes.State,
		"MR_TITLE":              hookMR.ObjectAttributes.Title,
		"MR_DESCRIPTION":        hookMR.ObjectAttributes.Description,
		"MR_IID":                strconv.FormatInt(hookMR.ObjectAttributes.IID, 10),
		"MR_URL":                hookMR.ObjectAttributes.URL,
		"MR_ACTION":             hookMR.ObjectAttributes.Action,
		"MR_WIP":                strconv.FormatBool(hookMR.ObjectAttributes.WorkInProgress),
	}
	if hookMR.Assignee != nil {
		variables["MR_ASSIGNEE"] = hookMR.Assignee.UserName
	}
	c.JSON(s.GitSvr.TriggerPipeline(c, appKey, 0, hookMR.ObjectAttributes.SourceBranch, variables))
}

func publishDepandency(c *bm.Context) {
	var (
		params                             = c.Request.Form
		appKey, gitName, versionCode, send string
		pkgType, gitType                   int
		res                                = map[string]interface{}{}
		err                                error
		buildID, depGitlabJobId            int64
		shouldNotify                       bool
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
	if versionCode = params.Get("version_code"); versionCode == "" {
		res["message"] = "version_code 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "pkg_type 为空"
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
	shouldNotify, _ = strconv.ParseBool(c.Request.FormValue("notify_group"))
	if get := params.Get("dep_gitlab_job_id"); get != "" {
		if depGitlabJobId, err = strconv.ParseInt(get, 10, 64); err != nil {
			log.Error("parse dep_gitlab_job_id[%s] error:%v", get, err)
			return
		}
	}
	send = params.Get("send")
	if buildID, err = s.CiSvr.CreateBuildPack(c, appKey, send, pkgType, gitType, gitName, userName, "", "", "", shouldNotify, depGitlabJobId); err != nil {
		res["message"] = "CI 创建失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var variables = map[string]string{
		"APP_KEY":      appKey,
		"PKG_TYPE":     strconv.Itoa(pkgType),
		"BUILD_ID":     strconv.FormatInt(buildID, 10),
		"FAWKES":       "1",
		"FAWKES_USER":  userName,
		"VERSION_CODE": versionCode,
		"TASK":         "publish",
	}
	c.JSON(s.GitSvr.TriggerPipeline(context.Background(), appKey, gitType, gitName, variables))
}

func packReportInfo(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		jobID  int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if jobID, err = strconv.ParseInt(params.Get("job_id"), 10, 64); err != nil {
		res["message"] = "job_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CiSvr.PackReportInfo(c, appKey, jobID))
}

func ciCrontabList(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		pn, ps int
		res    = map[string]interface{}{}
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil || ps < 1 {
		ps = 20
	}
	c.JSON(s.CiSvr.CrontabList(c, appKey, pn, ps))
}

func ciCrontabAdd(c *bm.Context) {
	var (
		params                                                = c.Request.Form
		appKey, gitName, stime, tick, send, envVars, userName string
		res                                                   = map[string]interface{}{}
		gitType, pkgType                                      int
		err                                                   error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if stime = params.Get("stime"); stime == "" {
		res["message"] = "stime 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if tick = params.Get("tick"); tick == "" {
		res["message"] = "tick 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if _, err = time.ParseDuration(tick); err != nil {
		res["message"] = "tick 参数错误"
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
		gitType = 0
	}
	if gitName = params.Get("git_name"); gitName == "" {
		res["message"] = "git_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pkgTypeStr := params.Get("pkg_type"); pkgTypeStr != "" {
		if pkgType, err = strconv.Atoi(pkgTypeStr); err != nil {
			res["message"] = "pkg_type 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		pkgType = 0
	}
	send = params.Get("send")
	envVars = params.Get("ci_env_vars")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CiSvr.CrontabAdd(c, appKey, stime, tick, gitType, gitName, pkgType, send, envVars, userName))
}

func ciCrontabStatus(c *bm.Context) {
	var (
		params = c.Request.Form
		id     int64
		status int
		res    = map[string]interface{}{}
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil || (status != cimdl.CronStop && status != cimdl.CronWait) {
		res["message"] = "status 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CiSvr.CrontabStatus(c, id, status))
}

func ciCrontabDel(c *bm.Context) {
	var (
		params = c.Request.Form
		id     int64
		res    = map[string]interface{}{}
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CiSvr.CrontabDel(c, id))
}

func getMonkeyList(c *bm.Context) {
	var (
		params  = c.Request.Form
		appKey  string
		buildId int64
		pn, ps  int
		res     = map[string]interface{}{}
		err     error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildId, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn < 1 {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil || ps < 1 {
		ps = 20
	}
	c.JSON(s.CiSvr.GetMonkeyList(c, appKey, buildId, pn, ps))
}

func addMonkey(c *bm.Context) {
	var (
		params                   = c.Request.Form
		appKey, osver, schemeUrl string
		buildId                  int64
		execDuration             int
		res                      = map[string]interface{}{}
		err                      error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if osver = params.Get("osver"); osver == "" {
		res["message"] = "osver异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildId, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if execDuration, err = strconv.Atoi(params.Get("exec_duration")); err != nil {
		res["message"] = "exec_duration异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	schemeUrl = params.Get("scheme_url")
	c.JSON(nil, s.CiSvr.AddMonkey(c, appKey, osver, schemeUrl, userName, userName, buildId, execDuration))
}

func updateMonkeyStatus(c *bm.Context) {
	var (
		res        = map[string]interface{}{}
		bs         []byte
		externalID int64
		err        error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	var epParams *cimdl.EPMonkeyCallbackBody
	if err = json.Unmarshal(bs, &epParams); err != nil {
		log.Error("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if epParams.AppKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if externalID, err = strconv.ParseInt(epParams.ExternalID, 10, 64); err != nil {
		res["message"] = "external_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CiSvr.UpdateMonkeyStatus(c, epParams.AppKey, epParams.ReportUrl, epParams.Emulator, externalID, epParams.Status))
}

func ciEnvList(c *bm.Context) {
	var (
		params                   = c.Request.Form
		envKey, appKey, platform string
	)
	appKey = params.Get("app_key")
	platform = params.Get("platform")
	envKey = params.Get("env_key")
	c.JSON(s.CiSvr.CiEnvList(c, envKey, appKey, platform))
}

func addCiEnv(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		envKey, envValues string
		envType           int
		err               error
	)
	if envKey = params.Get("env_key"); envKey == "" {
		res["message"] = "env_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if envValues = params.Get("env_values"); envValues == "" {
		res["message"] = "env_values 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if envType, err = strconv.Atoi(params.Get("env_type")); err != nil {
		res["message"] = "env_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CiSvr.AddCiEnv(c, envKey, envValues, userName, envType))
}

func UpdateCiEnv(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		envKey, envValues string
		envType           int
		err               error
	)
	if envKey = params.Get("env_key"); envKey == "" {
		res["message"] = "env_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if envValues = params.Get("env_values"); envValues == "" {
		res["message"] = "env_values 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if envType, err = strconv.Atoi(params.Get("env_type")); err != nil {
		res["message"] = "env_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CiSvr.UpdateCiEnv(c, envKey, envValues, userName, envType))
}

func DeleteCiEnv(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		id     int64
		envKey string
		err    error
	)
	idStr := params.Get("id")
	envKey = params.Get("env_key")
	if idStr == "" && envKey == "" {
		res["message"] = "id和env_key不能同时为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if idStr != "" {
		if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
			res["message"] = "id 异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	c.JSON(nil, s.CiSvr.DeleteCiEnv(c, id, envKey))
}

func DeleteCiEnvByAppKey(c *bm.Context) {
	var (
		params          = c.Request.Form
		res             = map[string]interface{}{}
		envKey, appKeys string
	)
	if envKey = params.Get("env_key"); envKey == "" {
		res["message"] = "env_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKeys = params.Get("app_keys"); appKeys == "" {
		res["message"] = "app_keys 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.CiSvr.DeleteCiEnvByAppKey(c, envKey, appKeys, userName))
}

func parseBBR(c *bm.Context) {
	var (
		params                      = c.Request.Form
		buildId                     int64
		appKey, buildIdStr, feature string
		exFilter                    []string
		res                         = map[string]interface{}{}
		err                         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildIdStr = params.Get("build_id"); buildIdStr == "" {
		res["message"] = "build_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if exFilterStr := params.Get("exclude_filter"); exFilterStr != "" {
		exFilter = strings.Split(exFilterStr, cimdl.Comma)
	}
	if buildId, err = strconv.ParseInt(buildIdStr, 10, 64); err != nil {
		res["message"] = "build_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	feature = params.Get("feature")
	c.JSON(s.CiSvr.ParseBBR(c, appKey, buildId, feature, exFilter))
}

func ciJobRecord(c *bm.Context) {
	var (
		err    error
		params = &cimdl.JobRecordParam{}
		res    = map[string]interface{}{}
	)
	if err = c.Bind(params); err != nil {
		return
	}
	if params.JobName == "" {
		res["message"] = "job_name 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CiSvr.RecordCIJob(c, params))
}

func ciJobInfo(c *bm.Context) {
	var (
		params   = c.Request.Form
		typeName string
		res      = map[string]interface{}{}
	)
	if typeName = params.Get("type"); typeName == "" {
		res["message"] = "type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CiSvr.CIJobInfo(c, typeName))
}

func ciCompileRecord(c *bm.Context) {
	var (
		err    error
		params = &cimdl.CICompileRecordParam{}
		res    = map[string]interface{}{}
	)
	if err = c.Bind(params); err != nil {
		return
	}
	if params.BuildLogURL == "" {
		res["message"] = "build_log_url 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.PkgType == 0 {
		res["message"] = "pkg_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.BuildEnv == 0 {
		res["message"] = "build_env 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.StartTime == 0 {
		res["message"] = "start_time 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if params.EndTime == 0 {
		res["message"] = "end_time 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.CiSvr.RecordCICompile(c, params))
}

func buildPackSubRepoList(c *bm.Context) {
	var (
		params         = c.Request.Form
		appKey, commit string
		res            = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if commit = params.Get("commit"); commit == "" {
		res["message"] = "commit 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.CiSvr.BuildPackSubRepoList(c, appKey, commit))
}

func getAppBuildPackVersionInfo(c *bm.Context) {
	p := new(cimdl.GetAppBuildPackVersionInfoReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := s.CiSvr.GetAppBuildPackVersionInfo(c, p)
	c.JSON(resp, err)
}

func ciTrack(c *bm.Context) {
	p := new(cimdl.TrackMessage)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(nil, s.CiSvr.CITrack(c, p))
}
