package mod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	"go-common/library/sync/errgroup.v2"

	xecode "go-gateway/app/app-svr/fawkes/ecode"
	bcmdl "go-gateway/app/app-svr/fawkes/service/model/broadcast"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

func (s *Service) OpenPoolList(ctx context.Context, modPoolKey, appKey string) ([]*mod.Pool, error) {
	poolNames, ok := s.c.Mod.PoolKey[modPoolKey]
	if !ok {
		return nil, xecode.InvalidModPoolKey
	}
	pool, err := s.fkDao.ModPoolList(ctx, appKey)
	if err != nil {
		return nil, err
	}
	var res []*mod.Pool
	for _, val := range poolNames {
		for _, p := range pool {
			if p.Name == val {
				res = append(res, p)
			}
		}
	}
	return res, nil
}

func (s *Service) OpenModuleList(ctx context.Context, modPoolKey string, poolID int64) ([]*mod.Module, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, xecode.InvalidModPoolKey
	}
	return s.fkDao.ModModuleList(ctx, poolID)
}

func (s *Service) OpenVersion(ctx context.Context, modPoolKey string, versionID int64) (*mod.Version, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, xecode.InvalidModPoolKey
	}
	r, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if r.ReleaseTime < 0 {
		r.ReleaseTime = 0
	}
	var id int64
	switch r.Env {
	case mod.EnvTest:
		id = r.ID
	case mod.EnvProd:
		id = r.FromVerID
	}
	var fm map[int64][]*mod.File
	if id > 0 {
		if fm, err = s.fkDao.ModVersionFile(ctx, []int64{id}); err != nil {
			return nil, err
		}
	}
	fs, ok := fm[id]
	if !ok {
		return r, nil
	}
	var count int
	for _, f := range fs {
		if !f.IsPatch {
			r.File = f
			continue
		}
		count++
	}
	r.Patch = &mod.VersionPatch{Count: count}
	return r, nil
}

func (s *Service) OpenVersionList(ctx context.Context, modPoolKey string, moduleID int64, env mod.Env, pn, ps int64) ([]*mod.Version, *mod.Page, bool, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, nil, false, xecode.InvalidModPoolKey
	}
	offset := (pn - 1) * ps
	limit := ps
	var (
		res  []*mod.Version
		page *mod.Page
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) (err error) {
		res, err = s.fkDao.ModVersionList(ctx, moduleID, env, offset, limit)
		return err
	})
	g.Go(func(ctx context.Context) (err error) {
		count, err := s.fkDao.ModVersionCount(ctx, moduleID, env)
		if err != nil && err != xsql.ErrNoRows {
			return err
		}
		page = &mod.Page{Total: count, Pn: pn, Ps: ps}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, false, err
	}
	if len(res) == 0 {
		return nil, page, false, nil
	}
	var ids []int64
	for _, r := range res {
		switch r.Env {
		case mod.EnvTest:
			ids = append(ids, r.ID)
		case mod.EnvProd:
			ids = append(ids, r.FromVerID)
		}
	}
	var (
		fm  map[int64][]*mod.File
		err error
	)
	if len(ids) != 0 {
		if fm, err = s.fkDao.ModVersionFile(ctx, ids); err != nil {
			return nil, nil, false, err
		}
	}
	var polling bool
	for _, r := range res {
		if r.ReleaseTime < 0 {
			r.ReleaseTime = 0
		}
		var id int64
		switch r.Env {
		case mod.EnvTest:
			id = r.ID
		case mod.EnvProd:
			id = r.FromVerID
		}
		fs, ok := fm[id]
		if !ok {
			continue
		}
		var count int
		for _, f := range fs {
			if !f.IsPatch {
				r.File = f
				continue
			}
			count++
		}
		if r.State.IsPolling() {
			polling = true
		}
		r.Patch = &mod.VersionPatch{Count: count}
	}
	return res, page, polling, nil
}

func (s *Service) OpenPatchList(ctx context.Context, modPoolKey string, versionID int64) ([]*mod.Patch, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, xecode.InvalidModPoolKey
	}
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.Env == mod.EnvProd {
		versionID = version.FromVerID
	}
	return s.fkDao.ModPatchList(ctx, versionID)
}

