package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model/apm"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_ckDatasourceType = "Clickhouse"
	_ckDatasourceName = "Clickhouse_datacenter_olap_ck_mobile_infra_replica" // "uat:Clickhouse_test_cluster"
	_ckTTLUnit        = "DAY"
	_ckPubLevel       = 1   // 数据部分公开
	_ckDataLevel      = "C" // "业务线一般数据"
	_ckEngine         = "ReplicatedMergeTree"
	_ckModelLevel     = 2709 // 2709 数据应用层，表名前加ads_,uat:474
	_ckBusTags        = "平台"
)

func (s *Service) ApmEvent(c context.Context, eventID int64) (res *apm.Event, err error) {
	var (
		eventFields  []*apm.EventField
		commonFields []*apm.EventField
	)
	if res, err = s.fkDao.ApmEvent(c, eventID); err != nil {
		return
	}
	if res == nil {
		return
	}
	if commonFields, err = s.fkDao.ApmEventFieldList(c, apm.EventFieldCommonType); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	if eventFields, err = s.fkDao.ApmEventFieldList(c, eventID); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	res.EventFields = eventFields
	res.CommonEventFields = commonFields
	return
}

func (s *Service) ApmEventList(c context.Context, name, appKey, logID, busName, topic, tableName, orderBy, dwdTableName string, ps, pn int, busId, appId int64, activity, dtCondition, state int8) (res *apm.ResultEventList, err error) {
	var (
		total       int
		eventList   []*apm.Event
		fieldModify map[int64]int64
	)
	res = &apm.ResultEventList{}
	if total, err = s.fkDao.ApmEventListCount(c, name, appKey, logID, busName, topic, "", tableName, "", dwdTableName, busId, appId, activity, dtCondition, state); err != nil {
		log.Errorc(c, "ApmEventListCount %v", err)
		return
	}
	pageInfo := &apm.Page{Total: total, PageNum: pn, PageSize: ps}
	res.PageInfo = pageInfo
	if total < 1 {
		return
	}
	if eventList, err = s.fkDao.ApmEventList(c, name, appKey, logID, busName, topic, "", tableName, "", orderBy, dwdTableName, ps, pn, busId, appId, activity, dtCondition, state); err != nil {
		log.Errorc(c, "ApmEventList %v", err)
		return
	}
	if fieldModify, err = s.fkDao.ApmEventFieldModifyCount(c); err != nil {
		log.Errorc(c, "ApmEventFieldModifyCount %v", err)
		return
	}
	for _, event := range eventList {
		if _, ok := fieldModify[event.ID]; !ok {
			event.IsReviewed = true
		}
	}
	res.Items = eventList
	return
}

func (s *Service) ApmEventAdd(c context.Context, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string, shared, sampleRate int, busId, datacenterAppID, dataCount int64, lowestSampleRate float64, activity, level, isWideTable, isIgnoreBillions int8) (re interface{}, err error) {
	var (
		datacenterBus     *apm.Bus
		datacenterEventID int64
	)
	if datacenterBus, err = s.fkDao.ApmBusByID(c, busId); err != nil {
		return
	}
	// 添加完成后. 推送企微消息
	defer reportEventCreateResult(&err, &name, &description, &owner, &appKey, &userName, &datacenterAppID, datacenterBus, s.fkDao, s.c)
	if datacenterBus == nil {
		log.Errorc(c, "datacenter bus is nil")
		return
	}
	// 日志平台
	if isIgnoreBillions == 0 {
		if err = s.ApmBillionsEventAdd(c, name, description, userName); err != nil {
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("日志平台技术埋点添加失败,%v", err))
			return
		}
	}
	if logID == apm.LogIdTrackT {
		// 数据平台
		if datacenterEventID, err = s.ApmDataCenterEventAdd(c, logID, name, description, strconv.FormatInt(datacenterAppID, 10), datacenterBus.DatacenterBusKey, topic, strconv.FormatInt(int64(activity), 10), dwdTableName); err != nil {
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("数据平台技术埋点添加失败,%v", err))
			return
		}
	}
	// fawkes增加
	var lastID int64
	if lastID, err = s.ApmFawkesEventAdd(c, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName, level, isWideTable, shared, sampleRate, busId, datacenterEventID, datacenterAppID, dataCount, lowestSampleRate); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Fawkes平台技术埋点添加失败,%v", err))
		return
	}
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if err = s.fkDao.TxApmEventMonitorNotifyConfigSet(tx, lastID, appKey, apm.EventMonitorNotifyOn, apm.EventMonitorNotifyMuteOff, time.Time{}, time.Time{}, userName); err != nil {
			log.Errorc(c, "event monitor notify config set error %v", err)
		}
		return err
	})
	return struct {
		LastID int64 `json:"id"`
	}{LastID: lastID}, err
}

func reportEventCreateResult(err *error, name, description, owner, appKey, userName *string, appId *int64, datacenterBus *apm.Bus, d *fkdao.Dao, c *conf.Config) {
	var (
		contents    string
		messageLink string
	)
	if datacenterBus == nil {
		return
	}
	if (*err) != nil {
		link := fmt.Sprintf("http://ops-log.bilibili.co/app/kibana#/discover?_a=(index:'billions-main.app-svr.fawkes-admin-@*',query:(query_string:(query:'%v')))", (*err).Error())
		u, _ := url.Parse(link)
		u.RawQuery = u.Query().Encode()
		messageLink = u.String()
		contents = fmt.Sprintf("技术埋点【%s】创建失败\n"+
			"[Error]: %s\n"+
			"[应用ID]: %v\n"+
			"[BusName]: %v\n"+
			"[BusDesc]: %v\n"+
			"[描述信息]: %s\n"+
			"[Owner]: %s\n"+
			"[Operator]: %s\n"+
			"[AppKey]: %s", *name, (*err).Error(), *appId, datacenterBus.Name, datacenterBus.Description, *description, *owner, *userName, *appKey)
	} else {
		contents = fmt.Sprintf("技术埋点【%s】创建成功\n"+
			"[应用ID]: %v\n"+
			"[BusName]: %v\n"+
			"[BusDesc]: %v\n"+
			"[描述信息]: %s\n"+
			"[Owner]: %s\n"+
			"[Operator]: %s\n"+
			"[AppKey]: %s", *name, *appId, datacenterBus.Name, datacenterBus.Description, *description, *owner, *userName, *appKey)
		messageLink = c.Host.Fawkes + "/#/apm-manager/apm-event"
	}
	contents = fmt.Sprintf("%s\n[链接]: %s", contents, messageLink)
	if err1 := d.WechatMessageNotify(contents, strings.Join(c.AlarmReceiver.EventMonitorReceiver, "|"), c.Comet.FawkesAppID); err1 != nil {
		log.Error("ApmEventAdd s.fkDao.WechatMessageNotify error(%v)", err1)
	}
}

// ApmBillionsEventAdd 日志平台事件新增
func (s *Service) ApmBillionsEventAdd(c context.Context, appID, appName, servicePrincipal string) (err error) {
	var originCommonFields []*apm.EventField
	if err = s.fkDao.ApmBillionsEventAdd(c, s.c.Billions, appID, appName, servicePrincipal); err != nil {
		return
	}
	if originCommonFields, err = s.fkDao.ApmEventFieldList(c, apm.EventFieldCommonType); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	var commonFields []*apm.EventField
	for _, eventField := range originCommonFields {
		if eventField.ISElasticsearchIndex == 1 {
			commonFields = append(commonFields, eventField)
		}
	}
	// 日志平台监控事件基础字段新增
	err = s.ApmBillionsEventFieldAdd(c, appID, commonFields)
	return
}

// ApmDataCenterEventAdd 数据平台事件新增
func (s *Service) ApmDataCenterEventAdd(c context.Context, logID, eventCode, eventName, proID, bizLine, topic, eventStatus, dwdTableName string) (datacenterEventID int64, err error) {
	datacenterEvent := &apm.DatacenterEvent{
		LogID:              logID,
		EventCode:          eventCode,
		EventName:          eventName,
		EventType:          "track",
		ProID:              proID,
		BizLine:            bizLine,
		Topic:              topic,
		DataWarehouseTable: dwdTableName,
		EventStatus:        eventStatus,
	}
	datacenterEventID, err = s.fkDao.ApmDatacenterEventAdd(c, datacenterEvent, s.c.Datacenter)
	return
}

