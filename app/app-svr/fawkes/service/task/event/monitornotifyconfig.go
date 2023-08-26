package event

import (
	"context"
	"time"

	"go-common/library/database/sql"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

type MonitorNotifyConfigTask struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func (t *MonitorNotifyConfigTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		configs []*apm.EventMonitorNotifyConfig
		err     error
	)
	if configs, err = t.fkDao.ApmEventMonitorNotifyConfigList(ctx, 0, "", 1, 1, 0, 0); err != nil {
		log.Errorc(ctx, "notify config list query error %v", err)
		return railgun.MsgPolicyFailure
	}
	var (
		targetIds  []int64
		now        = time.Now()
		targetTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	)
	for _, config := range configs {
		if targetTime > config.MuteEndTime.Unix() {
			targetIds = append(targetIds, config.Id)
		}
	}
	if err = t.fkDao.Transact(ctx, func(tx *sql.Tx) error {
		if err = t.fkDao.TxApmEventMonitorNotifyMuteUpdate(tx, targetIds); err != nil {
			log.Errorc(ctx, "mute state update error %v", err)
			return err
		}
		return err
	}); err != nil {
		return railgun.MsgPolicyFailure
	}
	return railgun.MsgPolicyNormal
}

func (t *MonitorNotifyConfigTask) TaskName() string {
	return t.name
}

func NewMonitorNotifyConfigTask(c *conf.Config, fkDao *fawkes.Dao, name string) *MonitorNotifyConfigTask {
	r := &MonitorNotifyConfigTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return r
}