func (s *Service) OpenModuleState(ctx context.Context, modPoolKey, username string, moduleID int64, state mod.ModuleState) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return xecode.ForbiddenOperateMod
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	if err := s.fkDao.ModModuleState(ctx, moduleID, state); err != nil {
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,状态:%v", pool.Name, module.Name, state)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源上下线", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) OpenModuleAdd(ctx context.Context, modPoolKey, username string, poolID int64, name, remark string, isWiFI bool, compress mod.Compress) (*mod.Module, error) {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return nil, err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, xecode.ForbiddenOperateMod
	}
	id, err := s.fkDao.ModModuleExist(ctx, poolID, name)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if id != 0 {
		return nil, xecode.ExistMod
	}
	if id, err = s.fkDao.ModModuleAdd(ctx, poolID, name, remark, isWiFI, compress, mod.ModuleOnline, false); err != nil {
		return nil, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,备注:%v,仅WIFI可下载:%v,解压方式:%v,资源ID:%v", pool.Name, name, remark, isWiFI, compress, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源新增", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Module{
		ID: id,
	}, nil
}

func (s *Service) OpenModulePushOffline(ctx context.Context, modPoolKey, username string, moduleID int64) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	if module.State == mod.ModuleOffline {
		return xecode.OfflineVersionNoPush
	}
	ok, err := s.fkDao.ReleasedVersionExists(ctx, moduleID)
	if err != nil {
		return err
	}
	if !ok {
		return xecode.DisableVersionNoPush
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	var ok1 bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok1 = true
			break
		}
	}
	if !ok1 {
		return xecode.ForbiddenOperateMod
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	if appInfo == nil {
		return xecode.NoExistAppKey
	}
	if ok, err = s.fkDao.ModulePushOffline(ctx, moduleID, pool.AppKey, pool.Name, module.Name); err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := s.pushModule(ctx, bcmdl.Action_DELETE, appInfo.MobiApp, pool.AppKey, pool.Name, module.Name); err != nil {
		log.Error("日志告警 broadcast push fail,action=%v,mobiApp=%v,appKey=%v,poolName=%v,moduleName=%v", bcmdl.Action_DELETE, appInfo.MobiApp, pool.AppKey, pool.Name, module.Name)
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,资源ID:%v", pool.Name, module.Name, module.ID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源推送下线", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) OpenVersionAdd(ctx context.Context, modPoolKey, username string, moduleID int64, env mod.Env, remark string, filename string, fileData []byte) (*mod.Version, error) {
	if env != mod.EnvTest {
		return nil, ecode.Error(ecode.RequestErr, "禁止非测试环境上传资源")
	}
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, xecode.ForbiddenOperateMod
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return nil, xecode.DisableMod
	}
	contentType, md5, size, err := parseFileData(fileData)
	if err != nil {
		return nil, err
	}
	if module.Compress == mod.CompressUnzip {
		// 校验 文件类型
		// application/zip application/x-gzip
		var ok bool
		for _, v := range s.c.ZipContentType {
			if contentType == v {
				ok = true
				break
			}
		}
		if !ok {
			return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("校验出上传文件的content-type是%v,该文件不是zip类型,如有疑问咨询fawkes管理员", contentType))
		}
	}
	url, err := s.fileUpload(ctx, filename, md5, fileData)
	if err != nil {
		return nil, err
	}
	file := &mod.File{
		Name:        filename,
		ContentType: contentType,
		Md5:         md5,
		Size:        size,
		URL:         url,
		IsPatch:     false,
		FromVer:     0,
	}
	version, err := s.fkDao.ModVersionAdd(ctx, moduleID, env, remark, file)
	if err != nil {
		return nil, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,版本ID:%v,版本号:%v", pool.Name, module.Name, version.ID, version.Version)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(env), "mod", "上传资源", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return version, nil
}

