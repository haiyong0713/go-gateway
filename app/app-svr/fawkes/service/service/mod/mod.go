package mod

import (
	"context"
	// nolint:gosec
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	"go-common/library/sync/errgroup.v2"

	bcmdl "go-gateway/app/app-svr/fawkes/service/model/broadcast"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	_type "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

const _whitelistID = "whitelist"

var (
	_emptyConfig = &mod.Config{}
	_emptyGray   = &mod.Gray{}
)

func (s *Service) PoolList(ctx context.Context, appKey string) ([]*mod.Pool, error) {
	return s.fkDao.ModPoolList(ctx, appKey)
}

func (s *Service) ModuleList(ctx context.Context, poolID int64) ([]*mod.Module, error) {
	return s.fkDao.ModModuleList(ctx, poolID)
}

func (s *Service) VersionList(ctx context.Context, moduleID int64, env mod.Env, pn, ps int64) ([]*mod.Version, *mod.Page, bool, error) {
	offset := (pn - 1) * ps
	limit := ps
	var (
		res  []*mod.Version
		page *mod.Page
	)
	g := errgroup.WithCancel(ctx)
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
	var (
		ids         []int64
		versionIDs  []int64
		prodVersion = make(map[int64]struct{})
	)
	for _, r := range res {
		switch r.Env {
		case mod.EnvTest:
			ids = append(ids, r.ID)
		case mod.EnvProd:
			ids = append(ids, r.FromVerID)
			prodVersion[r.Version] = struct{}{}
		}
		versionIDs = append(versionIDs, r.ID)
	}
	var (
		fm map[int64][]*mod.File
		am map[int64]*mod.VersionApply
		cm map[int64]*mod.ConfigApply
		gm map[int64]*mod.GrayApply
	)
	g = errgroup.WithCancel(ctx)
	if len(ids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			fm, err = s.fkDao.ModVersionFile(ctx, ids)
			return err
		})
	}
	if len(versionIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			am, err = s.fkDao.ModVersionApplyByVersionID(ctx, versionIDs)
			return err
		})
		g.Go(func(ctx context.Context) (err error) {
			cm, err = s.fkDao.ModVersionConfigApplyByVersionID(ctx, versionIDs)
			return err
		})
		g.Go(func(ctx context.Context) (err error) {
			gm, err = s.fkDao.ModVersionGrayApplyByVersionID(ctx, versionIDs)
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, nil, false, err
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
		if a, ok := am[r.ID]; ok {
			r.ApplyState = a.State
		}
		if c, ok := cm[r.ID]; ok {
			r.ConfigApplyState = c.State
		}
		if g, ok := gm[r.ID]; ok {
			r.GrayApplyState = g.State
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
			if env == mod.EnvProd {
				if _, ok := prodVersion[f.FromVer]; !ok {
					continue
				}
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

func (s *Service) PatchList(ctx context.Context, versionID int64, env mod.Env) ([]*mod.Patch, error) {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.Env == mod.EnvProd {
		versionID = version.FromVerID
	}
	modFile, err := s.fkDao.ModFile(ctx, versionID)
	if err != nil {
		return nil, err
	}
	list, err := s.fkDao.ModPatchList(ctx, versionID)
	if err != nil {
		return nil, err
	}
	var fromVers []int64
	for _, v := range list {
		i, _ := strconv.ParseInt(v.FromVer, 10, 64)
		fromVers = append(fromVers, i)
	}
	if env == mod.EnvProd {
		var prodVersion = make(map[int64]*mod.Version)
		versionList, err := s.fkDao.ModVersionList2(ctx, version.ModuleID, env, fromVers)
		if err != nil {
			return nil, err
		}
		for _, v := range versionList {
			prodVersion[v.Version] = v
		}
		var filterList []*mod.Patch
		for _, l := range list {
			i, _ := strconv.ParseInt(l.FromVer, 10, 64)
			if _, ok := prodVersion[i]; ok {
				filterList = append(filterList, l)
			}
		}
		list = filterList
	}
	for _, f := range list {
		if f.Size >= modFile.Size {
			f.Declare = "增量包体积大于原包"
		}
	}
	return list, nil
}

func (s *Service) VersionAdd(ctx context.Context, username string, moduleID int64, env mod.Env, remark string, filename string, fileData []byte) (*mod.Version, error) {
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
	if s.isDisableModule(pool.Name, module.Name) {
		return nil, ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
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
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本ID:%v,版本号:%v", pool.Name, module.Name, version.ID, version.Version)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(env), "mod", "上传资源", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return version, nil
}

// nolint:gocognit
func (s *Service) VersionConfig(ctx context.Context, username string, versionID int64) (config, onlineConifig *mod.VersionConfig, err error) {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, nil, err
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return nil, nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, nil, err
	}
	configFunc := func(config *mod.Config) (*mod.VersionConfig, error) {
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
	if version.Env == mod.EnvProd {
		perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
		if err != nil {
			return nil, nil, err
		}
		if perm == mod.FawkesUserPerm {
			g := errgroup.WithCancel(ctx)
			g.Go(func(ctx context.Context) error {
				c, err := s.fkDao.ModVersionConfigApply(ctx, versionID)
				if err != nil {
					if err == xsql.ErrNoRows {
						return nil
					}
					return err
				}
				if c.State != mod.ApplyStateChecking {
					return nil
				}
				config, err = configFunc(&c.Config)
				if err != nil {
					log.Error("%v", err)
				}
				return nil
			})
			g.Go(func(ctx context.Context) error {
				c, err := s.fkDao.ModVersionConfig(ctx, versionID)
				if err != nil {
					if err == xsql.ErrNoRows {
						return nil
					}
					return err
				}
				onlineConifig, err = configFunc(c)
				if err != nil {
					log.Error("%v", err)
				}
				return nil
			})
			if err := g.Wait(); err != nil {
				return nil, nil, err
			}
			if config == nil {
				config = onlineConifig
			}
			if version.Released {
				return config, onlineConifig, nil
			}
			return config, nil, nil
		}
	}
	c, err := s.fkDao.ModVersionConfig(ctx, versionID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if config, err = configFunc(c); err != nil {
		return nil, nil, err
	}
	return config, nil, nil
}

func (s *Service) VersionConfigAdd(ctx context.Context, username string, param *mod.ConfigParam) error {
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
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	appVer, err := checkVer(param.AppVer)
	if err != nil {
		return ecode.Error(ecode.RequestErr, "app_ver"+err.Error())
	}
	sysVer, err := checkVer(param.SysVer)
	if err != nil {
		return ecode.Error(ecode.RequestErr, "sys_ver"+err.Error())
	}
	config := &mod.Config{
		VersionID: param.VersionID,
		Priority:  param.Priority,
		AppVer:    appVer,
		SysVer:    sysVer,
		Stime:     param.Stime,
		Etime:     param.Etime,
	}
	oldConfig, err := s.fkDao.ModVersionConfig(ctx, param.VersionID)
	if err != nil && err != xsql.ErrNoRows {
		return ecode.Error(ecode.ServerErr, fmt.Sprintf("mod version error: %v", err.Error()))
	}
	var id int64
	if err := func() error {
		if version.Env == mod.EnvProd {
			perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
			if err != nil {
				return err
			}
			if perm == mod.FawkesUserPerm {
				applyID, err := s.fkDao.ModVersionApplyExist(ctx, config.VersionID)
				if err != nil && err != xsql.ErrNoRows {
					return err
				}
				if applyID > 0 {
					return ecode.Error(ecode.RequestErr, "已存在未处理的申请单，先联系资源池管理员处理")
				}
				id, err = s.fkDao.ModVersionConfigApplyAdd(ctx, config)
				return err
			}
		}
		id, err = s.fkDao.ModVersionConfigAdd(ctx, config)
		return err
	}(); err != nil {
		return err
	}
	s.event.Publish(VersionConfigChange, &VersionConfigChangeArgs{Ctx: utils.CopyTrx(ctx), old: oldConfig, new: config, UserName: username})
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,配置ID:%v,参数:%+v", pool.Name, module.Name, version.Version, id, param)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "配置修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionGray(ctx context.Context, username string, versionID int64) (gray, onlineGray *mod.Gray, err error) {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, nil, err
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return nil, nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, nil, err
	}
	if version.Env == mod.EnvProd {
		perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
		if err != nil {
			return nil, nil, err
		}
		if perm == mod.FawkesUserPerm {
			g := errgroup.WithCancel(ctx)
			g.Go(func(ctx context.Context) error {
				g, err := s.fkDao.ModVersionGrayApply(ctx, versionID)
				if err != nil {
					if err == xsql.ErrNoRows {
						return nil
					}
					return err
				}
				if g.State != mod.ApplyStateChecking {
					return nil
				}
				gray = &g.Gray
				return nil
			})
			g.Go(func(ctx context.Context) error {
				if onlineGray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil {
					if err == xsql.ErrNoRows {
						return nil
					}
					return err
				}
				return nil
			})
			if err := g.Wait(); err != nil {
				return nil, nil, err
			}
			if gray == nil {
				gray = onlineGray
			}
			if version.Released {
				return gray, onlineGray, nil
			}
			return gray, nil, nil
		}
	}
	if gray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	return gray, nil, nil
}

func (s *Service) VersionGrayAdd(ctx context.Context, username string, param *mod.GrayParam) error {
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
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
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
	var id int64
	if err := func() error {
		if version.Env == mod.EnvProd {
			perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
			if err != nil {
				return err
			}
			if perm == mod.FawkesUserPerm {
				applyID, err := s.fkDao.ModVersionApplyExist(ctx, gray.VersionID)
				if err != nil && err != xsql.ErrNoRows {
					return err
				}
				if applyID > 0 {
					return ecode.Error(ecode.RequestErr, "已存在未处理的申请单，先联系资源池管理员处理")
				}
				id, err = s.fkDao.ModVersionGrayApplyAdd(ctx, gray)
				return err
			}
		}
		id, err = s.fkDao.ModVersionGrayAdd(ctx, gray)
		return err
	}(); err != nil {
		return err
	}
	s.event.Publish(VersionGrayChange, &VersionGrayChangeArgs{Ctx: utils.CopyTrx(ctx), old: oldGray, new: gray, UserName: username})
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,灰度ID:%v,参数:%+v", pool.Name, module.Name, version.Version, id, param)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "灰度修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionRelease(ctx context.Context, username string, versionID int64, released bool, now time.Time) error {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return err
	}
	if version.State == mod.VersionProcessing {
		return ecode.Error(ecode.RequestErr, "增量包尚未构建完成,请稍后再试")
	}
	if version.State == mod.VersionDisable {
		return ecode.Error(ecode.RequestErr, "该版本已永久下线")
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	if version.Env == mod.EnvProd {
		file, err := s.fkDao.ModFile(ctx, version.FromVerID)
		if err != nil {
			return err
		}
		perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
		if err != nil {
			return err
		}
		if file.Size > 20*1024*1024 && perm != mod.FawkesSuperAdminPerm {
			return ecode.Error(ecode.RequestErr, "超过20M大小的文件,fawkes超级管理员权限才能操作")
		}
		if perm == mod.FawkesUserPerm {
			return ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
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
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,是否生效:%v,生效时间:%v", pool.Name, module.Name, version.Version, released, releaseTime)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "版本生效", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionPush(ctx context.Context, username string, versionID int64, pushConfig, pushGray bool) (*mod.Version, error) {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.Env != mod.EnvTest {
		return nil, ecode.Error(ecode.RequestErr, "禁止非测试环境资源推送到正式")
	}
	if version.State == mod.VersionProcessing {
		return nil, ecode.Error(ecode.RequestErr, "增量包尚未构建完成,请稍后再试")
	}
	if version.State == mod.VersionDisable {
		return nil, ecode.Error(ecode.RequestErr, "该版本已永久下线")
	}
	module, err := s.fkDao.ModModuleByID(ctx, version.ModuleID)
	if err != nil {
		return nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, err
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return nil, ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	id, err := s.fkDao.ModProdVersionExist(ctx, version.ModuleID, version.Version)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if id != 0 {
		return nil, ecode.Error(ecode.RequestErr, "已推送到线上,请勿重复推送")
	}
	var (
		config *mod.Config
		gray   *mod.Gray
	)
	if pushConfig {
		if config, err = s.fkDao.ModVersionConfig(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			return nil, err
		}
	}
	if pushGray {
		if gray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			return nil, err
		}
	}
	if id, err = s.fkDao.ModVersionPushWithConfig(ctx, version, config, gray); err != nil {
		return nil, err
	}
	s.event.Publish(VersionPush, &VersionPushArgs{Ctx: utils.CopyTrx(ctx), VersionId: id, UserName: username})
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,param:%+v,版本ID:%v", pool.Name, module.Name, version.Version, version, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(version.Env), "mod", "推送正式", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Version{
		ID: id,
	}, nil
}

func (s *Service) GrayWhitelistUpload(ctx context.Context, filename string, fileData []byte) (string, error) {
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

func (s *Service) ModuleDelete(ctx context.Context, username string, moduleID int64) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm == mod.FawkesUserPerm {
		return ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
	}
	if err := s.fkDao.ModModuleDelete(ctx, moduleID); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v", pool.Name, module.Name)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源删除", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) ModuleState(ctx context.Context, username string, moduleID int64, state mod.ModuleState) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm == mod.FawkesUserPerm {
		return ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
	}
	if err := s.fkDao.ModModuleState(ctx, moduleID, state); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,状态:%v", pool.Name, module.Name, state)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源上下线", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) ModuleUpdate(ctx context.Context, username string, moduleID int64, remark string, isWiFI, zipCheck bool, compress mod.Compress) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	if s.isDisableModule(pool.Name, module.Name) {
		return ecode.Error(ecode.RequestErr, "限制类型资源禁止修改操作")
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm == mod.FawkesUserPerm {
		return ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
	}
	if compress != module.Compress {
		prodCount, err := s.fkDao.ModVersionCount(ctx, moduleID, mod.EnvProd)
		if err != nil {
			return err
		}
		if prodCount > 0 {
			return ecode.Error(ecode.RequestErr, "该资源已经推送到正式环境，不可以修改文件格式")
		}
	}
	if err := s.fkDao.ModModuleUpdate(ctx, moduleID, remark, isWiFI, zipCheck, compress); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,备注:%v,仅WIFI可下载:%v,校验:%v", pool.Name, module.Name, remark, isWiFI, zipCheck)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) ModuleAdd(ctx context.Context, username string, poolID int64, name, remark string, isWiFI bool, compress mod.Compress, zipCheck bool) (*mod.Module, error) {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return nil, err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return nil, err
	}
	if perm == mod.FawkesUserPerm {
		return nil, ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
	}
	id, err := s.fkDao.ModModuleExist(ctx, poolID, name)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if id != 0 {
		return nil, ecode.Error(ecode.RequestErr, "已存在相同mod名,不可重复创建")
	}
	if id, err = s.fkDao.ModModuleAdd(ctx, poolID, name, remark, isWiFI, compress, mod.ModuleOnline, zipCheck); err != nil {
		return nil, err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,备注:%v,仅WIFI可下载:%v,解压方式:%v,资源ID:%v,是否解压:%v", pool.Name, name, remark, isWiFI, compress, id, zipCheck)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源新增", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Module{
		ID: id,
	}, nil
}

func (s *Service) ModulePushOffline(ctx context.Context, username string, moduleID int64) error {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}
	if module.State == mod.ModuleOffline {
		return ecode.Error(ecode.RequestErr, "已经下线的资源不能进行推送下线操作")
	}
	ok, err := s.fkDao.ReleasedVersionExists(ctx, moduleID)
	if err != nil {
		return err
	}
	if !ok {
		return ecode.Error(ecode.RequestErr, "资源不存在生效的版本不能进行推送下线操作")
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm == mod.FawkesUserPerm {
		return ecode.Error(ecode.RequestErr, "资源池管理员或更高权限才能操作")
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	if appInfo == nil {
		return ecode.Error(ecode.RequestErr, "appKey不存在")
	}
	if ok, err = s.fkDao.ModulePushOffline(ctx, moduleID, pool.AppKey, pool.Name, module.Name); err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := s.pushModule(ctx, bcmdl.Action_DELETE, appInfo.MobiApp, pool.AppKey, pool.Name, module.Name); err != nil {
		log.Error("日志告警 broadcast push fail,action=%v,mobiApp=%v,appKey=%v,poolName=%v,moduleName=%v,error=%+v", bcmdl.Action_DELETE, appInfo.MobiApp, pool.AppKey, pool.Name, module.Name, err)
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,资源ID:%v", pool.Name, module.Name, module.ID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源推送下线", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) pushModule(ctx context.Context, action bcmdl.Action, mobiApp, appKey, poolName, moduleName string) error {
	pb := &bcmdl.ModuleNotifyReply{
		Atcion:     action,
		AppKey:     appKey,
		PoolName:   poolName,
		ModuleName: moduleName,
	}
	body, err := types.MarshalAny(pb)
	if err != nil {
		return err
	}
	msg := &_type.Message{
		TargetPath: s.c.BroadcastGrpc.Module.TargetPath,
		Body:       body,
	}
	key := fmt.Sprintf("lock_push_mod_%d_%s_%s_%s_%s", action, mobiApp, appKey, poolName, moduleName)
	ok, err := s.fkDao.TryLock(ctx, key, 600)
	if err != nil {
		return err
	}
	if !ok {
		return ecode.Error(ecode.RequestErr, "10分钟内仅能发起一次推送")
	}
	if err = s.fkDao.BroadcastPushAll(ctx, "", msg); err != nil {
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.fkDao.UnLock(ctx, key); err != nil {
				log.Error("%+v", err)
			}
		})
		return err
	}
	return nil
}

func (s *Service) PoolAdd(ctx context.Context, username, appKey, name, remark string, moduleCountLimit, moduleSizeLimit int64) (*mod.Pool, error) {
	perm, err := s.fawkesPerm(ctx, username, appKey, 0)
	if err != nil {
		return nil, err
	}
	if perm < mod.FawkesAppAdminPerm {
		return nil, ecode.Error(ecode.RequestErr, "fawkes app管理员及以上权限才能操作")
	}
	id, err := s.fkDao.ModPoolExist(ctx, appKey, name)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if id != 0 {
		return nil, ecode.Error(ecode.RequestErr, "禁止创建同名资源池")
	}
	if id, err = s.fkDao.ModPoolAdd(ctx, appKey, name, remark, moduleCountLimit, moduleSizeLimit); err != nil {
		return nil, err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,备注:%v,资源数量限制:%v,资源大小限制:%v,资源池ID:%v", name, remark, moduleCountLimit, moduleSizeLimit, id)
		if _, err := s.fkDao.AddLog(ctx, appKey, string(mod.EnvProd), "mod", "资源池新增", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Pool{
		ID: id,
	}, nil
}
func (s *Service) PoolUpdate(ctx context.Context, username string, poolID int64, moduleCountLimit, moduleSizeLimit int64) error {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return err
	}
	// super admin 才能操作
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm != mod.FawkesSuperAdminPerm {
		return ecode.Error(ecode.RequestErr, "fawkes 超级管理员权限才能操作")
	}
	if err := s.fkDao.ModPoolUpdate(ctx, poolID, moduleCountLimit, moduleSizeLimit); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源数量限制:%v,资源大小限制:%v", pool.Name, moduleCountLimit, moduleSizeLimit)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "资源池修改", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) PermissionList(ctx context.Context, poolID int64) ([]*mod.Permission, error) {
	return s.fkDao.ModPermissionList(ctx, poolID)
}

func (s *Service) PermissionAdd(ctx context.Context, username string, param *mod.PermissionParam) error {
	pool, err := s.fkDao.ModPoolByID(ctx, param.PoolID)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesAppAdminPerm {
		return ecode.Error(ecode.RequestErr, "fawkes app管理员及以上权限才能操作")
	}
	id, err := s.fkDao.ModPermissionAdd(ctx, param.Username, param.PoolID, param.Permission)
	if err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,参数:%+v,权限ID:%v", pool.Name, param, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限新增", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) PermissionDelete(ctx context.Context, username string, permissionID int64) error {
	permission, err := s.fkDao.ModPermissionByID(ctx, permissionID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, permission.PoolID)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesAppAdminPerm {
		return ecode.Error(ecode.RequestErr, "fawkes app管理员及以上权限才能操作")
	}
	if err := s.fkDao.ModPermissionDelete(ctx, permissionID); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,用户:%v,权限:%v", pool.Name, permission.Username, permission.Permission)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限删除", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) PermissionRole(ctx context.Context, username string, poolID int64) (*mod.PermissionRole, error) {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return nil, err
	}
	res := &mod.PermissionRole{Username: username}
	for _, v := range s.c.Whitelist {
		if username == v {
			res.Role = mod.RoleWhitelistAdmin
			return res, nil
		}
	}
	// 超管权限
	supervisorRole, err := s.fkDao.AuthSupervisor(ctx, username)
	if err != nil {
		return nil, err
	}
	if len(supervisorRole) > 0 {
		res.Role = mod.RoleSuperAdmin
		return res, nil
	}
	// App 管理员权限
	user, err := s.fkDao.AuthUser(ctx, pool.AppKey, username)
	if err != nil {
		return nil, err
	}
	// 角色：ADMIN-1,DEV-2,TEST-3,DEVOPS-4,VISITOR-5
	if user.Role == 1 {
		res.Role = mod.RoleAppAdmin
		return res, nil
	}
	permission, err := s.fkDao.ModPermissionByUsername(ctx, username, poolID)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	if permission == mod.PermAdmin {
		res.Role = mod.RoleModAdmin
		return res, nil
	}
	res.Role = mod.RoleUser
	return res, nil
}

func (s *Service) GlobalPush(ctx context.Context, username, appKey string, now time.Time) error {
	// super admin 才能操作
	perm, err := s.fawkesPerm(ctx, username, appKey, 0)
	if err != nil {
		return err
	}
	if perm != mod.FawkesSuperAdminPerm {
		return ecode.Error(ecode.RequestErr, "fawkes 超级管理员权限才能操作")
	}
	appInfo, err := s.fkDao.AppPass(ctx, appKey)
	if err != nil {
		return err
	}
	if appInfo == nil {
		return ecode.Error(ecode.RequestErr, "appKey不存在")
	}
	key := fmt.Sprintf("lock_%s", appInfo.MobiApp)
	ok, err := s.fkDao.TryLock(ctx, key, 1200)
	if err != nil {
		return err
	}
	if !ok {
		return ecode.Error(ecode.RequestErr, "20分钟内仅能发起一次推送")
	}
	if err := s.fkDao.BroadcastPush(ctx, appInfo.MobiApp, now); err != nil {
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.fkDao.UnLock(ctx, key); err != nil {
				log.Error("%+v", err)
			}
		})
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		if _, err := s.fkDao.AddLog(ctx, appKey, string(mod.EnvProd), "mod", "全局推送更新", "", username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) fawkesPerm(ctx context.Context, username, appKey string, poolID int64) (mod.FawkesPerm, error) {
	// 超管权限
	if username != "" {
		supervisorRole, err := s.fkDao.AuthSupervisor(ctx, username)
		if err != nil {
			return 0, err
		}
		if len(supervisorRole) > 0 {
			return mod.FawkesSuperAdminPerm, nil
		}
		if appKey != "" {
			// App 管理员权限
			user, err := s.fkDao.AuthUser(ctx, appKey, username)
			if err != nil {
				return 0, err
			}
			// 角色：ADMIN-1,DEV-2,TEST-3,DEVOPS-4,VISITOR-5
			if user != nil && user.Role == 1 {
				return mod.FawkesAppAdminPerm, nil
			}
		}
	}
	if poolID != 0 {
		permission, err := s.fkDao.ModPermissionByUsername(ctx, username, poolID)
		if err != nil && err != xsql.ErrNoRows {
			return 0, err
		}
		if permission == mod.PermAdmin {
			return mod.FawkesModAdminPerm, nil
		}
	}
	return mod.FawkesUserPerm, nil
}

func (s *Service) isDisableModule(poolName, moduleName string) bool {
	for _, v := range s.c.Mod.DisableModule[poolName] {
		if moduleName == v {
			return true
		}
	}
	return false
}

func checkVer(verParam string) (string, error) {
	if verParam == "" {
		return "", nil
	}
	var val []map[mod.Condition]int64
	if err := json.Unmarshal([]byte(verParam), &val); err != nil {
		return "", err
	}
	var v []map[mod.Condition]int64
	for _, verm := range val { // 处理 [{}]
		if len(verm) == 0 {
			continue
		}
		v = append(v, verm)
	}
	if len(v) == 0 {
		return "", nil
	}
	//nolint:gomnd
	if len(v) > 10 {
		return "", errors.New("不要超过十组配置")

	}
	if ok := func() bool {
		for _, vs := range v {
			//nolint:gomnd
			if len(vs) > 2 {
				return false
			}
			var ltValue, gtValue int64
			for condition, value := range vs {
				if value < 0 {
					return false
				}
				if !condition.Valid() {
					return false
				}
				switch condition {
				case mod.ConditionLt:
					ltValue = value
				case mod.ConditionGt:
					gtValue = value
				case mod.ConditionLe:
					ltValue = value + 1
				case mod.ConditionGe:
					gtValue = value - 1
				}
				if ltValue != 0 && ltValue < gtValue {
					return false
				}
			}
		}
		return true
	}(); !ok {
		return "", errors.New("参数错误")
	}
	bs, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (s *Service) RoleApplyList(ctx context.Context, appKey, username string, state mod.ApplyState, pn, ps int64) ([]*mod.RoleApply, *mod.Page, error) {
	perm, err := s.fawkesPerm(ctx, username, appKey, 0)
	if err != nil {
		return nil, nil, err
	}
	if perm > mod.FawkesUserPerm {
		username = ""
	}
	offset := (pn - 1) * ps
	limit := ps
	var (
		res  []*mod.RoleApply
		page *mod.Page
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		res, err = s.fkDao.ModRoleApplyList(ctx, appKey, username, state, offset, limit)
		return err
	})
	g.Go(func(ctx context.Context) error {
		count, err := s.fkDao.ModRoleApplyCount(ctx, appKey, username, state)
		if err != nil && err != xsql.ErrNoRows {
			return err
		}
		page = &mod.Page{Total: count, Pn: pn, Ps: ps}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}
	if len(res) == 0 {
		return nil, page, nil
	}
	var poolIDs []int64
	poolExist := map[int64]struct{}{}
	for _, r := range res {
		if _, ok := poolExist[r.PoolID]; !ok {
			poolIDs = append(poolIDs, r.PoolID)
			poolExist[r.PoolID] = struct{}{}
		}
	}
	poolm, err := s.fkDao.ModPoolByPoolIDs(ctx, poolIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range res {
		r.Pool = poolm[r.PoolID]
	}
	return res, page, nil
}

func (s *Service) RoleAdd(ctx context.Context, operator string, poolID int64, username string, permission mod.Perm) error {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, operator, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesAppAdminPerm {
		return ecode.Error(ecode.RequestErr, "fawkes管理员及以上权限才能操作")
	}
	id, err := s.fkDao.ModPermissionAdd(ctx, username, poolID, permission)
	if err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("权限管理:%v,参数:%v,%v,%v,权限ID:%v", pool.Name, username, poolID, permission, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限新增", logInfo, operator); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) RoleApplyAdd(ctx context.Context, username string, poolID int64, permission mod.Perm, operator string) error {
	pool, err := s.fkDao.ModPoolByID(ctx, poolID)
	if err != nil {
		return err
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	curPerm, err := s.fkDao.ModPermissionByUsername(ctx, username, poolID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	if permission == curPerm {
		return ecode.Error(ecode.RequestErr, "当前权限已申请成功，请勿重复申请")
	}
	// 通知申请人
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s) mod:%s\"的管理员 \"%s\" 已收到了您的 %s 权限申请, 请耐心等待审核结果。",
		appInfo.Name, appInfo.AppKey, pool.Name, operator, permission), username, s.c.Comet.FawkesAppID)
	// 通知管理员
	_ = s.fkDao.WechatCardMessageNotify(
		"mod资源池权限申请提醒",
		fmt.Sprintf("%s 提交了一个mod资源池权限申请\n应用：%s(%s)\n资源池:%s %s \n审核员：%s", username, appInfo.Name, appInfo.AppKey, pool.Name, permission, operator),
		fmt.Sprintf("http://fawkes.bilibili.co/#/mod-manage/auth?app_key=%s&state=checking&pn=1", appInfo.AppKey),
		"",
		operator,
		s.c.Comet.FawkesAppID)
	id, err := s.fkDao.ModRoleApplyAdd(ctx, pool.AppKey, username, poolID, permission, operator)
	if err != nil {
		log.Error("%v", err)
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("权限管理:%v,参数:%v,%v,%v,权限ID:%v", pool.Name, username, poolID, permission, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限申请", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) RoleApplyProcess(ctx context.Context, username string, poolID int64) (*mod.RoleApply, error) {
	res, err := s.fkDao.ModRoleApplyByUsernamePoolID(ctx, username, poolID)
	if err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	return res, nil
}

func (s *Service) RoleOperatorList(ctx context.Context, poolID int64) ([]*mod.Permission, error) {
	return s.fkDao.ModRoleOperatorList(ctx, poolID)
}

func (s *Service) RoleApplyPass(ctx context.Context, username string, applyID int64) error {
	apply, err := s.fkDao.ModRoleApply(ctx, applyID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, apply.PoolID)
	if err != nil {
		return err
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesModAdminPerm {
		return ecode.Error(ecode.RequestErr, "对应资源池管理员及以上权限才能操作")
	}
	if _, err := s.fkDao.ModRoleApplyPass(ctx, apply); err != nil {
		return err
	}
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 通过了您的mod发布申请",
		appInfo.Name, appInfo.AppKey, username), apply.Username, s.c.Comet.FawkesAppID)
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,申请人:%v,申请ID:%v", pool.Name, apply.Username, applyID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限申请通过", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) RoleApplyRefuse(ctx context.Context, username string, applyID int64) error {
	apply, err := s.fkDao.ModRoleApply(ctx, applyID)
	if err != nil {
		return err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, apply.PoolID)
	if err != nil {
		return err
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesModAdminPerm {
		return ecode.Error(ecode.RequestErr, "对应资源池管理员及以上权限才能操作")
	}
	if err := s.fkDao.ModRoleApplyRefuse(ctx, applyID); err != nil {
		return err
	}
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 拒绝了您的mod发布申请",
		appInfo.Name, appInfo.AppKey, username), apply.Username, s.c.Comet.FawkesAppID)
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,申请人:%v,申请ID:%v", pool.Name, apply.Username, applyID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "权限申请拒绝", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionApplyAdd(ctx context.Context, username string, versionID int64, operator, remark string) error {
	version, err := s.fkDao.ModVersionByID(ctx, versionID)
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
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	applyID, err := s.fkDao.ModVersionApplyExist(ctx, versionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	if applyID > 0 {
		return ecode.Error(ecode.RequestErr, "已存在未处理的申请单，先联系资源池管理员处理")
	}
	configApplyID, err := s.fkDao.ModVersionConfigApplyExist(ctx, versionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	grayApplyID, err := s.fkDao.ModVersionGrayApplyExist(ctx, versionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	if configApplyID == 0 && grayApplyID == 0 && version.Released {
		return ecode.Error(ecode.RequestErr, "未有任何变更，无需发布")
	}
	id, err := s.fkDao.ModVersionApplyAdd(ctx, pool.AppKey, username, versionID, operator, remark, time.Now())
	if err != nil {
		return err
	}
	// 通知申请人
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 已收到了您的mod(%s)版本(%d)发布申请, 请耐心等待审核结果。",
		appInfo.Name, appInfo.AppKey, operator, module.Name, version.Version), username, s.c.Comet.FawkesAppID)
	// 通知管理员
	_ = s.fkDao.WechatCardMessageNotify(
		"mod发布申请提醒",
		fmt.Sprintf("%s 提交了一个mod发布申请\n应用：%s(%s)\nmod:%s 版本:%d \n审核员：%s", username, appInfo.Name, appInfo.AppKey, module.Name, version.Version, operator),
		fmt.Sprintf("http://fawkes.bilibili.co/#/mod-manage/index?app_key=%s", appInfo.AppKey),
		"",
		operator,
		s.c.Comet.FawkesAppID)
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,版本ID:%v,申请ID:%v", pool.Name, module.Name, version.Version, version.ID, id)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "版本发布申请", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionApplyNotify(ctx context.Context, username, appKey string) (int64, error) {
	return s.fkDao.ModVersionApplyNotify(ctx, appKey, username)
}

func (s *Service) VersionApplyList(ctx context.Context, username, appKey string) ([]*mod.VersionApply, error) {
	perm, err := s.fawkesPerm(ctx, username, appKey, 0)
	if err != nil {
		return nil, err
	}
	var tmpUsername string
	if perm < mod.FawkesAppAdminPerm {
		tmpUsername = username
	}
	res, err := s.fkDao.ModVersionApplyList(ctx, appKey, tmpUsername)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(res) == 0 {
		return nil, nil
	}
	var versionIDs []int64
	versionExist := map[int64]struct{}{}
	for _, val := range res {
		if _, ok := versionExist[val.VersionID]; !ok {
			versionIDs = append(versionIDs, val.VersionID)
			versionExist[val.VersionID] = struct{}{}
		}
	}
	version, err := s.fkDao.ModVersionByVersionIDs(ctx, versionIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var moduleIDs []int64
	moduleExist := map[int64]struct{}{}
	for _, val := range version {
		if _, ok := moduleExist[val.ModuleID]; !ok {
			moduleIDs = append(moduleIDs, val.ModuleID)
			moduleExist[val.ModuleID] = struct{}{}
		}
	}
	module, err := s.fkDao.ModModuleByModuleIDs(ctx, moduleIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var poolIDs []int64
	poolExist := map[int64]struct{}{}
	for _, val := range module {
		if _, ok := poolExist[val.PoolID]; !ok {
			poolIDs = append(poolIDs, val.PoolID)
			poolExist[val.PoolID] = struct{}{}
		}
	}
	pool, err := s.fkDao.ModPoolByPoolIDs(ctx, poolIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, val := range res {
		v, ok := version[val.VersionID]
		if !ok {
			continue
		}
		val.Version = v
		m, ok := module[v.ModuleID]
		if !ok {
			continue
		}
		val.Module = m
		p, ok := pool[m.PoolID]
		if !ok {
			continue
		}
		val.Pool = p
	}
	return res, nil
}

// nolint:gocognit
func (s *Service) VersionApplyOverview(ctx context.Context, applyID int64) (*mod.VersionOverView, error) {
	apply, err := s.fkDao.ModVersionApply(ctx, applyID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var (
		version     *mod.Version
		module      *mod.Module
		pool        *mod.Pool
		config      *mod.Config
		gray        *mod.Gray
		configApply *mod.ConfigApply
		grayApply   *mod.GrayApply
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		if version, err = s.fkDao.ModVersionByID(ctx, apply.VersionID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		if module, err = s.fkDao.ModModuleByID(ctx, version.ModuleID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		if pool, err = s.fkDao.ModPoolByID(ctx, module.PoolID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if config, err = s.fkDao.ModVersionConfig(ctx, apply.VersionID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if gray, err = s.fkDao.ModVersionGray(ctx, apply.VersionID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if configApply, err = s.fkDao.ModVersionConfigApply(ctx, apply.VersionID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		if configApply.State != mod.ApplyStateChecking {
			configApply = nil
			return nil
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if grayApply, err = s.fkDao.ModVersionGrayApply(ctx, apply.VersionID); err != nil {
			if err == xsql.ErrNoRows {
				return nil
			}
			return err
		}
		if grayApply.State != mod.ApplyStateChecking {
			grayApply = nil
			return nil
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	var onlineConfig *mod.OnlineConfig
	if version.Released {
		onlineConfig = &mod.OnlineConfig{
			Config: config,
			Gray:   gray,
		}
	}
	bs, _ := json.Marshal(&mod.VersionOverView{
		Version:      version,
		OnlineConfig: onlineConfig,
	})
	if configApply == nil && config != nil {
		configApply = &mod.ConfigApply{Config: *config}
	}
	if grayApply == nil && gray != nil {
		grayApply = &mod.GrayApply{Gray: *gray}
	}
	toOnlineConfig := &mod.ToOnlineConfig{
		Config: configApply,
		Gray:   grayApply,
	}
	return &mod.VersionOverView{
		ID:             apply.ID,
		AppKey:         apply.AppKey,
		Username:       apply.Username,
		VersionID:      apply.VersionID,
		Operator:       apply.Operator,
		Remark:         apply.Remark,
		Ctime:          apply.Ctime,
		Mtime:          apply.Mtime,
		Version:        version,
		Module:         module,
		Pool:           pool,
		OnlineConfig:   onlineConfig,
		ToOnlineConfig: toOnlineConfig,
		OnlineHash:     hash(bs),
	}, nil
}

func (s *Service) VersionApplyPass(ctx context.Context, username string, applyID int64, onlineHash string, now time.Time) error {
	apply, err := s.fkDao.ModVersionApply(ctx, applyID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil
		}
		return err
	}
	if apply.State != mod.ApplyStateChecking {
		return ecode.Error(ecode.RequestErr, "发布单已被处理")
	}
	version, err := s.fkDao.ModVersionByID(ctx, apply.VersionID)
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
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesModAdminPerm {
		return ecode.Error(ecode.RequestErr, "对应资源池管理员及以上权限才能操作")
	}
	config, err := s.fkDao.ModVersionConfigApply(ctx, apply.VersionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	if config != nil && config.State != mod.ApplyStateChecking {
		config = nil
	}
	gray, err := s.fkDao.ModVersionGrayApply(ctx, apply.VersionID)
	if err != nil && err != xsql.ErrNoRows {
		return err
	}
	if gray != nil && gray.State != mod.ApplyStateChecking {
		gray = nil
	}
	var onlineConfig *mod.OnlineConfig
	if version.Released {
		onlineConfig = &mod.OnlineConfig{}
		g := errgroup.WithCancel(ctx)
		g.Go(func(ctx context.Context) error {
			config, err := s.fkDao.ModVersionConfig(ctx, apply.VersionID)
			if err != nil {
				if err == xsql.ErrNoRows {
					return nil
				}
				return err
			}
			onlineConfig.Config = config
			return nil
		})
		g.Go(func(ctx context.Context) error {
			gray, err := s.fkDao.ModVersionGray(ctx, apply.VersionID)
			if err != nil {
				if err == xsql.ErrNoRows {
					return nil
				}
				return err
			}
			onlineConfig.Gray = gray
			return nil
		})
		if err := g.Wait(); err != nil {
			return err
		}
	}
	bs, _ := json.Marshal(&mod.VersionOverView{
		Version:      version,
		OnlineConfig: onlineConfig,
	})
	expectHash := hash(bs)
	if onlineHash != hash(bs) {
		log.Error("VersionApplyPass hash got:%v,expect:%v", onlineHash, expectHash)
		return ecode.Error(ecode.RequestErr, "线上配置已变更，请重新预览")
	}
	if err := s.fkDao.ModVersionApplyPass(ctx, apply, config, gray, true, xtime.Time(now.Unix())); err != nil {
		return err
	}
	s.event.Publish(VersionRelease, &VersionReleaseArgs{Ctx: utils.CopyTrx(ctx), VersionId: apply.VersionID, UserName: username})
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 通过了您的权限申请",
		appInfo.Name, appInfo.AppKey, username), apply.Username, s.c.Comet.FawkesAppID)
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,版本ID:%v,申请ID:%v", pool.Name, module.Name, version.Version, version.ID, applyID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "版本发布通过", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) VersionApplyRefuse(ctx context.Context, username string, applyID int64) error {
	apply, err := s.fkDao.ModVersionApply(ctx, applyID)
	if err != nil {
		if err == xsql.ErrNoRows {
			return nil
		}
		return err
	}
	if apply.State != mod.ApplyStateChecking {
		return ecode.Error(ecode.RequestErr, "发布单已被处理")
	}
	version, err := s.fkDao.ModVersionByID(ctx, apply.VersionID)
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
	perm, err := s.fawkesPerm(ctx, username, pool.AppKey, pool.ID)
	if err != nil {
		return err
	}
	if perm < mod.FawkesModAdminPerm {
		return ecode.Error(ecode.RequestErr, "对应资源池管理员及以上权限才能操作")
	}
	appInfo, err := s.fkDao.AppPass(ctx, pool.AppKey)
	if err != nil {
		return err
	}
	if err := s.fkDao.ModVersionApplyRefuse(ctx, apply.ID, apply.VersionID); err != nil {
		return err
	}
	_ = s.fkDao.WechatMessageNotify(fmt.Sprintf("\"%s(%s)\"的管理员 \"%s\" 拒绝了您的权限申请",
		appInfo.Name, appInfo.AppKey, username), apply.Username, s.c.Comet.FawkesAppID)
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,资源:%v,版本号:%v,版本ID:%v,申请ID:%v", pool.Name, module.Name, version.Version, version.ID, applyID)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvProd), "mod", "版本发布拒绝", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

func (s *Service) SyncPool(ctx context.Context, appKey string, moduleID int64) (*mod.SyncPool, error) {
	module, err := s.fkDao.ModModuleByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, err
	}
	toPool, err := s.fkDao.ModPoolByName(ctx, appKey, pool.Name)
	if err != nil {
		if err != xsql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}
	res := &mod.SyncPool{
		ID:   toPool.ID,
		Name: toPool.Name,
	}
	toModule, err := s.fkDao.ModModuleByName(ctx, toPool.ID, module.Name)
	if err != nil {
		if err != xsql.ErrNoRows {
			return nil, err
		}
		return res, nil
	}
	res.Module = &mod.SyncModule{
		ID:   toModule.ID,
		Name: toModule.Name,
	}
	return res, nil
}

func (s *Service) SyncVersionList(ctx context.Context, moduleID int64) ([]*mod.SyncVersion, error) {
	version, err := s.fkDao.ModVersionList(ctx, moduleID, mod.EnvTest, 0, -1)
	if err != nil {
		return nil, err
	}
	var res []*mod.SyncVersion
	for _, val := range version {
		r := &mod.SyncVersion{
			ID:      val.ID,
			Version: val.Version,
		}
		res = append(res, r)
	}
	return res, nil
}

func (s *Service) SyncAdd(ctx context.Context, username string, sync *mod.SyncParam) (toVersion *mod.Version, err error) {
	appVer, err := checkVer(sync.ConfigAppVer)
	if err != nil {
		return nil, ecode.Error(ecode.RequestErr, "config_app_ver"+err.Error())
	}
	sysVer, err := checkVer(sync.ConfigSysVer)
	if err != nil {
		return nil, ecode.Error(ecode.RequestErr, "config_sys_ver"+err.Error())
	}
	if sync.GrayWhitelistURL != "" && !strings.Contains(sync.GrayWhitelistURL, _whitelistID) {
		return nil, ecode.Error(ecode.RequestErr, "whitelist_url 参数错误")
	}
	version, err := s.fkDao.ModVersionByID(ctx, sync.VersionID)
	if err != nil {
		return nil, err
	}
	if version.Env != mod.EnvTest {
		return nil, ecode.Error(ecode.RequestErr, "非测试版本不能进行同步")
	}
	module, err := s.fkDao.ModModuleByID(ctx, sync.ToModuleID)
	if err != nil {
		return nil, err
	}
	pool, err := s.fkDao.ModPoolByID(ctx, module.PoolID)
	if err != nil {
		return nil, err
	}
	c := &mod.Config{
		Priority: sync.ConfigPriority,
		AppVer:   appVer,
		SysVer:   sysVer,
		Stime:    sync.ConfigStime,
		Etime:    sync.ConfigEtime,
	}
	if reflect.DeepEqual(c, _emptyConfig) {
		c = nil
	}
	g := &mod.Gray{
		Strategy:       sync.GrayStrategy,
		Salt:           sync.GraySalt,
		BucketStart:    sync.GrayBucketStart,
		BucketEnd:      sync.GrayBucketEnd,
		Whitelist:      xstr.JoinInts(sync.GrayWhitelist),
		WhitelistURL:   sync.GrayWhitelistURL,
		ManualDownload: sync.GrayManualDownload,
	}
	if reflect.DeepEqual(g, _emptyGray) {
		g = nil
	}
	toVersionID, err := s.fkDao.ModSyncAdd(ctx, sync.ToModuleID, sync.ToVersionID, version, c, g)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		logInfo := fmt.Sprintf("资源池:%v,原始版本ID:%v,目标资源ID:%v,目标版本ID:%v,配置:%+v,灰度:%+v", pool.Name, sync.ToModuleID, sync.ToVersionID, toVersionID, c, g)
		if _, err := s.fkDao.AddLog(ctx, pool.AppKey, string(mod.EnvTest), "mod", "点对点同步", logInfo, username); err != nil {
			log.Error("%+v", err)
		}
	})
	return &mod.Version{
		ID: toVersionID,
	}, nil
}

func (s *Service) SyncVersionInfo(ctx context.Context, versionID int64) (*mod.SyncVersionInfo, error) {
	var (
		version *mod.Version
		file    *mod.File
		config  *mod.Config
		gray    *mod.Gray
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		version, err = s.fkDao.ModVersionByID(ctx, versionID)
		return err
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if file, err = s.fkDao.ModFile(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if config, err = s.fkDao.ModVersionConfig(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		var err error
		if gray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	if version.Env != mod.EnvTest {
		return nil, ecode.Error(ecode.RequestErr, "非测试版本不能进行同步")
	}
	return &mod.SyncVersionInfo{
		Version: version,
		File:    file,
		Config:  config,
		Gray:    gray,
	}, nil
}

// GetProdMod 获取生产环境MOD全量信息
func (s *Service) GetProdMod(ctx context.Context, versionID int64) (m *mod.Mod, err error) {
	var (
		pool        *mod.Pool
		module      *mod.Module
		version     *mod.Version
		config      *mod.Config
		gray        *mod.Gray
		file        *mod.File
		patches     []*mod.Patch
		versionList []*mod.Version
	)
	if version, err = s.fkDao.ModVersionByID(ctx, versionID); err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "%v", err)
		return
	}
	if version.Env == mod.EnvTest {
		log.Infoc(ctx, "不支持测试环境查询")
		return
	}
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		if module, err = s.fkDao.ModModuleByID(ctx, version.ModuleID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if module == nil {
			return nil
		}
		if pool, err = s.fkDao.ModPoolByID(ctx, module.PoolID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if file, err = s.fkDao.ModFile(ctx, version.FromVerID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		if patches, err = s.fkDao.ModPatchList(ctx, version.FromVerID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if versionList, err = s.fkDao.ModVersionList(ctx, version.ModuleID, mod.EnvProd, 0, -1); err != nil {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if config, err = s.fkDao.ModVersionConfig(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if gray, err = s.fkDao.ModVersionGray(ctx, versionID); err != nil && err != xsql.ErrNoRows {
			log.Errorc(ctx, "%v", err)
			return err
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if file == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("versionID:%v 找不到文件信息", versionID))
		log.Errorc(ctx, err.Error())
		return
	}
	m = &mod.Mod{
		Pool:    pool,
		Module:  module,
		Version: version,
		File:    file,
		Patches: getProdPatch(versionList, patches),
		Config:  config,
		Gray:    gray,
	}
	return
}

func hash(data []byte) string {
	h := sha1.New()
	_, _ = h.Write(data)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}
