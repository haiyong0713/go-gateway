package bugly

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	buglymdl "go-gateway/app/app-svr/fawkes/service/model/bugly"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var statusMap = map[int64]string{
	0: "未解决",
	1: "已解决",
	2: "无需处理",
	3: "正在解决",
}

func (s *Service) CrashIndexList(c context.Context, column string, eventId int64, matchOption *apm.MatchOption) (res *buglymdl.CrashIndexRes, err error) {
	var (
		event                 *apm.Event
		eventVedaConfig       *apm.EventVedaConfig
		apmCountInfoList      []*apm.CountInfo
		crashIndexMessageList []*buglymdl.CrashIndex
		commands              []string
		total                 int64
	)

	// 根据eventId获取database和table
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if eventVedaConfig, err = s.fkDao.ApmEventVedaConfig(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	// 查询不同hash总数量
	if total, err = s.fkDao.ApmMoniDistinctCount(c, buglymdl.ErrorStackHashWithoutUseless, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	pageInfo := &buglymdl.PageInfo{Total: int(total), Ps: matchOption.Ps, Pn: matchOption.Pn}
	res = &buglymdl.CrashIndexRes{PageInfo: pageInfo}
	// 查询clickhouse数量
	if apmCountInfoList, err = s.fkDao.ApmMoniCountInfoList(c, event.Databases, event.DistributedTableName, buglymdl.ErrorStackHashWithoutUseless, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	if len(apmCountInfoList) == 0 {
		return
	}
	// 补全错误类型错误信息
	var stackHashList []string
	for _, row := range apmCountInfoList {
		stackHashList = append(stackHashList, row.Command)
	}
	if crashIndexMessageList, err = s.fkDao.CrashIndexMessageByHashList(c, matchOption.AppKey, eventVedaConfig.VedaIndexTable, eventVedaConfig.VedaStackTable, stackHashList); err != nil {
		log.Error("%v", err)
		return
	}
	for _, row := range apmCountInfoList {
		for _, messageInfo := range crashIndexMessageList {
			if row.Command == messageInfo.ErrorStackHashWithoutUseless {
				crashIndex := &buglymdl.CrashIndex{
					AppKey:                       matchOption.AppKey,
					ErrorStackHashWithoutUseless: messageInfo.ErrorStackHashWithoutUseless,
					ErrorStackBeforeHash:         messageInfo.ErrorStackBeforeHash,
					AnalyseErrorStack:            messageInfo.AnalyseErrorStack,
					ErrorType:                    messageInfo.ErrorType,
					ErrorMsg:                     messageInfo.ErrorMsg,
					HappenTime:                   messageInfo.HappenTime,
					Count:                        row.Count,
					DistinctBuvidCount:           row.BuvidDistinctCount,
					AssignOperator:               messageInfo.AssignOperator,
					SolveStatus:                  messageInfo.SolveStatus,
					SolveOperator:                messageInfo.SolveOperator,
					SolveVersionCode:             messageInfo.SolveVersionCode,
					SolveDescription:             messageInfo.SolveDescription,
					HappenNewestVersionCode:      messageInfo.HappenNewestVersionCode,
					HappenOldestVersionCode:      messageInfo.HappenOldestVersionCode,
					CTime:                        messageInfo.CTime,
					MTime:                        messageInfo.MTime,
				}
				res.Items = append(res.Items, crashIndex)
				break
			}
		}
	}
	return
}

func (s *Service) CrashInfoList(c context.Context, isLaser, eventId int64, matchOption *apm.MatchOption) (res *buglymdl.CrashRes, err error) {
	var (
		lasers      []*appmdl.Laser
		event       *apm.Event
		crashInfo   []*buglymdl.CrashInfo
		pages       *buglymdl.PageInfo
		deCrashInfo = make(map[string]*buglymdl.CrashInfo)
	)
	if lasers, err = s.fkDao.LaserWithCrash(c, matchOption.ErrorStackHashWithoutUseless); err != nil {
		log.Error("CrashInfoList s.fkDao.LaserWithCrash error(%v)", err)
		return
	}
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	res = &buglymdl.CrashRes{}
	if isLaser != 0 {
		if len(lasers) == 0 {
			return
		}
		matchOption.Pn = -1
		matchOption.Ps = -1
		// 查询时间范围[laser最新时间-3天，laser最新时间]
		endTime := lasers[0].CTime
		endTimeFormat := time.Unix(endTime, 0)
		startTime := endTimeFormat.AddDate(0, 0, -3).Unix()
		matchOption.StartTime = startTime * apm.SecToMilliUnit
		matchOption.EndTime = endTime * apm.SecToMilliUnit
		var buvids []string
		for _, laser := range lasers {
			buvids = append(buvids, laser.Buvid)
		}
		buvidStr := strings.Join(buvids, ",")
		filter := &apm.Filter{AndType: "AND", Column: "buvid", EqualType: "=", Values: buvidStr, ValueType: "STRING"}
		matchOption.Filters = append(matchOption.Filters, filter)
	}
	if crashInfo, err = s.fkDao.CrashInfoList(c, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("CrashInfoList s.fkDao.CrashInfoList error(%v)", err)
		return
	}
	if isLaser != 0 {
		var deRes []*buglymdl.CrashInfo
		// 异常问题去重并选择最新数据
		for _, info := range crashInfo {
			data, ok := deCrashInfo[info.Buvid]
			if ok {
				if info.TimeISO > data.TimeISO {
					deCrashInfo[info.Buvid] = info
				}
			} else {
				deCrashInfo[info.Buvid] = info
			}
		}
		for _, vCrash := range deCrashInfo {
			deRes = append(deRes, vCrash)
		}
		sort.Slice(deRes, func(i, j int) bool {
			return deRes[i].TimeISO > deRes[j].TimeISO
		})
		res.Items = deRes
		pages = &buglymdl.PageInfo{Total: len(deRes), Pn: 1, Ps: len(deRes)}
	} else {
		res.Items = crashInfo
		var (
			total int64
		)
		if total, err = s.fkDao.ApmMoniCount(c, event.Databases, event.DistributedTableName, matchOption); err != nil {
			log.Error("%v", err)
			return
		}
		pages = &buglymdl.PageInfo{Total: int(total), Pn: matchOption.Pn, Ps: matchOption.Ps}
	}
	// 异常问题与laser进行匹配
	for _, item := range res.Items {
		for _, laser := range lasers {
			if item.Buvid == laser.Buvid {
				item.Laser = laser
				break
			}
		}
	}
	res.PageInfo = pages
	return
}

func (s *Service) JankIndexList(c context.Context, column string, eventId int64, matchOption *apm.MatchOption) (res *buglymdl.JankIndexRes, err error) {
	var (
		event                *apm.Event
		JankIndexList        []*buglymdl.JankIndex
		jankIndexMessageList []*buglymdl.JankIndex
		commands             []string
		total                int64
	)
	// 根据eventId获取database和table
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if total, err = s.fkDao.ApmMoniDistinctCount(c, column, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	pageInfo := &buglymdl.PageInfo{Total: int(total), Ps: matchOption.Ps, Pn: matchOption.Pn}
	res = &buglymdl.JankIndexRes{PageInfo: pageInfo}
	// 查询clickhouse数量
	if JankIndexList, err = s.fkDao.ApmMoniJankIndexCountInfoList(c, event.Databases, event.DistributedTableName, column, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	if len(JankIndexList) == 0 {
		return
	}
	// 补全错误类型错误信息
	var stackHashList []string
	for _, row := range JankIndexList {
		stackHashList = append(stackHashList, row.Command)
	}
	if jankIndexMessageList, err = s.fkDao.JankIndexMessageByHashList(c, matchOption.AppKey, stackHashList); err != nil {
		log.Error("%v", err)
		return
	}
	for _, row := range JankIndexList {
		for _, messageInfo := range jankIndexMessageList {
			if row.Command == messageInfo.AnalyseJankStackHash {
				jankIndex := &buglymdl.JankIndex{
					Command:                 row.Command,
					AppKey:                  matchOption.AppKey,
					HappenTime:              messageInfo.HappenTime,
					AnalyseJankStackHash:    messageInfo.AnalyseJankStackHash,
					DurationQuantile80:      row.DurationQuantile80,
					AnalyseJankStack:        messageInfo.AnalyseJankStack,
					Count:                   row.Count,
					BuvidDistinctCount:      row.BuvidDistinctCount,
					SolveStatus:             messageInfo.SolveStatus,
					SolveOperator:           messageInfo.SolveOperator,
					SolveVersionCode:        messageInfo.SolveVersionCode,
					SolveDescription:        messageInfo.SolveDescription,
					HappenNewestVersionCode: messageInfo.HappenNewestVersionCode,
					HappenOldestVersionCode: messageInfo.HappenOldestVersionCode,
					CTime:                   messageInfo.CTime,
					MTime:                   messageInfo.MTime,
				}
				res.Items = append(res.Items, jankIndex)
				break
			}
		}
	}
	return
}

func (s *Service) OOMIndexList(c context.Context, column string, eventId int64, matchOption *apm.MatchOption) (res *buglymdl.OOMIndexRes, err error) {
	var (
		event               *apm.Event
		OOMIndexList        []*buglymdl.OOMIndex
		OOMIndexMessageList []*buglymdl.OOMIndex
		commands            []string
		total               int64
	)
	// 根据eventId获取database和table
	if event, err = s.fkDao.ApmEvent(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if event == nil {
		log.Errorc(c, "event is nil, eventId = %v", eventId)
		return
	}
	// 查询不同hash总数量
	if total, err = s.fkDao.ApmMoniDistinctCount(c, column, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	pageInfo := &buglymdl.PageInfo{Total: int(total), Ps: matchOption.Ps, Pn: matchOption.Pn}
	res = &buglymdl.OOMIndexRes{PageInfo: pageInfo}
	// 查询clickhouse数量
	if OOMIndexList, err = s.fkDao.ApmMoniOOMIndexCountInfoList(c, event.Databases, event.DistributedTableName, column, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	if len(OOMIndexList) == 0 {
		return
	}
	// 补全错误类型错误信息
	var stackHashList []string
	for _, row := range OOMIndexList {
		stackHashList = append(stackHashList, row.Command)
	}
	if OOMIndexMessageList, err = s.fkDao.OOMIndexMessageByHashList(c, matchOption.AppKey, stackHashList); err != nil {
		log.Error("%v", err)
		return
	}
	for _, row := range OOMIndexList {
		for _, messageInfo := range OOMIndexMessageList {
			if row.Command == messageInfo.Hash {
				oomIndex := &buglymdl.OOMIndex{
					Command:                 row.Command,
					AppKey:                  matchOption.AppKey,
					HappenTime:              messageInfo.HappenTime,
					Hash:                    messageInfo.Hash,
					AnalyseStack:            messageInfo.AnalyseStack,
					RetainedSizeQuantile80:  row.RetainedSizeQuantile80,
					LeakReason:              messageInfo.LeakReason,
					GcRoot:                  messageInfo.GcRoot,
					Count:                   row.Count,
					BuvidDistinctCount:      row.BuvidDistinctCount,
					SolveStatus:             messageInfo.SolveStatus,
					SolveOperator:           messageInfo.SolveOperator,
					SolveVersionCode:        messageInfo.SolveVersionCode,
					SolveDescription:        messageInfo.SolveDescription,
					HappenNewestVersionCode: messageInfo.HappenNewestVersionCode,
					HappenOldestVersionCode: messageInfo.HappenOldestVersionCode,
					CTime:                   messageInfo.CTime,
					MTime:                   messageInfo.MTime,
				}
				res.Items = append(res.Items, oomIndex)
				break
			}
		}
	}
	return
}

func (s *Service) OOMInfoList(c context.Context, matchOption *apm.MatchOption) (res *buglymdl.OOMInfoRes, err error) {
	var (
		event   *apm.Event
		total   int64
		oomInfo []*buglymdl.OOMInfo
	)
	if event, err = s.fkDao.ApmEvent(c, matchOption.EventID); err != nil {
		log.Error("%v", err)
		return
	}
	if total, err = s.fkDao.ApmMoniCount(c, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	if oomInfo, err = s.fkDao.OOMInfoList(c, matchOption); err != nil {
		log.Error("%v", err)
	}
	page := &buglymdl.PageInfo{Total: int(total), Pn: matchOption.Pn, Ps: matchOption.Ps}
	res = &buglymdl.OOMInfoRes{Items: oomInfo, PageInfo: page}
	return
}

func (s *Service) JankInfoList(c context.Context, matchOption *apm.MatchOption) (res *buglymdl.JankInfoRes, err error) {
	var (
		event    *apm.Event
		total    int64
		jankInfo []*buglymdl.JankInfo
	)
	if event, err = s.fkDao.ApmEvent(c, matchOption.EventID); err != nil {
		log.Error("%v", err)
		return
	}
	if total, err = s.fkDao.ApmMoniCount(c, event.Databases, event.DistributedTableName, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	if jankInfo, err = s.fkDao.JankInfoList(c, matchOption); err != nil {
		log.Error("%v", err)
	}
	page := &buglymdl.PageInfo{Total: int(total), Pn: matchOption.Pn, Ps: matchOption.Ps}
	res = &buglymdl.JankInfoRes{Items: jankInfo, PageInfo: page}
	return
}

func (s *Service) UpdateIndex(c context.Context, hash, appKey, assignOperator, solveOperator, operator, solveDescription string, solveVersionCode, solveStatus, eventId int64, wxNotify bool) (err error) {
	var (
		eventVedaConfig *apm.EventVedaConfig
		indexStatus     *buglymdl.IndexStatus
	)
	if eventVedaConfig, err = s.fkDao.ApmEventVedaConfig(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if indexStatus, err = s.fkDao.IndexStatusByHash(c, eventVedaConfig.HashColumn, eventVedaConfig.VedaIndexTable, hash, appKey); err != nil {
		log.Error("IndexStatusByHash %v", err)
		return
	}
	err = s.fkDao.VedaTransact(c, func(tx *sql.Tx) (txError error) {
		if _, txError = s.fkDao.TxIndexUpdate(tx, hash, appKey, assignOperator, solveOperator, solveDescription, eventVedaConfig.VedaIndexTable, eventVedaConfig.HashColumn, solveVersionCode, solveStatus); err != nil {
			log.Error("%v", err)
			return
		}
		return
	})
	if err != nil {
		log.Error("TxCrashIndexUpdate %v", err)
		return
	}
	logText := getLogText(indexStatus, assignOperator, solveOperator, operator, solveDescription, solveVersionCode, solveStatus)
	if logText != "" {
		if err = s.fkDao.CrashLogAdd(c, hash, appKey, operator, logText); err != nil {
			log.Error("s.fkDao.CrashLogAdd %v", err)
			return
		}
	}
	if wxNotify {
		var (
			crashIndexList     []*buglymdl.CrashIndex
			hashList, userList []string
			crashIndex         *buglymdl.CrashIndex
		)
		hashList = append(hashList, hash)
		if crashIndexList, err = s.fkDao.CrashIndexMessageByHashList(c, appKey, eventVedaConfig.VedaIndexTable, eventVedaConfig.VedaStackTable, hashList); err != nil {
			log.Error("%v", err)
			return
		}
		if len(crashIndexList) > 0 {
			crashIndex = crashIndexList[0]
		} else {
			log.Error("crashIndexList error")
			return
		}
		if assignOperator != "" {
			userList = append(userList, assignOperator)
		}

		if solveOperator != "" {
			userList = append(userList, solveOperator)
		}
		_ = s.fkDao.WechatCardMessageNotify(
			"崩溃通知",
			fmt.Sprintf("应用：%s\n解决状态：%s\n指派人：%s\n指定处理人：%s\n影响版本为：%d-%d\n错误类型：%s\n错误信息：%s\n异常堆栈：%s\n", appKey, statusMap[solveStatus], assignOperator, solveOperator, crashIndex.HappenOldestVersionCode, crashIndex.HappenNewestVersionCode, crashIndex.ErrorType, crashIndex.ErrorMsg, crashIndex.AnalyseErrorStack),
			fmt.Sprintf("https://fawkes.bilibili.co/#/apm/crash/analyse-crash-trend-of-visit?error_stack_hash_without_useless=%s&app_key=%s&event_id=9&start_time=%d&end_time=%d", hash, appKey, crashIndex.HappenTime*1000-time.Hour.Milliseconds()*24, crashIndex.HappenTime*1000+time.Hour.Milliseconds()),
			"",
			strings.Join(userList, "|"),
			s.c.Comet.FawkesAppID)
	}
	return
}

func getLogText(preStatus *buglymdl.IndexStatus, assignOperator, solveOperator, operator, solveDescription string, solveVersionCode, solveStatus int64) (res string) {
	if preStatus.SolveStatus != solveStatus {
		res += fmt.Sprintf("解决状态变更为【%s】 ", statusMap[solveStatus])
	}
	if preStatus.AssignOperator != assignOperator {
		res += fmt.Sprintf("指派人变更为【%s】 ", assignOperator)
	}
	if preStatus.SolveOperator != solveOperator {
		res += fmt.Sprintf("处理人变更为【%s】 ", solveOperator)
	}
	if preStatus.SolveVersionCode != solveVersionCode {
		if solveVersionCode != 0 {
			res += fmt.Sprintf("解决版本变更为【%d】 ", solveVersionCode)
		} else {
			res += "解决版本变更为【空】 "
		}
	}
	if preStatus.SolveDescription != solveDescription && solveDescription != "" {
		res += fmt.Sprintf("备注【%s】,", statusMap[solveStatus])
	}
	if res != "" {
		res += fmt.Sprintf("操作人%s", operator)
	}
	return
}

func (s *Service) UpdateCrashIndex(c context.Context, errorStackHash, appKey, solveOperator, solveDescription string, solveVersionCode, solveStatus int, eventId int64) (err error) {
	var eventVedaConfig *apm.EventVedaConfig
	if eventVedaConfig, err = s.fkDao.ApmEventVedaConfig(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	err = s.fkDao.VedaTransact(c, func(tx *sql.Tx) (txError error) {
		if _, txError = s.fkDao.TxCrashIndexUpdate(tx, errorStackHash, appKey, solveOperator, solveDescription, eventVedaConfig.VedaIndexTable, solveVersionCode, solveStatus); err != nil {
			log.Error("%v", err)
			return
		}
		return
	})
	if err != nil {
		log.Error("TxCrashIndexUpdate %v", err)
		return
	}
	return
}

func (s *Service) UpdateJankIndex(c context.Context, analyseJankStackHash, appKey, solveOperator, solveDescription string, solveVersionCode, solveStatus int) (err error) {
	err = s.fkDao.VedaTransact(c, func(tx *sql.Tx) (txError error) {
		if _, txError = s.fkDao.TxJankIndexUpdate(tx, analyseJankStackHash, appKey, solveOperator, solveDescription, solveVersionCode, solveStatus); err != nil {
			log.Error("%v", err)
			return
		}
		return
	})
	if err != nil {
		log.Error("TxCrashIndexUpdate: %v", err)
		return
	}
	return
}

func (s *Service) UpdateOOMIndex(c context.Context, hash, appKey, solveOperator, solveDescription string, solveVersionCode, solveStatus int) (err error) {
	err = s.fkDao.VedaTransact(c, func(tx *sql.Tx) (txError error) {
		if _, txError = s.fkDao.TxOOMIndexUpdate(tx, hash, appKey, solveOperator, solveDescription, solveVersionCode, solveStatus); err != nil {
			log.Error("%v", err)
			return
		}
		return
	})
	if err != nil {
		log.Error("TxOOMIndexUpdate: %v", err)
		return
	}
	return
}

func (s *Service) JankIndexMessageByHashList(c context.Context, appKey string, stackHashList []string) (res []*buglymdl.JankIndex, err error) {
	if res, err = s.fkDao.JankIndexMessageByHashList(c, appKey, stackHashList); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) CrashIndexMessageByHashList(c context.Context, appKey string, eventId int64, stackHashList []string) (res []*buglymdl.CrashIndex, err error) {
	var eventVedaConfig *apm.EventVedaConfig
	if eventVedaConfig, err = s.fkDao.ApmEventVedaConfig(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.CrashIndexMessageByHashList(c, appKey, eventVedaConfig.VedaIndexTable, eventVedaConfig.VedaStackTable, stackHashList); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

func (s *Service) CrashLaserRelationAdd(c context.Context, laserId int64, errorStackHashWithoutUseless, operator string) (err error) {
	if err = s.fkDao.CrashLaserRelationAdd(c, laserId, errorStackHashWithoutUseless, operator); err != nil {
		log.Error("CrashLaserRelationAdd s.fkDao.CrashLaserRelationAdd error(%v)", err)
	}
	return
}

func (s *Service) SolveStatus(c context.Context, eventId int64, hash, appKey string) (res *buglymdl.IndexStatus, err error) {
	var (
		eventVedaConfig *apm.EventVedaConfig
	)
	if eventVedaConfig, err = s.fkDao.ApmEventVedaConfig(c, eventId); err != nil {
		log.Error("%v", err)
		return
	}
	if res, err = s.fkDao.IndexStatusByHash(c, eventVedaConfig.HashColumn, eventVedaConfig.VedaIndexTable, hash, appKey); err != nil {
		log.Error("IndexStatusByHash %v", err)
		return
	}
	return
}

func (s *Service) LogList(c context.Context, hash, appKey string) (res []*buglymdl.LogText, err error) {
	if res, err = s.fkDao.LogList(c, hash, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	return
}