func (s *Service) OpenVersionRelease(ctx context.Context, modPoolKey, username string, versionID int64, released bool, now time.Time) error {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return err
	}
	if version.State == mod.VersionProcessing {
		return xecode.ProcessingVersion
	}
	if version.State == mod.VersionDisable {
		return xecode.DisableVersion
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return xecode.ForbiddenOperateMod
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return xecode.DisableMod
	}
	if version.Env == mod.EnvProd {
		file, err := s.fkDao.ModFile(ctx, version.FromVerID)
		if err != nil {
			return err
		}
		if file.Size > 20*1024*1024 {
			return ecode.Error(ecode.RequestErr, "超过20M大小的文件禁止通过第三方接口上传")
		}
	}
	releaseTime := version.ReleaseTime
	if released {
		releaseTime = xtime.Time(now.Unix())
	}
	if err = s.fkDao.ModVersionRelease(ctx, versionID, released, releaseTime); err != nil {
		return err
	}
	if released {
		s.event.Publish(VersionRelease, &VersionReleaseArgs{Ctx: utils.CopyTrx(ctx), VersionId: versionID, UserName: username})
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,版本号:%v,是否生效:%v,生效时间:%v", pool.Name, module.Name, version.Version, released, releaseTime)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "版本生效", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) OpenVersionPush(ctx context.Context, modPoolKey, username string, versionID int64) (*mod.Version, error) {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.Env != mod.EnvTest {
		return nil, ecode.Error(ecode.RequestErr, "禁止非测试环境资源推送到正式")
	}
	if version.State == mod.VersionProcessing {
		return nil, xecode.ProcessingVersion
	}
	if version.State == mod.VersionDisable {
		return nil, xecode.DisableVersion
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, xecode.ForbiddenOperateMod
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return nil, xecode.DisableMod
	}
	id, err := s.fkDao.ModProdVersionExist(ctx, version.ModuleID, version.Version)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if id != 0 {
		return &mod.Version{
			ID: id,
		}, nil
	}
	if id, err = s.fkDao.ModVersionPush(ctx, version); err != nil {
		return nil, err
	}
	s.event.Publish(VersionPush, &VersionPushArgs{Ctx: utils.CopyTrx(ctx), VersionId: id, UserName: username})
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,版本号:%v,param:%+v,版本ID:%v", pool.Name, module.Name, version.Version, version, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "推送正式", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Version{
		ID: id,
	}, nil
}

func (s *Service) OpenVersionConfig(ctx context.Context, modPoolKey string, versionID int64) (*mod.VersionConfig, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, xecode.InvalidModPoolKey
	}
	config, err := s.fkDao.ModVersionConfig(ctx, versionID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var appVer, sysVer []map[mod.Condition]int64
	if config.AppVer != "" {
		if err := json.Unmarshal([]byte(config.AppVer), &appVer); err != nil {
			return nil, err
		}
	}
	if config.SysVer != "" {
		if err := json.Unmarshal([]byte(config.SysVer), &sysVer); err != nil {
			return nil, err
		}
	}
	stime := config.Stime
	if config.Stime < 0 {
		stime = 0
	}
	etime := config.Etime
	if config.Etime < 0 {
		etime = 0
	}
	return &mod.VersionConfig{
		ID:       config.ID,
		Priority: config.Priority,
		AppVer:   appVer,
		SysVer:   sysVer,
		Stime:    stime,
		Etime:    etime,
	}, nil
}

