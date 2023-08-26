package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// ApmEventAlertAdd 告警规则添加
func (s *Service) ApmEventAlertAdd(c context.Context, req *apm.EventAlertAddReq) (err error) {
	var (
		alertId             int64
		billionsAlertAddReq *apm.EventAlertBillionsReq
		alertList           []*apm.EventAlertDB
	)
	if alertList, err = s.fkDao.ApmEventAlertList(c, req.EventId, req.DatacenterAppId, req.Title, "", 0, 0, 0, 0); err != nil {
		log.Errorc(c, "ApmEventAlertList error %v", alertList)
		return
	}
	if len(alertList) > 0 {
		err = ecode.Error(ecode.RequestErr, "存在同名告警规则，请勿重复添加")
		return
	}
	if billionsAlertAddReq, err = s.convert2BillionsAlertAdd(c, req); err != nil {
		log.Errorc(c, "convert2BillionsAlertAdd error %v", err)
		return
	}
	// 日志平台添加告警
	if alertId, err = s.fkDao.BillionsAlertAdd(c, billionsAlertAddReq, s.c.BillionsAlert); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("日志平台添加失败,%v", err))
		log.Errorc(c, "BillionsAlertAdd error %v", err)
		return
	}
	// fawkes平台添加告警
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.ApmEventAlertAdd(tx, req.Title, req.Description, req.TimeField, req.AggField, req.FilterQuery, req.DenominatorFilterQuery,
			req.TriggerCondition, req.GroupField, req.NotifyFields, req.Channels, strings.Join(req.Targets, "|"), req.BotWebhook, req.Webhook, req.MutePeriod, req.Creator, req.Operator,
			req.EventId, alertId, req.Intervals, req.TimeFrame, req.NotifyDuration, apm.AlertInitVersion, req.MinLogCount, req.DatacenterAppId, req.Cluster, req.Level, req.AggType, req.MuteType, apm.EventAlertOn, req.AggPercentile, req.IsLogDetail); err != nil {
			log.Errorc(c, "ApmEventAlertAdd error %v", err)
			return err
		}
		return err
	})
	return
}

// ApmEventAlertUpdate 告警规则修改
func (s *Service) ApmEventAlertUpdate(c context.Context, req *apm.EventAlertUpdateReq) (err error) {
	var (
		alert                  *apm.EventAlertDB
		billionsAlertUpdateReq *apm.EventAlertBillionsReq
		alertList              []*apm.EventAlertDB
	)
	if alert, err = s.fkDao.ApmEventAlert(c, req.Id); err != nil {
		log.Errorc(c, "ApmEventAlert error %v", alertList)
		return
	}
	if alert == nil {
		return
	}
	if alertList, err = s.fkDao.ApmEventAlertList(c, alert.EventId, req.DatacenterAppId, req.Title, "", 0, 0, 0, 0); err != nil {
		log.Errorc(c, "ApmEventAlertList error %v", alertList)
		return
	}
	if len(alertList) == 1 && alertList[0].Id != req.Id {
		err = ecode.Error(ecode.RequestErr, "存在同名告警规则，请勿重复添加")
		return
	}
	if billionsAlertUpdateReq, err = s.convert2BillionsAlertUpdate(c, req); err != nil {
		log.Errorc(c, "convert2BillionsAlertUpdate error %v", err)
		return
	}
	// 日志平台更新告警
	if err = s.fkDao.BillionsAlertUpdate(c, alert.BillionId, billionsAlertUpdateReq, s.c.BillionsAlert); err != nil {
		log.Errorc(c, "BillionsAlertUpdate error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.ApmEventAlertUpdate(tx, req.Title, req.Description, req.TimeField, req.AggField, req.FilterQuery, req.DenominatorFilterQuery,
			req.TriggerCondition, req.GroupField, req.NotifyFields, req.Channels, strings.Join(req.Targets, "|"), req.BotWebhook, req.Webhook,
			req.MutePeriod, req.Operator, req.Intervals, req.TimeFrame, req.NotifyDuration, req.Version+1, req.MinLogCount, req.DatacenterAppId, req.Id, req.Cluster, req.Level, req.AggType, req.MuteType, req.IsLogDetail, req.AggPercentile); err != nil {
			log.Errorc(c, "ApmEventAlertUpdate error %v", err)
		}
		return err
	})
	return
}

