package fawkes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/app-svr/fawkes/service/dao/database"
	clickSql "go-gateway/app/app-svr/fawkes/service/dao/database"
	apmmdl "go-gateway/app/app-svr/fawkes/service/model/apm"
	"go-gateway/app/app-svr/fawkes/service/model/bugly"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	// 通用业务
	_db    = "bilibili_mobile_monitor.ads_"
	_webdb = "bilibili_web_monitor.web_"

	_selectCount                   = `count()`
	_selectDistinctMidCount        = `count(DISTINCT mid)`
	_selectDistinctBuvidCount      = `count(DISTINCT buvid)`
	_selectNetErrorCount           = `sum(CASE WHEN http_code < 200 AND http_code >= 300 THEN 1 ELSE 0 END)`
	_selectNetSuccessRate          = `if(sum(CASE WHEN network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 THEN 1 ELSE 0 END) > 0, sum(CASE WHEN ((platform = 3 AND exception_msg = ' ') OR platform != 3 ) AND http_code >= 200 AND http_code < 300 AND network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 THEN (1/rate) ELSE 0 END)/sum(CASE WHEN network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 THEN (1/rate) ELSE 0 END), 0)`
	_selectNetErrorRate            = `1 - ` + _selectNetSuccessRate
	_selectNetSuccessRateDowngrade = `if(sum(CASE WHEN network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 AND downgrade = 0 THEN 1 ELSE 0 END) > 0, sum(CASE WHEN ((platform = 3 AND exception_msg = ' ') OR platform != 3 ) AND http_code >= 200 AND http_code < 300 AND network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 THEN (1/rate) ELSE 0 END)/sum(CASE WHEN network not in (-1, 3) AND total_time < 600000 AND rate > 0 AND tunnel = 0 AND downgrade = 0 THEN (1/rate) ELSE 0 END), ` + _selectNetSuccessRate + `)`
	_selectNetErrorRateDowngrade   = `1 - ` + _selectNetSuccessRateDowngrade
	_selectNetTotalTimeQuantile80  = `if(quantile(0.8)(total_time) > 0, quantile(0.8)(total_time), 0)`
	_selectNetTotalTimeQuantile95  = `if(quantile(0.95)(total_time) > 0, quantile(0.95)(total_time), 0)`
	_selectNetAvgReqSize           = `if(avg(req_size) > 0, avg(req_size), 0)`
	_selectNetAvgRecvSize          = `if(avg(recv_size) > 0, avg(recv_size), 0)`
	_selectNetBizErrorCount        = `sum(CASE WHEN response_code != 0 THEN 1 ELSE 0 END)`
	_selectNetBizParseErrorCount   = `sum(CASE WHEN biz_code = 1001 THEN 1 ELSE 0 END)`
	_selectStatisticsSuccessRate   = `if(count() > 0, sum(CASE WHEN status_code = 0 THEN 1 ELSE 0 END) / count(), 0)`
	_selectStatisticsErrorRate     = `if(count() > 0, sum(CASE WHEN status_code != 0 THEN 1 ELSE 0 END) / count(), 0)`
	_selectNetBizSuccessRate       = `if(sum(CASE WHEN network not in (-1, 3) AND biz_code != -250 THEN 1 ELSE 0 END) > 0, sum(CASE WHEN network not in (-1, 3) AND biz_code = 0 AND biz_code != -250 THEN 1 ELSE 0 END) / sum(CASE WHEN network not in (-1, 3) AND biz_code != -250 THEN 1 ELSE 0 END), 0)`
	_selectNetBizErrorRate         = `if(sum(CASE WHEN network not in (-1, 3) AND biz_code != -250 THEN 1 ELSE 0 END) > 0, sum(CASE WHEN network not in (-1, 3) AND biz_code != 0 AND biz_code != -250 THEN 1 ELSE 0 END) / sum(CASE WHEN network not in (-1, 3) AND biz_code != -250 THEN 1 ELSE 0 END), 0)`

	// 其他业务
	_getRouterList         = `SELECT real_name_from, name_to, count(), quantile(0.9)(memory) FROM bilibili_mobile_monitor.ads_infra_route_flux_rt WHERE app_key=? AND version_code=toInt64(?) AND time_iso>? AND time_iso<? AND real_name_from!='' AND name_to!='' GROUP BY real_name_from, name_to`
	_getSetupDetailList    = `SELECT toUnixTimestamp(time_iso),mid,buvid,model,brand,osver,app_key,version,version_code,province,isp,ff_version,config_version FROM bilibili_mobile_monitor.ads_infra_setup_rt WHERE %v ORDER BY time_iso DESC`
	_getBuglyCrashInfoList = `SELECT toUnixTimestamp(time_iso) AS time_iso,ip,mid,buvid,brand,chid,model,network,oid,version,version_code,crash_version,platform,osver,ff_version,config_version,app_key,country,province,city,isp,process,thread,crash_type,error_type,error_msg,error_stack,last_activity,top_activity,macho,all_macho,analyse_error_code,analyse_error_stack,error_stack_hash_without_useless_v2 AS error_stack_hash_without_useless,crash_time,lifetime,build_id,mem_free,storage_free,sdcard_free,manufacturer,domestic_rom_ver
				FROM %s.%s %s`
	_getBuglyJankInfoList = `SELECT event_id,toUnixTimestamp(time_iso) AS time_iso,ip,mid,buvid,brand,device_id,uid,chid,model,fts,network,oid,app_id,version,version_code,platform,osver,ff_version,config_version,abi,app_key,rate,country,province,city,isp,internal_version,process,thread,stacktrace_count,duration,jank_stack,jank_stack_count_json,jank_stack_max_count,analyse_jank_code,analyse_jank_stack,analyse_jank_stack_hash,build_id,route
				FROM bilibili_mobile_monitor.ads_public_apm_jank_monitor_analyse_rt %s ORDER BY time_iso DESC LIMIT ?,?`
	_getBuglyOOMInfoList = `SELECT event_id,toUnixTimestamp(time_iso) AS time_iso,ip,mid,buvid,brand,device_id,uid,chid,model,fts,network,oid,app_id,version,version_code,platform,osver,ff_version,config_version,abi,app_key,rate,country,province,city,isp,internal_version,process,thread,last_activity,top_activity,app_memory,app_memory_rate,device_ram,dump_time,file_size,file_url,session_id,hash,stack,analyse_stack,instance_count,leak_reason,gc_root,signature,retained_size,path
				FROM bilibili_mobile_monitor.ads_public_apm_oom_monitor_analyse_rt %s ORDER BY time_iso DESC LIMIT ?,?`
	// // Web数据埋点
	// _insertWebTrackPv    = `INSERT INTO bilibili_fawkes_monitor.fawkes_track_pv(timestamp,app_key,username,browser_name,browser_code,browser_version,navigator_platform,route_path,route_name,route_from) VALUES (?,?,?,?,?,?,?,?,?,?)`
	// _insertWebTrackError = `INSERT INTO bilibili_fawkes_monitor.fawkes_track_error(timestamp,app_key,username,browser_name,browser_code,browser_version,navigator_platform,route_path,error_msg) VALUES (?,?,?,?,?,?,?,?,?)`

	// 自定义计算函数
	_ffBarrelQuery = `if(modulo(javaHash(upper(hex(MD5(?)))), 1000) >= 0, modulo(javaHash(upper(hex(MD5(?)))), 1000), modulo(javaHash(upper(hex(MD5(?)))), 1000) + 1000)`
)

