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

type ChannelDeleteTask struct {
	fkDao *fawkes.Dao
	name  string
}

func (t *ChannelDeleteTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	channelCfg := conf.Conf.Task.NasClean.ChannelDelete
	keys, err := t.getAppKeys(ctx, channelCfg.ExcludeAppKey)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return railgun.MsgPolicyFailure
	}
	for _, v := range keys {
		var (
			generates []*cdmdl.Generate
			err       error
			res       taskmdl.DeleteResult
		)
		// 筛选出满足状态的渠道包
		if generates, err = t.fkDao.GenerateByAppKeyAndStatus(ctx, v, []int{cdmdl.GenerateSuccess, cdmdl.GenerateUpload, cdmdl.GenerateTest, cdmdl.GeneratePublish}, taskmdl.Active); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		for _, g := range generates {
			canDelete, err := t.deleteCheck(ctx, g)
			if err != nil {
				log.Errorc(ctx, "%v", err)
				return railgun.MsgPolicyFailure
			}
			if !canDelete {
				continue
			}
			var fileBytes int64
			id, pPath, appKey := g.ID, g.GeneratePath, g.AppKey
			if fileBytes, err = utils.DirSizeB(pPath); err != nil {
				log.Errorc(ctx, "id %d, filePath: %s calc file size error.", id, pPath)
				return railgun.MsgPolicyFailure
			}
			if err = os.Remove(pPath); err != nil {
				res.FailedId = append(res.FailedId, g.ID)
				log.Errorc(ctx, "id[%d]，appKey[%s],delete file: %s FAILED! err: %+v", id, appKey, pPath, err)
			} else {
				log.Infoc(ctx, "id[%d]，appKey[%s],delete file: %s SUCCESS!", id, appKey, pPath)
				res.DeletedId = append(res.DeletedId, g.ID)
				_metricCleanNasSize.Add(float64(fileBytes), appKey, "CHANNEL")
				_metricCleanNasCount.Inc(appKey, "CHANNEL")
			}
		}
		// update删除状态
		var affected int64
		if affected, err = t.fkDao.UpdateChannelPackState(ctx, res.DeletedId, taskmdl.Deleted); err != nil {
			log.Errorc(ctx, "%v", err)
			return railgun.MsgPolicyFailure
		}
		res.AffectedRows = res.AffectedRows + affected
		log.Infoc(ctx, "appKey: %s, deleted count: %d, failed count: %d, statistical result: %+v", v, len(res.DeletedId), len(res.FailedId), res)
	}
	return railgun.MsgPolicyNormal
}

func (t *ChannelDeleteTask) deleteCheck(ctx context.Context, g *cdmdl.Generate) (canDelete bool, err error) {
	monthAgo := time.Now().AddDate(0, -conf.Conf.Task.NasClean.ChannelDelete.Persistence, 0)
	if time.Unix(g.CTime, 0).After(monthAgo) {
		log.Infoc(ctx, fmt.Sprintf("最近%d月内创建的渠道包-id:%d，不可以删除", conf.Conf.Task.NasClean.ChannelDelete.Persistence, g.ID))
		return false, nil
	}
	if g.PackState == int8(taskmdl.Deleted) {
		log.Infoc(ctx, fmt.Sprintf("渠道包-id:%d，已删除", g.ID))
		return false, nil
	}
	id, pPath, appKey := g.ID, g.GeneratePath, g.AppKey
	if len(pPath) == 0 {
		log.Infoc(ctx, "id[%d]，appKey[%s] filePath is empty", id, appKey)
		return false, nil
	}
	if !utils.FileExists(pPath) {
		log.Infoc(ctx, "id[%d]，appKey[%s] filePath: %s doesn't exist.", id, appKey, pPath)
		if _, err = t.fkDao.UpdateChannelPackState(ctx, []int64{id}, taskmdl.FileNotExist); err != nil {
			log.Errorc(ctx, "%v", err)
			return
		}
		return false, nil
	}
	return true, nil
}
func (t *ChannelDeleteTask) TaskName() string {
	return t.name
}

func NewChannelDeleteTask(fkDao *fawkes.Dao, name string) *ChannelDeleteTask {
	t := &ChannelDeleteTask{
		fkDao: fkDao,
		name:  name,
	}
	return t
}

func (t *ChannelDeleteTask) getAppKeys(ctx context.Context, excludeAppKey []string) (appKey []string, err error) {
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