// ApmFawkesEventAdd  fawkes事件增加
func (s *Service) ApmFawkesEventAdd(c context.Context, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string, level, isWideTable int8, shared, sampleRate int, busId, datacenterEventID, datacenterAppID, dataCount int64, lowestSampleRate float64) (lastID int64, err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if lastID, err = s.fkDao.TxApmEventAdd(tx, name, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName, level, isWideTable, shared, sampleRate, busId, datacenterEventID, datacenterAppID, dataCount, lowestSampleRate); err != nil {
			return err
		}
		if err = s.fkDao.TxApmAppEventRelAdd(tx, lastID, datacenterAppID, datacenterEventID, userName); err != nil {
			return err
		}
		return err
	})
	return
}

func (s *Service) ApmEventDel(c context.Context, eventID int64) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if _, err = s.fkDao.TxApmEventDel(tx, eventID); err != nil {
			log.Errorc(c, "TxApmEventDel %v", err)
			return err
		}
		if _, err = s.fkDao.TxApmEventFieldDelByEventID(tx, eventID, apm.EventDelete); err != nil {
			log.Errorc(c, "TxApmEventFieldDelByEventID %v", err)
			return err
		}
		return err
	})
	return
}

func (s *Service) ApmEventUpdate(c context.Context, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName string, activity, state, level, isWideTable int8, shared, sampleRate int, eventId, datacenterAppID, busID, dataCount int64, lowestSampleRate float64) (err error) {
	var (
		event         *apm.Event
		datacenterBus *apm.Bus
		datacenterRel []*apm.EventDatacenterRel
	)
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event == nil {
		return
	}
	if datacenterBus, err = s.fkDao.ApmBusByID(c, busID); err != nil {
		log.Errorc(c, "ApmBusByID error %v", err)
		return
	}
	if datacenterRel, err = s.fkDao.ApmAppEventRelList(c, eventId, 0); err != nil {
		log.Errorc(c, "ApmAppEventRelList error %v", err)
		return
	}
	for _, rel := range datacenterRel {
		if rel.DatacenterEventId != 0 {
			//	数据平台
			if err = s.ApmDatacenterEventUpdate(c, eventId, rel.DatacenterEventId, logID, name, description, strconv.FormatInt(rel.DatacenterAppId, 10), datacenterBus.DatacenterBusKey, topic, strconv.FormatInt(int64(activity), 10), dwdTableName); err != nil {
				err = ecode.Error(ecode.ServerErr, fmt.Sprintf("数据平台技术埋点更新失败,%v", err))
				return
			}
		}
	}
	//	fawkes
	if err = s.ApmFawkesEventUpdate(c, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName, activity, state, level, isWideTable, shared, sampleRate, eventId, datacenterAppID, busID, event.DatacenterEventID, dataCount, lowestSampleRate); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Fawkes平台技术埋点更新失败,%v", err))
	}
	return
}

// ApmFawkesEventUpdate fawkes事件更新
func (s *Service) ApmFawkesEventUpdate(c context.Context, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName string, activity, state, level, isWideTable int8, shared, sampleRate int, eventId, datacenterAppID, busID, datacenterEventID, dataCount int64, lowestSampleRate float64) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		//	fawkes
		if _, err = s.fkDao.TxApmEventUpdate(tx, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName, activity, state, level, isWideTable, shared, sampleRate, eventId, datacenterAppID, busID, datacenterEventID, dataCount, lowestSampleRate); err != nil {
			log.Errorc(c, "TxApmEventUpdate error %v", err)
		}
		return err
	})
	return
}

