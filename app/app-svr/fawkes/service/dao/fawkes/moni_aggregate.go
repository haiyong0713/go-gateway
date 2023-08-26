package fawkes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_getNetList   = `SELECT toUnixTimestamp(timestamp), command, app_key, count, error_rate, total_time_quantile_80, avg_req_size, avg_recv_size FROM bilibili_mobile_monitor_aggregate.ads_infra_url_net_info_interval_rt WHERE app_key=? AND command=? %s`
	_getSetupList = `SELECT toUnixTimestamp(timestamp) AS timestamp,app_key,version_code,count,distinct_buvid_count FROM bilibili_mobile_monitor_aggregate.ads_infra_setup_%s_rt WHERE app_key=? %s`

	_getCrashList = `SELECT toUnixTimestamp(timestamp) AS timestamp,app_key,version_code,count,distinct_buvid_count FROM bilibili_mobile_monitor_aggregate.ads_infra_crash_%s_rt WHERE app_key=? %s`
	_getANRList   = `SELECT toUnixTimestamp(timestamp) AS timestamp,app_key,version_code,count,distinct_buvid_count FROM bilibili_mobile_monitor_aggregate.ads_infra_anr_%s_rt WHERE app_key=? %s`

	_getAlertCrashInfo = `SELECT toUnixTimestamp(timestamp),app_key,version_code,count,distinct_buvid_count FROM bilibili_mobile_monitor_aggregate.ads_infra_crash_interval_rt WHERE app_key=? AND version_code=? AND interval_time = 5 AND timestamp <= ? ORDER BY timestamp desc LIMIT 1`
	_getAlertSetupInfo = `SELECT toUnixTimestamp(timestamp),app_key,version_code,count,distinct_buvid_count FROM bilibili_mobile_monitor_aggregate.ads_infra_setup_interval_rt WHERE app_key=? AND version_code=? AND interval_time = 5 AND timestamp <= ? ORDER BY timestamp desc LIMIT 1`

	_getTFPackUser = `SELECT count(DISTINCT buvid) FROM bilibili_mobile_monitor_aggregate.ads_infra_testflight_setup_interval_rt WHERE app_key=? AND version_code=? AND timestamp >= toDateTime(?)`
)

// 解析 queryType
func parseQueryType(queryType string) (queryKey string, queryOption int, err error) {
	queryComps := strings.Split(queryType, "@")
	queryKey = queryComps[0]
	if len(queryComps) > 1 {
		if queryOption, err = strconv.Atoi(queryComps[1]); err != nil {
			log.Error("parseQueryType error %v", err)
		}
	}
	return
}

