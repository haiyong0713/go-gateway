package event

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	"go-gateway/app/app-svr/fawkes/service/model/app"
	apmSvr "go-gateway/app/app-svr/fawkes/service/service/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_operator = "system"
)

type CompletionTask struct {
	conf  *conf.Config
	svr   *apmSvr.Service
	fkDao *fawkes.Dao
	name  string
}

func (t *CompletionTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		eventsCompletion []*apm.EventCompletion
		err              error
	)
	logDate := time.Now().AddDate(0, 0, -1).Format("20060102")
	if eventsCompletion, err = t.fkDao.ApmEventCompletionList(ctx, logDate); err != nil {
		log.Errorc(ctx, "ApmEventCompletionList error %v", err)
		return railgun.MsgPolicyFailure
	}
	for _, evc := range eventsCompletion {
		var (
			event *apm.Event
			bus   *apm.Bus
		)
		if event, err = t.fkDao.ApmEventByName(ctx, evc.DatacenterEventName); err != nil {
			log.Errorc(ctx, "ApmEventByName error %v", err)
			return railgun.MsgPolicyAttempts
		}
		if event == nil {
			log.Warnc(ctx, "fawkes event is null %v", evc.DatacenterEventName)
			continue
		}
		if event.LogID != apm.LogIdTrackT {
			continue
		}
		if bus, err = t.fkDao.ApmBusByID(ctx, event.BusID); err != nil {
			log.Errorc(ctx, "ApmBusByID error %v", err)
			return railgun.MsgPolicyAttempts
		}
		if bus == nil {
			log.Warnc(ctx, "business is null")
			continue
		}
		if err = t.EventCompletion(ctx, event, evc.DatacenterAppId, bus.DatacenterBusKey); err != nil {
			log.Errorc(ctx, "EventCompletion error %v", err)
			return railgun.MsgPolicyAttempts
		}
		if err = t.MonitorNotifyConfigCompletion(ctx, event); err != nil {
			log.Errorc(ctx, "MonitorNotifyConfigCompletion error %v", err)
			return railgun.MsgPolicyAttempts
		}
	}
	return railgun.MsgPolicyNormal
}

// EventCompletion 技术埋点数据补全
func (t *CompletionTask) EventCompletion(ctx context.Context, event *apm.Event, datacenterAppId int64, datacenterBusKey string) (err error) {
	var dcEventId int64
	if dcEventId, err = t.svr.ApmDataCenterEventAdd(ctx, event.LogID, event.Name, event.Description, strconv.FormatInt(datacenterAppId, 10), datacenterBusKey,
		event.Topic, strconv.Itoa(int(event.Activity)), event.DatacenterDwdTableName); err != nil {
		if !strings.Contains(err.Error(), "数据重复") {
			return
		}
		var rel []*apm.EventDatacenterRel
		if rel, err = t.fkDao.ApmAppEventRelList(ctx, event.ID, datacenterAppId); err != nil {
			log.Errorc(ctx, "ApmAppEventRelList error %v", err)
			return
		}
		if len(rel) == 1 {
			log.Warnc(ctx, "已存在关联，[event_name]:%v,[datacenter_app_id]:%v", event.Name, datacenterAppId)
		} else {
			log.Warnc(ctx, "不存在关联，[event_name]:%v,[datacenter_app_id]:%v", event.Name, datacenterAppId)
		}
		return
	}
	if dcEventId == 0 {
		log.Errorc(ctx, "datacenter event id is nil")
		return
	}
	if err = t.fkDao.Transact(ctx, func(tx *sql.Tx) error {
		if err = t.fkDao.TxApmAppEventRelAdd(tx, event.ID, datacenterAppId, dcEventId, _operator); err != nil {
			log.Errorc(ctx, "TxApmAppEventRelAdd error %v", err)
		}
		return err
	}); err != nil {
		return
	}
	if err = t.svr.ApmDatacenterEventUpdate(ctx, event.ID, dcEventId, event.LogID, event.Name, event.Description, strconv.FormatInt(datacenterAppId, 10), datacenterBusKey,
		event.Topic, strconv.Itoa(int(event.Activity)), event.DatacenterDwdTableName); err != nil {
		return
	}
	return
}

// MonitorNotifyConfigCompletion 监测通知配置补全
func (t *CompletionTask) MonitorNotifyConfigCompletion(ctx context.Context, event *apm.Event) (err error) {
	var (
		apps    []*app.APP
		configs []*apm.EventMonitorNotifyConfig
	)
	if apps, err = t.fkDao.AppListByDatacenterAppId(ctx, event.DatacenterAppID); err != nil {
		log.Errorc(ctx, "AppListByDatacenterAppId error %v", err)
		return
	}
	for _, app := range apps {
		config := &apm.EventMonitorNotifyConfig{
			EventId:  event.ID,
			AppKey:   app.AppKey,
			IsNotify: apm.EventMonitorNotifyOn,
			IsMute:   apm.EventMonitorNotifyMuteOff,
			Operator: _operator,
		}
		configs = append(configs, config)
	}
	err = t.fkDao.Transact(ctx, func(tx *sql.Tx) error {
		if err = t.fkDao.TxApmEventMonitorNotifyConfigBatchSet(tx, configs); err != nil {
			log.Errorc(ctx, "batch set error %v", err)
		}
		return err
	})
	return
}

func (t *CompletionTask) TaskName() string {
	return t.name
}

func NewCompletionTask(c *conf.Config, fkDao *fawkes.Dao, name string) *CompletionTask {
	r := &CompletionTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
		svr:   apmSvr.New(c),
	}
	return r
}
