package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/sync/errgroup"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	confmdl "go-gateway/app/app-svr/fawkes/service/model/config"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// ConfigVersionAdd add app config version.
func (s *Service) ConfigVersionAdd(c context.Context, appKey, env string, version string, versionCode int64, userName string) (err error) {
	var count int
	if count, err = s.fkDao.ExistConfigVersion(c, appKey, env, version, versionCode); err != nil {
		log.Error("%v", err)
		return
	}
	if count > 0 {
		return
	}
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var newCVID int64
	// 添加版本 返回新版本自增ID
	if newCVID, err = s.fkDao.TxSetConfigVersion(tx, appKey, env, version, versionCode, userName); err != nil {
		log.Error("%v", err)
		return
	}
	var cvids []int64
	// 获取默认版本ID
	if cvids, err = s.fkDao.ConfigVersionIDs(c, appKey, env, "default", 0); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cvid := range cvids {
		var ps []*confmdl.Publish
		// 获取默认版本的发布信息
		if ps, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, -1, -1); err != nil {
			log.Error("%v", err)
			return
		}
		for _, p := range ps {
			// 发布状态分为历史版本(-1)和当前版本(1)，此处获取当前发布版本
			if p.State == confmdl.ConfigPublishStateNow {
				var defaultConfig []*confmdl.Config
				// 获取已归档的默认default配置
				if defaultConfig, err = s.fkDao.ConfigFile(c, appKey, env, p.CV); err != nil {
					log.Error("%v", err)
					return
				}
				var (
					sqls []string
					args []interface{}
				)
				for _, dc := range defaultConfig {
					if dc.State == confmdl.ConfigStatDel {
						continue
					}
					sqls = append(sqls, "(?,?,?,?,?,?,?,?,?)")
					args = append(args, appKey, env, newCVID, dc.Group, dc.Key, dc.Value, confmdl.ConfigStatAdd, userName, dc.Desc)
				}
				// 将已经归档的default中有效(非删除)配置填充到新添加的版本
				if _, err = s.fkDao.TxAddConfigs(tx, sqls, args); err != nil {
					log.Error("%v", err)
					return
				}
			}
		}
	}
	return
}

