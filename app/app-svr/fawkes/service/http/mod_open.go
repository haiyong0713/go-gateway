package http

import (
	"io/ioutil"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
)

func modOpenPoolList(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		AppKey     string `form:"app_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	pool, err := s.ModSvr.OpenPoolList(ctx, param.ModPoolKey, param.AppKey)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["pool"] = pool
	ctx.JSON(res, nil)
}

func modOpenModuleList(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		PoolID     int64  `form:"pool_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	module, err := s.ModSvr.OpenModuleList(ctx, param.ModPoolKey, param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["module"] = module
	ctx.JSON(res, nil)
}

func modOpenVersion(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	version, err := s.ModSvr.OpenVersion(ctx, param.ModPoolKey, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modOpenVersionList(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string  `form:"mod_pool_key" validate:"required"`
		ModuleID   int64   `form:"module_id" validate:"min=1"`
		Env        mod.Env `form:"env" validate:"required"`
		Pn         int64   `form:"pn" default:"1" validate:"min=1"`
		Ps         int64   `form:"ps" default:"20" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Env.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	version, page, polling, err := s.ModSvr.OpenVersionList(ctx, param.ModPoolKey, param.ModuleID, param.Env, param.Pn, param.Ps)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"], res["page"], res["polling"] = version, page, polling
	ctx.JSON(res, nil)
}

func modOpenPatchList(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	patch, err := s.ModSvr.OpenPatchList(ctx, param.ModPoolKey, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["patch"] = patch
	ctx.JSON(res, nil)
}

func modOpenModuleAdd(ctx *bm.Context) {
	param := new(struct {
		Username   string       `form:"username"`
		ModPoolKey string       `form:"mod_pool_key" validate:"required"`
		PoolID     int64        `form:"pool_id" validate:"min=1"`
		Name       string       `form:"name" validate:"required"`
		Remark     string       `form:"remark" validate:"required"`
		IsWIFI     bool         `form:"is_wifi"`
		Compress   mod.Compress `form:"compress" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Compress.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	if !nameValid(ctx, param.Name) {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	module, err := s.ModSvr.OpenModuleAdd(ctx, param.ModPoolKey, username, param.PoolID, param.Name, param.Remark, param.IsWIFI, param.Compress)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["module"] = module
	ctx.JSON(res, nil)
}

func modOpenModulePushOffline(ctx *bm.Context) {
	param := new(struct {
		Username   string `form:"username"`
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		ModuleID   int64  `form:"module_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	ctx.JSON(nil, s.ModSvr.OpenModulePushOffline(ctx, param.ModPoolKey, username, param.ModuleID))
}

func modOpenModuleState(ctx *bm.Context) {
	param := new(struct {
		Username   string          `form:"username"`
		ModPoolKey string          `form:"mod_pool_key" validate:"required"`
		ModuleID   int64           `form:"module_id" validate:"min=1"`
		State      mod.ModuleState `form:"state"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	ctx.JSON(nil, s.ModSvr.OpenModuleState(ctx, param.ModPoolKey, param.Username, param.ModuleID, param.State))
}

func modOpenVersionAdd(ctx *bm.Context) {
	param := new(struct {
		Username   string  `form:"username"`
		ModPoolKey string  `form:"mod_pool_key" validate:"required"`
		Env        mod.Env `form:"env" validate:"required"`
		ModuleID   int64   `form:"module_id" validate:"min=1"`
		Remark     string  `form:"remark" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Env.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	defer file.Close()
	filename := header.Filename
	if !filenameValid(ctx, filename) {
		return
	}
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if len(fileData) == 0 {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "空数据文件"))
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	version, err := s.ModSvr.OpenVersionAdd(ctx, param.ModPoolKey, username, param.ModuleID, param.Env, param.Remark, filename, fileData)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modOpenVersionRelease(ctx *bm.Context) {
	param := new(struct {
		Username   string `form:"username"`
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
		Released   bool   `form:"released"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	ctx.JSON(nil, s.ModSvr.OpenVersionRelease(ctx, param.ModPoolKey, username, param.VersionID, param.Released, time.Now()))
}

func modOpenVersionPush(ctx *bm.Context) {
	param := new(struct {
		Username   string `form:"username"`
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	version, err := s.ModSvr.OpenVersionPush(ctx, param.ModPoolKey, username, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modOpenVersionConfig(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	config, err := s.ModSvr.OpenVersionConfig(ctx, param.ModPoolKey, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["config"] = config
	ctx.JSON(res, nil)
}

func modOpenVersionConfigAdd(ctx *bm.Context) {
	param := new(struct {
		Username   string `form:"username"`
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		mod.ConfigParam
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if param.Etime != 0 && param.Etime < param.Stime {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "结束时间必须大于开始时间"))
		return
	}
	if !param.Priority.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	ctx.JSON(nil, s.ModSvr.OpenVersionConfigAdd(ctx, param.ModPoolKey, username, &param.ConfigParam))
}

func modOpenVersionGray(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		VersionID  int64  `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	gray, err := s.ModSvr.OpenVersionGray(ctx, param.ModPoolKey, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["gray"] = gray
	ctx.JSON(res, nil)
}

func modOpenVersionGrayAdd(ctx *bm.Context) {
	param := new(struct {
		Username   string `form:"username"`
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
		*mod.GrayParam
	})
	param.GrayParam = &mod.GrayParam{}
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username := param.Username
	if u, ok := ctx.Get("username"); ok {
		username = u.(string)
	}
	if username == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "username参数为空"))
		return
	}
	ctx.JSON(nil, s.ModSvr.OpenVersionGrayAdd(ctx, param.ModPoolKey, username, param.GrayParam))
}

func modOpenGrayWhitelistUpload(ctx *bm.Context) {
	param := new(struct {
		ModPoolKey string `form:"mod_pool_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	defer file.Close()
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	filename := header.Filename
	if !filenameValid(ctx, filename) {
		return
	}
	whitelistURL, err := s.ModSvr.OpenGrayWhitelistUpload(ctx, param.ModPoolKey, filename, fileData)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["whitelist_url"] = whitelistURL
	ctx.JSON(res, nil)
}
