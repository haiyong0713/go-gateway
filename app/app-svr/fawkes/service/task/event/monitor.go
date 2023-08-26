package event

import (
	"context"
	"text/template"
	"time"

	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

type MonitorTask struct {
	conf  *conf.Config
	fkDao *fawkes.Dao
	name  string
}

func (t *MonitorTask) HandlerFunc(ctx context.Context) railgun.MsgPolicy {
	var (
		events        []*apm.Event
		monitorMap    map[string][]*apm.EventMonitor
		eventsMonitor []*apm.EventMonitor
		err           error
	)
	log.Warnc(ctx, "railgun apm event start")
	ctx = utils.CopyTrx(ctx)
	if events, err = t.fkDao.ApmEventList(ctx, "", "", "", "", "", "", "", "", "", "", 0, 0, 0, 0, 0, 0, 0); err != nil {
		log.Errorc(ctx, "ApmEventList %v", err)
		return railgun.MsgPolicyFailure
	}
	logDate := time.Now().AddDate(0, 0, -2).Format("20060102")
	if monitorMap, err = t.fkDao.ApmEventStorageList(ctx, logDate); err != nil {
		log.Errorc(ctx, "ApmEventStorageList %v", err)
		return railgun.MsgPolicyFailure
	}
	for _, event := range events {
		var (
			count    int64
			capacity int64
		)
		if monitor, ok := monitorMap[event.Name]; ok {
			for _, m := range monitor {
				count += m.StorageCount
				capacity += m.StorageCapacity
			}
			eventMonitor := &apm.EventMonitor{
				EventId:         event.ID,
				EventName:       event.Name,
				StorageCount:    count,
				StorageCapacity: capacity,
			}
			eventsMonitor = append(eventsMonitor, eventMonitor)
		}
	}
	if err = t.fkDao.ApmEventStorageUpdate(ctx, eventsMonitor); err != nil {
		log.Errorc(ctx, "ApmEventStorageUpdate %v", err)
		return railgun.MsgPolicyFailure
	}
	log.Warnc(ctx, "railgun apm event end")
	return railgun.MsgPolicyNormal
}

func (t *MonitorTask) BillionsQueryBodyGenerate(pn, ps int64, query string, queryRange *apm.BillionsQueryRange, querySort *apm.BillionsQuerySort, queryAggs []*apm.BillionsQueryAggs) (body string, err error) {
	if queryRange == nil {
		yesterday := time.Now().AddDate(0, 0, -1)
		startTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()).Unix() * apm.SecToMilliUnit
		endTime := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, yesterday.Location()).Unix() * apm.SecToMilliUnit
		queryRange = &apm.BillionsQueryRange{StartTime: startTime, EndTime: endTime}
	}
	funcMap := template.FuncMap{
		"maxIndex": func() int {
			return len(queryAggs) - 1
		},
	}
	billionsQueryBody := &apm.BillionsQueryBodyTemplate{From: pn, Size: ps, Sort: querySort, Query: query, RangeFiled: queryRange, Aggs: queryAggs}
	body, err = t.fkDao.TemplateAlterFunc(billionsQueryBody, funcMap, apm.BillionsTemplateQueryBody)
	return
}

func (t *MonitorTask) TaskName() string {
	return t.name
}

func NewMonitorTask(c *conf.Config, fkDao *fawkes.Dao, name string) *MonitorTask {
	r := &MonitorTask{
		conf:  c,
		fkDao: fkDao,
		name:  name,
	}
	return r
}
