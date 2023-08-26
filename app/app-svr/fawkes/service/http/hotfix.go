package http

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func hotfixPushEnv(c *bm.Context) {
	var request = new(appmdl.HfPushEnvReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, s.AppSvr.HotfixPushEnv(c, request))
}

func hotfixConfSet(c *bm.Context) {
	var request = new(appmdl.HfConfSetReq)
	if err := c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.HotfixConfSet(c, request, userName))
}

func hotfixConfGet(c *bm.Context) {
	var request = new(appmdl.HfConfGetReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(s.AppSvr.HotfixConfGet(c, request))
}

func hotfixEffect(c *bm.Context) {
	var request = new(appmdl.HfEffectReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, s.AppSvr.HotfixEffect(c, request))
}

func hotfixList(c *bm.Context) {
	var request = new(appmdl.HfListReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(s.AppSvr.HotfixList(c, request))
}

func hotfixBuild(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		envVars      = make(map[string]string)
		request      = new(appmdl.HfBuildReq)
		patchBuildID int64
		err          error
	)
	if err = c.Bind(request); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	gitType := params.Get("git_type")
	if gitType == "" {
		res["message"] = "git_type为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var envVarStr string
	if envVarStr = params.Get("env_vars"); envVarStr != "" {
		if err = json.Unmarshal([]byte(envVarStr), &envVars); err != nil {
			log.Errorc(c, "unmarshal ci env vars error: [%v]", err)
			return
		}
	}
	if patchBuildID, err = s.AppSvr.HotfixBuild(c, request, userName, envVarStr); err != nil {
		res["message"] = "hotfix 创建失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	var variables = map[string]string{
		"APP_KEY":               request.AppKey,
		"PKG_TYPE":              "2",
		"BUILD_ID":              strconv.FormatInt(patchBuildID, 10),
		"FAWKES":                "1",
		"FAWKES_USER":           userName,
		"TASK":                  "hotfix",
		"INTERNAL_VERSION_CODE": strconv.FormatInt(request.InternalVersionCode, 10),
	}
	variables = combineMap(variables, envVars)
	if _, err = s.GitSvr.TriggerPipeline(context.Background(), request.AppKey, request.GitType, request.GitName, variables); err != nil {
		res["message"] = "hotfix trigger pipeline 失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, err)
}

func hotfixUpdate(c *bm.Context) {
	var request = new(appmdl.HfUpdateReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, s.AppSvr.HotfixUpdate(c, request))
}

func hotfixUpload(c *bm.Context) {
	var (
		request         = new(appmdl.HfUploadReq)
		header          *multipart.FileHeader
		file            multipart.File
		patchBuildID    int64
		patchName, fmd5 string
		err             error
		res             = map[string]interface{}{}
	)
	if file, header, err = c.Request.FormFile("file"); err != nil {
		res["message"] = "file 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if patchBuildID, err = strconv.ParseInt(c.Request.FormValue("patch_build_id"), 10, 64); err != nil {
		res["message"] = "patch_build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fmd5 = c.Request.FormValue("md5"); fmd5 == "" {
		res["message"] = "md5 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if patchName = c.Request.FormValue("patch_name"); patchName == "" {
		res["message"] = "patchName 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// md5 校验
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, file); err != nil {
		c.JSON(nil, err)
		return
	}
	// reset file pointer
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
	request.PatchBuildID = patchBuildID
	request.PatchName = patchName
	c.JSON(nil, s.AppSvr.HotfixUpload(c, request, file, header))
}

func hotfixOrigGet(c *bm.Context) {
	var request = new(appmdl.HfOriginInfoReq)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(s.AppSvr.HotfixOrigGet(c, request))
}

func hotfixCancel(c *bm.Context) {
	var request = new(appmdl.HfCancelReq)
	if err := c.Bind(request); err != nil {
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelHotfixJob(context.Background(), request.PatchBuildID)
	}()
	c.JSON(nil, s.AppSvr.HotfixCancel(c, request))
}

func hotfixDel(c *bm.Context) {
	var request = new(appmdl.HfDelReq)
	if err := c.Bind(request); err != nil {
		return
	}
	// nolint:biligowordcheck
	go func() {
		_ = s.GitSvr.CancelHotfixJob(context.Background(), request.PatchBuildID)
	}()
	c.JSON(nil, s.AppSvr.HotfixDel(c, request))
}

func combineMap(variables map[string]string, envVar map[string]string) map[string]string {
	for k, v := range envVar {
		variables[k] = v
	}
	return variables
}