// ApmDatacenterEventUpdate  数据平台事件更新
func (s *Service) ApmDatacenterEventUpdate(c context.Context, eventId, id int64, logID, eventCode, eventName, proID, bizLine, topic, eventStatus, dwdTableName string) (err error) {
	var (
		fv       int64
		files    []*apm.EventFieldFile
		fields   []*apm.EventField
		dcFields []*apm.DatacenterField
	)
	// 查询已发布的字段
	if fv, err = s.fkDao.ApmEventFieldFileLastFV(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if files, err = s.fkDao.ApmEventFieldFileList(c, eventId, fv); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	for _, file := range files {
		if file.IsClickhouse != 0 && file.FieldState != apm.EventFieldStateDelete {
			fields = append(fields, file.FileConvertToField())
		}
	}
	for index, field := range fields {
		dcField := apmDataCenterEventFieldMapping(index, field)
		dcFields = append(dcFields, dcField)
	}
	datacenterEvent := &apm.DatacenterEvent{
		ID:                 id,
		LogID:              logID,
		EventCode:          eventCode,
		EventName:          eventName,
		EventType:          "track",
		ProID:              proID,
		BizLine:            bizLine,
		Topic:              topic,
		DataWarehouseTable: dwdTableName,
		EventStatus:        eventStatus,
		Fields:             dcFields,
	}
	err = s.fkDao.ApmDatacenterEventUpdate(c, datacenterEvent, s.c.Datacenter)
	return
}

func (s *Service) ApmEventFieldList(c context.Context, eventId int64) (eventFields *apm.EventFieldGroup, err error) {
	var (
		commonFields   []*apm.EventField
		extendedFields []*apm.EventField
	)
	if commonFields, err = s.fkDao.ApmEventFieldList(c, apm.EventFieldCommonType); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	if extendedFields, err = s.fkDao.ApmEventFieldList(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	eventFields = &apm.EventFieldGroup{CommonFields: commonFields, ExtendedFields: extendedFields}
	return
}

func (s *Service) ApmEventFieldSet(c context.Context, req *apm.EventFieldReq) (err error) {
	if err = s.ApmFawkesEventFieldOperation(c, req.EventID, req.Operator, req.Fields); err != nil {
		log.Errorc(c, "ApmFawkesEventFieldOperation error %v", err)
	}
	return
}

// ApmEventFieldSetOperation	判断set动作：add，del，update
func (s *Service) ApmEventFieldSetOperation(c context.Context, eventID int64, newFields []*apm.EventField) (as, ds, ups []*apm.EventField, err error) {
	var originFields []*apm.EventField
	if originFields, err = s.fkDao.ApmEventFieldList(c, eventID); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	// add && update
	for _, newField := range newFields {
		var (
			isAppend = true
			isUpdate = false
			isModify = true
		)
		newField.EventID = eventID
		for _, origin := range originFields {
			newField.State = origin.State
			if newField.Key == origin.Key {
				// 无变化
				if newField.Example == origin.Example && newField.DefaultValue == origin.DefaultValue && newField.Description == origin.Description && newField.Type == origin.Type &&
					newField.IsClickhouse == origin.IsClickhouse && newField.ISElasticsearchIndex == origin.ISElasticsearchIndex && newField.ElasticSearchFieldType == origin.ElasticSearchFieldType {
					isModify = false
					break
				}
				newField.ID = origin.ID
				isAppend = false
				isUpdate = true
				break
			}
		}
		if isAppend && isModify {
			as = append(as, newField)
		}
		if isUpdate && isModify {
			ups = append(ups, newField)
		}
	}
	// delete
	for _, origin := range originFields {
		var isDelete = true
		for _, newField := range newFields {
			if newField.Key == origin.Key {
				isDelete = false
				break
			}
		}
		if isDelete {
			ds = append(ds, origin)
		}
	}
	return
}

// ApmBillionsEventFieldPublish 日志平台字段操作
func (s *Service) ApmBillionsEventFieldPublish(c context.Context, eventID int64, eventFields []*apm.EventField) (err error) {
	var event *apm.Event
	if event, err = s.fkDao.ApmEvent(c, eventID); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event == nil {
		log.Errorc(c, "event is nil, eventId = %v", eventID)
		return
	}
	// 日志平台监控事件字段新增
	var (
		originCommonFields []*apm.EventField
		commonFields       []*apm.EventField
		extendedFields     []*apm.EventField
		billionsFields     []*apm.EventField
	)
	for _, eventField := range eventFields {
		eventField.EventID = eventID
		if eventField.ISElasticsearchIndex == 1 && eventField.State != apm.EventFieldStateDelete {
			extendedFields = append(extendedFields, eventField)
		}
	}
	if originCommonFields, err = s.fkDao.ApmEventFieldList(c, apm.EventFieldCommonType); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	for _, eventField := range originCommonFields {
		if eventField.ISElasticsearchIndex == 1 {
			commonFields = append(commonFields, eventField)
		}
	}
	billionsFields = append(append(billionsFields, commonFields...), extendedFields...)
	err = s.ApmBillionsEventFieldAdd(c, event.Name, billionsFields)
	return
}

// ApmDatacenterEventFieldPublish 数据平台字段发布
func (s *Service) ApmDatacenterEventFieldPublish(c context.Context, eventId, datacenterEventId, datacenterAppId int64, eventFields []*apm.EventField) (err error) {
	var (
		event    *apm.Event
		bus      *apm.Bus
		dcFields []*apm.DatacenterField
	)
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event == nil {
		log.Errorc(c, "event is nil, eventId = %v", eventId)
		return
	}
	if bus, err = s.fkDao.ApmBusByID(c, event.BusID); err != nil {
		log.Errorc(c, "ApmBusByID error %v", err)
		return
	}
	for index, field := range eventFields {
		if field.State != apm.EventFieldStateDelete {
			dcField := apmDataCenterEventFieldMapping(index, field)
			dcFields = append(dcFields, dcField)
		}
	}
	datacenterEvent := &apm.DatacenterEvent{
		ID:          datacenterEventId,
		LogID:       event.LogID,
		EventCode:   event.Name,
		EventName:   event.Description,
		EventType:   "track",
		ProID:       strconv.FormatInt(datacenterAppId, 10),
		BizLine:     bus.DatacenterBusKey,
		Topic:       event.Topic,
		EventStatus: strconv.FormatInt(int64(event.Activity), 10),
		Fields:      dcFields,
	}
	if err = s.fkDao.ApmDatacenterEventUpdate(c, datacenterEvent, s.c.Datacenter); err != nil {
		log.Errorc(c, "ApmDatacenterEventUpdate error %v", err)
	}
	return
}

func (s *Service) ApmFawkesEventFieldOperation(c context.Context, eventID int64, operator string, eventFields []*apm.EventField) (err error) {
	var as, ds, ups []*apm.EventField
	as, ds, ups, err = s.ApmEventFieldSetOperation(c, eventID, eventFields)
	if err != nil {
		log.Errorc(c, "ApmEventFieldSetOperation error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		// fawkes add
		if len(as) > 0 {
			if _, err = s.fkDao.TxApmAddEventField(tx, eventID, as, operator); err != nil {
				log.Errorc(c, "TxApmAddEventField error %v", err)
				return err
			}
		}
		//	fawkes update
		if len(ups) > 0 {
			if err = s.ApmEventFieldModify(c, eventID, operator, ups); err != nil {
				log.Errorc(c, "ApmEventFieldModify error %v", err)
				return err
			}
		}
		// fawkes del
		if len(ds) > 0 {
			for _, d := range ds {
				// 未审核状态
				// add -> del
				if d.State == apm.EventFieldStateAdd {
					if err = s.fkDao.TxApmEventFieldDelById(tx, d.ID); err != nil {
						log.Errorc(c, "TxApmEventFieldDelById error %v", err)
						return err
					}
				} else {
					if err = s.fkDao.TxApmEventFieldStateUpdateById(tx, d.ID, apm.EventFieldStateDelete); err != nil {
						log.Errorc(c, "TxApmEventFieldDelByUpdate error %v", err)
						return err
					}
				}
			}
		}
		return err
	})
	return
}

func (s *Service) ApmEventFieldModify(c context.Context, eventId int64, operator string, ups []*apm.EventField) (err error) {
	var (
		fv    int64
		files []*apm.EventFieldFile
	)
	if fv, err = s.fkDao.ApmEventFieldFileLastFV(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if files, err = s.fkDao.ApmEventFieldFileList(c, eventId, fv); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	fileMap := make(map[string]*apm.EventFieldFile)
	for _, file := range files {
		if file.FieldState == apm.EventFieldStateDelete {
			continue
		}
		fileMap[file.FieldKey] = file
	}
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		for _, up := range ups {
			// 未审核状态  字段状态A->字段状态B->字段状态A  判断
			if field, ok := fileMap[up.Key]; ok {
				if field.Example == up.Example && field.FieldType == up.Type && field.DefaultValue == up.DefaultValue && field.Description == up.Description &&
					field.IsElasticsearchIndex == up.ISElasticsearchIndex && field.IsClickhouse == up.IsClickhouse && field.ElasticsearchFieldType == up.ElasticSearchFieldType {
					if err = s.fkDao.TxApmEventFieldStateUpdateById(tx, field.FieldId, apm.EventFieldStateReviewed); err != nil {
						log.Errorc(c, "TxApmEventFieldStateUpdateById error %v", err)
						return err
					}
					return err
				}
			}
			// 未审核状态
			// add -> update
			// del -> update
			if up.State == apm.EventFieldStateAdd {
				if _, err = s.fkDao.TxApmEventFieldUpdate(tx, up.Example, up.Description, up.DefaultValue, operator, up.IsClickhouse, up.ISElasticsearchIndex, apm.EventFieldStateAdd, up.ElasticSearchFieldType, up.Index, up.ID, up.Type, up.Mode); err != nil {
					log.Errorc(c, "TxApmEventFieldUpdate error %v", err)
					return err
				}
			} else if up.State == apm.EventFieldStateDelete {
				return err
			} else {
				if _, err = s.fkDao.TxApmEventFieldUpdate(tx, up.Example, up.Description, up.DefaultValue, operator, up.IsClickhouse, up.ISElasticsearchIndex, apm.EventFieldStateModify, up.ElasticSearchFieldType, up.Index, up.ID, up.Type, up.Mode); err != nil {
					log.Errorc(c, "TxApmEventFieldUpdate error %v", err)
					return err
				}
			}
		}
		return err
	})
	return
}

// ApmBillionsEventFieldAdd 日志平台监控事件字段新增
func (s *Service) ApmBillionsEventFieldAdd(c context.Context, appId string, eventFields []*apm.EventField) (err error) {
	var mapping = &apm.BillionsEventFieldMapping{}
	mapping.AppID = appId
	if len(eventFields) > 0 {
		for _, eventField := range eventFields {
			billionsEventField := apmBillionsEventFieldMapping(eventField)
			mapping.Fields = append(mapping.Fields, billionsEventField)
		}
	}
	if len(mapping.Fields) == 0 {
		return
	}
	//	日志平台新增字段
	err = s.fkDao.ApmBillionsAddEventField(c, s.c.Billions, mapping)
	return
}

// apmBillionsEventFieldMapping 日志平台字段类型映射
func apmBillionsEventFieldMapping(eventField *apm.EventField) (billionsEventField *apm.BillionsEventField) {
	billionsEventField = &apm.BillionsEventField{}
	switch eventField.ElasticSearchFieldType {
	case apm.EventFieldTypeByte:
		billionsEventField.Type = "byte"
	case apm.EventFieldTypeInteger:
		billionsEventField.Type = "integer"
	case apm.EventFieldTypeLong:
		billionsEventField.Type = "long"
	case apm.EventFieldTypeDouble:
		billionsEventField.Type = "double"
	case apm.EventFieldTypeDateTime:
		billionsEventField.Type = "date"
	case apm.EventFieldTypeFloat:
		billionsEventField.Type = "float"
	case apm.EventFieldTypeString_Text:
		billionsEventField.Type = "text"
	default:
		billionsEventField.Type = "keyword"
	}
	if eventField.Mode == 0 && eventField.EventID == 0 {
		billionsEventField.Name = eventField.Key
	} else {
		billionsEventField.Name = fmt.Sprintf("extended_fields.%s", eventField.Key)
	}
	return
}

// apmDataCenterEventFieldMapping 数据平台字段映射
func apmDataCenterEventFieldMapping(index int, eventField *apm.EventField) (datacenterField *apm.DatacenterField) {
	datacenterField = &apm.DatacenterField{}
	switch eventField.Type {
	case apm.EventFieldTypeByte, apm.EventFieldTypeInteger, apm.EventFieldTypeLong:
		datacenterField.FieldType = "int"
	case apm.EventFieldTypeDateTime:
		datacenterField.FieldType = "datetime"
	default:
		datacenterField.FieldType = "string"
	}
	datacenterField.FieldId = eventField.Key
	// 数据平台默认属性显示名格式=index.description
	datacenterField.FieldName = fmt.Sprintf("%v.%v", index, eventField.Description)
	if len(datacenterField.FieldName) > apm.DatacenterFieldNameMaxLen {
		datacenterField.FieldName = datacenterField.FieldName[:apm.DatacenterFieldNameMaxLen]
	}
	if len(eventField.Example) > apm.DatacenterFieldDescMaxLen {
		eventField.Example = eventField.Example[:apm.DatacenterFieldDescMaxLen]
	}
	datacenterField.FieldDesc = eventField.Example
	return
}

func (s *Service) ApmEventSql(c context.Context, eventID int64) (sqls interface{}, err error) {
	var (
		event                                *apm.Event
		extendedFields, commonFields, fields []*apm.EventFieldSql
	)
	if event, err = s.ApmEvent(c, eventID); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	if event.DistributedTableName == "" {
		return "", err
	}
	if commonFields, extendedFields, err = s.apmEventFieldSqlGenerator(c, eventID); err != nil {
		log.Errorc(c, "apmEventFieldsGenerator error %v", err)
		return
	}
	fields = append(append(fields, commonFields...), extendedFields...)
	sqls = s.apmEventSqlGenerator(event.Databases, event.TableName, event.DistributedTableName, fields)
	return
}

func (s *Service) apmEventFieldSqlGenerator(c context.Context, eventId int64) (commonFields, extendedFields []*apm.EventFieldSql, err error) {
	var (
		extendedFv, commonFv       int64
		extendedFiles, commonFiles []*apm.EventFieldFile
	)
	// 从file归档表中读最新的字段集合 扩展字段
	if extendedFv, err = s.fkDao.ApmEventFieldFileLastFV(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if extendedFiles, err = s.fkDao.ApmEventFieldFileList(c, eventId, extendedFv); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	// 基础字段
	if commonFv, err = s.fkDao.ApmEventFieldFileLastFV(c, apm.EventFieldCommonType); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if commonFiles, err = s.fkDao.ApmEventFieldFileList(c, apm.EventFieldCommonType, commonFv); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	for _, file := range extendedFiles {
		//	跳过不进入clickhouse的字段
		if file.IsClickhouse == 0 {
			continue
		}
		// 跳过删除的字段
		if file.FieldState == apm.EventFieldStateDelete {
			continue
		}
		extendedField := apmEventFieldSqlMapping(file.FieldKey, file.Description, file.FieldType, file.FieldIndex)
		extendedField.Name = fmt.Sprintf("`%s`", strings.TrimSpace(extendedField.Name))
		extendedFields = append(extendedFields, extendedField)
	}
	for _, file := range commonFiles {
		//	跳过不进入clickhouse的字段
		if file.IsClickhouse == 0 {
			continue
		}
		// 跳过删除的字段
		if file.FieldState == apm.EventFieldStateDelete {
			continue
		}
		commonField := apmEventFieldSqlMapping(file.FieldKey, file.Description, file.FieldType, file.FieldIndex)
		commonField.Name = fmt.Sprintf("`%s`", strings.TrimSpace(commonField.Name))
		commonFields = append(commonFields, commonField)
	}
	sort.Slice(commonFields, func(i, j int) bool {
		return commonFields[i].Index < commonFields[j].Index
	})
	return
}

func (s *Service) apmEventSqlGenerator(dbName, tableName, distributedTableName string, fields []*apm.EventFieldSql) (sqls *apm.EventSqlRes) {
	eventSqlTemplate := &apm.EventSqlTemplate{DBName: dbName, TableName: tableName, DisTableName: distributedTableName, Fields: fields}
	sqls = &apm.EventSqlRes{}
	funcMap := template.FuncMap{
		"maxIndex": func() int {
			return len(fields) - 1
		},
	}
	sqls.CreateSql, _ = s.fkDao.TemplateAlterFunc(eventSqlTemplate, funcMap, apm.EventTemplateCreateSql)
	sqls.CreateSqlDis, _ = s.fkDao.TemplateAlterFunc(eventSqlTemplate, funcMap, apm.EventTemplateCreateSqlDis)
	return
}

// apmEventFieldSqlMapping 监控事件字段和sql映射
func apmEventFieldSqlMapping(fieldKey, desc string, fieldType int8, fieldIndex int64) (sql *apm.EventFieldSql) {
	sql = &apm.EventFieldSql{}
	sql.Name = fieldKey
	sql.Index = fieldIndex
	sql.Desc = desc
	switch fieldType {
	case apm.EventFieldTypeString_Keyword, apm.EventFieldTypeString_Text:
		sql.Type = "String"
	case apm.EventFieldTypeByte:
		sql.Type = "Int8"
	case apm.EventFieldTypeInteger:
		sql.Type = "Int32"
	case apm.EventFieldTypeLong:
		sql.Type = "Int64"
	case apm.EventFieldTypeDouble:
		sql.Type = "Float64"
	case apm.EventFieldTypeDateTime:
		sql.Type = "DateTime"
	case apm.EventFieldTypeFloat:
		sql.Type = "Float32"
	case apm.EventFieldTypeUInt64:
		sql.Type = "UInt64"
	case apm.EventFieldTypeMap:
		sql.Type = "Map(String,String)"
	}
	return
}

func (s *Service) ApmEventAdvancedList(c context.Context, eventID int64) (res interface{}, err error) {
	var (
		items []*apm.EventAdvanced
		count int64
	)
	if count, err = s.fkDao.ApmEventAdvancedCount(c, eventID); err != nil {
		log.Errorc(c, "ApmEventAdvancedCount error %v", err)
		return
	}
	if items, err = s.fkDao.ApmEventAdvancedList(c, eventID); err != nil {
		log.Errorc(c, "ApmEventAdvancedList error %v", err)
		return
	}
	return struct {
		Items []*apm.EventAdvanced `json:"items"`
		Count int64                `json:"count"`
	}{
		Items: items,
		Count: count,
	}, err
}

func (s *Service) ApmEventAdvancedAdd(c context.Context, eventID, displayType int64, fieldName, title, description, queryType, mappingGroup, customSql, operator string) (err error) {
	if err = s.fkDao.ApmEventAdvancedAdd(c, eventID, displayType, fieldName, title, description, queryType, mappingGroup, customSql, operator); err != nil {
		log.Errorc(c, "ApmEventAdvancedAdd error %v", err)
	}
	return
}

func (s *Service) ApmEventAdvancedDel(c context.Context, id int64) (err error) {
	if err = s.fkDao.ApmEventAdvancedDel(c, id); err != nil {
		log.Errorc(c, "ApmEventAdvancedDel error %v", err)
	}
	return
}

func (s *Service) ApmEventAdvancedUpdate(c context.Context, id, displayType int64, title, description, queryType, mappingGroup, customSql, operator string) (err error) {
	if err = s.fkDao.ApmEventAdvancedUpdate(c, id, displayType, title, description, queryType, mappingGroup, customSql, operator); err != nil {
		log.Errorc(c, "ApmEventAdvancedUpdate error %v", err)
	}
	return
}

//// BusApmEventAdd fawkes外部接口监控事件增加
//func (s *Service) BusApmEventAdd(c context.Context, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName string, level, isWideTable int8, shared, sampleRate int, busId, datacenterAppID, datacenterEventID, dataCount int64) (res interface{}, err error) {
//	var lastID int64
//	if lastID, err = s.ApmFawkesEventAdd(c, name, appKey, appKeys, description, owner, userName, logID, dbName, tableName, distributedTableName, topic, dwdTableName, level, isWideTable, shared, sampleRate, busId, datacenterEventID, datacenterAppID, dataCount); err != nil {
//		log.Errorc(c, "ApmFawkesEventAdd error %v", err)
//		return
//	}
//	return struct {
//		LastID int64 `json:"id"`
//	}{LastID: lastID}, err
//}
//
//// BusApmEventUpdate fawkes外部接口监控事件更新
//func (s *Service) BusApmEventUpdate(c context.Context, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName string, activity, state, level, isWideTable int8, shared, sampleRate int, eventId, datacenterAppID, busID, datacenterEventID, dataCount int64) (err error) {
//	if err = s.ApmFawkesEventUpdate(c, appKeys, description, Owner, userName, logID, dbName, tableName, distributedTableName, topic, name, dwdTableName, activity, state, level, isWideTable, shared, sampleRate, eventId, datacenterAppID, busID, datacenterEventID, dataCount); err != nil {
//		log.Errorc(c, "ApmFawkesEventUpdate error %v", err)
//	}
//	return
//}

// BusApmEventFieldSet fawkes外部接口监控事件扩展字段增加
func (s *Service) BusApmEventFieldSet(c context.Context, req *apm.EventFieldReq) (err error) {
	// fawkes
	if err = s.ApmFawkesEventFieldOperation(c, req.EventID, req.Operator, req.Fields); err != nil {
		log.Errorc(c, "ApmFawkesEventFieldOperation error %v", err)
	}
	return
}

// ApmAppEventList 指定app下的event列表
func (s *Service) ApmAppEventList(c context.Context, appKey, name, busName, logId, orderBy string, pn, ps int, state int8) (res *apm.ResultEventList, err error) {
	var (
		app         *appmdl.APP
		events      []*apm.Event
		total       int
		fieldModify map[int64]int64
	)
	res = &apm.ResultEventList{}
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "AppPass error %v", err)
		return
	}
	if app == nil || app.DataCenterAppID == 0 {
		return
	}
	if total, err = s.fkDao.ApmEventListCountWithAppId(c, name, busName, logId, app.DataCenterAppID, state); err != nil {
		log.Error("ApmEventListCountWithAppKey error %v", err)
		return
	}
	page := &apm.Page{Total: total, PageSize: ps, PageNum: pn}
	res.PageInfo = page
	if total == 0 {
		return
	}
	if events, err = s.fkDao.ApmEventListWithAppId(c, name, busName, logId, orderBy, app.DataCenterAppID, pn, ps, state); err != nil {
		log.Error("ApmEventListByAppKey error %v", err)
		return
	}
	if fieldModify, err = s.fkDao.ApmEventFieldModifyCount(c); err != nil {
		log.Errorc(c, "ApmEventFieldModifyCount %v", err)
		return
	}
	for _, event := range events {
		if _, ok := fieldModify[event.ID]; !ok {
			event.IsReviewed = true
		}
	}
	res.Items = events
	return
}

