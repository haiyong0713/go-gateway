package monitor

import (
	"context"
	"strings"

	xsql "go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) ApmBusList(c context.Context, appKey, filterKey string, ps, pn int) (res *apm.ResultBusList, err error) {
	var (
		total   int
		busList []*apm.Bus
	)
	if total, err = s.fkDao.ApmBusListCount(c, appKey, filterKey); err != nil {
		log.Error("%v", err)
		return
	}
	pageInfo := &apm.Page{Total: total, PageNum: pn, PageSize: ps}
	res = &apm.ResultBusList{PageInfo: pageInfo}
	if total < 1 {
		return
	}
	if busList, err = s.fkDao.ApmBusList(c, appKey, filterKey, ps, pn); err != nil {
		log.Error("%v", err)
		return
	}
	res.Items = busList
	return
}

func (s *Service) ApmBusAdd(c context.Context, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string, shared int) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if _, err := s.fkDao.TxApmBusAdd(tx, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames, shared); err != nil {
			log.Error("%v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmBusDel(c context.Context, busId int64, username string) (err error) {
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
	if _, err = s.fkDao.TxApmBusDel(tx, busId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmBusUpdate(c context.Context, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames string, BusId int64, shared int) (err error) {
	err = s.fkDao.Transact(c, func(tx *xsql.Tx) error {
		if _, err := s.fkDao.TxApmBusUpdate(tx, name, appKeys, description, owner, datacenterBusinessKey, userName, datacenterDwdTableNames, BusId, shared); err != nil {
			log.Error("%v", err)
		}
		return err
	})
	return
}

func (s *Service) ApmCommandGroupAdvancedList(c context.Context, appKey string, eventId, groupId int64) (res []*apm.CommandGroupAdvanced, err error) {
	if res, err = s.fkDao.ApmCommandGroupAdvancedList(c, appKey, eventId, groupId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmCommandGroupAdvancedAdd(c context.Context, appKey, fieldName, title, description, queryType, mapping, operator string, displayType int, eventId, groupId int64) (err error) {
	var (
		tx *xsql.Tx
	)
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
	if _, err = s.fkDao.TxApmCommandGroupAdvancedAdd(tx, appKey, fieldName, title, description, queryType, mapping, operator, displayType, eventId, groupId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmCommandGroupAdvancedDel(c context.Context, appKey string, eventId, groupId, itemId int64) (err error) {
	var (
		tx *xsql.Tx
	)
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
	if _, err = s.fkDao.TxApmCommandGroupAdvancedDel(tx, appKey, eventId, groupId, itemId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmCommandGroupAdvancedUpdate(c context.Context, appKey, title, description, queryType, mapping, operator string, displayType int, eventId, groupId, itemId int64) (err error) {
	var (
		tx *xsql.Tx
	)
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
	if _, err = s.fkDao.TxApmCommandGroupAdvancedUpdate(tx, appKey, title, description, queryType, mapping, operator, displayType, eventId, groupId, itemId); err != nil {
		log.Error("%v", err)
	}
	return
}

// / 通过参数获取Urls
func apmGetCommands(s *Service, c context.Context, appKey string, eventID, busID, commandGroupId int64) (event *apm.Event, urls []string, err error) {
	var (
		commandGroups []*apm.CommandGroup
	)
	// 判断是否为有效event_id
	if event, err = s.fkDao.ApmEvent(c, eventID); err != nil {
		return
	}
	if event == nil {
		return
	}
	// 业务组通配查询
	if busID != 0 && commandGroupId == 0 {
		commandGroups, _ = s.fkDao.ApmCommandGroupByBusID(c, appKey, eventID, busID)
	} else if commandGroupId != 0 {
		commandGroups, _ = s.fkDao.ApmCommandByGroupID(c, appKey, eventID, commandGroupId)
	}
	for _, group := range commandGroups {
		commands, _ := s.fkDao.ApmCommandList(c, appKey, "", eventID, group.ID)
		for _, command := range commands {
			if command == nil {
				continue
			}
			if command.Command == "" {
				continue
			}
			urls = append(urls, command.Command)
		}
	}
	return
}

func (s *Service) ApmMoniCalculate(c context.Context, cType string, matchOption *apm.MatchOption) (res *apm.Moni, err error) {
	var event *apm.Event
	if event, err = s.fkDao.ApmEvent(c, matchOption.EventID); err != nil {
		log.Error("%v", err)
		return
	}
	if event == nil {
		if res, err = s.fkDao.ApmMoniCalculate(c, cType, matchOption); err != nil {
			log.Error("%v", err)
		}
	} else {
		res = &apm.Moni{}
		if res.Value, err = s.fkDao.ApmMoniAggregateCalculate(c, cType, event.Databases, event.DistributedTableName, matchOption); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

func (s *Service) ApmMoniLine(c context.Context, appKey, cType string, eventID, busID, commandGroupId int64, baseMatchOptions *apm.MatchOption) (res []*apm.Moni, err error) {
	var (
		commands []string
		event    *apm.Event
	)
	if event, commands, err = apmGetCommands(s, c, appKey, eventID, busID, commandGroupId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.ApmMoniLine(c, event.Databases, event.DistributedTableName, cType, commands, baseMatchOptions); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmMoniPie(c context.Context, appKey, cType, column string, eventID, busID, commandGroupId int64,
	matchOption *apm.MatchOption) (res []*apm.Moni, err error) {
	var (
		commands []string
		event    *apm.Event
	)
	if event, commands, err = apmGetCommands(s, c, appKey, eventID, busID, commandGroupId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.ApmMoniPie(c, event.Databases, event.DistributedTableName, cType, column, commands, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmMoniNetInfoList(c context.Context, appKey, column string, eventID, busID, commandGroupId int64, matchOption *apm.MatchOption) (res []*apm.NetInfo, err error) {
	var (
		commands []string
		event    *apm.Event
	)
	if event, commands, err = apmGetCommands(s, c, appKey, eventID, busID, commandGroupId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.ApmMoniNetInfoList(c, event.Databases, event.DistributedTableName, column, commands, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmMoniMetricInfoList(c context.Context, column string, matchOption *apm.MatchOption) (res []*apm.MetricInfo, err error) {
	var (
		event *apm.Event
	)
	if event, err = s.fkDao.ApmEvent(c, matchOption.EventID); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.ApmMoniMetricInfoList(c, event.Databases, event.DistributedTableName, column, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmMoniCountInfoList(c context.Context, column string, eventID, busID, subEventGID int64, matchOption *apm.MatchOption) (res []*apm.CountInfo, err error) {
	var (
		commands []string
		event    *apm.Event
	)
	if event, err = s.fkDao.ApmEvent(c, eventID); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if event == nil {
		log.Infoc(c, "apm_event event_id:%v 不存在", eventID)
		return
	}
	if res, err = s.fkDao.ApmMoniCountInfoList(c, event.Databases, event.DistributedTableName, column, commands, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmMoniStatisticsInfoList(c context.Context, appKey, column string, eventID, busID, commandGroupId int64, matchOption *apm.MatchOption) (res []*apm.StatisticsInfo, err error) {
	var (
		commands []string
		event    *apm.Event
	)
	if event, commands, err = apmGetCommands(s, c, appKey, eventID, busID, commandGroupId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.ApmMoniStatisticsInfoList(c, event.Databases, event.DistributedTableName, column, commands, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmCommandGroupList(c context.Context, appKey, filterKey string, eventId, busId int64, ps, pn int) (res *apm.ResultCommandGroupList, err error) {
	var (
		total            int
		commandGroupList []*apm.CommandGroup
	)
	pageInfo := &apm.Page{Total: total, PageNum: pn, PageSize: ps}
	res = &apm.ResultCommandGroupList{PageInfo: pageInfo}
	if total, err = s.fkDao.ApmCommandGroupListCount(c, appKey, filterKey, eventId, busId, ps, pn); err != nil {
		log.Error("%v", err)
		return
	}
	if total < 1 {
		return
	}
	if commandGroupList, err = s.fkDao.ApmCommandGroupList(c, appKey, filterKey, eventId, busId, ps, pn); err != nil {
		log.Error("%v", err)
		return
	}
	for _, commandGroup := range commandGroupList {
		commandGroup.Commands, _ = s.fkDao.ApmCommandList(c, appKey, "", eventId, commandGroup.ID)
	}
	res.PageInfo.Total = total
	res.Items = commandGroupList
	return
}

// 添加事件组
func (s *Service) ApmCommandGroupAdd(c context.Context, appKey, name, urls, description, userName string, busId, eventId int64) (err error) {
	var (
		tx      *xsql.Tx
		groupId int64
	)
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
	if groupId, err = s.fkDao.TxApmCommandGroupAdd(tx, appKey, name, description, userName, busId, eventId); err != nil {
		log.Error("%v", err)
		return
	}

	commands := strings.Split(urls, ",")
	if len(commands) > 0 {
		var (
			sqls []string
			args []interface{}
		)
		for _, command := range commands {
			sqls = append(sqls, "(?,?,?,?)")
			args = append(args, appKey, groupId, command, userName)
		}
		if _, err = s.fkDao.TxApmCommandAdd(tx, sqls, args); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

func (s *Service) ApmCommandGroupDel(c context.Context, appKey string, id, eventId int64) (err error) {
	var (
		tx *xsql.Tx
	)
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
	if _, err = s.fkDao.TxApmCommandGroupDel(tx, appKey, id, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxApmCommandsDel(tx, appKey, id); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmCommandGroupUpdate(c context.Context, appKey, urls, description, userName string, id, eventId int64) (err error) {
	var (
		tx       *xsql.Tx
		commands []*apm.Command
	)
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
	if _, err = s.fkDao.TxApmCommandGroupUpdate(tx, appKey, description, userName, id, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	newUrls := strings.Split(urls, ",")
	if commands, err = s.fkDao.ApmCommandList(c, appKey, "", eventId, id); err != nil {
		log.Error("%v", err)
		return
	}
	// 删除
	var as []string
	var ds []*apm.Command
	for _, command := range commands {
		var isDelete = true
		for _, url := range newUrls {
			if command.Command == url && command.AppKey == appKey && command.GroupId == id {
				isDelete = false
				break
			}
		}
		if isDelete {
			ds = append(ds, command)
		}
	}
	for _, url := range newUrls {
		var isAppend = true
		for _, command := range commands {
			if command.Command == url && command.AppKey == appKey && command.GroupId == id {
				isAppend = false
				break
			}
		}
		if isAppend {
			as = append(as, url)
		}
	}
	if len(as) > 0 {
		var (
			sqls []string
			args []interface{}
		)
		for _, url := range as {
			sqls = append(sqls, "(?,?,?,?)")
			args = append(args, appKey, id, url, userName)
		}
		if _, err = s.fkDao.TxApmCommandAdd(tx, sqls, args); err != nil {
			log.Error("%v", err)
		}
	}
	if len(ds) > 0 {
		for _, d := range ds {
			if _, err = s.fkDao.TxApmCommandDel(tx, d.AppKey, d.ID); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
	return
}

func (s *Service) ApmCommandList(c context.Context, appKey, filterKey string, eventId, groupId int64) (res []*apm.Command, err error) {
	if res, err = s.fkDao.ApmCommandList(c, appKey, filterKey, eventId, groupId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmAggregateNetList(c context.Context, appKey, command, queryType string, startTime, endTime int64) (res []*apm.AggregateNetInfo, err error) {
	if res, err = s.fkDao.ApmAggregateNetList(c, appKey, command, queryType, startTime, endTime); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmAggregateCrashList(c context.Context, appKey, versionCode, queryType string, isAllVersion, dataType int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	if res, err = s.fkDao.ApmAggregateCrashList(c, appKey, versionCode, queryType, isAllVersion, dataType, startTime, endTime); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmAggregateANRList(c context.Context, appKey, versionCode, queryType string, isAllVersion, dataType int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	if res, err = s.fkDao.ApmAggregateANRList(c, appKey, versionCode, queryType, isAllVersion, dataType, startTime, endTime); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmAggregateSetupList(c context.Context, appKey, versionCode, queryType string, isAllVersion int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	if res, err = s.fkDao.ApmAggregateSetupList(c, appKey, versionCode, queryType, isAllVersion, startTime, endTime); err != nil {
		log.Error("%v", err)
	}
	return
}
func (s *Service) ApmFlowmapRouteList(c context.Context, matchOption *apm.MatchOption) (res []*apm.FlowmapRoute, err error) {
	if res, err = s.fkDao.ApmFlowmapRouteList(c, matchOption); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmFlowmapRouteAliasList(c context.Context, appKey, filterKey string, busID int64) (res []*apm.FlowmapRouteAlias, err error) {
	if res, err = s.fkDao.ApmFlowmapRouteAliasList(c, appKey, filterKey, busID); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmFlowmapRouteAliasAdd(c context.Context, appKey, routeName, routeAlias, userName string, busID int64) (err error) {
	var (
		tx *xsql.Tx
	)
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("ApmFlowmapRouteAliasAdd->s.fkDao.BeginTran() error(%v)", err)
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

	if _, err = s.fkDao.TxApmFlowMapRouteAliasAdd(tx, appKey, routeName, routeAlias, userName, busID); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmFlowmapRouteAliasUpdate(c context.Context, id int64, routeName, routeAlias, userName string, busID int64) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("ApmFlowmapRouteAliasUpdate->s.fkDao.BeginTran() error(%v)", err)
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
	if _, err = s.fkDao.TxApmFlowMapRouteAliasUpdate(tx, id, routeName, routeAlias, userName, busID); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmFlowmapRouteAliasDel(c context.Context, id int64, userName string) (err error) {
	var tx *xsql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("ApmFlowmapRouteAliasDel->s.fkDao.BeginTran() error(%v)", err)
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
	if _, err = s.fkDao.TxApmFlowMapRouteAliasDel(tx, id, userName); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmWebTrack(c context.Context, trackParams *apm.WebTrackParams) (err error) {
	var (
		pvModels    []*apm.WebTrackModel
		errorModels []*apm.WebTrackModel
	)
	for _, model := range trackParams.Models {
		switch model.EventId {
		case apm.CLICKHOUSE_WEB_TRACK_PV:
			pvModels = append(pvModels, model)
		case apm.CLICKHOUSE_WEB_TRACK_ERROR:
			errorModels = append(errorModels, model)
		}
	}
	if len(pvModels) > 0 {
		err = s.fkDao.AddWebTracePv(c, pvModels)
	}
	if len(errorModels) > 0 {
		err = s.fkDao.AddWebTraceError(c, errorModels)
	}
	return
}

func (s *Service) ApmEventSetting(c context.Context, appKey string, eventId int64) (res []*apm.ApmEventSetting, err error) {
	if res, err = s.fkDao.ApmEventSetting(c, appKey, eventId); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) ApmDetailSetup(c context.Context, matchOption *apm.MatchOption) (res []*apm.ApmDetailSetup, err error) {
	if res, err = s.fkDao.ApmDetailSetup(c, matchOption); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}

func (s *Service) ApmCrashRule(c context.Context, id int64) (res *apm.CrashRule, err error) {
	if res, err = s.fkDao.ApmCrashRule(c, id); err != nil {
		log.Error("ApmCrashRule s.fkDao.ApmCrashRule error(%v)", err)
	}
	return
}

func (s *Service) ApmCrashRuleList(c context.Context, req *apm.CrashRuleReq) (res *apm.CrashRuleRes, err error) {
	var (
		total      int
		crashRules []*apm.CrashRule
	)
	if total, err = s.fkDao.ApmCrashRuleListCount(c, req.KeyWords, req.PageKeyWords, req.AppKeys, req.BusID); err != nil {
		log.Error("ApmCrashRuleList s.fkDao.ApmCrashListCount error(%v)", err)
		return
	}
	if total == 0 {
		return
	}
	pageInfo := &apm.Page{
		Total:    total,
		PageNum:  req.Pn,
		PageSize: req.Ps,
	}
	res = &apm.CrashRuleRes{PageInfo: pageInfo}
	if crashRules, err = s.fkDao.ApmCrashRuleList(c, req.KeyWords, req.PageKeyWords, req.AppKeys, req.BusID, req.Pn, req.Ps); err != nil {
		log.Error("ApmCrashRuleList s.fkDao.ApmCrashRuleList error(%v)", err)
		return
	}
	res.Items = crashRules
	return
}

func (s *Service) ApmCrashRuleAdd(c context.Context, req *apm.CrashRuleReq) (err error) {
	if err = s.fkDao.ApmCrashRuleAdd(c, req.AppKeys, req.RuleName, req.KeyWords, req.PageKeyWords, req.Operator, req.Description, req.BusID); err != nil {
		log.Error("ApmCrashRuleAdd s.fkDao.ApmCrashRuleAdd error(%v)", err)
	}
	return
}

func (s *Service) ApmCrashRuleDel(c context.Context, req *apm.CrashRuleReq) (err error) {
	if err = s.fkDao.ApmCrashRuleDel(c, req.ID); err != nil {
		log.Error("ApmCrashRuleDel s.fkDao.ApmCrashRuleDel error(%v)", err)
	}
	return
}

func (s *Service) ApmCrashRuleUpdate(c context.Context, req *apm.CrashRuleReq) (err error) {
	if err = s.fkDao.ApmCrashRuleUpdate(c, req.AppKeys, req.RuleName, req.KeyWords, req.PageKeyWords, req.Operator, req.Description, req.BusID, req.ID); err != nil {
		log.Error("ApmCrashRuleUpdate s.fkDao.ApmCrashRuleUpdate error(%v)", err)
	}
	return
}
