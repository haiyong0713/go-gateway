package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) ApmAlertRuleList(c context.Context, req *apm.AlertRuleListReq) (res *apm.AlertRuleRes, err error) {
	var majorRuleId int64
	// 子规则 -> 父规则
	if majorRuleId, err = s.getMajorRuleIdByAdjustId(c, req.HawkeyeId); err != nil {
		log.Errorc(c, "getMajorRuleIdByAdjustId error %v", err)
		return
	}
	// 参数为父规则id
	if majorRuleId == 0 {
		majorRuleId = req.HawkeyeId
	}
	var (
		count   int
		ruleRes []*apm.AlertRule
	)
	if count, err = s.fkDao.ApmAlertRuleCount(c, majorRuleId, req.Name, req.Species, req.MetricName, apm.AlertRuleTypeMajor); err != nil {
		log.Errorc(c, "ApmAlertRuleCount error %v", err)
		return
	}
	if count < 1 {
		return
	}
	var majorRules []*apm.AlertRule
	if majorRules, err = s.fkDao.ApmAlertRuleList(c, majorRuleId, req.Name, req.Species, req.MetricName, apm.AlertRuleTypeMajor, req.Pn, req.Ps); err != nil {
		log.Errorc(c, "ApmAlertRuleList error %v", err)
		return
	}
	var (
		majorRuleIds  []int64
		adjustRuleIds []int64
		majorRuleMap  = make(map[int64]*apm.AlertRule)
		adjustRuleMap = make(map[int64]*apm.AlertRule)
		rels          []*apm.AlertRuleRel
	)
	for _, majorRule := range majorRules {
		majorRuleMap[majorRule.HawkeyeId] = majorRule
		majorRuleIds = append(majorRuleIds, majorRule.HawkeyeId)
	}
	if rels, err = s.fkDao.ApmAlertRuleRelByRuleIds(c, majorRuleIds); err != nil {
		log.Errorc(c, "ApmAlertRuleRelList error %v", err)
		return
	}
	for _, rel := range rels {
		adjustRuleIds = append(adjustRuleIds, rel.AdjustRuleId)
	}
	if adjustRuleMap, err = s.fkDao.ApmAlertRuleByHawkeyeIds(c, adjustRuleIds); err != nil {
		log.Errorc(c, "ApmAlertRuleByHawkeyeIds error %v", err)
		return
	}
	for _, rel := range rels {
		majorRule, ok := majorRuleMap[rel.RuleId]
		if ok {
			majorRule.AdjustRule = append(majorRule.AdjustRule, adjustRuleMap[rel.AdjustRuleId])
		}
	}
	for _, majorRule := range majorRuleMap {
		ruleRes = append(ruleRes, majorRule)
	}
	page := &apm.Page{
		PageNum:  req.Pn,
		PageSize: req.Ps,
		Total:    count,
	}
	res = &apm.AlertRuleRes{
		Items:    ruleRes,
		PageInfo: page,
	}
	return
}

func (s *Service) ApmAlertRuleSet(c context.Context, req *apm.AlertRuleSetReq) (err error) {
	var ruleId = req.HawkeyeId
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if req.RuleType == apm.AlertRuleTypeAdjust {
			if err = s.fkDao.TxApmAlertRuleRelAdd(tx, req.HawkeyeId, req.HawkeyeAdjustId, req.Operator); err != nil {
				log.Errorc(c, "TxApmAlertRuleRelAdd error %v", err)
				return err
			}
			ruleId = req.HawkeyeAdjustId
		}
		if err = s.fkDao.TxApmAlertRuleAdd(tx, ruleId, req.Name, req.TriggerCondition, req.Species, req.QueryExprs, req.Operator, req.RuleType); err != nil {
			log.Errorc(c, "TxApmAlertRuleAdd error %v", err)
			return err
		}
		return err
	})
	return
}