// ApmAppEventRelAdd app和event关系增加¸
func (s *Service) ApmAppEventRelAdd(c context.Context, req *apm.EventDatacenterRel) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if err = s.fkDao.TxApmAppEventRelAdd(tx, req.EventId, req.DatacenterAppId, req.DatacenterEventId, req.Operator); err != nil {
			log.Errorc(c, "TxAppEventRelAdd error %v", err)
		}
		return err
	})
	return
}

// ApmEventFieldBillionsSync fawkes平台技术埋点字段同步到日志平台
func (s *Service) ApmEventFieldBillionsSync(c context.Context, eventId int64) (err error) {
	var (
		extendedFields []*apm.EventField
	)
	if extendedFields, err = s.fkDao.ApmEventFieldList(c, eventId); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	// 日志平台监控事件字段刷新
	if err = s.ApmBillionsEventFieldPublish(c, eventId, extendedFields); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	return
}

func (s *Service) BillionsQueryBodyGenerate(pn, ps int64, query string, queryRange *apm.BillionsQueryRange, querySort *apm.BillionsQuerySort, queryAggs []*apm.BillionsQueryAggs) (body string, err error) {
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
	body, err = s.fkDao.TemplateAlterFunc(billionsQueryBody, funcMap, apm.BillionsTemplateQueryBody)
	return
}