// ApmEventAlertDel 告警规则删除
func (s *Service) ApmEventAlertDel(c context.Context, id int64) (err error) {
	if err = s.BillionsAlertRuleOpt(c, id, apm.AlertOptDelete); err != nil {
		log.Errorc(c, "BillionsAlertRuleOpt error %v", err)
		return
	}
	// 日志平台删除告警
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.ApmEventAlertDel(tx, id); err != nil {
			log.Errorc(c, "ApmEventAlertDel error %v", err)
		}
		return err
	})
	return
}

// ApmEventAlertSwitch 告警是否启用
func (s *Service) ApmEventAlertSwitch(c context.Context, isEnable int8, id int64) (err error) {
	var enabled string
	// 日志平台操作
	if isEnable == 1 {
		enabled = apm.AlertOptEnable
	} else {
		enabled = apm.AlertOptDisable
	}
	if err = s.BillionsAlertRuleOpt(c, id, enabled); err != nil {
		log.Errorc(c, "BillionsAlertRuleOpt error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.ApmEventAlertSwitch(tx, isEnable, id); err != nil {
			log.Errorc(c, "ApmEventAlertSwitch error %v", err)
		}
		return err
	})
	return
}

// ApmEventAlert 告警规则查询
func (s *Service) ApmEventAlert(c context.Context, id int64) (res *apm.EventAlertResp, err error) {
	var (
		event    *apm.Event
		queryRes *apm.EventAlertDB
	)
	if queryRes, err = s.fkDao.ApmEventAlert(c, id); err != nil {
		log.Errorc(c, "ApmEventAlertInfo error %v", err)
	}
	if queryRes == nil {
		return nil, err
	}
	if event, err = s.fkDao.ApmEvent(c, queryRes.EventId); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event == nil {
		return nil, err
	}
	res = queryRes.Convert2Resp()
	return
}

// ApmEventAlertList 告警规则列表
func (s *Service) ApmEventAlertList(c context.Context, req *apm.EventAlertQueryReq) (res *apm.EventAlertList, err error) {
	var (
		count    int
		queryRes []*apm.EventAlertDB
		resp     []*apm.EventAlertResp
	)
	res = &apm.EventAlertList{}
	if count, err = s.fkDao.ApmEventAlertCount(c, req.EventId, req.DatacenterAppId, req.Title, req.EventName, req.IsEnable, req.Level); err != nil {
		log.Errorc(c, "ApmEventAlertCount error %v", err)
		return
	}
	page := &apm.Page{Total: count, PageSize: req.Ps, PageNum: req.Pn}
	res.PageInfo = page
	if count < 1 {
		return
	}
	if queryRes, err = s.fkDao.ApmEventAlertList(c, req.EventId, req.DatacenterAppId, req.Title, req.EventName, req.IsEnable, req.Level, req.Pn, req.Ps); err != nil {
		log.Errorc(c, "ApmEventAlertList error %v", err)
		return
	}
	for _, query := range queryRes {
		resp = append(resp, query.Convert2Resp())
	}
	res.Items = resp
	return
}

// convert2BillionsAlertAdd 告警添加请求转化 fawkes->billions
func (s *Service) convert2BillionsAlertAdd(c context.Context, alertAdd *apm.EventAlertAddReq) (billionsAlertReq *apm.EventAlertBillionsReq, err error) {
	if billionsAlertReq, err = s.convert2BillionsReqOperation(c, alertAdd.Title, alertAdd.Description, alertAdd.TimeField, alertAdd.AggField, alertAdd.FilterQuery, alertAdd.DenominatorFilterQuery,
		alertAdd.TriggerCondition, alertAdd.GroupField, alertAdd.NotifyFields, alertAdd.Channels, alertAdd.BotWebhook, alertAdd.Webhook, alertAdd.MutePeriod,
		alertAdd.Targets, alertAdd.EventId, alertAdd.Intervals, alertAdd.TimeFrame, alertAdd.NotifyDuration, apm.AlertInitVersion, alertAdd.MinLogCount, alertAdd.DatacenterAppId, alertAdd.Cluster, alertAdd.Level, alertAdd.AggType, alertAdd.MuteType, alertAdd.AggPercentile, alertAdd.IsLogDetail); err != nil {
		log.Errorc(c, "convert2BillionsReqOperation error %v", err)
	}
	return
}

