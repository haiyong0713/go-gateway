package pack

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	taskmdl "go-gateway/app/app-svr/fawkes/service/model/task"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

type PatchDeleteTask struct {
	fkDao *fawkes.Dao
	name  string
}

func (t *PatchDeleteTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	pdCfg := conf.Conf.Task.NasClean.PatchDelete
	keys, err := t.getAppKeys(ctx, pdCfg.ExcludeAppKey)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return railgun.MsgPolicyFailure
	}
	for _, v := range keys {
		var (
			patchPersisted map[string]*cdmdl.Patch
			patch          []*cdmdl.Patch
			err            error
		)
		log.Infoc(ctx, "appKey: %s", v)
		if patchPersisted, err = t.getPersistedPatch(ctx, v); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		if patch, err = t.fkDao.PatchByAppKey(ctx, v, taskmdl.Active); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		res := new(taskmdl.DeleteResult)
		// 逐条删除满足条件的patch文件
		for _, p := range patch {
			canDelete, err := t.deleteCheck(ctx, p, patchPersisted)
			if err != nil {
				log.Errorc(ctx, "%v", err)
				return railgun.MsgPolicyFailure
			}
			if !canDelete {
				continue
			}
			var fileBytes int64
			id, pPath, appKey := p.ID, p.PatchPath, p.AppKey
			if fileBytes, err = utils.DirSizeB(pPath); err != nil {
				log.Errorc(ctx, "patchId %d, filePath: %s calc file size error.", id, pPath)
			}
			if err = os.Remove(pPath); err != nil {
				res.FailedId = append(res.FailedId, p.ID)
				log.Errorc(ctx, "patchId[%d]，appKey[%s],delete file: %s FAILED! err: %+v", id, appKey, pPath, err)
			} else {
				log.Infoc(ctx, "patchId[%d]，appKey[%s],delete file: %s SUCCESS!", id, appKey, pPath)
				res.DeletedId = append(res.DeletedId, p.ID)
				_metricCleanNasSize.Add(float64(fileBytes), appKey, "PATCH")
				_metricCleanNasCount.Inc(appKey, "PATCH")
			}
		}
		// update删除状态
		if _, err = t.fkDao.UpdatePatchState(ctx, res.DeletedId, taskmdl.Deleted); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		log.Infoc(ctx, "appKey: %s, deleted count: %d, failed count: %d, statistical result: %+v", v, len(res.DeletedId), len(res.FailedId), res)
	}
	return railgun.MsgPolicyNormal
}

func (t *PatchDeleteTask) deleteCheck(ctx context.Context, p *cdmdl.Patch, patchPersisted map[string]*cdmdl.Patch) (canDelete bool, err error) {
	monthAgo := time.Now().AddDate(0, -conf.Conf.Task.NasClean.PatchDelete.Persistence, 0)
	if pp, ok := patchPersisted[fmt.Sprintf("%v_%v", p.TargetBuildID, p.OriginBuildID)]; ok {
		log.Infoc(ctx, fmt.Sprintf("网关需要使用的patch包-id:%d，不可以删除", pp.ID))
		return false, nil
	}
	if p.CTime.After(monthAgo) {
		log.Infoc(ctx, fmt.Sprintf("最近%d月内创建的patch包-id:%d，不可以删除", conf.Conf.Task.NasClean.PatchDelete.Persistence, p.ID))
		return false, nil
	}
	if p.PatchState == int64(taskmdl.Deleted) {
		log.Infoc(ctx, fmt.Sprintf("patch包-id:%d，已删除", p.ID))
		return false, nil
	}
	id, pPath, appKey := p.ID, p.PatchPath, p.AppKey
	if len(pPath) == 0 {
		log.Infoc(ctx, "patchId[%d]，appKey[%s] filePath is empty", id, appKey)
		return false, nil
	}
	if !utils.FileExists(pPath) {
		log.Infoc(ctx, "patchId[%d]，appKey[%s] filePath: %s doesn't exist.", id, appKey, pPath)
		if _, err = t.fkDao.UpdatePatchState(ctx, []int64{id}, taskmdl.FileNotExist); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
		return false, nil
	}
	return true, nil
}

// getPersistedPatch 需要继续保存的patch包
func (t *PatchDeleteTask) getPersistedPatch(ctx context.Context, appKey string) (res map[string]*cdmdl.Patch, err error) {
	var (
		versionIdsTest, versionIdsProd, buildIdsTest, buildIdsProd []int64
	)
	buildIds := make([]int64, 0)
	//取最近十个版本
	if versionIdsTest, err = t.fkDao.LastPackVersionIds(ctx, appKey, "test"); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if len(versionIdsTest) > 0 {
		if buildIdsTest, err = t.fkDao.PackBuildIdsByVersions(ctx, appKey, "test", versionIdsTest); err != nil {
			log.Error("%v", err)
			return
		}
		buildIds = append(buildIds, buildIdsTest...)
	}
	if versionIdsProd, err = t.fkDao.LastPackVersionIds(ctx, appKey, "prod"); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if len(versionIdsProd) > 0 {
		if buildIdsProd, err = t.fkDao.PackBuildIdsByVersions(ctx, appKey, "prod", versionIdsProd); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
		buildIds = append(buildIds, buildIdsProd...)
	}
	if len(buildIds) > 0 {
		res = make(map[string]*cdmdl.Patch)
		var row []*cdmdl.Patch
		if row, err = t.fkDao.PatchAll3(ctx, buildIds); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
		for _, v := range row {
			res[fmt.Sprintf("%v_%v", v.TargetBuildID, v.OriginBuildID)] = v
		}
	}
	return
}

func (t *PatchDeleteTask) TaskName() string {
	return t.name
}

func NewPatchDeleteTask(fkDao *fawkes.Dao, name string) *PatchDeleteTask {
	t := &PatchDeleteTask{
		fkDao: fkDao,
		name:  name,
	}
	return t
}

func (t *PatchDeleteTask) getAppKeys(ctx context.Context, excludeAppKey []string) (appKey []string, err error) {
	allApp, err := t.fkDao.AppAll(ctx)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	m := make(map[string]interface{})
	for _, v := range allApp {
		m[v.AppKey] = interface{}(nil)
	}
	for _, v := range excludeAppKey {
		delete(m, v)
	}
	for k := range m {
		appKey = append(appKey, k)
	}
	return
}