// ApmEventFieldPublish 技术埋点字段发布
func (s *Service) ApmEventFieldPublish(c context.Context, eventId int64, isIgnoreBillions int8, operator string) (err error) {
	var (
		fields        []*apm.EventField
		datacenterRel []*apm.EventDatacenterRel
	)
	if fields, err = s.fkDao.ApmEventFieldList(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	// 日志平台
	if isIgnoreBillions == 0 {
		if err = s.ApmBillionsEventFieldPublish(c, eventId, fields); err != nil {
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("日志平台字段发布失败,%v", err))
			return
		}
	}
	// 数据平台
	if datacenterRel, err = s.fkDao.ApmAppEventRelList(c, eventId, 0); err != nil {
		log.Errorc(c, "ApmAppEventRelList error %v", err)
		return
	}
	for _, rel := range datacenterRel {
		if rel.DatacenterEventId != 0 {
			if err = s.ApmDatacenterEventFieldPublish(c, eventId, rel.DatacenterEventId, rel.DatacenterAppId, fields); err != nil {
				err = ecode.Error(ecode.ServerErr, fmt.Sprintf("数据平台字段发布失败,%v", err))
				return
			}
		}
	}
	// fawkes
	if err = s.ApmFawkesEventFieldPublish(c, eventId, fields, operator); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Fawkes平台字段发布失败,%v", err))
		return
	}
	return
}

// ApmFawkesEventFieldPublish fawkes技术埋点字段发布
func (s *Service) ApmFawkesEventFieldPublish(c context.Context, eventId int64, fields []*apm.EventField, operator string) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		// nolint:gomnd
		fv := time.Now().UnixNano() / 1e6
		if err = s.fkDao.TxApmEventFieldFileAdd(tx, fv, fields, operator); err != nil {
			log.Errorc(c, "TxApmEventFieldFileAdd error %v", err)
			return err
		}
		if err = s.fkDao.TxApmEventFieldPublish(tx, eventId, fv, operator); err != nil {
			log.Errorc(c, "TxApmEventFieldPublish error %v", err)
			return err
		}
		if _, err = s.fkDao.TxApmEventFieldDelByEventID(tx, eventId, apm.EventFieldStateDelete); err != nil {
			log.Errorc(c, "ApmEventFieldPublishDel error %v", err)
			return err
		}
		if err = s.fkDao.TxApmEventFieldStateUpdate(tx, eventId, apm.EventFieldStateReviewed); err != nil {
			log.Errorc(c, "TxApmEventFieldStateUpdate error %v", err)
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "ApmEventFieldPublish error %v", err)
		return
	}
	return
}