// convert2BillionsAlertUpdate 告警修改请求转化 fawkes->billions
func (s *Service) convert2BillionsAlertUpdate(c context.Context, alertUpdate *apm.EventAlertUpdateReq) (billionsAlertReq *apm.EventAlertBillionsReq, err error) {
	var (
		eventAlert *apm.EventAlertDB
	)
	if eventAlert, err = s.fkDao.ApmEventAlert(c, alertUpdate.Id); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if eventAlert == nil {
		return
	}
	if billionsAlertReq, err = s.convert2BillionsReqOperation(c, alertUpdate.Title, alertUpdate.Description, alertUpdate.TimeField, alertUpdate.AggField, alertUpdate.FilterQuery, alertUpdate.DenominatorFilterQuery,
		alertUpdate.TriggerCondition, alertUpdate.GroupField, alertUpdate.NotifyFields, alertUpdate.Channels, alertUpdate.BotWebhook, alertUpdate.Webhook, alertUpdate.MutePeriod,
		alertUpdate.Targets, eventAlert.EventId, alertUpdate.Intervals, alertUpdate.TimeFrame, alertUpdate.NotifyDuration, eventAlert.Version, alertUpdate.MinLogCount, alertUpdate.DatacenterAppId, alertUpdate.Cluster, alertUpdate.Level, alertUpdate.AggType, alertUpdate.MuteType, alertUpdate.AggPercentile, alertUpdate.IsLogDetail); err != nil {
		log.Errorc(c, "convert2BillionsReqOperation error %v", err)
	}
	return
}

var (
	billionsCluster  = []string{"main", "mobile", "intl"}
	billionsMuteType = []string{"no_mute", "custom"}
	billionsSeverity = []string{"record", "mention", "major", "disaster"}
	billionsAggType  = []string{"count", "percent_count", "average", "min", "max", "distinct_count", "sum", "percentiles"}
	billionsTrigOpt  = map[string]int{ // 操作符：规则拼接长度
		"gte":           2,
		"lt":            2,
		"in_range":      3,
		"not_in_range":  3,
		"relative_up":   2,
		"relative_down": 2,
	}
)