func (s *Service) OpenVersionConfigAdd(ctx context.Context, modPoolKey, username string, param *mod.ConfigParam) error {
	version, err := s.fkDao.ModVersionByID(ctx, param.VersionID)
	if err != nil {
		return err
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return xecode.ForbiddenOperateMod
	}
	if version.Env == mod.EnvProd && param.Priority != mod.PriorityLow {
		if ok := func() bool {
			vals, ok := s.c.Mod.PriorityMod[pool.Name]
			if !ok {
				return false
			}
			if len(vals) == 0 {
				return true
			}
			for _, val := range vals {
				if module.Name == val {
					return true
				}
			}
			return false
		}(); !ok {
			return ecode.Error(ecode.RequestErr, "禁止通过第三方接口设置正式资源为低以上的优先级")
		}
		if param.Priority == mod.PriorityHigh {
			return ecode.Error(ecode.RequestErr, "禁止通过第三方接口设置正式资源为中以上的优先级")
		}
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return xecode.DisableMod
	}
	appVer, err := checkVer(param.AppVer)
	if err != nil {
		return ecode.Error(ecode.RequestErr, "app_ver"+err.Error())
	}
	sysVer, err := checkVer(param.SysVer)
	if err != nil {
		return ecode.Error(ecode.RequestErr, "sys_ver"+err.Error())
	}
	oldConfig, err := s.fkDao.ModVersionConfig(ctx, param.VersionID)
	if err != nil && err != xsql.ErrNoRows {
		return ecode.Error(ecode.ServerErr, "mod version error: "+err.Error())
	}
	config := &mod.Config{
		VersionID: param.VersionID,
		Priority:  param.Priority,
		AppVer:    appVer,
		SysVer:    sysVer,
		Stime:     param.Stime,
		Etime:     param.Etime,
	}
	id, err := s.fkDao.ModVersionConfigAdd(ctx, config)
	if err != nil {
		return err
	}
	s.event.Publish(VersionConfigChange, &VersionConfigChangeArgs{Ctx: utils.CopyTrx(ctx), old: oldConfig, new: config, UserName: username})
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,版本号:%v,配置ID:%v,参数:%+v", pool.Name, module.Name, version.Version, id, param)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "配置修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) OpenVersionGray(ctx context.Context, modPoolKey string, versionID int64) (*mod.Gray, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return nil, xecode.InvalidModPoolKey
	}
	gray, err := s.fkDao.ModVersionGray(ctx, versionID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return gray, nil
}

func (s *Service) OpenVersionGrayAdd(ctx context.Context, modPoolKey, username string, param *mod.GrayParam) error {
	if param.WhitelistURL != "" && !strings.Contains(param.WhitelistURL, _whitelistID) {
		return ecode.Error(ecode.RequestErr, "whitelist_url 参数错误")
	}
	version, err := s.fkDao.ModVersionByID(ctx, param.VersionID)
	if err != nil {
		return err
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	var ok bool
	for _, val := range s.c.Mod.PoolKey[modPoolKey] {
		if pool.Name == val {
			ok = true
			break
		}
	}
	if !ok {
		return xecode.ForbiddenOperateMod
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return xecode.DisableMod
	}
	oldGray, err := s.fkDao.ModVersionGray(ctx, param.VersionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	gray := &mod.Gray{
		VersionID:      param.VersionID,
		Strategy:       param.Strategy,
		Salt:           param.Salt,
		BucketStart:    param.BucketStart,
		BucketEnd:      param.BucketEnd,
		Whitelist:      xstr.JoinInts(param.Whitelist),
		WhitelistURL:   param.WhitelistURL,
		ManualDownload: param.ManualDownload,
	}
	id, err := s.fkDao.ModVersionGrayAdd(ctx, gray)
	if err != nil {
		return err
	}
	s.event.Publish(VersionGrayChange, &VersionGrayChangeArgs{Ctx: utils.CopyTrx(ctx), old: oldGray, new: gray, UserName: username})
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("openAPI,资源池:%v,资源:%v,版本号:%v,灰度ID:%v,参数:%+v", pool.Name, module.Name, version.Version, id, param)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "灰度修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) OpenGrayWhitelistUpload(ctx context.Context, modPoolKey, filename string, fileData []byte) (string, error) {
	if _, ok := s.c.Mod.PoolKey[modPoolKey]; !ok {
		return "", xecode.InvalidModPoolKey
	}
	if _, err := xstr.SplitInts(string(fileData)); err != nil {
		log.Error("%+v", err)
		return "", ecode.Error(ecode.RequestErr, "白名单文件中填写mid,添加多个可用英文逗号隔开,请勿空格或者换行")
	}
	_, md5, _, err := parseFileData(fileData)
	if err != nil {
		return "", err
	}
	return s.fileUpload(ctx, fmt.Sprintf("%s_%s", _whitelistID, filename), md5, fileData)
}