func (s *Service) ApmAlertRuleMDUpdate(c context.Context, id int64, markdown string) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertRuleMDUpdate(tx, id, markdown); err != nil {
			log.Errorc(c, "TxApmAlertRuleMDUpdate error %v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmAlertRuleDel(c context.Context, id int64) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertRuleDel(tx, id); err != nil {
			log.Errorc(c, "TxApmAlertRuleDel error %v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmAlertList(c context.Context, req *apm.AlertListReq) (res *apm.AlertRes, err error) {
	var (
		ruleIds   []int64
		adjustIds []int64
	)
	// 父规则 -> 子规则
	if adjustIds, err = s.getMajorRuleRelIds(c, req.RuleId); err != nil {
		log.Errorc(c, "ApmAlertCount error %v", err)
		return
	}
	if req.RuleId > 0 {
		ruleIds = append(append(ruleIds, req.RuleId), adjustIds...)
	}
	var (
		count     int
		alertsRes []*apm.Alert
	)
	if count, err = s.fkDao.ApmAlertCount(c, req.AppKey, req.Env, ruleIds, req.Type, req.Status, req.AlertMd5, req.StartTime, req.EndTime); err != nil {
		log.Errorc(c, "ApmAlertCount error %v", err)
		return
	}
	if count < 1 {
		return
	}
	var alerts []*apm.Alert
	if alerts, err = s.fkDao.ApmAlertList(c, req.AppKey, req.Env, ruleIds, req.Type, req.Status, req.AlertMd5, req.StartTime, req.EndTime, req.Pn, req.Ps); err != nil {
		log.Errorc(c, "ApmAlertList error %v", err)
		return
	}
	var (
		set       = make(map[int64]struct{})
		ruleIdSet []int64
		alertMap  = make(map[*apm.Alert]int64)
	)
	for _, alert := range alerts {
		if _, ok := set[alert.RuleId]; !ok {
			set[alert.RuleId] = struct{}{}
			// 防止 查询条件中没有rule_id
			ruleIdSet = append(ruleIdSet, alert.RuleId)
		}
		alertMap[alert] = alert.RuleId
	}
	var ruleMap = make(map[int64]*apm.AlertRule)
	if ruleMap, err = s.fkDao.ApmAlertRuleByHawkeyeIds(c, ruleIdSet); err != nil {
		log.Errorc(c, "ApmAlertRuleByHawkeyeIds error %v", err)
		return
	}
	for alert, ruleId := range alertMap {
		if rule, ok := ruleMap[ruleId]; ok {
			alert.Rule = rule
		}
		alertsRes = append(alertsRes, alert)
	}
	page := &apm.Page{
		PageNum:  req.Pn,
		PageSize: req.Ps,
		Total:    count,
	}
	res = &apm.AlertRes{
		Items:    alerts,
		PageInfo: page,
	}
	return
}

func (s *Service) ApmAlertAdd(c context.Context, req *apm.AlertAddReq) (err error) {
	labels := strings.ReplaceAll(req.Labels, "'", "\"")
	var alertLabels map[string]string
	if err = json.Unmarshal([]byte(labels), &alertLabels); err != nil {
		log.Errorc(c, "json.Unmarshal error %v", err)
	}
	appKey := alertLabels["app_key"]
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertAdd(tx, req.RuleId, appKey, req.Env, req.Duration, req.AlertMd5, req.Labels, req.Operator, apm.AlertTypeUncategorized, req.Status, req.TriggerValue, req.StartTime); err != nil {
			log.Errorc(c, "ApmAlertAdd error %v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmAlertUpdate(c context.Context, req *apm.AlertUpdateReq) (err error) {
	var (
		duration int64
		alert    *apm.Alert
	)
	if alert, err = s.fkDao.ApmAlertById(c, req.Id); err != nil {
		log.Errorc(c, "ApmAlertById error %v", err)
		return
	}
	if alert == nil {
		log.Errorc(c, "alert is nil")
		return
	}
	duration = alert.Duration
	// 已解决状态 -> 未解决状态后，再改为已解决状态，持续时间累加
	if req.Status == apm.AlertStatusResolved && alert.Status == apm.AlertStatusUnresolved {
		duration += time.Now().Unix() - alert.CTime.Unix()
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertUpdate(tx, req.Type, req.Status, req.Description, req.Operator, duration, req.Id); err != nil {
			log.Errorc(c, "ApmAlertUpdate error %v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmAlertIndicatorInfo(c context.Context, req *apm.AlertIndicatorReq) (res *apm.AlertIndicator, err error) {
	var (
		ruleIds   []int64
		adjustIds []int64
	)
	// 父规则 -> 子规则
	if adjustIds, err = s.getMajorRuleRelIds(c, req.RuleId); err != nil {
		log.Errorc(c, "ApmAlertCount error %v", err)
		return
	}
	if req.RuleId > 0 {
		ruleIds = append(append(ruleIds, req.RuleId), adjustIds...)
	}
	var (
		count  int
		alerts []*apm.Alert
	)
	if count, err = s.fkDao.ApmAlertCount(c, req.AppKey, req.Env, ruleIds, req.Type, req.Status, "", req.StartTime, req.EndTime); err != nil {
		log.Errorc(c, "ApmAlertCount error %v", err)
		return
	}
	if count < 1 {
		return
	}
	if alerts, err = s.fkDao.ApmAlertList(c, req.AppKey, req.Env, ruleIds, req.Type, req.Status, "", req.StartTime, req.EndTime, 0, 0); err != nil {
		log.Errorc(c, "ApmAlertList error %v", err)
		return
	}
	var (
		falseAlert, errorAlert, uncategorizedAlert                 int64
		falseAlertRate, errorAlertRate, uncategorizedRate, quality float64
	)
	for _, alert := range alerts {
		if alert.Type == apm.AlertTypeFalse {
			falseAlert++
		}
		if alert.Type == apm.AlertTypeError {
			errorAlert++
		}
		if alert.Type == apm.AlertTypeUncategorized {
			uncategorizedAlert++
		}
	}
	if falseAlertRate, err = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(falseAlert)/float64(count)), 64); err != nil {
		log.Errorc(c, "falseAlertRate error %v", err)
		return
	}
	if errorAlertRate, err = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(errorAlert)/float64(count)), 64); err != nil {
		log.Errorc(c, "errorAlertRate error %v", err)
		return
	}
	if uncategorizedRate, err = strconv.ParseFloat(fmt.Sprintf("%.4f", float64(uncategorizedAlert)/float64(count)), 64); err != nil {
		log.Errorc(c, "uncategorizedRate error %v", err)
		return
	}
	if quality, err = strconv.ParseFloat(fmt.Sprintf("%.4f", 1-falseAlertRate-errorAlertRate-uncategorizedRate), 64); err != nil {
		log.Errorc(c, "errorAlertRate error %v", err)
		return
	}
	res = &apm.AlertIndicator{
		FalseAlertRate:         falseAlertRate,
		ErrorAlertRate:         errorAlertRate,
		UncategorizedAlertRate: uncategorizedRate,
		Quality:                quality,
	}
	return
}

func (s *Service) getMajorRuleRelIds(c context.Context, ruleId int64) (adjustIds []int64, err error) {
	if ruleId < 1 {
		return
	}
	var (
		ruleIds = []int64{ruleId}
		rels    []*apm.AlertRuleRel
	)
	// 查询 是否是 主规则
	if rels, err = s.fkDao.ApmAlertRuleRelByRuleIds(c, ruleIds); err != nil {
		log.Errorc(c, "ApmAlertAdjustRuleRel error %v", err)
	}
	if len(rels) < 1 {
		return
	}
	for _, rel := range rels {
		adjustIds = append(adjustIds, rel.AdjustRuleId)
	}
	return
}

func (s *Service) getMajorRuleIdByAdjustId(c context.Context, adjustId int64) (ruleId int64, err error) {
	if adjustId < 1 {
		return
	}
	var rel *apm.AlertRuleRel
	if rel, err = s.fkDao.ApmAlertRuleRelByAdjustId(c, adjustId); err != nil {
		log.Errorc(c, "ApmAlertRuleRelByAdjustId error %v", err)
	}
	if rel == nil {
		return
	}
	ruleId = rel.RuleId
	return
}

func (s *Service) ApmAlertReason(c context.Context, alertMd5 string) (content string, err error) {
	var alert *apm.Alert
	if alert, err = s.fkDao.ApmAlertByMd5(c, alertMd5); err != nil {
		log.Errorc(c, "ApmAlertById error %v", err)
		return
	}
	if alert == nil {
		log.Errorc(c, "alert is nil")
		return
	}
	fileName := fmt.Sprintf("%s.json", alert.AlertMd5)
	alertUrl := fmt.Sprintf("%s/%s", s.c.LocalPath.LocalDomain, filepath.Join("mobile-ep", "slime-alert", alert.StartTime.Format("20060102"), alert.AppKey, strconv.FormatInt(alert.RuleId, 10), fileName))
	if content, err = s.fkDao.MacrossFileInfo(c, alertUrl); err != nil {
		log.Errorc(c, "MacrossFileInfo error %v", err)
	}
	return
}

func (s *Service) ApmAlertReasonConfig(c context.Context, req *apm.AlertReasonConfigReq) (res *apm.AlertReasonConfigResp, err error) {
	var configs []*apm.AlertReasonConfig
	if configs, err = s.fkDao.ApmAlertReasonConfig(c, req.RuleId); err != nil {
		log.Errorc(c, "ApmAlertReasonConfig error %v", err)
		return
	}
	if res, err = s.convert2AlertReasonConfigResp(c, configs); err != nil {
		log.Errorc(c, "alert reason config convert2Resp error %v", err)
	}
	return
}

func (s *Service) ApmAlertReasonConfigAdd(c context.Context, req *apm.AlertReasonConfigAddReq) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertReasonConfigAdd(tx, req.RuleId, req.EventId, req.QueryType, req.CustomQuerySql, req.QueryCondition, req.ImpactFactorFields, req.Description, req.Operator); err != nil {
			log.Errorc(c, "TxApmAlertReasonConfigAdd error %v", err)
			return err
		}
		return nil
	})
	return
}