// ApmEventFieldPublishDiff 技术埋点字段发布的diff
func (s *Service) ApmEventFieldPublishDiff(c context.Context, eventId int64) (res []*apm.EventFieldDiff, err error) {
	var (
		newFields    []*apm.EventField
		originFields []*apm.EventField
	)
	if newFields, err = s.fkDao.ApmEventFieldList(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	var (
		fv    int64
		files []*apm.EventFieldFile
	)
	if fv, err = s.fkDao.ApmEventFieldFileLastFV(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldLastFV error %v", err)
		return
	}
	if fv > 0 {
		if files, err = s.fkDao.ApmEventFieldFileList(c, eventId, fv); err != nil {
			log.Errorc(c, "ApmEventFieldFileList error %v", err)
			return
		}
	}
	for _, file := range files {
		originFields = append(originFields, file.FileConvertToField())
	}
	res = s.ApmEventFieldDiffGenerator(originFields, newFields)
	return
}

// ApmEventFieldPublishHistory 技术埋点字段发布历史
func (s *Service) ApmEventFieldPublishHistory(c context.Context, eventId int64, pn, ps int) (res *apm.EventFieldPublishResp, err error) {
	var (
		count       int
		publishInfo []*apm.EventFieldPublish
	)
	if count, err = s.fkDao.ApmEventFieldPublishHistoryCount(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldPublishHistoryCount error %v", err)
		return
	}
	page := &apm.Page{Total: count, PageNum: pn, PageSize: ps}
	res = &apm.EventFieldPublishResp{PageInfo: page}
	if count < 1 {
		return
	}
	if publishInfo, err = s.fkDao.ApmEventFieldPublishHistory(c, eventId, pn, ps); err != nil {
		log.Errorc(c, "ApmEventFieldPublishHistory error %v", err)
	}
	res.Items = publishInfo
	return
}

// ApmEventFieldDiff 技术埋点字段发布历史diff
func (s *Service) ApmEventFieldDiff(c context.Context, eventId, version int64) (res []*apm.EventFieldDiff, err error) {
	var (
		newFiles     []*apm.EventFieldFile
		newFields    []*apm.EventField
		originFiles  []*apm.EventFieldFile
		originFields []*apm.EventField
		lastVersion  int64
	)
	if newFiles, err = s.fkDao.ApmEventFieldFileList(c, eventId, version); err != nil {
		log.Errorc(c, "ApmEventFieldFileList error %v", err)
		return
	}
	if lastVersion, err = s.fkDao.ApmEventFieldPublishLastVersion(c, eventId, version); err != nil {
		log.Errorc(c, "ApmEventFieldPublishLastVersion error %v", err)
		return
	}
	if lastVersion > 0 {
		if originFiles, err = s.fkDao.ApmEventFieldFileList(c, eventId, lastVersion); err != nil {
			log.Errorc(c, "ApmEventFieldFileList error %v", err)
			return
		}
	}
	for _, file := range newFiles {
		newFields = append(newFields, file.FileConvertToField())
	}
	for _, file := range originFiles {
		originFields = append(originFields, file.FileConvertToField())
	}
	res = s.ApmEventFieldDiffGenerator(originFields, newFields)
	return
}

// ApmEventFieldDiffGenerator 技术埋点新老字段diff判断
func (s *Service) ApmEventFieldDiffGenerator(originFields []*apm.EventField, newFields []*apm.EventField) (diff []*apm.EventFieldDiff) {
	originMap := make(map[string]*apm.EventField)
	for _, origin := range originFields {
		if origin.State == apm.EventFieldStateDelete {
			continue
		}
		originMap[origin.Key] = origin
	}
	for _, newField := range newFields {
		var (
			state int8
			oldF  *apm.EventField
			newF  *apm.EventField
		)
		if field, ok := originMap[newField.Key]; ok {
			// 新字段状态是已审核 代表 无变化
			if newField.State == apm.EventFieldStateReviewed {
				continue
			}
			// 更改 || 删除
			oldF = field
			state = newField.State
		} else {
			// 增加
			state = apm.EventFieldStateAdd
		}
		if state != apm.EventFieldStateDelete {
			newF = newField
		}
		re := &apm.EventFieldDiff{
			Old:   oldF,
			New:   newF,
			State: state,
		}
		diff = append(diff, re)
	}
	return
}

func (s *Service) ApmEventFieldTypeSync(c context.Context, eventId int64) (err error) {
	var (
		fields []*apm.EventField
	)
	if fields, err = s.fkDao.ApmEventFieldList(c, eventId); err != nil {
		log.Errorc(c, "ApmEventFieldList error %v", err)
		return
	}
	for _, field := range fields {
		field.ElasticSearchFieldType = field.Type
		if err = s.fkDao.ApmEventFieldStateSync(c, field.ID, field.ElasticSearchFieldType); err != nil {
			log.Errorc(c, "ApmEventFieldStateSync error %v", err)
			return
		}
	}
	return
}

// ApmAppCommonFieldGroupAdd App下基础字段组添加
func (s *Service) ApmAppCommonFieldGroupAdd(c context.Context, appKey, name, description, operator string, isDefault int8, fields []*apm.AppEventCommonField) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		var groupId int64
		if groupId, err = s.fkDao.TxApmAppEventCommonFieldGroupAdd(tx, appKey, name, description, operator, isDefault); err != nil {
			log.Errorc(c, "TxApmAppEventCommonFieldGroupAdd error %v", err)
			return err
		}
		if err = s.fkDao.TxApmAppEventCommonFieldAdd(tx, appKey, groupId, operator, fields); err != nil {
			log.Errorc(c, "TxApmAppEventCommonFieldAdd error %v", err)
			return err
		}
		return err
	})
	return
}

// ApmAppCommonFieldGroupUpdate App下基础字段组更新
func (s *Service) ApmAppCommonFieldGroupUpdate(c context.Context, appKey, name, description, operator string, isDefault int8, groupId int64, fields []*apm.AppEventCommonField) (err error) {
	var (
		as, ds, ups []*apm.AppEventCommonField
	)
	if as, ds, ups, err = s.ApmAppCommonFieldGroupSetOperation(c, groupId, fields); err != nil {
		log.Errorc(c, "ApmAppCommonFieldGroupSetOperation error %v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if err = s.fkDao.TxApmAppEventCommonFieldGroupUpdate(tx, name, description, operator, isDefault, groupId); err != nil {
			log.Errorc(c, "TxApmAppEventCommonFieldGroupUpdate error %v", err)
			return err
		}
		if len(as) > 0 {
			if err = s.fkDao.TxApmAppEventCommonFieldAdd(tx, appKey, groupId, operator, as); err != nil {
				log.Errorc(c, "TxApmAppEventCommonFieldAdd error %v", err)
				return err
			}
		}
		if len(ups) > 0 {
			for _, up := range ups {
				if err = s.fkDao.TxApmAppEventCommonFieldUpdate(tx, up.Description, up.DefaultValue, operator, up.Type, up.IsClickhouse, up.IsElasticsearchIndex, up.ElasticsearchFieldType, up.Id, up.Index); err != nil {
					log.Errorc(c, "TxApmAppEventCommonFieldUpdate error %v", err)
					return err
				}
			}
		}
		if len(ds) > 0 {
			for _, d := range ds {
				if err = s.fkDao.TxApmAppEventCommonFieldDel(tx, d.Id); err != nil {
					log.Errorc(c, "TxApmAppEventCommonFieldDel error %v", err)
					return err
				}
			}
		}
		return err
	})
	return
}

// ApmAppCommonFieldGroupSetOperation App下的基础字段组的增删改判断
func (s *Service) ApmAppCommonFieldGroupSetOperation(c context.Context, groupId int64, newFields []*apm.AppEventCommonField) (as, ds, ups []*apm.AppEventCommonField, err error) {
	var (
		originFields []*apm.AppEventCommonField
	)
	if originFields, err = s.fkDao.ApmAppEventCommonFieldList(c, groupId); err != nil {
		log.Errorc(c, "ApmAppEventComFieldList error %v", err)
		return
	}
	// add && update
	for _, newField := range newFields {
		var (
			isAppend = true
			isUpdate = false
		)
		for _, origin := range originFields {
			if newField.Key == origin.Key {
				// 没有变化
				if newField.DefaultValue == origin.DefaultValue && newField.Description == origin.Description && newField.Type == origin.Type && newField.Index == origin.Index &&
					newField.IsClickhouse == origin.IsClickhouse && newField.IsClickhouse == origin.IsElasticsearchIndex && newField.ElasticsearchFieldType == origin.ElasticsearchFieldType {
					isAppend = false
					isUpdate = false
					break
				}
				newField.Id = origin.Id
				isAppend = false
				isUpdate = true
			}
		}
		if isAppend {
			as = append(as, newField)
		}
		if isUpdate {
			ups = append(ups, newField)
		}
	}
	// del
	for _, origin := range originFields {
		var isDelete = true
		for _, newField := range newFields {
			if newField.Key == origin.Key {
				isDelete = false
				break
			}
		}
		if isDelete {
			ds = append(ds, origin)
		}
	}
	return
}