const (
	WeeklyIntervalTimeStamp = 604800
)

// 根据类型选择、实际的逻辑SQL
func getColumnSQL(cType string, customQuerySql ...interface{}) (column string, err error) {
	var columnSQL string
	switch cType {
	case "count":
		columnSQL = _selectCount
	case "distinct_mid_count":
		columnSQL = _selectDistinctMidCount
	case "distinct_buvid_count":
		columnSQL = _selectDistinctBuvidCount
	case "http_error_count":
		columnSQL = _selectNetErrorCount
	case "http_success_rate":
		columnSQL = _selectNetSuccessRate
	case "http_error_rate":
		columnSQL = _selectNetErrorRate
	case "http_success_rate_downgrade":
		columnSQL = _selectNetSuccessRateDowngrade
	case "http_error_rate_downgrade":
		columnSQL = _selectNetErrorRateDowngrade
	case "http_biz_error_rate":
		columnSQL = _selectNetBizErrorRate
	case "http_biz_success_rate":
		columnSQL = _selectNetBizSuccessRate
	case "total_time_quantile_80":
		columnSQL = _selectNetTotalTimeQuantile80
	case "total_time_quantile_95":
		columnSQL = _selectNetTotalTimeQuantile95
	case "req_size_avg":
		columnSQL = _selectNetAvgReqSize
	case "recv_size_avg":
		columnSQL = _selectNetAvgRecvSize
	case "biz_error_count":
		columnSQL = _selectNetBizErrorCount
	case "biz_parse_error_count":
		columnSQL = _selectNetBizParseErrorCount
	case "statistics_success_rate":
		columnSQL = _selectStatisticsSuccessRate
	case "statistics_error_rate":
		columnSQL = _selectStatisticsErrorRate
	case "ff_barrel_num":
		columnSQL = _ffBarrelQuery
	case "custom_query_sql":
		column = fmt.Sprintf("%v", customQuerySql...)
		return
	default:
		if columnSQL, err = getCustomColumnSQL(cType); err != nil {
			log.Error("%v", err)
			return
		}
	}
	cType = strings.ReplaceAll(strings.Replace(cType, "@", "_", 1), ":", "_")
	column = fmt.Sprintf("%s AS %s", columnSQL, cType)
	return
}

// 自定义数据
func getCustomColumnSQL(cType string) (columnSQL string, err error) {
	comps := strings.Split(cType, "@")
	//nolint:gomnd
	if len(comps) != 2 {
		columnSQL = cType
		return
	}
	var queryKey = comps[0]
	var fieldParam = comps[1]
	switch {
	case strings.HasPrefix(queryKey, "quantile"):
		var fieldValueFloat float64
		keys := strings.Split(queryKey, "_")
		if fieldValueFloat, err = strconv.ParseFloat(keys[1], 64); err != nil {
			log.Error("%v", err)
			return
		}
		columnSQL = fmt.Sprintf("if(quantile(%v)(%v) > 0, quantile(%v)(%v), 0)", fieldValueFloat/100, fieldParam, fieldValueFloat/100, fieldParam)
	case strings.HasPrefix(queryKey, "distinct_count"):
		columnSQL = fmt.Sprintf("count(DISTINCT %v)", fieldParam)
	case strings.HasPrefix(queryKey, "json"):
		fields := strings.Split(fieldParam, ":")
		columnSQL = fields[0]
		for _, field := range fields[1:] {
			columnSQL = fmt.Sprintf("JSONExtractRaw(%v, '%v')", columnSQL, field)
		}
	case strings.HasPrefix(queryKey, "avg"):
		columnSQL = fmt.Sprintf("if (avg(%v) > 0, avg(%v), 0)", fieldParam, fieldParam)
	default:
		columnSQL = fmt.Sprintf("%v(%v)", queryKey, fieldParam)
	}
	return
}

