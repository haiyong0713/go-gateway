package pack

import (
	"context"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"time"
)

type VedaUpdateTask struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func (t *VedaUpdateTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var err error
	log.Infoc(ctx, "开始更新堆栈")
	// 获取配置
	VedaUpdateConfig := t.conf.Task.VedaUpdate
	// appKey
	nowTime := time.Now()
	persistence := VedaUpdateConfig.Persistence
	endTime := nowTime.AddDate(0, -persistence, -1)

	appKeys := VedaUpdateConfig.Apps
	if len(appKeys) == 0 {
		log.Infoc(ctx, "没有配置app_keys")
		return railgun.MsgPolicyFailure
	}
	// 每个应用取6个月之前的数据刷未完成的100条
	log.Infoc(ctx, "本次更新堆栈信息的appKey:%+v", appKeys)
	for _, appKey := range appKeys {
		log.Infoc(ctx, "appKey:%v 开始执行", appKey)
		// 获取x月前最新版本
		var (
			maxVersionCode string
			hashList       []string
		)
		if maxVersionCode, err = t.fkDao.GetMaxVersionCodeByTime(ctx, appKey, endTime, VedaUpdateConfig.Count); err != nil {
			log.Infoc(ctx, "获取版本错误")
			return railgun.MsgPolicyFailure
		}
		if hashList, err = t.fkDao.HashListLimitVersionCode(ctx, "veda_crash_index_v2", "error_stack_hash_without_useless", appKey, maxVersionCode, VedaUpdateConfig.Count); err != nil {
			log.Infoc(ctx, "获取hash列表错误")
			return railgun.MsgPolicyFailure
		}
		if len(hashList) > 0 {
			for _, hash := range hashList {
				err = t.fkDao.VedaTransact(ctx, func(tx *sql.Tx) (txError error) {
					if _, txError = t.fkDao.TxIndexUpdate(tx, hash, appKey, "Fawkes小姐姐", "Fawkes小姐姐", "最近六个月版本未发生崩溃，自动标注解决", "veda_crash_index_v2", "error_stack_hash_without_useless", 0, 1); err != nil {
						log.Error("%v", err)
						return
					}
					return
				})
			}
		}
		if err != nil {
			log.Error("TxIndexUpdate %v", err)
			return railgun.MsgPolicyFailure
		}
		log.Infoc(ctx, "appKey:%s 执行结束 hash: %v", appKey, hashList)
	}
	return railgun.MsgPolicyNormal
}

func (t *VedaUpdateTask) TaskName() string {
	return t.name
}

func NewVedaUpdateTask(c *conf.Config, fkDao *fawkes.Dao, name string) *VedaUpdateTask {
	t := &VedaUpdateTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return t
}