// ApmAppCommonFieldGroupDel App下基础字段组删除
func (s *Service) ApmAppCommonFieldGroupDel(c context.Context, groupId int64) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if err = s.fkDao.TxApmAppEventCommonFieldGroupDel(tx, groupId); err != nil {
			log.Errorc(c, "TxApmAppEventCommonFieldGroupDel error %v", err)
			return err
		}
		if err = s.fkDao.TxApmAppEventCommonFieldDelByGroupId(tx, groupId); err != nil {
			log.Errorc(c, "TxApmAppEventCommonFieldDelByGroupId error %v", err)
			return err
		}
		return err
	})
	return
}

// ApmAppCommonFieldGroupList App下基础字段组查询list
func (s *Service) ApmAppCommonFieldGroupList(c context.Context, appKey string, pn, ps int) (res *apm.EventCommonFieldGroupResp, err error) {
	var (
		group []*apm.EventCommonFieldGroup
		total int64
	)
	if total, err = s.fkDao.ApmAppEventCommonFieldGroupCount(c, appKey); err != nil {
		log.Errorc(c, "ApmAppEventCommonFieldGroupCount error %v", err)
		return
	}
	page := &apm.Page{Total: int(total), PageNum: pn, PageSize: ps}
	res = &apm.EventCommonFieldGroupResp{PageInfo: page}
	if total < 1 {
		return
	}
	if group, err = s.fkDao.ApmAppEventCommonFieldGroupList(c, appKey, pn, ps); err != nil {
		log.Errorc(c, "ApmAppEventCommonFieldGroupList error %v", err)
		return
	}
	res.Items = group
	return
}

// ApmAppCommonFieldGroup 查询指定的基础字段组
func (s *Service) ApmAppCommonFieldGroup(c context.Context, groupId int64) (res *apm.EventCommonFieldGroup, err error) {
	if res, err = s.fkDao.ApmAppEventCommonFieldGroupById(c, groupId); err != nil {
		log.Errorc(c, "ApmAppEventCommonFieldGroupById error %v", err)
		return
	}
	var fields []*apm.AppEventCommonField
	if fields, err = s.fkDao.ApmAppEventCommonFieldList(c, groupId); err != nil {
		log.Errorc(c, "ApmAppEventCommonFieldList error %v", err)
		return
	}
	res.Fields = fields
	return
}

// ApmEventCKTableCreate 技术埋点clickhouse表创建
func (s *Service) ApmEventCKTableCreate(c context.Context, req *apm.CKTableCreateReq) (err error) {
	var (
		event                            *apm.Event
		deCommFields, deExtendFields     []*apm.EventFieldSql
		commCkCols, extendCkCols, ckCols []*apm.CKCol
	)
	if event, err = s.fkDao.ApmEvent(c, req.EventId); err != nil {
		log.Errorc(c, "ApmEvent error %v", err)
		return
	}
	if event.Databases == "" || event.DistributedTableName == "" {
		err = ecode.Error(ecode.RequestErr, "当前技术埋点未建立数据库/数据表")
		return
	}
	if deCommFields, deExtendFields, err = s.ApmEventDeFieldSql(c, event.Databases, event.DistributedTableName); err != nil {
		log.Errorc(c, "ApmEventDeFieldSql error %v", err)
		return
	}
	for _, deCommField := range deCommFields {
		commCkCol := &apm.CKCol{
			Name: deCommField.Name,
			Type: deCommField.Type,
			Desc: deCommField.Desc,
		}
		commCkCols = append(commCkCols, commCkCol)
	}
	for _, deExtendField := range deExtendFields {
		extendCkCol := &apm.CKCol{
			Name: deExtendField.Name,
			Type: deExtendField.Type,
			Desc: deExtendField.Desc,
		}
		commCkCols = append(commCkCols, extendCkCol)
	}
	ckCols = append(append(ckCols, commCkCols...), extendCkCols...)
	basic := &apm.CKBasicModule{
		DSName:      _ckDatasourceName,
		DBName:      event.Databases,
		TabName:     event.DistributedTableName,
		TabDesc:     req.Description,
		TTLUnit:     _ckTTLUnit,
		TTLDuration: req.TTLDur,
		TTLDataExpr: fmt.Sprintf("toDate(%s)", req.TTLExp),
		Operator:    req.Operator,
	}
	indexCfg := &apm.CKCustomCfg{
		Name:  "index_granularity",
		Value: "8192",
	}
	storageCfg := &apm.CKCustomCfg{
		Name:  "storage_policy",
		Value: "hot_and_cold",
	}
	customCfg := []*apm.CKCustomCfg{indexCfg, storageCfg}
	orderExp := &apm.CKCfgExp{Expression: req.OrderBy}
	partitionExp := &apm.CKCfgExp{Expression: req.PartitionBy}
	cfg := &apm.CKCfgModule{
		CustomSet:   customCfg,
		OrderBy:     []*apm.CKCfgExp{orderExp},
		PartitionBy: []*apm.CKCfgExp{partitionExp},
		ShardingKey: &apm.CKCfgShardingKey{Func: "rand"},
		Engine:      &apm.CKCfgEngine{Type: _ckEngine},
	}
	bus := &apm.CKBusModule{
		BusTag:   &apm.CKBusTag{Items: []string{_ckBusTags}},
		PubLevel: _ckPubLevel,
	}
	content := &apm.CKContentModule{DataLevel: _ckDataLevel}
	module := &apm.CKModelModule{ModeLevel: _ckModelLevel}
	tabData := &apm.CKTableCreateData{
		BasicModule:    basic,
		Cols:           ckCols,
		CfgModule:      cfg,
		BusModule:      bus,
		ContentModule:  content,
		ModelModule:    module,
		DataSourceType: _ckDatasourceType,
	}
	err = s.fkDao.ApmDataCenterCKTableCreate(c, tabData, s.c.Datacenter)
	return
}

// ApmEventDeFieldSql 相同db和table埋点的去重字段
func (s *Service) ApmEventDeFieldSql(c context.Context, dbName, disTabName string) (deCommFields, deExtendFields []*apm.EventFieldSql, err error) {
	var (
		events                     []*apm.Event
		commFields, extendedFields []*apm.EventFieldSql
		commMap                    = make(map[string]*apm.EventFieldSql) // 埋点字段去重 [name]*field
		extendedMap                = make(map[string]*apm.EventFieldSql)
	)
	if events, err = s.fkDao.ApmEventList(c, "", "", "", "", "", dbName, "", disTabName, "", "", 0, 0, 0, 0, 0, 0, 0); err != nil {
		log.Errorc(c, "ApmEventList error %v", err)
		return
	}
	mu := sync.Mutex{}
	group := errgroup.WithContext(c)
	for _, ev := range events {
		var (
			e                = ev
			err1             = err
			common, extended []*apm.EventFieldSql
		)
		group.Go(func(ctx context.Context) error {
			if common, extended, err1 = s.apmEventFieldSqlGenerator(c, e.ID); err != nil {
				log.Errorc(c, "apmEventFieldsGenerator error %v", err1)
				return err1
			}
			mu.Lock()
			commFields = append(commFields, common...)
			extendedFields = append(extendedFields, extended...)
			mu.Unlock()
			return err1
		})
	}
	if err = group.Wait(); err != nil {
		log.Errorc(c, "group.Wait %v", err)
		return
	}
	// 字段去重
	for _, com := range commFields {
		if _, ok := commMap[com.Name]; !ok {
			commMap[com.Name] = com
			deCommFields = append(deCommFields, com)
		}
	}
	sort.Slice(deCommFields, func(i, j int) bool {
		return deCommFields[i].Index < deCommFields[j].Index
	})
	for _, ex := range extendedFields {
		if _, ok := extendedMap[ex.Name]; !ok {
			extendedMap[ex.Name] = ex
			deExtendFields = append(deExtendFields, ex)
		}
	}
	return
}