func (s *Service) ApmAlertReasonConfigUpdate(c context.Context, req *apm.AlertReasonConfigUpdateReq) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertReasonConfigUpdate(tx, req.Id, req.EventId, req.QueryType, req.CustomQuerySql, req.QueryCondition, req.ImpactFactorFields, req.Description, req.Operator); err != nil {
			log.Errorc(c, "TxApmAlertReasonConfigUpdate error %v", err)
			return err
		}
		return nil
	})
	return
}

func (s *Service) ApmAlertReasonConfigDelete(c context.Context, req *apm.AlertReasonConfigDeleteReq) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if err = s.fkDao.TxApmAlertReasonConfigDelete(tx, req.Id); err != nil {
			log.Errorc(c, "TxApmAlertReasonConfigDelete error %v", err)
			return err
		}
		return nil
	})
	return
}

func (s *Service) convert2AlertReasonConfigResp(c context.Context, rows []*apm.AlertReasonConfig) (resp *apm.AlertReasonConfigResp, err error) {
	var (
		eventIds []int64
		events   map[int64]*apm.Event
	)
	for _, row := range rows {
		eventIds = append(eventIds, row.EventId)
	}
	if events, err = s.fkDao.ApmEventByIds(c, eventIds); err != nil {
		log.Errorc(c, "ApmEventByIds error %v", err)
		return
	}
	var items []*apm.AlertReasonConfigItem
	for _, row := range rows {
		item := &apm.AlertReasonConfigItem{
			Id:             row.Id,
			RuleId:         row.RuleId,
			EventId:        row.EventId,
			QueryType:      row.QueryType,
			QuerySql:       row.QuerySql,
			QueryCondition: row.QueryCondition,
			Description:    row.Description,
			Operator:       row.Operator,
			CTime:          row.CTime,
			MTime:          row.MTime,
		}
		var fields []*apm.AlertReasonField
		if row.ImpactFactorFields != "" {
			if err = json.Unmarshal([]byte(row.ImpactFactorFields), &fields); err != nil {
				return
			}
		}
		item.ImpactFactorFields = fields
		if event, ok := events[row.EventId]; ok {
			item.EventName = event.Name
			item.Databases = event.Databases
			item.DistributedTableName = event.DistributedTableName
		}
		items = append(items, item)
	}
	resp = &apm.AlertReasonConfigResp{
		Items: items,
	}
	return
}