func (s *Service) convert2BillionsReqOperation(c context.Context, title, description, timeField, aggField, filterQuery, denominatorFilterQuery, triggerCondition, groupField, notifyFields, channels, botWebhook, webhook, mutePeriod string, targets []string, eventId, intervals, timeFrame, notifyDur, version, minLogCount, datacenterAppId int64, cluster, level, aggType, muteType, percentile, isLogDetail int8) (billionsAlertReq *apm.EventAlertBillionsReq, err error) {
	var event *apm.Event
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event == nil {
		err = ecode.Error(ecode.RequestErr, "技术埋点为空")
		return
	}
	schedule := &apm.BillionsAlertSchedule{Interval: fmt.Sprintf("%vm", intervals)}
	search := &apm.BillionsAlertSearch{
		SearchTimeframe:    fmt.Sprintf("%vm", timeFrame),
		AggregationType:    billionsAggType[aggType-1],
		FieldToAggregateOn: aggField,
		Percentile:         percentile,
		Query:              &apm.BillionsAlertQuery{Query: filterQuery},
	}
	if denominatorFilterQuery != "" {
		search.DenominatorQuery = &apm.BillionsAlertQuery{Query: denominatorFilterQuery}
	}
	if groupField != "" {
		search.GroupBy = []string{groupField}
	}
	triggerArr := strings.Split(triggerCondition, ":")
	triggerLen := len(triggerArr)
	if triggerLen < 1 {
		err = ecode.Error(ecode.RequestErr, "trigger 格式不符合规范")
		return
	}
	operator := triggerArr[0]
	trigger := &apm.BillionsAlertTrigger{
		Operator: operator,
	}
	var (
		trigLen int
		ok      bool
	)
	if trigLen, ok = billionsTrigOpt[operator]; !ok {
		err = ecode.Error(ecode.RequestErr, "trigger 操作符不存在")
		return
	}
	if trigLen != triggerLen {
		err = ecode.Error(ecode.RequestErr, "trigger 格式不符合规范")
		return
	}
	switch operator {
	// trigger格式 操作符:值1
	case apm.AlertTrigOptGte, apm.AlertTrigOptLt, apm.AlertTrigOptUp, apm.AlertTrigOptDown:
		value, err1 := strconv.ParseInt(triggerArr[1], 10, 64)
		if err != nil {
			log.Errorc(c, "strconv.ParseInt error %v", err1)
		}
		trigger.Value = value
	// trigger格式 操作符:值1:值2
	case apm.AlertTrigOptInRange, apm.AlertTrigOptNotRange:
		rangeStart, err1 := strconv.ParseInt(triggerArr[1], 10, 64)
		if err1 != nil {
			log.Errorc(c, "strconv.ParseInt error %v", err1)
			return nil, err1
		}
		trigger.RangeStart = rangeStart
		rangeEnd, err1 := strconv.ParseInt(triggerArr[2], 10, 64)
		if err1 != nil {
			log.Errorc(c, "strconv.ParseInt error %v", err1)
			return nil, err1
		}
		trigger.RangeEnd = rangeEnd
	}
	trigger.AtLeastLogCount = minLogCount
	var alertTargets []*apm.BillionsAlertTarget
	for _, target := range targets {
		alertTarget := &apm.BillionsAlertTarget{}
		// target格式 type-vale
		targetArr := strings.Split(target, "-")
		// nolint:gomnd
		if len(targetArr) < 2 {
			err = ecode.Error(ecode.RequestErr, "target 格式不符合规范")
			return
		}
		alertTarget.Type = targetArr[0]
		alertTarget.Value = targetArr[1]
		alertTargets = append(alertTargets, alertTarget)
	}
	mute := &apm.BillionsAlertMute{Type: billionsMuteType[muteType-1]}
	if muteType != 1 {
		period := &apm.BillionsAlertMutePeriod{}
		if err = json.Unmarshal([]byte(mutePeriod), period); err != nil {
			log.Errorc(c, "json.Unmarshal error %v", err)
			return
		}
		mute.Period = period
	} else {
		mute.Period = nil
	}
	alert := &apm.BillionsAlert{
		SuppressAlertDuration: fmt.Sprintf("%vm", notifyDur),
		AlertChannels:         strings.Split(channels, ","),
		Targets:               alertTargets,
		WechatRobotWebhookUrl: botWebhook,
		WebhookUrl:            webhook,
		MutePeriod:            mute,
	}
	if isLogDetail == 0 {
		alert.CheckedIncludeLog = false
	} else {
		alert.CheckedIncludeLog = true
		if notifyFields == "" {
			alert.ShouldIncludeAllFields = true
			alert.IncludeLogFields = []string{}
		} else {
			alert.IncludeLogFields = strings.Split(notifyFields, ",")
		}
	}
	action := &apm.BillionsAlertAction{Alert: alert}
	return &apm.EventAlertBillionsReq{
		Title:         fmt.Sprintf("【Fawkes】%s(app_id=%v)", title, datacenterAppId),
		Description:   description,
		Schedule:      schedule,
		TimeField:     timeField,
		AppId:         event.Name,
		EsCluster:     billionsCluster[cluster-1],
		Severity:      billionsSeverity[level-1],
		Search:        search,
		Trigger:       trigger,
		Action:        action,
		SchemaVersion: version,
	}, err
}

func (s *Service) BillionsAlertRuleOpt(c context.Context, id int64, operation string) (err error) {
	var (
		alertDB  *apm.EventAlertDB
		alertOpt = &apm.BillionsAlertOpt{}
	)
	if alertDB, err = s.fkDao.ApmEventAlert(c, id); err != nil {
		log.Errorc(c, "ApmEventAlertInfo err %v", err)
		return
	}
	if alertDB == nil {
		return
	}
	alertOpt.RuleId = append(alertOpt.RuleId, strconv.FormatInt(alertDB.BillionId, 10))
	alertOpt.Operation = operation
	if err = s.fkDao.BillionsAlertRuleOpt(c, alertOpt, s.c.BillionsAlert); err != nil {
		log.Errorc(c, "BillionsAlertRuleOpt err %v", err)
	}
	return
}