// 判断是否有控制
func getRealValue(value string) (realValue string, isEmpty bool) {
	switch value {
	case "":
		isEmpty = true
	case "__UNKNOW__":
		isEmpty = false
		realValue = ""
	default:
		isEmpty = false
		realValue = value
	}
	return
}

func parseMatchOptionV2(matchOption *apmmdl.MatchOption) (condition string, args []interface{}, err error) {
	// 追加查询条件
	matchSqls, matchArgs := parseMatchOption(matchOption)
	if len(matchSqls) > 0 {
		condition = strings.Join(matchSqls, " AND ")
	}
	args = append(args, matchArgs...)
	// filters 处理
	var (
		filterCondition string
		filterArgs      []interface{}
	)
	if filterCondition, filterArgs, err = ParseFiltersOption(matchOption.Filters); err != nil {
		log.Error("ParseFiltersOption error %v", err)
		return
	}
	if filterCondition != "" {
		condition = condition + " AND " + "(" + filterCondition + ")"
	}
	args = append(args, filterArgs...)
	log.Warn("condition:%s args:%v", condition, args)
	return
}

func ParseFiltersOption(filters []*apmmdl.Filter) (condition string, args []interface{}, err error) {
	if len(filters) == 0 {
		return
	}
	for index, item := range filters {
		var (
			andTypeSql     string
			columnSql      string
			equalSql       string
			valueSql       string
			valueCondition string
		)
		// andType AND OR
		if item.AndType == "" {
			andTypeSql = " AND "
		} else {
			andTypeSql = " " + item.AndType + " "
		}
		if index == 0 {
			andTypeSql = ""
		}
		if columnSql, err = getCustomColumnSQL(item.Column); err != nil {
			log.Error("%v", err)
			return
		}
		// equalType = != < <= > >= null notnull like
		if item.EqualType == "" {
			equalSql = "="
		} else {
			equalSql = item.EqualType
		}
		condition += andTypeSql
		// valueType
		values := strings.Split(item.Values, ",")
		for index, value := range values {
			var arg interface{}
			switch item.ValueType {
			case "STRING":
				arg = string(value)
			case "INT64":
				if arg, err = strconv.ParseInt(value, 10, 64); err != nil {
					log.Error("value convert value(%s) err(%v)", item.Values, err)
					return
				}
			case "":
			default:
				arg = string(value)
			}
			switch equalSql {
			case "LIKE":
				valueSql = "?"
				arg = "%" + arg.(string) + "%"
				args = append(args, arg)
			case "IS NULL":
				valueSql = ""
			case "IS NOT NULL":
				valueSql = ""
			default:
				valueSql = "?"
				args = append(args, arg)
			}
			if index != 0 && index <= len(values)-1 {
				valueCondition += " OR "
			}
			valueCondition += fmt.Sprintf("%s %s %s", columnSql, equalSql, valueSql)
		}
		if len(values) > 0 {
			condition += "(" + valueCondition + ")"
		}
	}
	return
}