func (s *Service) ApmEventSampleRateAdd(ctx context.Context, req *apm.AddEventSampleRateReq) (err error) {
	switch req.LogId {
	case apm.LogIdTrackT:
		_, err = s.fkDao.ApmEventSampleRateAppAdd(ctx, req.AppKey, req.EventId, req.EventName, req.LogId, req.Rate)
		if err != nil {
			log.Errorc(ctx, err.Error())
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Add error: %v", err))
			return
		}
	case apm.LogIdPolaris:
		_, err = s.fkDao.ApmEventSampleRateAdd(ctx, req.DatacenterAppId, req.EventId, req.EventName, req.LogId, req.Rate)
		if err != nil {
			log.Errorc(ctx, err.Error())
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Add error: %v", err))
			return
		}
	default:
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("不支持的log_id %s", req.LogId))
	}
	return
}

func (s *Service) ApmEventSampleRateDelete(ctx context.Context, req *apm.DeleteEventSampleRateReq) (err error) {
	if len(req.Items) == 0 {
		return
	}
	logSet := make(map[string]interface{})
	for _, v := range req.Items {
		logSet[v.LogId] = struct{}{}
	}
	if len(logSet) > 1 {
		err = ecode.Error(ecode.RequestErr, "只能批量删除同一个logId的事件")
		return
	}
	switch req.Items[0].LogId {
	case apm.LogIdTrackT:
		_, err = s.fkDao.ApmEventSampleRateAppDelete(ctx, req.Items)
		if err != nil {
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Delete error: %v", err))
			return
		}
	case apm.LogIdPolaris:
		_, err = s.fkDao.ApmEventSampleRateDelete(ctx, req.Items)
		if err != nil {
			err = ecode.Error(ecode.ServerErr, fmt.Sprintf("Delete error: %v", err))
			return
		}
	default:
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("不支持的log_id %s", req.Items[0].LogId))
	}
	return
}

func (s *Service) ApmEventSampleRateList(ctx context.Context, req *apm.EventSampleRateListReq) (resp *apm.EventSampleRateListResp, err error) {
	var (
		sharedRateList []*apm.EventSampleRate
		appRateList    []*apm.EventSampleRateApp
		eventList      []*apm.Event
		app            *appmdl.APP
		logIDs         []string
	)
	if app, err = s.fkDao.AppPass(ctx, req.AppKey); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("appInfo err: %v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	if len(req.LogId) != 0 {
		logIDs = strings.Split(req.LogId, ",")
	}
	// 全局采样率配置
	if sharedRateList, err = s.fkDao.SelectApmEventSampleRate(ctx, app.DataCenterAppID, req.AppKey, req.EventId, logIDs); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("ApmEventSampleRateList err: %v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	// 独立APP采样率配置
	if appRateList, err = s.fkDao.SelectApmEventSampleRateApp(ctx, req.AppKey, req.EventId, logIDs); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("ApmEventSampleRateList err: %v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	// 高峰时间. 注入额外的采样配置
	if app.IsHighestPeak == 1 {
		// 应用关联的埋点数据
		if eventList, err = s.fkDao.ApmEventList(ctx, "", req.AppKey, "", "", "", "", "", "", "", "", 0, 0, 0, 0, 0, 0, 0); err != nil {
			log.Errorc(ctx, "ApmEventList %v", err)
			return
		}
		// 事件注入
		for _, e := range eventList {
			// 默认全量直接跳过
			if e.LowestSampleRate == 1 {
				continue
			}
			var existRate *apm.EventSampleRateApp
			for i := 0; i < len(appRateList); i++ {
				rate := appRateList[i]
				if rate.EventName == e.Name {
					existRate = rate
					break
				}
			}
			if existRate != nil {
				// ...
			} else {
				appRateList = append(appRateList, &apm.EventSampleRateApp{
					AppKey:      req.AppKey,
					EventId:     e.Name,
					SampleRate:  e.LowestSampleRate,
					EventName:   e.Description,
					LogId:       e.LogID,
					IsTemporary: 1,
				})
			}
		}
	}
	resp = convert2Resp(sharedRateList, appRateList)
	return
}

func (s *Service) ApmEventSampleRateConfig(ctx context.Context, req *apm.EventSampleRateConfigReq) (resp *apm.EventSampleRateConfigResp, err error) {
	var (
		listResp *apm.EventSampleRateListResp
		rateMap  = make(map[string]float64)
		eventID  []string
		rateByte []byte
	)
	if listResp, err = s.ApmEventSampleRateList(ctx, &apm.EventSampleRateListReq{AppKey: req.AppKey}); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("list error %v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	for _, v := range listResp.Items {
		rateMap[v.EventId] = v.Rate
		eventID = append(eventID, v.EventId)
	}
	if rateByte, err = json.Marshal(rateMap); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("list error %v", err))
		log.Errorc(ctx, err.Error())
		return
	}
	resp = &apm.EventSampleRateConfigResp{
		EventRates: string(rateByte),
	}
	return
}

func convert2Resp(sharedRateList []*apm.EventSampleRate, appRateList []*apm.EventSampleRateApp) *apm.EventSampleRateListResp {
	var (
		respItems []*apm.EventSampleRateItem
	)
	for _, r := range sharedRateList {
		respItems = append(respItems, &apm.EventSampleRateItem{
			AppKey:          r.AppKey,
			DatacenterAppId: r.DatacenterAppId,
			EventId:         r.EventId,
			EventName:       r.EventName,
			Rate:            r.SampleRate,
			Ctime:           r.Ctime.Unix(),
			Mtime:           r.Mtime.Unix(),
			LogId:           r.LogId,
		})
	}
	for _, r := range appRateList {
		respItems = append(respItems, &apm.EventSampleRateItem{
			AppKey:      r.AppKey,
			EventId:     r.EventId,
			EventName:   r.EventName,
			Rate:        r.SampleRate,
			Ctime:       r.Ctime.Unix(),
			Mtime:       r.Mtime.Unix(),
			LogId:       r.LogId,
			IsTemporary: r.IsTemporary,
		})
	}
	return &apm.EventSampleRateListResp{
		Items: respItems,
	}
}

func (s *Service) ApmEventMonitorNotifyConfig(c context.Context, req *apm.EventMonitorNotifyConfigReq) (resp *apm.EventMonitorNotifyConfig, err error) {
	if resp, err = s.fkDao.ApmEventMonitorNotifyConfig(c, req.EventId, req.AppKey); err != nil {
		log.Errorc(c, "monitor notify config query error %v", err)
	}
	return
}

func (s *Service) ApmEventMonitorNotifyConfigList(c context.Context, req *apm.EventMonitorNotifyConfigListReq) (resp *apm.EventMonitorNotifyConfigListResp, err error) {
	var (
		count   int
		configs []*apm.EventMonitorNotifyConfig
	)
	if count, err = s.fkDao.ApmEventMonitorNotifyConfigCount(c, req.EventId, req.AppKey, req.IsNotify, req.IsMute); err != nil {
		log.Errorc(c, "monitor notify config count query error %v", err)
		return
	}
	if count < 1 {
		return
	}
	if configs, err = s.fkDao.ApmEventMonitorNotifyConfigList(c, req.EventId, req.AppKey, req.IsNotify, req.IsMute, req.Pn, req.Ps); err != nil {
		log.Errorc(c, "monitor notify config list query error %v", err)
		return
	}
	page := &apm.Page{
		Total:    count,
		PageNum:  req.Pn,
		PageSize: req.Ps,
	}
	resp = &apm.EventMonitorNotifyConfigListResp{
		PageInfo: page,
		Items:    configs,
	}
	return
}

func (s *Service) ApmEventMonitorNotifyConfigSet(c context.Context, req *apm.EventMonitorNotifyConfigSetReq) (resp *apm.EventMonitorNotifyConfigSetReq, err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if err = s.fkDao.TxApmEventMonitorNotifyConfigSet(tx, req.EventId, req.AppKey, req.IsNotify, req.IsMute, req.MuteStartTime, req.MuteEndTime, req.Operator); err != nil {
			log.Errorc(c, "monitor notify config set error %v", err)
		}
		return err
	})
	return
}