func (d *Dao) ApmAggregateNetList(c context.Context, appKey, command, queryType string, startTime, endTime int64) (res []*apm.AggregateNetInfo, err error) {
	var (
		queryKey, sqlAdd string
		queryOption      int
		sqls             []string
		args             []interface{}
	)
	if queryKey, queryOption, err = parseQueryType(queryType); err != nil {
		log.Error("ApmAggregateCrashList error %v", err)
		return
	}
	args = append(args, appKey)
	args = append(args, command)
	if queryKey == "interval" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND interval_time=? ")
		}
	}
	args = append(args, startTime)
	args = append(args, endTime)
	sqls = append(sqls, " AND timestamp >= ? ")
	sqls = append(sqls, " AND timestamp <= ? ")
	sqlAdd = strings.Join(sqls, " ")
	rows, err := d.clickhouse.Query(c, fmt.Sprintf(_getNetList, sqlAdd), args...)
	if err != nil {
		log.Error("ApmAggregateNetList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateNetInfo{}
		if err = rows.Scan(&re.Timestamp, &re.Command, &re.AppKey, &re.Count, &re.ErrorRate, &re.TotalTimeQuantile80, &re.AvgReqSize, &re.AvgRecvSize); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAggregateCrashList(c context.Context, appKey, versionCode, queryType string, isAllVersion, dataType int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	var (
		queryKey, sqlAdd string
		queryOption      int
		sqls             []string
		args             []interface{}
	)
	if queryKey, queryOption, err = parseQueryType(queryType); err != nil {
		log.Error("ApmAggregateCrashList error %v", err)
		return
	}
	args = append(args, appKey)
	if queryKey == "addson" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND is_daily=? ")
		}
	} else if queryKey == "interval" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND interval_time=? ")
		}
	}
	if versionCode != "" {
		sqls = append(sqls, fmt.Sprintf(" AND version_code in (%s) ", versionCode))
	}
	args = append(args, isAllVersion)
	args = append(args, dataType)
	args = append(args, startTime)
	args = append(args, endTime)
	sqls = append(sqls, " AND is_all_version = ? ")
	sqls = append(sqls, " AND type = ? ")
	sqls = append(sqls, " AND timestamp >= ? ")
	sqls = append(sqls, " AND timestamp <= ? ")
	sqlAdd = strings.Join(sqls, " ")
	rows, err := d.clickhouse.Query(c, fmt.Sprintf(_getCrashList, queryKey, sqlAdd), args...)
	if err != nil {
		log.Error("ApmAggregateCrashList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateCountItem{}
		if err = rows.Scan(&re.Timestamp, &re.AppKey, &re.VersionCode, &re.Count, &re.DistinctBuvidCount); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAggregateANRList(c context.Context, appKey, versionCode, queryType string, isAllVersion, dataType int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	var (
		queryKey, sqlAdd string
		queryOption      int
		sqls             []string
		args             []interface{}
	)
	if queryKey, queryOption, err = parseQueryType(queryType); err != nil {
		log.Error("ApmAggregateANRList error %v", err)
		return
	}
	args = append(args, appKey)
	if queryKey == "addson" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND is_daily=? ")
		}
	} else if queryKey == "interval" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND interval_time=? ")
		}
	}
	if versionCode != "" {
		sqls = append(sqls, fmt.Sprintf(" AND version_code in (%s) ", versionCode))
	}
	args = append(args, isAllVersion)
	args = append(args, dataType)
	args = append(args, startTime)
	args = append(args, endTime)
	sqls = append(sqls, " AND is_all_version = ? ")
	sqls = append(sqls, " AND type = ? ")
	sqls = append(sqls, " AND timestamp >= ? ")
	sqls = append(sqls, " AND timestamp <= ? ")
	sqlAdd = strings.Join(sqls, " ")
	rows, err := d.clickhouse.Query(c, fmt.Sprintf(_getANRList, queryKey, sqlAdd), args...)
	if err != nil {
		log.Error("ApmAggregateANRList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateCountItem{}
		if err = rows.Scan(&re.Timestamp, &re.AppKey, &re.VersionCode, &re.Count, &re.DistinctBuvidCount); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAggregateSetupList(c context.Context, appKey string, versionCode, queryType string, isAllVersion int, startTime, endTime int64) (res []*apm.AggregateCountItem, err error) {
	var (
		queryKey, sqlAdd string
		queryOption      int
		sqls             []string
		args             []interface{}
	)
	if queryKey, queryOption, err = parseQueryType(queryType); err != nil {
		log.Error("ApmAggregateSetupList error %v", err)
		return
	}
	args = append(args, appKey)
	if queryKey == "addson" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND is_daily=? ")
		}
	} else if queryKey == "interval" {
		if queryOption != 0 {
			args = append(args, queryOption)
			sqls = append(sqls, " AND interval_time=? ")
		}
	}
	if versionCode != "" {
		sqls = append(sqls, fmt.Sprintf(" AND version_code in (%s) ", versionCode))
	}
	args = append(args, isAllVersion)
	args = append(args, startTime)
	args = append(args, endTime)
	sqls = append(sqls, " AND is_all_version = ? ")
	sqls = append(sqls, "AND timestamp >= ?")
	sqls = append(sqls, "AND timestamp <= ?")
	sqlAdd = strings.Join(sqls, " ")
	rows, err := d.clickhouse.Query(c, fmt.Sprintf(_getSetupList, queryKey, sqlAdd), args...)
	if err != nil {
		log.Error("ApmAggregateSetupList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateCountItem{}
		if err = rows.Scan(&re.Timestamp, &re.AppKey, &re.VersionCode, &re.Count, &re.DistinctBuvidCount); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAggregateCrashInfo(c context.Context, appKey string, versionCode, endTime int64) (res *apm.AggregateCountItem, err error) {
	rows, err := d.clickhouse.Query(c, _getAlertCrashInfo, appKey, versionCode, endTime)
	if err != nil {
		log.Error("ApmAggregateCrashInfo %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateCountItem{}
		if err = rows.Scan(&re.Timestamp, &re.AppKey, &re.VersionCode, &re.Count, &re.DistinctBuvidCount); err != nil {
			log.Error("%v", err)
			return
		}
		res = re
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmAggregateSetupInfo(c context.Context, appKey string, versionCode, endTime int64) (res *apm.AggregateCountItem, err error) {
	rows, err := d.clickhouse.Query(c, _getAlertSetupInfo, appKey, versionCode, endTime)
	if err != nil {
		log.Error("ApmAggregateSetupInfo %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apm.AggregateCountItem{}
		if err = rows.Scan(&re.Timestamp, &re.AppKey, &re.VersionCode, &re.Count, &re.DistinctBuvidCount); err != nil {
			log.Error("%v", err)
			return
		}
		res = re
	}
	err = rows.Err()
	return
}

// TestFlightPackUser get testflight user count of a package
func (d *Dao) TestFlightPackUser(c context.Context, appKey string, versionCode, createDate int64) (count int64, err error) {
	row := d.clickhouse.QueryRow(c, _getTFPackUser, appKey, versionCode, createDate)
	if err = row.Scan(&count); err != nil {
		log.Error("TestFlightPackUser %v", err)
	}
	return
}