// 解析前端的条件，用于SQL WHERE条件
// nolint:gocognit
func parseMatchOption(baseMatchOptions *apmmdl.MatchOption) (sqls []string, args []interface{}) {
	if realValue, isEmpty := getRealValue(baseMatchOptions.AppKey); !isEmpty {
		sqls = append(sqls, "app_key=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Country); !isEmpty {
		sqls = append(sqls, "country=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Province); !isEmpty {
		sqls = append(sqls, "province=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.City); !isEmpty {
		sqls = append(sqls, "city=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.NegotiatedProtocol); !isEmpty {
		sqls = append(sqls, "negotiated_protocol=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.InternetProtocolVersion); !isEmpty {
		sqls = append(sqls, "internet_protocol_version=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Network); !isEmpty {
		sqls = append(sqls, "network=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.VersionCode); !isEmpty {
		sqls = append(sqls, "version_code=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.TunnelSDK); !isEmpty {
		sqls = append(sqls, "tunnel_sdk=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.RealRequestUrl); !isEmpty {
		sqls = append(sqls, "real_request_url=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Isp); !isEmpty {
		sqls = append(sqls, "isp=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Model); !isEmpty {
		sqls = append(sqls, "upper(model)=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Brand); !isEmpty {
		sqls = append(sqls, "upper(brand)=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Platform); !isEmpty {
		sqls = append(sqls, "platform=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Oid); !isEmpty {
		sqls = append(sqls, "oid=toInt64(?)")
		args = append(args, realValue)
	}
	if baseMatchOptions.StartTime != 0 {
		sqls = append(sqls, "time_iso > ?")
		args = append(args, baseMatchOptions.StartTime/1000)
	}
	if baseMatchOptions.EndTime != 0 {
		sqls = append(sqls, "time_iso < ?")
		args = append(args, baseMatchOptions.EndTime/1000)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.HTTPCode); !isEmpty {
		sqls = append(sqls, "http_code=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.BizCode); !isEmpty {
		sqls = append(sqls, "biz_code=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Command); !isEmpty {
		sqls = append(sqls, "command=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Domain); !isEmpty {
		sqls = append(sqls, "domain=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ResponseCode); !isEmpty {
		sqls = append(sqls, "response_code=toInt32(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.OSVer); !isEmpty {
		sqls = append(sqls, "osver=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ErrorType); !isEmpty {
		sqls = append(sqls, "error_type=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ErrorMessage); !isEmpty {
		sqls = append(sqls, "error_msg=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Process); !isEmpty {
		sqls = append(sqls, "process=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Thread); !isEmpty {
		sqls = append(sqls, "thread=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ErrorStack); !isEmpty {
		sqls = append(sqls, "error_stack LIKE ?")
		args = append(args, "%"+realValue+"%")
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Hash); !isEmpty {
		sqls = append(sqls, "hash=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ErrorStackHashWithoutUseless); !isEmpty {
		sqls = append(sqls, "error_stack_hash_without_useless_v2=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.AnalyseErrorStack); !isEmpty {
		sqls = append(sqls, "analyse_error_stack LIKE ?")
		args = append(args, "%"+realValue+"%")
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.AnalyseJankStack); !isEmpty {
		sqls = append(sqls, "analyse_jank_stack LIKE ?")
		args = append(args, "%"+realValue+"%")
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.LastActivity); !isEmpty {
		sqls = append(sqls, "last_activity=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.TopActivity); !isEmpty {
		sqls = append(sqls, "top_activity=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.StatusCode); !isEmpty {
		sqls = append(sqls, "status_code=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ExternalNumber1); !isEmpty {
		sqls = append(sqls, "external_num1=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ExternalNumber2); !isEmpty {
		sqls = append(sqls, "external_num2=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ExternalNumber3); !isEmpty {
		sqls = append(sqls, "external_num3=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ExternalNumber4); !isEmpty {
		sqls = append(sqls, "external_num4=toInt64(?)")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.GroupKey); !isEmpty {
		sqls = append(sqls, "group_key=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.NameFrom); !isEmpty {
		sqls = append(sqls, "name_from=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.NameTo); !isEmpty {
		sqls = append(sqls, "name_to=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.RealNameFrom); !isEmpty {
		sqls = append(sqls, "real_name_from=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.AnalyseJankStackHash); !isEmpty {
		sqls = append(sqls, "analyse_jank_stack_hash=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.JankStackMaxCount); !isEmpty {
		sqls = append(sqls, "jank_stack_max_count>?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.FFVersion); !isEmpty {
		if realValue == "__NOT_NULL__" {
			sqls = append(sqls, "ff_version != ''")
		} else {
			sqls = append(sqls, "ff_version=? AND ff_version != ''")
			args = append(args, realValue)
		}
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.ConfigVersion); !isEmpty {
		if realValue == "__NOT_NULL__" {
			sqls = append(sqls, "config_version != ''")
		} else {
			sqls = append(sqls, "config_version=? AND config_version != ''")
			args = append(args, realValue)
		}
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Href); !isEmpty {
		sqls = append(sqls, "href=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Mid); !isEmpty {
		sqls = append(sqls, "mid=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Buvid); !isEmpty {
		sqls = append(sqls, "buvid=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.Chid); !isEmpty {
		sqls = append(sqls, "chid=?")
		args = append(args, realValue)
	}
	if realValue, isEmpty := getRealValue(baseMatchOptions.CrashType); !isEmpty {
		comps := strings.Split(realValue, "@")
		condition := ""
		if len(comps) > 1 {
			condition = comps[0]
			realValue = comps[1]
		}
		if condition == "~" {
			sqls = append(sqls, "crash_type not in (?)")
		} else {
			sqls = append(sqls, "crash_type in (?)")
		}
		args = append(args, realValue)
	}
	return
}

// 执行SQL. 若为WEB平台. 则进行数据库/表隔离
func execQuerySql(c context.Context, d *Dao, matchOption *apmmdl.MatchOption, querySQl string, args ...interface{}) (rows *database.Rows, err error) {
	// Web端数据源. 则使用老集群进行数据查询
	if strings.HasPrefix(matchOption.AppKey, "web") {
		appInfo, _ := d.AppPass(c, matchOption.AppKey)
		if appInfo.Platform == "web" {
			// web端换表
			newQuerySQl := strings.Replace(querySQl, _db, _webdb, -1)
			newQuerySQl = strings.Replace(newQuerySQl, "_rt", "_all", -1)
			rows, err = d.clickhouse2.Query(c, newQuerySQl, args...)
			log.Warnc(c, "clickhouse sql: %v, args: %v", newQuerySQl, args)
			return
		}
	}
	rows, err = d.clickhouse.Query(c, querySQl, args...)
	log.Warnc(c, "clickhouse sql: %v, args: %v", querySQl, args)
	return
}

func (d *Dao) ApmMoniCalculate(c context.Context, cType string, matchOption *apmmdl.MatchOption) (re *apmmdl.Moni, err error) {
	var (
		args []interface{}
	)
	for _, v := range strings.Split(matchOption.CalculateArgs, ",") {
		args = append(args, v)
	}
	columnSQLs, err := getColumnSQL(cType)
	if err != nil {
		log.Error("%v", err)
		return
	}
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(`SELECT %v`, columnSQLs), args...)
	if err != nil {
		log.Error("ApmMoniCalculate %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re = &apmmdl.Moni{}
		if err = rows.Scan(&re.Value); err != nil {
			log.Error("ApmMoniCalculate %v", err)
			return
		}
	}
	return
}

// 线形图专用函数
func (d *Dao) ApmMoniLine(c context.Context, database, tableName, cType string, commands []string, matchOption *apmmdl.MatchOption) (res []*apmmdl.Moni, err error) {
	var (
		querySQL        = `SELECT toUnixTimestamp(toStartOfInterval(time_iso, INTERVAL %s)) * 1000 as t,%s FROM %s GROUP BY t ORDER BY t`
		db              = fmt.Sprintf("%s.%s", database, tableName)
		args, matchArgs []interface{}
		condition       string
	)
	columnSQLs, err := getColumnSQL(cType)
	if err != nil {
		log.Error("%v", err)
		return
	}
	// command IN (?,?,?) ➡ args sqls
	commandArgs, commandSqls := parseCommands(commands)
	// parseMatchOptions
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("CrashInfoList parseBaseMatchOptionsNew err(%v)", err)
		return
	}
	if len(commandSqls) > 0 {
		if condition != "" {
			condition = strings.Join(commandSqls, " AND ") + " AND " + condition
		} else {
			condition = strings.Join(commandSqls, " AND ")
		}
	}
	if condition != "" {
		db += fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, commandArgs...)
	args = append(args, matchArgs...)
	execSql := fmt.Sprintf(querySQL, matchOption.IntervalTime, columnSQLs, db)
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.Moni{}
		if err = rows.Scan(&re.Title, &re.Value); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// 饼图专用函数
func (d *Dao) ApmMoniPie(c context.Context, database, tableName, cType, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*apmmdl.Moni, err error) {
	var (
		querySQL        = `SELECT %s,%s FROM %s GROUP BY %s ORDER BY %s LIMIT %s`
		db              = fmt.Sprintf("%s.%s", database, tableName)
		args, matchArgs []interface{}
		condition       string
	)
	columnSQLs, err := getColumnSQL(cType)
	if err != nil {
		log.Error("%v", err)
		return
	}
	// command IN (?,?,?) ➡ args sqls
	commandArgs, commandSqls := parseCommands(commands)
	// parseMatchOptions
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("CrashInfoList parseBaseMatchOptionsNew err(%v)", err)
		return
	}
	if len(commandSqls) > 0 {
		if condition != "" {
			condition = strings.Join(commandSqls, " AND ") + " AND " + condition
		} else {
			condition = strings.Join(commandSqls, " AND ")
		}
	}
	if condition != "" {
		db += fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, commandArgs...)
	args = append(args, matchArgs...)
	execSql := fmt.Sprintf(querySQL, column, columnSQLs, db, column, matchOption.OrderBy, strconv.Itoa(matchOption.Limit))
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.Moni{}
		if err = rows.Scan(&re.Title, &re.Value); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// 网络/图片 相关数据基础信息
func (d *Dao) ApmMoniNetInfoList(c context.Context, database, tableName, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*apmmdl.NetInfo, err error) {
	var (
		querySQL = `SELECT %s,%s FROM %s %s %s %s`
		db       = fmt.Sprintf("%s.%s", database, tableName)
		sqls     []string
		args     []interface{}
		groupBy  string
		orderBy  string
		limit    string
	)
	// 默认可查询的列表
	defaultColumns := []string{
		"count",
		"total_time_quantile_80",
		"total_time_quantile_95",
		"http_success_rate",
		"http_success_rate_downgrade",
		"http_biz_success_rate",
		"req_size_avg",
		"recv_size_avg",
	}
	// 前端请求需要查询的columns && 构建SQL组
	queryColumns := strings.Split(matchOption.QueryKeys, ",")
	columnSQLs := []string{}
	for _, column := range defaultColumns {
		contain := false
		for _, queryColumn := range queryColumns {
			if queryColumn == column {
				contain = true
				break
			}
		}
		if contain {
			var columnSQL string
			columnSQL, err = getColumnSQL(column)
			if err != nil {
				log.Error("%v", err)
				return
			}
			columnSQLs = append(columnSQLs, columnSQL)
		} else {
			columnSQLs = append(columnSQLs, "0")
		}
	}
	// command IN (?,?,?)
	if len(commands) > 0 {
		var sqlsTmp []string
		for _, command := range commands {
			sqlsTmp = append(sqlsTmp, "?")
			args = append(args, command)
		}
		sqls = append(sqls, fmt.Sprintf("command IN(%v)", strings.Join(sqlsTmp, ",")))
	}
	// 追加查询条件
	matchSqls, matchArgs := parseMatchOption(matchOption)
	args = append(args, matchArgs...)
	sqls = append(sqls, matchSqls...)
	if len(sqls) > 0 {
		condition := strings.Join(sqls, " AND ")
		db += fmt.Sprintf(" WHERE %v", condition)
	}
	if column != "" {
		groupBy = "GROUP BY " + column
	} else {
		column = "0"
	}
	if matchOption.OrderBy != "" {
		orderBy = "ORDER BY " + matchOption.OrderBy
	}
	if matchOption.Limit != 0 {
		limit = "LIMIT " + strconv.Itoa(matchOption.Limit)
	}
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(querySQL, column, strings.Join(columnSQLs, ","), db, groupBy, orderBy, limit), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.NetInfo{}
		if err = rows.Scan(&re.Command, &re.Count, &re.TotalTimeQuantile80, &re.TotalTimeQuantile95, &re.NetSuccessRate, &re.NetSuccessRateDowngrade, &re.NetBizSuccessRate, &re.ReqSizeAvg, &re.RecvSizeAvg); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) ApmMoniMetricInfoList(c context.Context, database, tableName, column string, matchOption *apmmdl.MatchOption) (res []*apmmdl.MetricInfo, err error) {
	var (
		execSql string
		args    []interface{}
	)
	if execSql, args, err = getMetricSQLs(database, tableName, column, nil, nil, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.MetricInfo{}
		if err = rows.Scan(&re.Command, &re.Value); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// 通配获取数量的函数
func (d *Dao) ApmMoniCountInfoList(c context.Context, database, tableName, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*apmmdl.CountInfo, err error) {
	var (
		execSql string
		args    []interface{}
	)
	// 默认可查询的列表
	defaultColumns := []string{
		"count",
		"distinct_mid_count",
		"distinct_buvid_count",
	}
	if execSql, args, err = getMetricSQLs(database, tableName, column, defaultColumns, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.CountInfo{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// 自定义埋点函数
func (d *Dao) ApmMoniStatisticsInfoList(c context.Context, database, tableName, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*apmmdl.StatisticsInfo, err error) {
	var (
		_arg    = `SELECT %s,%s FROM %s %s %s %s`
		db      = fmt.Sprintf("%s.%s", database, tableName)
		sqls    []string
		args    []interface{}
		groupBy string
		orderBy string
		limit   string
	)
	queryKeys := strings.Split(matchOption.QueryKeys, ",")
	columnSQLs := []string{}
	for i := 0; i < 6; i++ {
		if len(queryKeys) > i {
			var (
				key       = queryKeys[i]
				columnSQL string
			)
			columnSQL, err = getColumnSQL(key)
			if err != nil {
				log.Error("%v", err)
				return
			}
			columnSQLs = append(columnSQLs, columnSQL)
		} else {
			columnSQLs = append(columnSQLs, "0")
		}
	}
	// command IN (?,?,?)
	if len(commands) > 0 {
		var sqlsTmp []string
		for _, command := range commands {
			sqlsTmp = append(sqlsTmp, "?")
			args = append(args, command)
		}
		sqls = append(sqls, fmt.Sprintf("command IN(%v)", strings.Join(sqlsTmp, ",")))
	}
	// 追加查询条件
	matchSqls, matchArgs := parseMatchOption(matchOption)
	args = append(args, matchArgs...)
	sqls = append(sqls, matchSqls...)
	if len(sqls) > 0 {
		condition := strings.Join(sqls, " AND ")
		db += fmt.Sprintf(" WHERE %v", condition)
	}
	if column != "" {
		groupBy = "GROUP BY " + column
	} else {
		column = "0"
	}
	if matchOption.OrderBy != "" {
		orderBy = "ORDER BY " + matchOption.OrderBy
	}
	if matchOption.Limit != 0 {
		limit = "LIMIT " + strconv.Itoa(matchOption.Limit)
	}
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(_arg, column, strings.Join(columnSQLs, ","), db, groupBy, orderBy, limit), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.StatisticsInfo{}
		if err = rows.Scan(&re.Command, &re.Count, &re.Num1, &re.Num2, &re.Num3, &re.Num4, &re.Num5); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// 路由表函数
func (d *Dao) ApmFlowmapRouteList(c context.Context, matchOption *apmmdl.MatchOption) (res []*apmmdl.FlowmapRoute, err error) {
	rows, err := execQuerySql(c, d, matchOption, _getRouterList, matchOption.AppKey, matchOption.VersionCode, matchOption.StartTime, matchOption.EndTime)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.FlowmapRoute{}
		if err = rows.Scan(&re.NameFrom, &re.NameTo, &re.Count, &re.Memory); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// Fawkes平台 -- 前端埋点专用
func (d *Dao) AddWebTracePv(c context.Context, models []*apmmdl.WebTrackModel) (err error) {
	// var (
	//	tx, _   = d.clickhouse.GetConnect().Begin()
	//	stmt, _ = tx.Prepare(_insertWebTrackPv)
	// )
	// defer stmt.Close()
	// for _, model := range models {
	//	if _, err := stmt.Exec(model.Timestamp, model.AppKey, model.Username, model.BowerName, model.BowerCode, model.BowerVersion,
	//		model.NavigatorPlatform, model.RoutePath, model.RouteName, model.RouteFrom); err != nil {
	//		log.Error("clickhouse stmt.Exec: %v", err)
	//	}
	// }
	// if err = tx.Commit(); err != nil {
	//	log.Error("%v", err)
	// }
	return
}

func (d *Dao) AddWebTraceError(c context.Context, models []*apmmdl.WebTrackModel) (err error) {
	// var (
	//	tx, _   = d.clickhouse.GetConnect().Begin()
	//	stmt, _ = tx.Prepare(_insertWebTrackError)
	// )
	// defer stmt.Close()
	// for _, model := range models {
	//	if _, err := stmt.Exec(model.Timestamp, model.AppKey, model.Username, model.BowerName, model.BowerCode, model.BowerVersion,
	//		model.NavigatorPlatform, model.RoutePath, model.ErrorMessage); err != nil {
	//		log.Error("clickhouse stmt.Exec: %v", err)
	//	}
	// }
	// if err = tx.Commit(); err != nil {
	//	log.Error("%v", err)
	// }
	return
}

func (d *Dao) ApmDetailSetup(c context.Context, matchOption *apmmdl.MatchOption) (res []*apmmdl.ApmDetailSetup, err error) {
	var (
		sqls      []string
		args      []interface{}
		endTime   int64
		startTime int64
	)
	if matchOption.Mid == "" && matchOption.Buvid == "" {
		return
	}
	if matchOption.Mid != "" {
		args = append(args, matchOption.Mid)
		sqls = append(sqls, "mid=?")
	}
	if matchOption.Buvid != "" {
		args = append(args, matchOption.Buvid)
		sqls = append(sqls, "buvid=?")
	}
	if matchOption.StartTime == 0 && matchOption.EndTime == 0 {
		// 时间范围默认一天
		end := time.Now()
		start := end.AddDate(0, 0, -1)
		startTime = start.Unix()
		endTime = end.Unix()
	} else if (matchOption.EndTime - matchOption.StartTime) > WeeklyIntervalTimeStamp {
		// 时间间隔不能超过一周
		startTime = matchOption.EndTime - WeeklyIntervalTimeStamp
		endTime = matchOption.EndTime
	} else {
		startTime = matchOption.StartTime
		endTime = matchOption.EndTime
	}
	args = append(args, startTime)
	args = append(args, endTime)
	sqls = append(sqls, "toUnixTimestamp(time_iso)>?")
	sqls = append(sqls, "toUnixTimestamp(time_iso)<?")
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(_getSetupDetailList, strings.Join(sqls, " AND ")), args...)
	log.Infoc(c, "ApmDetailSetup exec query")
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &apmmdl.ApmDetailSetup{}
		if err = rows.Scan(&re.TimeISO, &re.Mid, &re.Buvid, &re.Model, &re.Brand, &re.Osver, &re.AppKey, &re.Version, &re.VersionCode, &re.Province, &re.ISP, &re.FFVersion, &re.ConfigVersion); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		res = append(res, re)
	}
	log.Infoc(c, "ApmDetailSetup scan")
	err = rows.Err()
	return
}

func (d *Dao) CrashInfoList(c context.Context, database, tableName string, matchOption *apmmdl.MatchOption) (res []*bugly.CrashInfo, err error) {
	var (
		condition       string
		args, matchArgs []interface{}
	)
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("CrashInfoList parseBaseMatchOptionsNew err(%v)", err)
		return
	}
	if condition != "" {
		condition = fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, matchArgs...)
	condition += " ORDER BY time_iso DESC"
	if matchOption.Pn != -1 && matchOption.Ps != -1 {
		args = append(args, (matchOption.Pn-1)*matchOption.Ps, matchOption.Ps)
		condition += " LIMIT ?,?"
	}
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(_getBuglyCrashInfoList, database, tableName, condition), args...)
	if err != nil {
		log.Error("CrashInfoList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.CrashInfo{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	return
}

// 卡顿数量
func (d *Dao) ApmMoniJankIndexCountInfoList(c context.Context, database, tableName, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*bugly.JankIndex, err error) {
	var (
		execSql string
		args    []interface{}
	)
	// 默认可查询的列表
	defaultColumns := []string{
		"count",
		"distinct_mid_count",
		"distinct_buvid_count",
		"quantile_80@duration",
	}
	if execSql, args, err = getMetricSQLs(database, tableName, column, defaultColumns, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.JankIndex{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// OOM数量
func (d *Dao) ApmMoniOOMIndexCountInfoList(c context.Context, database, tableName, column string, commands []string, matchOption *apmmdl.MatchOption) (res []*bugly.OOMIndex, err error) {
	var (
		execSql string
		args    []interface{}
	)
	// 默认可查询的列表
	defaultColumns := []string{
		"count",
		"distinct_mid_count",
		"distinct_buvid_count",
		"quantile_80@retained_size",
	}
	if execSql, args, err = getMetricSQLs(database, tableName, column, defaultColumns, commands, matchOption); err != nil {
		log.Error("%v", err)
		return
	}
	rows, err := execQuerySql(c, d, matchOption, execSql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.OOMIndex{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) JankInfoList(c context.Context, matchOption *apmmdl.MatchOption) (res []*bugly.JankInfo, err error) {
	var (
		condition       string
		args, matchArgs []interface{}
	)
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("CrashInfoList parseBaseMatchOptionsNew err(%v)", err)
		return
	}
	if condition != "" {
		condition = fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, matchArgs...)
	args = append(args, (matchOption.Pn-1)*matchOption.Ps, matchOption.Ps)
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(_getBuglyJankInfoList, condition), args...)
	if err != nil {
		log.Error("JankInfoList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.JankInfo{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	return
}

func (d *Dao) OOMInfoList(c context.Context, matchOption *apmmdl.MatchOption) (res []*bugly.OOMInfo, err error) {
	var (
		condition       string
		args, matchArgs []interface{}
	)
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("OOMInfoList parseBaseMatchOptionsNew err(%v)", err)
		return
	}
	if condition != "" {
		condition = fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, matchArgs...)
	args = append(args, (matchOption.Pn-1)*matchOption.Ps, matchOption.Ps)
	rows, err := execQuerySql(c, d, matchOption, fmt.Sprintf(_getBuglyOOMInfoList, condition), args...)
	if err != nil {
		log.Error("OOMInfoList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.OOMInfo{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	return
}

// 指标通用函数
// 特殊场景：
// defaultColumns = nil 则直接查询，不再merge queryKeys
// command = nil 则无需添加command配置
func getMetricSQLs(database, tableName, column string, defaultColumns, commands []string, matchOption *apmmdl.MatchOption) (execSQL string, execArgs []interface{}, err error) {
	var (
		db              = fmt.Sprintf("%s.%s", database, tableName)
		matchOptionSQL  string
		matchOptionArgs []interface{}
		conditionSQLs   []string
		conditionArgs   []interface{}
	)
	// 1. 获取 querySQL
	querySQL, err := getColumnQuerySQL(matchOption, column, defaultColumns)
	if err != nil {
		log.Error("%v", err)
		return
	}
	// 2. 解析 MatchOption，生成 WHERE 后置条件
	if matchOptionSQL, matchOptionArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Error("parseMatchOptionV2 err(%v)", err)
		return
	}
	conditionSQLs = append(conditionSQLs, matchOptionSQL)
	conditionArgs = append(conditionArgs, matchOptionArgs...)

	// 3. 兼容逻辑：若存在commands, 则注入SQL: command IN (?,?,?) ➡ args sqls
	commandArgs, commandSQLs := parseCommands(commands)
	conditionSQLs = append(conditionSQLs, commandSQLs...)
	conditionArgs = append(conditionArgs, commandArgs...)

	// conditionSQL
	if len(conditionSQLs) > 0 {
		db += fmt.Sprintf(" WHERE %v", strings.Join(conditionSQLs, " AND "))
	}

	// 4. GROUP BY / ORDER BY / LIMIT
	tailSQL := getTailSQL(matchOption, column)

	// 5. Generator SQL
	execSQL = fmt.Sprintf("SELECT %s FROM %s %s", querySQL, db, tailSQL)
	execArgs = conditionArgs
	return
}

// 生成 column, metric1, metric2 查询词条
func getColumnQuerySQL(matchOption *apmmdl.MatchOption, queryColumn string, defaultColumns []string) (querySQL string, err error) {
	var (
		columnSQLs []string
	)
	// 若column=""，则表示不会进行GROUP BY； 则该值默认为0直接返回前端即可
	if queryColumn == "" {
		queryColumn = "0"
	}
	// metricSQL
	queryKeys := strings.Split(matchOption.QueryKeys, ",")
	for _, queryKey := range queryKeys {
		contain := false
		// default = nil，则全量通过；
		if defaultColumns == nil {
			contain = true
		} else {
			for _, column := range defaultColumns {
				if column == queryKey {
					contain = true
					break
				}
			}
		}
		// 若不包含在queryKeys内，则也赋值为0
		if !contain {
			columnSQLs = append(columnSQLs, "0")
		} else {
			var columnSQL string
			columnSQL, err = getColumnSQL(queryKey)
			if err != nil {
				log.Error("%v", err)
				return
			}
			columnSQLs = append(columnSQLs, columnSQL)
		}
	}
	querySQL = fmt.Sprintf("%s AS command, %s", queryColumn, strings.Join(columnSQLs, ","))
	return
}

func parseCommands(commands []string) (args []interface{}, sqls []string) {
	if len(commands) > 0 {
		var sqlsTmp []string
		for _, command := range commands {
			sqlsTmp = append(sqlsTmp, "?")
			args = append(args, command)
		}
		sqls = append(sqls, fmt.Sprintf("command IN(%v)", strings.Join(sqlsTmp, ",")))
	}
	return
}

func getTailSQL(matchOption *apmmdl.MatchOption, column string) (sql string) {
	var (
		groupBy string
		limit   string
		orderBy string
	)
	if column != "" {
		groupBy = "GROUP BY " + column
	}
	if matchOption.OrderBy != "" {
		orderBy = "ORDER BY " + matchOption.OrderBy
	}
	if matchOption.Pn != 0 && matchOption.Ps != 0 {
		limit = "LIMIT " + strconv.Itoa((matchOption.Pn-1)*matchOption.Ps) + "," + strconv.Itoa(matchOption.Ps)
	} else if matchOption.Limit != 0 {
		limit = "LIMIT " + strconv.Itoa(matchOption.Limit)
	}
	sql = fmt.Sprintf("%s %s %s", groupBy, orderBy, limit)
	return
}

func (d *Dao) ApmMoniCount(c context.Context, database, tableName string, matchOption *apmmdl.MatchOption) (count int64, err error) {
	var countStr string
	if countStr, err = d.ApmMoniAggregateCalculate(c, "count", database, tableName, matchOption); err != nil {
		log.Error("ApmMoniCount d.ApmMoniAggregateCalculate error(%v)", err)
		return
	}
	if count, err = strconv.ParseInt(countStr, 10, 64); err != nil {
		log.Error("ApmMoniCount strconv.ParseInt error(%v)", err)
	}
	return
}

func (d *Dao) ApmMoniDistinctCount(c context.Context, fieldName, database, tableName string, matchOption *apmmdl.MatchOption) (count int64, err error) {
	var countStr string
	if countStr, err = d.ApmMoniAggregateCalculate(c, fmt.Sprintf("distinct_count@%s", fieldName), database, tableName, matchOption); err != nil {
		log.Error("ApmMoniDistinctCount d.ApmMoniAggregateCalculate error(%v)", err)
		return
	}
	if count, err = strconv.ParseInt(countStr, 10, 64); err != nil {
		log.Error("ApmMoniDistinctCount strconv.ParseInt error(%v)", err)
	}
	return
}

/**
 * -------------- 基础组件 --------------
 */

// ApmMoniAggregateCalculate ClickHouse聚合函数查询
func (d *Dao) ApmMoniAggregateCalculate(c context.Context, cType, database, tableName string, matchOption *apmmdl.MatchOption) (value string, err error) {
	var (
		args, matchArgs []interface{}
		condition       string
	)
	columnSQLs, err := getColumnSQL(cType)
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	db := fmt.Sprintf("%s.%s", database, tableName)
	if condition, matchArgs, err = parseMatchOptionV2(matchOption); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if condition != "" {
		condition = fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, matchArgs...)
	sql := fmt.Sprintf(`SELECT %v FROM %v %v`, columnSQLs, db, condition)
	var row *clickSql.Row
	row = d.clickhouse.QueryRow(c, sql, args...)
	log.Infoc(c, "clickhouse sql: %v, args: %v", sql, args)
	if err = row.Scan(&value); err != nil {
		log.Errorc(c, "%v", err)
	}
	return
}
