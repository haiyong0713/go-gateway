package fawkes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_getModDownloadSize = `SELECT sum(toUInt64OrZero(extended_fields['downloadSize'])) FROM bilibili_mobile_monitor_basesdk.ads_app_technology_rt WHERE event_id = 'public.modmanager.update.track' AND %s` // toDateTime(1657007749)

	_getActiveUserCount = `SELECT count(DISTINCT buvid) FROM bilibili_mobile_monitor.ads_infra_setup_rt WHERE %s` // toDateTime(1657007749)

)

func (d *Dao) ModDownloadSizeSum(c context.Context, appKey, poolName, modName string, startTime, endTime time.Time) (downloadSize float64, err error) {
	var (
		sqlAdd string
		sqls   []string
		args   []interface{}
	)
	if len(appKey) != 0 {
		args = append(args, appKey)
		sqls = append(sqls, "app_key=?")
	}
	if len(poolName) != 0 {
		args = append(args, poolName)
		sqls = append(sqls, "toString(extended_fields['pool'])=?")
	}
	if len(modName) != 0 {
		args = append(args, modName)
		sqls = append(sqls, "toString(extended_fields['mod'])=?")
	}
	args = append(args, startTime)
	sqls = append(sqls, "time_iso>=?")
	args = append(args, endTime)
	sqls = append(sqls, "time_iso<=?")
	sqlAdd = strings.Join(sqls, " AND ")
	sql := fmt.Sprintf(_getModDownloadSize, sqlAdd)
	log.Infoc(c, "ModDownloadSizeSum sql: %v, args: %v", sql, args)
	row := d.clickhouse.QueryRow(c, sql, args...)
	err = row.Scan(&downloadSize)
	return
}

func (d *Dao) ActiveUserCount(c context.Context, appKey string, appVerCondition []map[mod.Condition]int64, startTime, endTime time.Time) (activeUserCount int64, err error) {
	var (
		sqlAdd string
		sqls   []string
		args   []interface{}
	)
	if len(appKey) != 0 {
		args = append(args, appKey)
		sqls = append(sqls, "app_key=?")
	}
	if len(appVerCondition) != 0 {
		sql, arg := verConditionSql(appVerCondition)
		for _, v := range arg {
			args = append(args, v)
		}
		sqls = append(sqls, sql)
	}
	args = append(args, startTime)
	sqls = append(sqls, "time_iso>=?")
	args = append(args, endTime)
	sqls = append(sqls, "time_iso<=?")
	sqlAdd = strings.Join(sqls, " AND ")
	sql := fmt.Sprintf(_getActiveUserCount, sqlAdd)
	log.Infoc(c, "ActiveUserCount sql: %v, args: %v", sql, args)
	row := d.clickhouse.QueryRow(c, sql, args...)
	err = row.Scan(&activeUserCount)
	return
}

func verConditionSql(ver []map[mod.Condition]int64) (string, []int64) {
	var (
		sqls []string
		args []int64
	)
	for _, term := range ver {
		if len(term) == 0 {
			continue
		}
		for k, v := range term {
			switch k {
			case mod.ConditionLe:
				sqls = append(sqls, "toString(version_code) <= toString(?)")
				args = append(args, v)
			case mod.ConditionLt:
				sqls = append(sqls, "toString(version_code) < toString(?)")
				args = append(args, v)
			case mod.ConditionGe:
				sqls = append(sqls, "toString(version_code) >= toString(?)")
				args = append(args, v)
			case mod.ConditionGt:
				sqls = append(sqls, "toString(version_code) > toString(?)")
				args = append(args, v)
			}
		}
	}
	return strings.Join(sqls, " AND "), args
}
