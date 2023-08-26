package http

import (
	"io/ioutil"
	"regexp"
	"time"

	"go-gateway/app/app-svr/fawkes/service/model/mod"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func modPoolList(ctx *bm.Context) {
	param := new(struct {
		AppKey string `form:"app_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	pool, err := s.ModSvr.PoolList(ctx, param.AppKey)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["pool"] = pool
	ctx.JSON(res, nil)
}

func modModuleList(ctx *bm.Context) {
	param := new(struct {
		PoolID int64 `form:"pool_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	module, err := s.ModSvr.ModuleList(ctx, param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["module"] = module
	ctx.JSON(res, nil)
}

func modVersionList(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64   `form:"module_id" validate:"min=1"`
		Env      mod.Env `form:"env" validate:"required"`
		Pn       int64   `form:"pn" default:"1" validate:"min=1"`
		Ps       int64   `form:"ps" default:"20" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Env.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	version, page, polling, err := s.ModSvr.VersionList(ctx, param.ModuleID, param.Env, param.Pn, param.Ps)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"], res["page"], res["polling"] = version, page, polling
	ctx.JSON(res, nil)
}

func modPatchList(ctx *bm.Context) {
	param := new(struct {
		VersionID int64   `form:"version_id" validate:"min=1"`
		Env       mod.Env `form:"env" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	patch, err := s.ModSvr.PatchList(ctx, param.VersionID, param.Env)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["patch"] = patch
	ctx.JSON(res, nil)
}

func modVersionAdd(ctx *bm.Context) {
	param := new(struct {
		Env      mod.Env `form:"env" validate:"required"`
		ModuleID int64   `form:"module_id" validate:"min=1"`
		Remark   string  `form:"remark" validate:"required"`
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
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, err.Error()))
		return
	}
	defer file.Close()
	filename := header.Filename
	if !filenameValid(ctx, filename) {
		return
	}
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, err.Error()))
		return
	}
	if len(fileData) == 0 {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "空数据文件"))
		return
	}
	username, _ := ctx.Get("username")
	version, err := s.ModSvr.VersionAdd(ctx, username.(string), param.ModuleID, param.Env, param.Remark, filename, fileData)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modVersionRelease(ctx *bm.Context) {
	param := new(struct {
		VersionID int64 `form:"version_id" validate:"min=1"`
		Released  bool  `form:"released"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionRelease(ctx, username.(string), param.VersionID, param.Released, time.Now()))
}

func modVersionReleaseCheck(ctx *bm.Context) {
	param := new(struct {
		VersionID int64 `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(s.ModSvr.VersionReleaseCheck(ctx, param.VersionID, username.(string)))
}

func modVersionPush(ctx *bm.Context) {
	param := new(struct {
		VersionID  int64 `form:"version_id" validate:"min=1"`
		PushConfig bool  `form:"push_config"`
		PushGray   bool  `form:"push_gray"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	version, err := s.ModSvr.VersionPush(ctx, username.(string), param.VersionID, param.PushConfig, param.PushGray)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modVersionConfig(ctx *bm.Context) {
	param := new(struct {
		VersionID int64 `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	config, onlineConfig, err := s.ModSvr.VersionConfig(ctx, username.(string), param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["config"] = config
	res["online_config"] = onlineConfig
	ctx.JSON(res, nil)
}

func modVersionConfigAdd(ctx *bm.Context) {
	param := &mod.ConfigParam{}
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if param.Etime != 0 && param.Etime < param.Stime {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "结束时间必须大于开始时间"))
		return
	}
	if !param.Priority.Valid() {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "priority参数值错误"))
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionConfigAdd(ctx, username.(string), param))
}

func modVersionGray(ctx *bm.Context) {
	param := new(struct {
		VersionID int64 `form:"version_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	gray, onlineGray, err := s.ModSvr.VersionGray(ctx, username.(string), param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["gray"] = gray
	res["online_gray"] = onlineGray
	ctx.JSON(res, nil)
}

func modGrayWhitelistUpload(ctx *bm.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	defer file.Close()
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, err.Error()))
		return
	}
	filename := header.Filename
	if !filenameValid(ctx, filename) {
		return
	}
	whitelistURL, err := s.ModSvr.GrayWhitelistUpload(ctx, filename, fileData)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["whitelist_url"] = whitelistURL
	ctx.JSON(res, nil)
}

func modVersionGrayAdd(ctx *bm.Context) {
	param := &mod.GrayParam{}
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionGrayAdd(ctx, username.(string), param))
}

func modModuleDelete(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64 `form:"module_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.ModuleDelete(ctx, username.(string), param.ModuleID))
}

func modModuleState(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64           `form:"module_id" validate:"min=1"`
		State    mod.ModuleState `form:"state"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.State.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.ModuleState(ctx, username.(string), param.ModuleID, param.State))
}

func modModuleUpdate(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64        `form:"module_id" validate:"min=1"`
		Remark   string       `form:"remark" validate:"required"`
		Compress mod.Compress `form:"compress"`
		IsWIFI   bool         `form:"is_wifi"`
		ZipCheck bool         `form:"zip_check"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.ModuleUpdate(ctx, username.(string), param.ModuleID, param.Remark, param.IsWIFI, param.ZipCheck, param.Compress))
}

func modModuleAdd(ctx *bm.Context) {
	param := new(struct {
		PoolID   int64        `form:"pool_id" validate:"min=1"`
		Name     string       `form:"name" validate:"required"`
		Remark   string       `form:"remark" validate:"required"`
		IsWIFI   bool         `form:"is_wifi"`
		Compress mod.Compress `form:"compress" validate:"required"`
		ZipCheck bool         `form:"zip_check"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Compress.Valid() {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "compress参数值错误"))
		return
	}
	if !nameValid(ctx, param.Name) {
		return
	}
	username, _ := ctx.Get("username")
	module, err := s.ModSvr.ModuleAdd(ctx, username.(string), param.PoolID, param.Name, param.Remark, param.IsWIFI, param.Compress, param.ZipCheck)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["module"] = module
	ctx.JSON(res, nil)
}

func modModulePushOffline(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64 `form:"module_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.ModulePushOffline(ctx, username.(string), param.ModuleID))
}

func modPoolAdd(ctx *bm.Context) {
	param := new(struct {
		AppKey           string `form:"app_key" validate:"required"`
		Name             string `form:"name" validate:"required"`
		Remark           string `form:"remark" validate:"required"`
		ModuleCountLimit int64  `form:"module_count_limit"`
		ModuleSizeLimit  int64  `form:"module_size_limit"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !nameValid(ctx, param.Name) {
		return
	}
	username, _ := ctx.Get("username")
	pool, err := s.ModSvr.PoolAdd(ctx, username.(string), param.AppKey, param.Name, param.Remark, param.ModuleCountLimit, param.ModuleSizeLimit)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["pool"] = pool
	ctx.JSON(res, nil)
}

func modPoolUpdate(ctx *bm.Context) {
	param := new(struct {
		PoolID           int64 `form:"pool_id" validate:"min=1"`
		ModuleCountLimit int64 `form:"module_count_limit" validate:"required"`
		ModuleSizeLimit  int64 `form:"module_size_limit" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.PoolUpdate(ctx, username.(string), param.PoolID, param.ModuleCountLimit, param.ModuleSizeLimit))
}

func modPermissionList(ctx *bm.Context) {
	param := new(struct {
		PoolID int64 `form:"pool_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	permission, err := s.ModSvr.PermissionList(ctx, param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["permission"] = permission
	ctx.JSON(res, nil)
}

func modPermissionAdd(ctx *bm.Context) {
	param := &mod.PermissionParam{}
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Permission.Valid() {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.PermissionAdd(ctx, username.(string), param))
}

func modPermissionDelete(ctx *bm.Context) {
	param := new(struct {
		PermissionID int64 `form:"permission_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.PermissionDelete(ctx, username.(string), param.PermissionID))
}

func modPermissionRole(ctx *bm.Context) {
	param := new(struct {
		PoolID int64 `form:"pool_id" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	role, err := s.ModSvr.PermissionRole(ctx, username.(string), param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["role"] = role
	ctx.JSON(res, nil)
}

func modGlobalPush(ctx *bm.Context) {
	param := new(struct {
		AppKey string `form:"app_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.GlobalPush(ctx, username.(string), param.AppKey, time.Now()))
}

func modRoleApplyList(ctx *bm.Context) {
	param := new(struct {
		AppKey string         `form:"app_key" validate:"required"`
		State  mod.ApplyState `form:"state"`
		Pn     int64          `form:"pn" default:"1" validate:"min=1"`
		Ps     int64          `form:"ps" default:"20" validate:"min=1"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if param.State != "" && !param.State.Valid() {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "state字段值错误"))
		return
	}
	username, _ := ctx.Get("username")
	apply, page, err := s.ModSvr.RoleApplyList(ctx, param.AppKey, username.(string), param.State, param.Pn, param.Ps)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["apply"], res["page"] = apply, page
	ctx.JSON(res, nil)
}

func modRoleAdd(ctx *bm.Context) {
	param := new(struct {
		PoolID     int64    `form:"pool_id" validate:"required"`
		Username   string   `form:"username" validate:"required"`
		Permission mod.Perm `form:"permission" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Permission.Valid() {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "permission字段值错误"))
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.RoleAdd(ctx, username.(string), param.PoolID, param.Username, param.Permission))
}

func modRoleApplyAdd(ctx *bm.Context) {
	param := new(struct {
		PoolID     int64    `form:"pool_id" validate:"required"`
		Permission mod.Perm `form:"permission" validate:"required"`
		Operator   string   `form:"operator" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	if !param.Permission.Valid() {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "permission字段值错误"))
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.RoleApplyAdd(ctx, username.(string), param.PoolID, param.Permission, param.Operator))
}

func modRoleApplyProcess(ctx *bm.Context) {
	param := new(struct {
		PoolID int64 `form:"pool_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	apply, err := s.ModSvr.RoleApplyProcess(ctx, username.(string), param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["apply"] = apply
	ctx.JSON(res, nil)
}

func modRoleOperatorList(ctx *bm.Context) {
	param := new(struct {
		PoolID int64 `form:"pool_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	operator, err := s.ModSvr.RoleOperatorList(ctx, param.PoolID)
	if err != nil {
		ctx.JSON(nil, err)
	}
	res := map[string]interface{}{}
	res["operator"] = operator
	ctx.JSON(res, nil)
}

func modRoleApplyPass(ctx *bm.Context) {
	param := new(struct {
		ApplyID int64 `form:"apply_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.RoleApplyPass(ctx, username.(string), param.ApplyID))
}

func modRoleApplyRefuse(ctx *bm.Context) {
	param := new(struct {
		ApplyID int64 `form:"apply_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.RoleApplyRefuse(ctx, username.(string), param.ApplyID))
}

func modVersionApplyAdd(ctx *bm.Context) {
	param := new(struct {
		VersionID int64  `form:"version_id" validate:"required"`
		Operator  string `form:"operator" validate:"required"`
		Remark    string `form:"remark" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionApplyAdd(ctx, username.(string), param.VersionID, param.Operator, param.Remark))
}

func modVersionApplyNotify(ctx *bm.Context) {
	param := new(struct {
		AppKey string `form:"app_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	count, err := s.ModSvr.VersionApplyNotify(ctx, username.(string), param.AppKey)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["count"] = count
	ctx.JSON(res, nil)
}

func modVersionApplyList(ctx *bm.Context) {
	param := new(struct {
		AppKey string `form:"app_key" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	apply, err := s.ModSvr.VersionApplyList(ctx, username.(string), param.AppKey)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["apply"] = apply
	ctx.JSON(res, nil)
}

func modVersionApplyOverview(ctx *bm.Context) {
	param := new(struct {
		ApplyID int64 `form:"apply_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	overview, err := s.ModSvr.VersionApplyOverview(ctx, param.ApplyID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["overview"] = overview
	ctx.JSON(res, nil)
}

func modVersionApplyPass(ctx *bm.Context) {
	param := new(struct {
		ApplyID    int64  `form:"apply_id" validate:"required"`
		OnlineHash string `form:"online_hash" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionApplyPass(ctx, username.(string), param.ApplyID, param.OnlineHash, time.Now()))
}

func modVersionApplyRefuse(ctx *bm.Context) {
	param := new(struct {
		ApplyID int64 `form:"apply_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	ctx.JSON(nil, s.ModSvr.VersionApplyRefuse(ctx, username.(string), param.ApplyID))
}

func modSyncPool(ctx *bm.Context) {
	param := new(struct {
		AppKey   string `form:"app_key" validate:"required"`
		ModuleID int64  `form:"module_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	module, err := s.ModSvr.SyncPool(ctx, param.AppKey, param.ModuleID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["pool"] = module
	ctx.JSON(res, nil)
}

func modSyncVersionList(ctx *bm.Context) {
	param := new(struct {
		ModuleID int64 `form:"module_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	version, err := s.ModSvr.SyncVersionList(ctx, param.ModuleID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modSyncAdd(ctx *bm.Context) {
	param := new(mod.SyncParam)
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	username, _ := ctx.Get("username")
	version, err := s.ModSvr.SyncAdd(ctx, username.(string), param)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["version"] = version
	ctx.JSON(res, nil)
}

func modSyncVersionInfo(ctx *bm.Context) {
	param := new(struct {
		VersionID int64 `form:"version_id" validate:"required"`
	})
	err := ctx.Bind(param)
	if err != nil {
		return
	}
	info, err := s.ModSvr.SyncVersionInfo(ctx, param.VersionID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["info"] = info
	ctx.JSON(res, nil)
}

func nameValid(ctx *bm.Context, name string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if matched := reg.MatchString(name); !matched {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "名称不合法,只允许大小写字母和数字"))
		return false
	}
	return true
}

func filenameValid(ctx *bm.Context, filename string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if matched := reg.MatchString(filename); !matched {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "文件名不合法,只允许大小写字母和数字"))
		return false
	}
	return true
}