// ConfigVersionList get config version list.
func (s *Service) ConfigVersionList(c context.Context, appKey, env string, pn, ps int) (res *confmdl.VersionResult, err error) {
	var total int
	if total, err = s.fkDao.ConfigVersionCount(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	var cvs []*confmdl.Version
	if cvs, err = s.fkDao.ConfigVersionList(c, appKey, env, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	if len(cvs) < 1 {
		return
	}
	var cvids []int64
	for _, version := range cvs {
		cvids = append(cvids, version.ID)
	}
	if len(cvids) > 0 {
		var cps map[int64][]*confmdl.Publish
		if cps, err = s.fkDao.ConfigPublishs(c, appKey, env, cvids); err != nil {
			log.Error("%v", err)
			return
		}
		var modifyCounts map[int64]int
		if modifyCounts, err = s.fkDao.ConfigModifyCounts(c, appKey, env, cvids); err != nil {
			log.Error("%v", err)
			return
		}
		for _, cv := range cvs {
			if cp, ok := cps[cv.ID]; ok {
				cv.CV = cp[0].CV
				cv.Operator = cp[0].Operator
				cv.Desc = cp[0].Desc
				cv.PTime = cp[0].PTime
			}
			if cmc, ok := modifyCounts[cv.ID]; ok {
				cv.ModifyNum = cmc
			}
		}
	}
	res = &confmdl.VersionResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: cvs,
	}
	return
}

// ConfigVersionHistory get app config version history.
func (s *Service) ConfigVersionHistory(c context.Context, appKey, env string, cvid int64, pn, ps int) (res *confmdl.HistoryResult, err error) {
	var (
		historys []*confmdl.Publish
		total    int
	)
	if total, err = s.fkDao.ConfigPublishCountByCvid(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if historys, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &confmdl.HistoryResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: historys,
	}
	return
}

// AppConfigVersionHistoryByID get app config publish history by id
func (s *Service) AppConfigVersionHistoryByID(c context.Context, appKey string, cid int64) (res *confmdl.Publish, err error) {
	if res, err = s.fkDao.AppConfigVersionHistoryByID(c, appKey, cid); err != nil {
		log.Error("%v", err)
	}
	return
}

// ConfigVersionHistorys get app config version all.
func (s *Service) ConfigVersionHistorys(c context.Context, appKey, env string, pn, ps int) (res *confmdl.HistoryResult, err error) {
	var (
		historys []*confmdl.Publish
		total    int
	)
	if total, err = s.fkDao.ConfigPublishCount(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	if historys, err = s.fkDao.ConfigPublishAll(c, appKey, env, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &confmdl.HistoryResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: historys,
	}
	return
}

// ConfigVersionDel del app config version.
func (s *Service) ConfigVersionDel(c context.Context, cvid int64) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// TODO default con't del.
	var cv *confmdl.Version
	if cv, err = s.fkDao.ConfigVersionByID(c, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if cv.Version == "default" {
		return
	}
	if _, err = s.fkDao.TxDelConfigVersion(tx, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxDelAllConfig(tx, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxDelAllConfigPublish(tx, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxDelAllConfigFile(tx, cvid); err != nil {
		log.Error("%v", err)
	}
	return
}

// ConfigFastAdd fast add config.
// nolint:gocognit
func (s *Service) ConfigFastAdd(c context.Context, appKey, env string, cvid int64, group, key, value, userName, description string) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	group = strings.ToLower(group)
	key = strings.ToLower(key)
	var cvids []int64
	if cvid == 0 {
		if cvids, err = s.fkDao.ConfigVersionIDs(c, appKey, env, "*", 0); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		cvids = append(cvids, cvid)
	}
	for _, cvid := range cvids {
		var cs []*confmdl.Config
		if cs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil {
			log.Error("%v", err)
			return
		}
		for _, c := range cs {
			if c.Group == group && c.Key == key {
				if c.Value != value {
					if c.State == confmdl.ConfigStatAdd {
						if _, err = s.fkDao.TxAddConfig(tx, appKey, env, cvid, group, key, value, userName, description); err != nil {
							log.Error("%v", err)
						}
					} else {
						if _, err = s.fkDao.TxUpConfig(tx, appKey, env, cvid, group, key, value, userName, description); err != nil {
							log.Error("%v", err)
						}
					}
				} else {
					if c.State == confmdl.ConfigStatDel {
						if _, err = s.fkDao.TxUpConfigItemState(tx, appKey, env, cvid, group, key, confmdl.ConfigStatePublish); err != nil {
							log.Error("%v", err)
							return
						}
					}
					if c.Desc != description {
						if _, err = s.fkDao.TxUpConfigDesc(tx, appKey, env, cvid, description, group, key); err != nil {
							log.Error("%v", err)
						}
					}
				}
				return
			}
		}
		if _, err = s.fkDao.TxAddConfig(tx, appKey, env, cvid, group, key, value, userName, description); err != nil {
			log.Error("%v", err)
			return
		}
	}
	return
}

// ConfigPublishView publish view.
func (s *Service) ConfigPublishView(c context.Context, appKey, env string, cvid int64) (res []*confmdl.Diff, err error) {
	var ps []*confmdl.Publish
	if ps, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	var origins []*confmdl.Config
	if len(ps) > 0 {
		if cv := ps[0].CV; cv > 0 {
			if origins, err = s.fkDao.ConfigFile(c, appKey, env, cv); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
	var news []*confmdl.Config
	if news, err = s.fkDao.Config(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	for _, new := range news {
		if new.State == confmdl.ConfigStatePublish {
			continue
		}
		diff := &confmdl.Diff{
			AppKey:   new.AppKey,
			Env:      new.Env,
			CVID:     new.CVID,
			State:    new.State,
			Group:    new.Group,
			Key:      new.Key,
			Desc:     new.Desc,
			Operator: new.Operator,
			MTime:    new.MTime,
		}
		// 赋值 "修改后的value"
		if new.State != -1 {
			diff.New = &confmdl.Config{
				Value: new.Value,
			}
		}
		// 赋值 "修改前的value"
		for _, origin := range origins {
			if new.Group == origin.Group && new.Key == origin.Key {
				diff.Origin = &confmdl.Config{
					Value: origin.Value,
				}
			}
		}
		// 特殊情况判断
		if new.State == -1 && diff.Origin == nil {
			continue
		}
		res = append(res, diff)
	}
	return
}

// ConfigAdd add config.
// nolint:gocognit
func (s *Service) ConfigAdd(c context.Context, appKey, env string, cvid int64, params []*confmdl.Config, userName, description string) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var cs []*confmdl.Config
	if cs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	var ms, as, dsfalse, dstrue []*confmdl.Config
NEXT1:
	for _, param := range params {
		param.Group = strings.ToLower(param.Group)
		param.Key = strings.ToLower(param.Key)
		for _, c := range cs {
			if c.Group == param.Group && c.Key == param.Key {
				if c.Value != param.Value {
					if c.State == confmdl.ConfigStatAdd {
						as = append(as, param)
					} else {
						ms = append(ms, param)
					}
				} else if c.Desc != param.Desc {
					if _, err = s.fkDao.TxUpConfigDesc(tx, appKey, env, cvid, param.Desc, param.Group, param.Key); err != nil {
						log.Error("%v", err)
						return
					}
				}
				continue NEXT1
			}
		}
		as = append(as, param)
	}
NEXT2:
	for _, c := range cs {
		for _, param := range params {
			param.Group = strings.ToLower(param.Group)
			param.Key = strings.ToLower(param.Key)
			if c.Group == param.Group && c.Key == param.Key {
				// 兼容删除已存在的配置后，又新增了同名且同值的配置的情况
				if c.State == confmdl.ConfigStatDel {
					if _, err = s.fkDao.TxUpConfigItemState(tx, appKey, env, cvid, c.Group, c.Key, confmdl.ConfigStatePublish); err != nil {
						log.Error("%v", err)
						return
					}
					if c.Desc != param.Desc {
						if _, err = s.fkDao.TxUpConfigDesc(tx, appKey, env, cvid, param.Desc, param.Group, param.Key); err != nil {
							log.Error("%v", err)
							return
						}
					}
				}
				continue NEXT2
			}
		}
		if c.State != confmdl.ConfigStatAdd {
			dsfalse = append(dsfalse, c)
		} else {
			dstrue = append(dstrue, c)
			continue
		}
	}
	for _, m := range ms {
		if _, err = s.fkDao.TxUpConfig(tx, appKey, env, cvid, m.Group, m.Key, m.Value, userName, m.Desc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, a := range as {
		if _, err = s.fkDao.TxAddConfig(tx, appKey, env, cvid, a.Group, a.Key, a.Value, userName, a.Desc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, d := range dsfalse {
		if _, err = s.fkDao.TxDelConfig(tx, d.AppKey, d.Env, d.CVID, d.Group, d.Key, userName, d.Desc); err != nil {
			log.Error("%v", err)
		}
	}
	for _, d := range dstrue {
		if _, err = s.fkDao.TxDelConfig2(tx, d.AppKey, d.Env, d.CVID, d.Group, d.Key); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if _, err = s.fkDao.TxUpConfigVersionDesc(tx, cvid, description); err != nil {
		log.Error("%v", err)
	}
	return
}

// ConfigDiff get diff by cv.
func (s *Service) ConfigDiff(c context.Context, appKey, env string, cvid, cv int64) (res []*confmdl.Diff, err error) {
	var news []*confmdl.Config
	if news, err = s.fkDao.ConfigFile(c, appKey, env, cv); err != nil {
		log.Error("%v", err)
		return
	}
	var lastCV int64
	if lastCV, err = s.fkDao.ConfigLastCV(c, appKey, env, cvid, cv); err != nil {
		log.Error("%v", err)
		return
	}
	var origins []*confmdl.Config
	if lastCV > 0 {
		if origins, err = s.fkDao.ConfigFile(c, appKey, env, lastCV); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, new := range news {
		if new.State == confmdl.ConfigStatePublish {
			continue
		}
		if lastCV == 0 && new.State == confmdl.ConfigStatDel {
			continue
		}
		diff := &confmdl.Diff{
			AppKey:   new.AppKey,
			Env:      new.Env,
			CVID:     new.CVID,
			State:    new.State,
			Group:    new.Group,
			Key:      new.Key,
			Desc:     new.Desc,
			Operator: new.Operator,
			MTime:    new.MTime,
		}
		if new.State != -1 {
			diff.New = &confmdl.Config{
				Value: new.Value,
			}
		}
		for _, origin := range origins {
			if new.Group == origin.Group && new.Key == origin.Key {
				diff.Origin = &confmdl.Config{
					Value: origin.Value,
				}
			}
		}
		res = append(res, diff)
	}
	return
}

// ConfigPublish config publish.
// nolint:gocognit
func (s *Service) ConfigPublish(c context.Context, appKey, env string, cvid int64, userName string) (err error) {
	var cv = time.Now().UnixNano() / 1e6
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start", appKey, env, cvid, cv)
	var configs []*confmdl.Config
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config ", appKey, env, cvid, cv)
	if configs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil || len(configs) == 0 {
		log.Error("ConfigPublish %v or configs is nil", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get configs(%+v)", appKey, env, cvid, cv, configs)
	var (
		appInfo  *appmdl.APP
		contents = make(map[string][]byte)
		sqls     []string
		args     []interface{}
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start from config(%+v)", appKey, env, cvid, cv, configs)
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cTmp := range configs {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?)")
		args = append(args, cTmp.AppKey, cTmp.Env, cTmp.CVID, cv, cTmp.Group, cTmp.Key, cTmp.Value, cTmp.State, cTmp.Operator, cTmp.Desc)
		if cTmp.State != confmdl.ConfigStatDel {
			var configValue []byte
			// Web端使用Config功能的时候. 不会加密数据
			if appInfo.Platform == "web" && strings.HasPrefix(appKey, "web_") {
				configValue = []byte(cTmp.Value)
			} else {
				if configValue, err = s.AesEncrypt([]byte(cTmp.Value)); err != nil {
					log.Error("%v", err)
					return
				}
			}
			contents[fmt.Sprintf("%v.%v", cTmp.Group, cTmp.Key)] = configValue
		}
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success from contents(%+v)", appKey, env, cvid, cv, contents)
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config version info", appKey, env, cvid, cv)
	var version *confmdl.Version
	if version, err = s.fkDao.ConfigVersionByID(c, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get config version info(%+v)", appKey, env, cvid, cv, version)
	log.Error("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start insert config publish", appKey, env, cvid, cv)
	var mcv int64
	if mcv, err = s.fkDao.TxAddConfigPublish(tx, appKey, env, cvid, cv, version.Desc, userName); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success insert config publish mcv(%v)", appKey, env, cvid, cv, mcv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start file config(%+v)", appKey, env, cvid, cv, args)
	if _, err = s.fkDao.TxAddConfigFile(tx, sqls, args); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success file config", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start flush config(del state -1 and modify all to 3)", appKey, env, cvid, cv)
	if _, err = s.fkDao.TxFlushConfig(tx, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpConfigState(tx, appKey, env, cvid, confmdl.ConfigStatePublish); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success flush config", appKey, env, cvid, cv)
	var (
		mcvs   = strconv.FormatInt(mcv, 10)  // publish ID
		cvids  = strconv.FormatInt(cvid, 10) // version ID
		cvs    = strconv.FormatInt(cv, 10)   // time.Now().UnixNano() / 1e6
		folder = path.Join(s.c.LocalPath.LocalDir, "config", appKey, env, mcvs)
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write config file foler(%v)", appKey, env, cvid, cv, folder)
	var contentByte []byte
	if contentByte, err = json.Marshal(contents); err != nil {
		log.Error("%v", err)
		return
	}
	var zipFilePath string
	if zipFilePath, err = s.fkDao.WriteConfigFile(folder, fmt.Sprintf("%v_%v_%v", appKey, cvids, cvs), contentByte); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write config file zipFilePaht(%v)", appKey, env, cvid, cv, zipFilePath)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs zipFilePath(%v)", appKey, env, cvid, cv, zipFilePath)
	var cdnURL, md5Str string
	if cdnURL, md5Str, err = s.fkDao.UpBFSV2(path.Join("config", env), zipFilePath, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs cdnURL(%v),md5Str(%v)", appKey, env, cvid, cv, cdnURL, md5Str)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get last publish version", appKey, env, cvid, cv)
	var configPublishs []*confmdl.Publish
	if configPublishs, err = s.fkDao.ConfigPublish(context.Background(), appKey, env, cvid, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get last publish version configPublishs(%v)", appKey, env, cvid, cv, configPublishs)
	var diffs = make(map[int64]string)
	for _, configPublish := range configPublishs {
		var (
			patchPath     string
			patchFileName = fmt.Sprintf("%v_%v_%v_%v.patch", appKey, cvids, strconv.FormatInt(configPublish.CV, 10), cvs)
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start diff_file(%v) new(%v) old(%v)", appKey,
			env, cvid, cv, patchFileName, zipFilePath, configPublish.LocalPath)
		if patchPath, err = s.fkDao.DiffCmd(folder, patchFileName, zipFilePath, configPublish.LocalPath); err != nil {
			log.Error("%v", err)
			err = nil
			continue
		}
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success diff file %v", appKey, env, cvid, cv, patchPath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs diff file %v", appKey, env, cvid, cv, patchPath)
		patch := &model.Diff{}
		if patch.URL, patch.MD5, err = s.fkDao.UpBFSV2(path.Join("config", env), patchPath, appKey); err != nil {
			log.Error("%v", err)
			continue
		}
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs diff url(%v) md5(%v)", appKey,
			env, cvid, cv, patch.URL, patch.MD5)
		diffs[configPublish.CV] = patch.URL
		//nolint:gomnd
		if len(diffs) == 3 {
			break
		}
	}
	var ds []byte
	if len(diffs) > 0 {
		if ds, err = json.Marshal(diffs); err != nil {
			log.Error("%v", err)
			err = nil
		}
	}
	// update publish
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start update publish md5(%v) url(%v) path(%v) diff(%v)",
		appKey, env, cvid, cv, md5Str, cdnURL, zipFilePath, string(ds))
	if _, err = s.fkDao.TxUpConfigPublishFiles(tx, appKey, env, mcv, md5Str, cdnURL, zipFilePath, string(ds)); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success update publish md5 url path diff", appKey, env, cvid, cv)
	var (
		allNewPublishs map[int64]*confmdl.Publish
		app            *appmdl.APP
		cvids2         []int64
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get all publish to build total", appKey, env, cvid, cv)
	if allNewPublishs, err = s.fkDao.AllNewConfigPublish(context.Background(), appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get all publish to build total allNewPublishs(%+v)",
		appKey, env, cvid, cv, allNewPublishs)
	allNewPublishs[cvid] = &confmdl.Publish{
		CV:        cv,
		CVID:      cvid,
		URL:       cdnURL,
		LocalPath: zipFilePath,
		Diffs:     string(ds),
		MD5:       md5Str,
	}
	for _, anp := range allNewPublishs {
		cvids2 = append(cvids2, anp.CVID)
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config version(%cvids2) info and app info",
		appKey, env, cvid, cv, cvids2)
	var cpvs map[int64]*confmdl.Version
	if cpvs, err = s.fkDao.ConfigVersionByIDs(context.Background(), cvids2); err != nil {
		log.Error("%v", err)
		return
	}
	if app, err = s.fkDao.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get config version info(%+v) and app info(%+v)",
		appKey, env, cvid, cv, cpvs, app)
	var tf = &model.TotalFile{
		MCV:      strconv.FormatInt(mcv, 10),
		Platform: app.Platform,
		Version:  make(map[string]*model.FileVersion),
	}
	for cvid2, anp := range allNewPublishs {
		cpv, ok := cpvs[cvid2]
		if !ok {
			log.Error("cvid(%v) is not exist", cvid2)
			continue
		}
		var key string
		if cpv.Version == "default" {
			key = cpv.Version
		} else {
			key = strconv.FormatInt(cpv.VersionCode, 10)
		}
		if anp.MD5 == "" || anp.URL == "" || anp.CV == 0 {
			log.Warn("%v-%v publish faild md5(%v) url(%v) cv(%v)", cpv.Version, cpv.VersionCode, anp.MD5, anp.URL, anp.CV)
			continue
		}
		tfv := &model.FileVersion{
			Md5:     anp.MD5,
			URL:     anp.URL,
			Version: strconv.FormatInt(anp.CV, 10),
		}
		if anp.Diffs != "" {
			var d = make(map[string]string)
			if err = json.Unmarshal([]byte(anp.Diffs), &d); err != nil {
				log.Error("%v", err)
				err = nil
				continue
			}
			tfv.Diffs = make(map[string]string)
			tfv.Diffs = d
		}
		tf.Version[key] = tfv
	}
	var tfContentByte []byte
	if tfContentByte, err = json.Marshal(tf); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		TotalZipFilePath = make(map[string]string)
		TotalcdnURL      = make(map[string]string)
		mutexFile        sync.Mutex
		mutexBFS         sync.Mutex
	)
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() (err error) {
		var (
			totalFileName     = appKey
			totalPath, cdnURL string
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write total file %v %v %v", appKey, env,
			cvid, cv, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["lengqidong"] = totalPath
		mutexFile.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write total file TotalZipFilePath(%+v)",
			appKey, env, cvid, cv, TotalZipFilePath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs total file(%v)", appKey, env, cvid, cv, totalPath)
		if cdnURL, _, err = s.fkDao.UpBFSV2(path.Join("config", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["lengqidong"] = cdnURL
		mutexBFS.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs total file TotalcdnURL(%+v)",
			appKey, env, cvid, cv, TotalcdnURL)
		return
	})
	g.Go(func() (err error) {
		var (
			totalFileName     = fmt.Sprintf("%v_%v", appKey, mcv)
			totalPath, cdnURL string
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write total file %v %v %v", appKey, env,
			cvid, cv, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["gengxin"] = totalPath
		mutexFile.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write total file TotalZipFilePath(%+v)",
			appKey, env, cvid, cv, TotalZipFilePath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs total file(%v)", appKey, env, cvid, cv, totalPath)
		if cdnURL, _, err = s.fkDao.UpBFSV2(path.Join("config", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["gengxin"] = cdnURL
		mutexBFS.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs total file TotalcdnURL(%+v)",
			appKey, env, cvid, cv, TotalcdnURL)
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start form TotalZipFilePath(%+v) and TotalcdnURL(%+v)",
		appKey, env, cvid, cv, TotalZipFilePath, TotalcdnURL)
	var tfb, tub []byte
	if tfb, err = json.Marshal(TotalZipFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	if tub, err = json.Marshal(TotalcdnURL); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success form TotalZipFilePath(%v) and TotalcdnURL(%v)",
		appKey, env, cvid, cv, TotalZipFilePath, TotalcdnURL)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start update TotalZipFilePath(%v) and TotalcdnURL(%v)",
		appKey, env, cvid, cv, string(tfb), string(tub))
	if _, err = s.fkDao.TxUpConfigPublishTotal(tx, appKey, env, cvid, cv, string(tfb), string(tub)); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success update TotalZipFilePath and TotalcdnURL", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get last publish", appKey, env, cvid, cv)
	var lastcv int64
	if lastcv, err = s.fkDao.ConfigLastCV(context.Background(), appKey, env, cvid, cv); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) succtss get last publish id(%v)", appKey, env, cvid, cv, lastcv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start change last publish %v state to %v", appKey,
		env, cvid, cv, lastcv, confmdl.ConfigPublishStateHistory)
	if _, err = s.fkDao.TxUpConfigPublishState(tx, appKey, env, cvid, lastcv, confmdl.ConfigPublishStateHistory); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success change last publish state", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start change new publish %v state to %v", appKey,
		env, cvid, cv, cv, confmdl.ConfigPublishStateNow)
	if _, err = s.fkDao.TxUpConfigPublishState(tx, appKey, env, cvid, cv, confmdl.ConfigPublishStateNow); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success change new publish state", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success", appKey, env, cvid, cv)
	return
}

// Config get configs.
func (s *Service) Config(c context.Context, appKey, env string, cvid int64) (res map[string][]*confmdl.Config, err error) {
	var cs []*confmdl.Config
	if cs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[string][]*confmdl.Config)
	for _, c := range cs {
		if c.State == confmdl.ConfigStatDel {
			continue
		}
		res[c.Group] = append(res[c.Group], &confmdl.Config{
			Group:    c.Group,
			Key:      c.Key,
			Value:    c.Value,
			Operator: c.Operator,
			Desc:     c.Desc,
			MTime:    c.MTime,
		})
	}
	return
}

// ConfigPublishMultiple config publish multiple.
// nolint:gocognit
func (s *Service) ConfigPublishMultiple(c context.Context, appKey, env string, cvid int64, userName string, pubConf []*confmdl.PubConfig) (err error) {
	var cv = time.Now().UnixNano() / 1e6
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start", appKey, env, cvid, cv)
	var (
		configs        []*confmdl.Config
		publishConfigs []*confmdl.Config
		ps             []*confmdl.Publish
		origins        []*confmdl.Config
		filterSql      []string
		filterArg      []interface{}
	)
	filterArg = append(filterArg, appKey, env, cvid)
	// 查询远程配置
	if ps, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	if len(ps) > 0 {
		if cv := ps[0].CV; cv > 0 {
			if origins, err = s.fkDao.ConfigFile(c, appKey, env, cv); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config ", appKey, env, cvid, cv)
	if configs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil || len(configs) == 0 {
		log.Error("ConfigPublish %v or configs is nil", err)
		return
	}
	// 过滤出需要发布的配置
	for _, pConf := range pubConf {
		filterSql = append(filterSql, "(cgroup=? AND ckey=?)")
		filterArg = append(filterArg, pConf.Group, pConf.Key)
		for _, conf := range configs {
			if conf.Group == pConf.Group && conf.Key == pConf.Key {
				publishConfigs = append(publishConfigs, conf)
				break
			}
		}
	}
	// 获取远程配置中未变更的配置
	for _, oriConf := range origins {
		inOrigin := true
		for _, pConf := range pubConf {
			if oriConf.Group == pConf.Group && oriConf.Key == pConf.Key {
				inOrigin = false
				break
			}
		}
		if inOrigin {
			publishConfigs = append(publishConfigs, oriConf)
		}
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get configs(%+v)", appKey, env, cvid, cv, publishConfigs)
	var (
		appInfo  *appmdl.APP
		contents = make(map[string][]byte)
		sqls     []string
		args     []interface{}
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start from config(%+v)", appKey, env, cvid, cv, publishConfigs)
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cTmp := range publishConfigs {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?)")
		args = append(args, cTmp.AppKey, cTmp.Env, cTmp.CVID, cv, cTmp.Group, cTmp.Key, cTmp.Value, cTmp.State, cTmp.Operator, cTmp.Desc)
		if cTmp.State != confmdl.ConfigStatDel {
			var configValue []byte
			// Web端使用Config功能的时候. 不会加密数据
			if appInfo.Platform == "web" && strings.HasPrefix(appKey, "web_") {
				configValue = []byte(cTmp.Value)
			} else {
				if configValue, err = s.AesEncrypt([]byte(cTmp.Value)); err != nil {
					log.Error("%v", err)
					return
				}
			}
			contents[fmt.Sprintf("%v.%v", cTmp.Group, cTmp.Key)] = configValue
		}
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success from contents(%+v)", appKey, env, cvid, cv, contents)
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config version info", appKey, env, cvid, cv)
	var version *confmdl.Version
	if version, err = s.fkDao.ConfigVersionByID(c, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get config version info(%+v)", appKey, env, cvid, cv, version)
	var mcv int64
	if mcv, err = s.fkDao.TxAddConfigPublish(tx, appKey, env, cvid, cv, version.Desc, userName); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success insert config publish mcv(%v)", appKey, env, cvid, cv, mcv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start file config(%+v)", appKey, env, cvid, cv, args)
	if _, err = s.fkDao.TxAddConfigFile(tx, sqls, args); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success file config", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start flush config(del state -1 and modify all to 3)", appKey, env, cvid, cv)
	if _, err = s.fkDao.TxFlushConfigMultiple(tx, filterSql, filterArg); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxUpConfigStateMultiple(tx, filterSql, filterArg, confmdl.ConfigStatePublish); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success flush config", appKey, env, cvid, cv)
	var (
		mcvs   = strconv.FormatInt(mcv, 10)  // publish ID
		cvids  = strconv.FormatInt(cvid, 10) // version ID
		cvs    = strconv.FormatInt(cv, 10)   // time.Now().UnixNano() / 1e6
		folder = path.Join(s.c.LocalPath.LocalDir, "config", appKey, env, mcvs)
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write config file foler(%v)", appKey, env, cvid, cv, folder)
	var contentByte []byte
	if contentByte, err = json.Marshal(contents); err != nil {
		log.Error("%v", err)
		return
	}
	var zipFilePath string
	if zipFilePath, err = s.fkDao.WriteConfigFile(folder, fmt.Sprintf("%v_%v_%v", appKey, cvids, cvs), contentByte); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write config file zipFilePaht(%v)", appKey, env, cvid, cv, zipFilePath)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs zipFilePath(%v)", appKey, env, cvid, cv, zipFilePath)
	var cdnURL, md5Str string
	if cdnURL, md5Str, err = s.fkDao.UpBFSV2(path.Join("config", env), zipFilePath, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs cdnURL(%v),md5Str(%v)", appKey, env, cvid, cv, cdnURL, md5Str)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get last publish version", appKey, env, cvid, cv)
	var configPublishs []*confmdl.Publish
	if configPublishs, err = s.fkDao.ConfigPublish(context.Background(), appKey, env, cvid, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get last publish version configPublishs(%v)", appKey, env, cvid, cv, configPublishs)
	var diffs = make(map[int64]string)
	for _, configPublish := range configPublishs {
		var (
			patchPath     string
			patchFileName = fmt.Sprintf("%v_%v_%v_%v.patch", appKey, cvids, strconv.FormatInt(configPublish.CV, 10), cvs)
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start diff_file(%v) new(%v) old(%v)", appKey,
			env, cvid, cv, patchFileName, zipFilePath, configPublish.LocalPath)
		if patchPath, err = s.fkDao.DiffCmd(folder, patchFileName, zipFilePath, configPublish.LocalPath); err != nil {
			log.Error("%v", err)
			err = nil
			continue
		}
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success diff file %v", appKey, env, cvid, cv, patchPath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs diff file %v", appKey, env, cvid, cv, patchPath)
		patch := &model.Diff{}
		if patch.URL, patch.MD5, err = s.fkDao.UpBFSV2(path.Join("config", env), patchPath, appKey); err != nil {
			log.Error("%v", err)
			continue
		}
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs diff url(%v) md5(%v)", appKey,
			env, cvid, cv, patch.URL, patch.MD5)
		diffs[configPublish.CV] = patch.URL
		//nolint:gomnd
		if len(diffs) == 3 {
			break
		}
	}
	var ds []byte
	if len(diffs) > 0 {
		if ds, err = json.Marshal(diffs); err != nil {
			log.Error("%v", err)
			err = nil
		}
	}
	// update publish
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start update publish md5(%v) url(%v) path(%v) diff(%v)",
		appKey, env, cvid, cv, md5Str, cdnURL, zipFilePath, string(ds))
	if _, err = s.fkDao.TxUpConfigPublishFiles(tx, appKey, env, mcv, md5Str, cdnURL, zipFilePath, string(ds)); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success update publish md5 url path diff", appKey, env, cvid, cv)
	var (
		allNewPublishs map[int64]*confmdl.Publish
		app            *appmdl.APP
		cvids2         []int64
	)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get all publish to build total", appKey, env, cvid, cv)
	if allNewPublishs, err = s.fkDao.AllNewConfigPublish(context.Background(), appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get all publish to build total allNewPublishs(%+v)",
		appKey, env, cvid, cv, allNewPublishs)
	allNewPublishs[cvid] = &confmdl.Publish{
		CV:        cv,
		CVID:      cvid,
		URL:       cdnURL,
		LocalPath: zipFilePath,
		Diffs:     string(ds),
		MD5:       md5Str,
	}
	for _, anp := range allNewPublishs {
		cvids2 = append(cvids2, anp.CVID)
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get config version(%cvids2) info and app info",
		appKey, env, cvid, cv, cvids2)
	var cpvs map[int64]*confmdl.Version
	if cpvs, err = s.fkDao.ConfigVersionByIDs(context.Background(), cvids2); err != nil {
		log.Error("%v", err)
		return
	}
	if app, err = s.fkDao.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success get config version info(%+v) and app info(%+v)",
		appKey, env, cvid, cv, cpvs, app)
	var tf = &model.TotalFile{
		MCV:      strconv.FormatInt(mcv, 10),
		Platform: app.Platform,
		Version:  make(map[string]*model.FileVersion),
	}
	for cvid2, anp := range allNewPublishs {
		cpv, ok := cpvs[cvid2]
		if !ok {
			log.Error("cvid(%v) is not exist", cvid2)
			continue
		}
		var key string
		if cpv.Version == "default" {
			key = cpv.Version
		} else {
			key = strconv.FormatInt(cpv.VersionCode, 10)
		}
		if anp.MD5 == "" || anp.URL == "" || anp.CV == 0 {
			log.Warn("%v-%v publish faild md5(%v) url(%v) cv(%v)", cpv.Version, cpv.VersionCode, anp.MD5, anp.URL, anp.CV)
			continue
		}
		tfv := &model.FileVersion{
			Md5:     anp.MD5,
			URL:     anp.URL,
			Version: strconv.FormatInt(anp.CV, 10),
		}
		if anp.Diffs != "" {
			var d = make(map[string]string)
			if err = json.Unmarshal([]byte(anp.Diffs), &d); err != nil {
				log.Error("%v", err)
				err = nil
				continue
			}
			tfv.Diffs = make(map[string]string)
			tfv.Diffs = d
		}
		tf.Version[key] = tfv
	}
	var tfContentByte []byte
	if tfContentByte, err = json.Marshal(tf); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		TotalZipFilePath = make(map[string]string)
		TotalcdnURL      = make(map[string]string)
		mutexFile        sync.Mutex
		mutexBFS         sync.Mutex
	)
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() (err error) {
		var (
			totalFileName     = appKey
			totalPath, cdnURL string
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write total file %v %v %v", appKey, env,
			cvid, cv, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["lengqidong"] = totalPath
		mutexFile.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write total file TotalZipFilePath(%+v)",
			appKey, env, cvid, cv, TotalZipFilePath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs total file(%v)", appKey, env, cvid, cv, totalPath)
		if cdnURL, _, err = s.fkDao.UpBFSV2(path.Join("config", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["lengqidong"] = cdnURL
		mutexBFS.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs total file TotalcdnURL(%+v)",
			appKey, env, cvid, cv, TotalcdnURL)
		return
	})
	g.Go(func() (err error) {
		var (
			totalFileName     = fmt.Sprintf("%v_%v", appKey, mcv)
			totalPath, cdnURL string
		)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start write total file %v %v %v", appKey, env,
			cvid, cv, folder, totalFileName, string(tfContentByte))
		if totalPath, err = s.fkDao.WriteConfigFile(folder, totalFileName, tfContentByte); err != nil {
			log.Error("%v", err)
			return
		}
		mutexFile.Lock()
		TotalZipFilePath["gengxin"] = totalPath
		mutexFile.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success write total file TotalZipFilePath(%+v)",
			appKey, env, cvid, cv, TotalZipFilePath)
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start up bfs total file(%v)", appKey, env, cvid, cv, totalPath)
		if cdnURL, _, err = s.fkDao.UpBFSV2(path.Join("config", env), totalPath, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		mutexBFS.Lock()
		TotalcdnURL["gengxin"] = cdnURL
		mutexBFS.Unlock()
		log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success up bfs total file TotalcdnURL(%+v)",
			appKey, env, cvid, cv, TotalcdnURL)
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start form TotalZipFilePath(%+v) and TotalcdnURL(%+v)",
		appKey, env, cvid, cv, TotalZipFilePath, TotalcdnURL)
	var tfb, tub []byte
	if tfb, err = json.Marshal(TotalZipFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	if tub, err = json.Marshal(TotalcdnURL); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success form TotalZipFilePath(%v) and TotalcdnURL(%v)",
		appKey, env, cvid, cv, TotalZipFilePath, TotalcdnURL)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start update TotalZipFilePath(%v) and TotalcdnURL(%v)",
		appKey, env, cvid, cv, string(tfb), string(tub))
	if _, err = s.fkDao.TxUpConfigPublishTotal(tx, appKey, env, cvid, cv, string(tfb), string(tub)); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success update TotalZipFilePath and TotalcdnURL", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start get last publish", appKey, env, cvid, cv)
	var lastcv int64
	if lastcv, err = s.fkDao.ConfigLastCV(context.Background(), appKey, env, cvid, cv); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) succtss get last publish id(%v)", appKey, env, cvid, cv, lastcv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start change last publish %v state to %v", appKey,
		env, cvid, cv, lastcv, confmdl.ConfigPublishStateHistory)
	if _, err = s.fkDao.TxUpConfigPublishState(tx, appKey, env, cvid, lastcv, confmdl.ConfigPublishStateHistory); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success change last publish state", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) start change new publish %v state to %v", appKey,
		env, cvid, cv, cv, confmdl.ConfigPublishStateNow)
	if _, err = s.fkDao.TxUpConfigPublishState(tx, appKey, env, cvid, cv, confmdl.ConfigPublishStateNow); err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success change new publish state", appKey, env, cvid, cv)
	log.Info("ConfigPublish appkey(%v) env(%v) cvid(%v) cv(%v) success", appKey, env, cvid, cv)
	return
}

// AppConfigFile watch app config file.
func (s *Service) AppConfigFile(c context.Context, appKey, env string, cvid, cv int64) (res string, err error) {
	var cps []*confmdl.Publish
	if cps, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	var localPath string
	for _, cp := range cps {
		if cp.CV != cv {
			continue
		}
		localPath = strings.TrimRight(cp.LocalPath, ".zip")
		break
	}
	// TODO read zip.
	if localPath != "" {
		var f *os.File
		if f, err = os.Open(localPath); err != nil {
			log.Error("%v", err)
			return
		}
		var rb []byte
		if rb, err = ioutil.ReadAll(f); err != nil {
			log.Error("%v", err)
			return
		}
		var contents = make(map[string][]byte)
		if err = json.Unmarshal(rb, &contents); err != nil {
			log.Error("%v", err)
			return
		}
		var appInfo *appmdl.APP
		if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
			log.Error("%v", err)
			return
		}
		var rem = make(map[string]string)
		for gk, v := range contents {
			var v2 []byte
			if appInfo.Platform == "web" && strings.HasPrefix(appKey, "web_") {
				v2 = v
			} else {
				if v2, err = s.AesDecrypt(v); err != nil {
					log.Error("%v", err)
					return
				}
			}
			rem[gk] = string(v2)
		}
		var resb []byte
		if resb, err = json.Marshal(rem); err != nil {
			log.Error("%v", err)
			return
		}
		res = string(resb)
	}
	return
}

func (s *Service) AppConfigKeyPublishHistory(c context.Context, appKey, env, ckey, cgroup string, cvid int64) (res []*confmdl.Config, err error) {
	var (
		tmpValue string
		cfs      []*confmdl.Config
		filters  []*confmdl.Config
	)
	if cfs, err = s.fkDao.ConfigKeyPublishHistory(c, appKey, env, ckey, cgroup, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	for i := len(cfs) - 1; i >= 0; i-- {
		row := cfs[i]
		if row.Value != tmpValue {
			tmpValue = row.Value
			filters = append(filters, row)
		}
	}
	sort.Slice(filters, func(i, j int) bool {
		return filters[i].MTime-filters[j].MTime > 0
	})
	res = filters
	return
}

func (s *Service) AppConfigModifyCount(c context.Context, appKey string) (res *confmdl.ModifyCount, err error) {
	modifyItem := &confmdl.ModifyCount{Test: 0, Prod: 0}
	if modifyItem.Test, err = s.fkDao.ConfigModifyCountsAll(c, appKey, "test"); err != nil {
		log.Error("%v", err)
		return
	}
	if modifyItem.Prod, err = s.fkDao.ConfigModifyCountsAll(c, appKey, "prod"); err != nil {
		log.Error("%v", err)
		return
	}
	res = modifyItem
	return
}

// AddConfigBusinessV2 批量写入
func (s *Service) AddConfigBusinessV2(c context.Context, appKey, env string, params []*confmdl.Config, groupName, userName, description string) (err error) {
	appKeys := strings.Split(appKey, ",")
	for _, k := range appKeys {
		if err = s.AddConfigBusiness(c, k, env, params, groupName, userName, description); err != nil {
			log.Error("%v", err)
			break
		}
	}
	return
}

// AddConfigBusiness 添加业务Config项（限制写入单个Group组）
// nolint:gocognit
func (s *Service) AddConfigBusiness(c context.Context, appKey, env string, params []*confmdl.Config, groupName, userName, description string) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// 1. get default version_code cvid
	var (
		cvs  []*confmdl.Version
		cvid int64
	)
	if cvs, err = s.fkDao.GetDefaultConfigVersion(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cv := range cvs {
		if cv.Version == "default" {
			cvid = cv.ID
		}
	}
	if cvid == 0 {
		return
	}
	// 2. get diff conf
	var cs []*confmdl.Config
	if cs, err = s.fkDao.Config(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	var ms, as, dsfalse, dstrue []*confmdl.Config
NEXT1:
	for _, param := range params {
		param.Group = strings.ToLower(param.Group)
		param.Key = strings.ToLower(param.Key)
		for _, c := range cs {
			// 指定GROUP下的配置修改
			if param.Group != groupName {
				continue NEXT1
			}
			if c.Group == param.Group && c.Key == param.Key {
				if c.Value != param.Value {
					if c.State == confmdl.ConfigStatAdd {
						as = append(as, param)
					} else {
						ms = append(ms, param)
					}
				} else if c.Desc != param.Desc {
					if _, err = s.fkDao.TxUpConfigDesc(tx, appKey, env, cvid, param.Desc, param.Group, param.Key); err != nil {
						log.Error("%v", err)
						return
					}
				}
				continue NEXT1
			}
		}
		as = append(as, param)
	}
	for _, m := range ms {
		if _, err = s.fkDao.TxUpConfig(tx, appKey, env, cvid, m.Group, m.Key, m.Value, userName, m.Desc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, a := range as {
		if _, err = s.fkDao.TxAddConfig(tx, appKey, env, cvid, a.Group, a.Key, a.Value, userName, a.Desc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	for _, d := range dsfalse {
		if _, err = s.fkDao.TxDelConfig(tx, d.AppKey, d.Env, d.CVID, d.Group, d.Key, userName, d.Desc); err != nil {
			log.Error("%v", err)
		}
	}
	for _, d := range dstrue {
		if _, err = s.fkDao.TxDelConfig2(tx, d.AppKey, d.Env, d.CVID, d.Group, d.Key); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if _, err = s.fkDao.TxUpConfigVersionDesc(tx, cvid, description); err != nil {
		log.Error("%v", err)
	}
	return
}

// ConfigVersionHistory get app config version history.
func (s *Service) DefaultConfigHistory(c context.Context, appKey, env string, pn, ps int) (res *confmdl.HistoryResult, err error) {
	var (
		historys []*confmdl.Publish
		total    int
	)
	// 1. get default version_code cvid
	var (
		cvs  []*confmdl.Version
		cvid int64
	)
	if cvs, err = s.fkDao.GetDefaultConfigVersion(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cv := range cvs {
		if cv.Version == "default" {
			cvid = cv.ID
		}
	}
	if cvid == 0 {
		return
	}
	if total, err = s.fkDao.ConfigPublishCountByCvid(c, appKey, env, cvid); err != nil {
		log.Error("%v", err)
		return
	}
	if historys, err = s.fkDao.ConfigPublish(c, appKey, env, cvid, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &confmdl.HistoryResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: historys,
	}
	return
}

// DefaultConfig get app config
func (s *Service) DefaultConfig(c context.Context, appKey, env, business string) (res []*confmdl.Config, err error) {
	var (
		cvid int64
		cvs  []*confmdl.Version
	)
	if cvs, err = s.fkDao.GetDefaultConfigVersion(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cv := range cvs {
		if cv.Version == "default" {
			cvid = cv.ID
		}
	}
	if cvid == 0 {
		return
	}
	if res, err = s.fkDao.ConfigGroup(c, appKey, env, business, cvid); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppConfigSetKV set app config key value
func (s *Service) AppConfigSetKV(c context.Context, appKey, env, cgroup, ckey, value, userName string) (err error) {
	var (
		configItem *confmdl.Config
		cvs        []*confmdl.Version
		cvid       int64
	)
	if cvs, err = s.fkDao.GetDefaultConfigVersion(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cv := range cvs {
		if cv.Version == "default" {
			cvid = cv.ID
		}
	}
	if cvid == 0 {
		return
	}
	if configItem, err = s.fkDao.ConfigItem(c, appKey, env, cgroup, ckey, cvid); err != nil {
		return
	}
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// 如果没有索引到，则进行添加操作
	if configItem == nil {
		if _, err = s.fkDao.TxAddConfig(tx, appKey, env, cvid, cgroup, ckey, value, userName, "openapi添加. 缺省备注信息"); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		// 进行添加 OR 更新操作
		if _, err = s.fkDao.TxUpConfig(tx, appKey, env, cvid, cgroup, ckey, value, userName, configItem.Desc); err != nil {
			log.Error("%v", err)
			return
		}
	}
	return
}
func (s *Service) ConfigPublishDefault(c context.Context, appKey, env, username string) (err error) {
	// 1. get default version_code cvid
	var (
		cvs  []*confmdl.Version
		cvid int64
	)
	if cvs, err = s.fkDao.GetDefaultConfigVersion(c, appKey, env); err != nil {
		log.Error("%v", err)
		return
	}
	for _, cv := range cvs {
		if cv.Version == "default" {
			cvid = cv.ID
		}
	}
	if cvid == 0 {
		return
	}
	return s.ConfigPublish(c, appKey, env, cvid, username)
}

//
//func (s *Service) AppPaladinFeConfig(c context.Context) (res interface{}, err error) {
//	res = struct {
//		Conf string `json:"conf"`
//	}{feConfig}
//	return
//}
//
//func load() (err error) {
//	var ok bool
//	if feConfig, ok = confClient.Value2("fawkes-fe.json"); !ok {
//		err = errors.New("load fawkes-fe.json failed")
//	}
//	return
//}

//func reload() {
//	confClient.Watch("fawkes-fe.json")
//	// nolint:biligowordcheck
//	go func() {
//		for range confClient.Event() {
//			log.Info("config reload")
//			if err := load(); err != nil {
//				log.Error("config reload error (%v)", err)
//			}
//		}
//	}()
//}
