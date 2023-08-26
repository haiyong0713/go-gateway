package fawkes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	statisticsmdl "go-gateway/app/app-svr/fawkes/service/model/statistics"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var (
	_BUILD_TIME_RANGE = "unix_timestamp(build_end_time)-unix_timestamp(build_start_time)"
)

var (
	_mysql_ciBuildTime_avg     = fmt.Sprintf("IFNULL(avg(IF(%s<0,0,IF(%s>86400, 0, %s))), 0) |opt= unix_timestamp(build_end_time)>0 AND unix_timestamp(build_start_time)>0", _BUILD_TIME_RANGE, _BUILD_TIME_RANGE, _BUILD_TIME_RANGE)
	_mysql_selectCount         = `count(*)`
	_mysql_selectSum           = `IFNULL(sum(%s),0)`
	_mysql_selectMax           = `IFNULL(max(%s),0)`
	_mysql_selectMin           = `IFNULL(min(%s),0)`
	_mysql_selectAvg           = `IFNULL(avg(NULLIF(%s,0)),0)`
	_mysql_ciBuildSuccessRate  = `IFNULL(sum(status=3)/sum(status in (-1,3)), 0)`
	_mysql_ciBuildFailedRate   = `IFNULL(sum(status=-1)/sum(status in (-1,3)), 0)`
	_mysql_ciJobTime_avg       = `IFNULL(avg(unix_timestamp(job_end_time)-unix_timestamp(job_start_time)), 0) |opt= unix_timestamp(job_end_time)>0 AND unix_timestamp(job_start_time)>0`
	_mysql_ciJobSuccessRate    = `IFNULL(sum(job_status=1)/sum(job_status in (1,2)), 0)`
	_mysql_ciJobFailedRate     = `IFNULL(sum(job_status=2)/sum(job_status in (1, 2)), 0)`
	_mysql_laserFailedRate     = `IFNULL(sum(status in (-1,-2))/sum(status in (-1,-2,3)), 0)`
	_mysql_laserSuccessRate    = `IFNULL(sum(status=3)/sum(status in (-1,-2,3)), 0)`
	_mysql_laserAllSuccessRate = `IFNULL(sum(status=3)/sum(status LIKE '%'), 0)`
	_mysql_ciCompileTime_avg   = `IFNULL(avg(unix_timestamp(end_time)-unix_timestamp(start_time)), 0) |opt= unix_timestamp(end_time)>0 AND unix_timestamp(start_time)>0`
	_mysql_compileSuccessRate  = `IFNULL(sum(status=1)/sum(status in (1,2)), 0)`
	_mysql_compileFailedRate   = `IFNULL(sum(status=2)/sum(status in (1,2)), 0)`
)

// EventID映射
func getFawkesTableName(eventID int64) (tableName string) {
	switch eventID {
	case statisticsmdl.CIBUILD:
		tableName = "build_pack"
	case statisticsmdl.CIJOB:
		tableName = "ci_job_time"
	case statisticsmdl.LASER:
		tableName = "app_laser"
	case statisticsmdl.CICOMPILE:
		tableName = "ci_compile_time"
	case statisticsmdl.TECHNOLOGYSTORAGE:
		tableName = "(apm_event_technology_storage INNER JOIN apm_event ON event_id=name)"
	case statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED:
		tableName = "apm_event_technology_storage"
	case statisticsmdl.TECHNOLOGYQUANTITY:
		tableName = "(apm_event_quantity_statistics INNER JOIN apm_event ON datacenter_event_name=name)"
	}
	return
}

// 根据类型选择、实际的逻辑SQL
func fawkesGetColumn(eventId int64, cType string) (column string, options string) {
	var (
		columnSQL  string
		optionsSQL string
	)
	switch cType {
	case "count":
		columnSQL = _mysql_selectCount
	case "ci_build_success_rate":
		columnSQL = _mysql_ciBuildSuccessRate
	case "ci_build_failed_rate":
		columnSQL = _mysql_ciBuildFailedRate
	case "ci_build_time_avg":
		comps := strings.Split(_mysql_ciBuildTime_avg, "|opt=")
		columnSQL = comps[0]
		optionsSQL = comps[1]
	case "ci_job_time_avg":
		comps := strings.Split(_mysql_ciJobTime_avg, "|opt=")
		columnSQL = comps[0]
		optionsSQL = comps[1]
	case "ci_job_success_rate":
		columnSQL = _mysql_ciJobSuccessRate
	case "ci_job_failed_rate":
		columnSQL = _mysql_ciJobFailedRate
	case "laser_failed_rate":
		columnSQL = _mysql_laserFailedRate
	case "laser_success_rate":
		columnSQL = _mysql_laserSuccessRate
	case "laser_all_success_rate":
		columnSQL = _mysql_laserAllSuccessRate
	case "ci_compile_time_avg":
		comps := strings.Split(_mysql_ciCompileTime_avg, "|opt=")
		columnSQL = comps[0]
		optionsSQL = comps[1]
	case "ci_compile_failed_rate":
		columnSQL = _mysql_compileFailedRate
	case "ci_compile_success_rate":
		columnSQL = _mysql_compileSuccessRate
	default:
		// 函数@表名.字段名
		columnSQL = getFawkesCustomColumnSQL(eventId, cType)
		cType = getFawkesCustomColumnName(cType)
	}
	cType = strings.ReplaceAll(cType, "@", "_")
	column = fmt.Sprintf("(%s) AS %s", columnSQL, cType)
	options = optionsSQL
	return
}

func getFawkesCustomColumnSQL(eventId int64, cType string) (columnSQL string) {
	cType = getColumnNameMapping(eventId, cType)
	comps := strings.Split(cType, "@")
	//nolint:gomnd
	if len(comps) != 2 {
		columnSQL = cType
		return
	}
	var queryKey = comps[0]
	var fieldParam = comps[1]
	switch {
	case strings.HasPrefix(queryKey, "sum"):
		columnSQL = fmt.Sprintf(_mysql_selectSum, fieldParam)
	case strings.HasPrefix(queryKey, "max"):
		columnSQL = fmt.Sprintf(_mysql_selectMax, fieldParam)
	case strings.HasPrefix(queryKey, "min"):
		columnSQL = fmt.Sprintf(_mysql_selectMin, fieldParam)
	case strings.HasPrefix(queryKey, "avg"):
		columnSQL = fmt.Sprintf(_mysql_selectAvg, fieldParam)
	default:
		columnSQL = fmt.Sprintf("%v(%v)", queryKey, fieldParam)
	}
	return
}

func getFawkesCustomColumnName(cType string) (columnName string) {
	comps := strings.Split(cType, ".")
	//nolint:gomnd
	if len(comps) != 2 {
		columnName = cType
		return
	}
	columnName = comps[1]
	return
}

// getColumnNameMapping 兼容前端的column传值
func getColumnNameMapping(eventId int64, cType string) (columnName string) {
	switch eventId {
	case statisticsmdl.TECHNOLOGYSTORAGE, statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED:
		switch cType {
		case "event_name":
			columnName = "apm_event_technology_storage.event_id"
		case "datacenter_app_id":
			columnName = "apm_event_technology_storage.datacenter_app_id"
		case "created_datacenter_app_id":
			columnName = "apm_event.datacenter_app_id"
		default:
			columnName = cType
		}
	default:
		columnName = cType
	}
	return
}

// getDateFormatColumn 时间字段格式化
func getDateFormatColumn(eventId int64) (column, querySql, dateFormat string) {
	switch eventId {
	case statisticsmdl.TECHNOLOGYSTORAGE, statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED, statisticsmdl.TECHNOLOGYQUANTITY:
		column = "log_date"
		querySql = "log_date"
		dateFormat = "FROM_UNIXTIME(?,'%Y%m%d')"
	default:
		column = "ctime"
		querySql = "DATE_FORMAT(ctime,'%Y%m%d')"
		dateFormat = "FROM_UNIXTIME(?,'%Y-%m-%d %H:%i:%s')"
	}
	return
}

func getDefaultColInfo(eventID int64) (defaultColumns []string) {
	switch eventID {
	case statisticsmdl.CIBUILD:
		defaultColumns = []string{
			"count",
			"ci_build_failed_rate",
			"ci_build_success_rate",
			"ci_build_time_avg",
		}
	case statisticsmdl.CIJOB:
		defaultColumns = []string{
			"count",
			"ci_job_failed_rate",
			"ci_job_success_rate",
			"ci_job_time_avg",
		}
	case statisticsmdl.LASER:
		defaultColumns = []string{
			"count",
			"laser_failed_rate",
			"laser_success_rate",
			"laser_all_success_rate",
		}
	case statisticsmdl.CICOMPILE:
		defaultColumns = []string{
			"count",
			"ci_compile_failed_rate",
			"ci_compile_success_rate",
			"ci_compile_time_avg",
			"sum@steps_count",
			"sum@uptodate_count",
			"sum@executed_count",
			"sum@cache_count",
			"sum@fast_total",
			"sum@fast_remote",
			"sum@fast_local",
			"avg@after_sync_task",
			"avg@build_source_local",
			"avg@build_source_remote",
		}
	case statisticsmdl.TECHNOLOGYSTORAGE, statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED:
		defaultColumns = []string{
			"sum@cnt",
			"sum@part_real_size",
		}
	case statisticsmdl.TECHNOLOGYQUANTITY:
		defaultColumns = []string{
			"max@cnt",
			"min@cnt",
			"avg@cnt",
		}
	}
	return
}

// 解析前端的条件，用于SQL WHERE条件
func parseFawkesMatchOptions(matchOptions *statisticsmdl.FawkesMatchOption) (args []string, sqls []interface{}) {
	if realValue, isEmpty := getRealValue(matchOptions.AppKey); !isEmpty {
		args = append(args, "app_key=?")
		sqls = append(sqls, realValue)
	}
	if matchOptions.PkgType != 0 {
		args = append(args, "pkg_type=?")
		sqls = append(sqls, matchOptions.PkgType)
	}
	if matchOptions.Status != 0 {
		args = append(args, "status=?")
		sqls = append(sqls, matchOptions.Status)
	}
	if matchOptions.CIEnv != "" {
		likeOrNot := "LIKE"
		sqlEnv := matchOptions.CIEnv
		if fChar := matchOptions.CIEnv[:1]; fChar == "!" {
			likeOrNot = "NOT LIKE"
			sqlEnv = matchOptions.CIEnv[1:]
		}
		args = append(args, fmt.Sprintf("ci_env_vars %v ?", likeOrNot))
		sqls = append(sqls, fmt.Sprintf("%%%v%%", sqlEnv))
	}
	if matchOptions.PipelineID != 0 {
		args = append(args, "pipeline_id=?")
		sqls = append(sqls, matchOptions.PipelineID)
	}
	if matchOptions.JobName != "" {
		args = append(args, "job_name=?")
		sqls = append(sqls, matchOptions.JobName)
	}
	if matchOptions.JobStage != "" {
		args = append(args, "stage=?")
		sqls = append(sqls, matchOptions.JobStage)
	}
	if matchOptions.JobStatus != 0 {
		args = append(args, "job_status=?")
		sqls = append(sqls, matchOptions.JobStatus)
	}
	if matchOptions.JobID != 0 {
		args = append(args, "job_id=?")
		sqls = append(sqls, matchOptions.JobID)
	}
	if matchOptions.JobURL != "" {
		args = append(args, "job_url=?")
		sqls = append(sqls, matchOptions.JobURL)
	}
	if matchOptions.JobTagList != "" {
		args = append(args, "tag_list=?")
		sqls = append(sqls, matchOptions.JobTagList)
	}
	if matchOptions.JobRunnerInfo != "" {
		args = append(args, "runner_info=?")
		sqls = append(sqls, matchOptions.JobRunnerInfo)
	}
	if matchOptions.LaserSilenceStatus != 0 {
		args = append(args, "silence_status=?")
		sqls = append(sqls, matchOptions.LaserSilenceStatus)
	}
	if matchOptions.LaserParseStatus != 0 {
		args = append(args, "parse_status=?")
		sqls = append(sqls, matchOptions.LaserParseStatus)
	}
	if matchOptions.Operator != "" {
		args = append(args, "operator LIKE ?")
		sqls = append(sqls, "%"+matchOptions.Operator+"%")
	}
	if matchOptions.BuildEnv != 0 {
		args = append(args, "build_env=?")
		sqls = append(sqls, matchOptions.BuildEnv)
	}
	if matchOptions.BuildLogURL != "" {
		args = append(args, "build_log_url=?")
		sqls = append(sqls, matchOptions.BuildLogURL)
	}
	if matchOptions.EventID == statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED {
		args = append(args, "event_id NOT IN(SELECT name FROM apm_event)")
	}
	if matchOptions.TechnologyName != "" {
		args = append(args, "apm_event_technology_storage.event_id=?")
		sqls = append(sqls, matchOptions.TechnologyName)
	}
	if matchOptions.BusId != 0 {
		args = append(args, "bus_id=?")
		sqls = append(sqls, matchOptions.BusId)
	}
	if matchOptions.TechnologyTopic != "" {
		args = append(args, "kafka_topic=?")
		sqls = append(sqls, matchOptions.TechnologyTopic)
	}
	if matchOptions.DatacenterAppId != 0 {
		args = append(args, "apm_event_technology_storage.datacenter_app_id=?")
		sqls = append(sqls, matchOptions.DatacenterAppId)
	}
	if matchOptions.CreatedDatacenterAppId != 0 {
		args = append(args, "apm_event.datacenter_app_id=?")
		sqls = append(sqls, matchOptions.CreatedDatacenterAppId)
	}
	if matchOptions.TechnologyHiveTable != "" {
		args = append(args, "datacenter_dwd_table_name=?")
		sqls = append(sqls, matchOptions.TechnologyHiveTable)
	}
	if matchOptions.TechnologyOwner != "" {
		args = append(args, "owner=?")
		sqls = append(sqls, matchOptions.TechnologyOwner)
	}
	if matchOptions.OptimizeLevel != 0 {
		args = append(args, "optimize_level=?")
		sqls = append(sqls, matchOptions.OptimizeLevel)
	}
	if matchOptions.StartTime != 0 {
		// FROM_UNIXTIME
		column, _, dateFormat := getDateFormatColumn(matchOptions.EventID)
		args = append(args, fmt.Sprintf("%s > %s", column, dateFormat))
		sqls = append(sqls, matchOptions.StartTime/1000)
	}
	if matchOptions.EndTime != 0 {
		column, _, dateFormat := getDateFormatColumn(matchOptions.EventID)
		args = append(args, fmt.Sprintf("%s < %s", column, dateFormat))
		sqls = append(sqls, matchOptions.EndTime/1000)
	}
	return
}

func parseFawkesMatchOptionsV2(matchOption *statisticsmdl.FawkesMatchOption) (condition string, args []interface{}, err error) {
	// 追加查询条件
	matchSqls, matchArgs := parseFawkesMatchOptions(matchOption)
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

func (d *Dao) FawkesMoniLine(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []*statisticsmdl.FawkesMoni, err error) {
	var (
		db               = getFawkesTableName(matchOption.EventID)
		_arg             = `SELECT %s as t, %s FROM %s GROUP BY t ORDER BY t`
		column1, column2 string
		options          string
		args             []interface{}
		conditions       []string
	)
	_, column1, _ = getDateFormatColumn(matchOption.EventID)
	column2, options = fawkesGetColumn(matchOption.EventID, matchOption.ClassType)
	var (
		condition string
		matchArgs []interface{}
	)
	if condition, matchArgs, err = parseFawkesMatchOptionsV2(matchOption); err != nil {
		log.Error("parseFawkesMatchOptionsV2 error %v", err)
		return
	}
	conditions = append(conditions, condition)
	if options != "" {
		conditions = append(conditions, options)
	}
	if len(conditions) > 0 {
		db += fmt.Sprintf(" WHERE %v", strings.Join(conditions, " AND "))
	}
	args = append(args, matchArgs...)
	querySql := fmt.Sprintf(_arg, column1, column2, db)
	log.Infoc(c, "FawkesMoniLine sql:\n %v\n args:\n %v", querySql, args)
	rows, err := d.db.Query(c, querySql, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.FawkesMoni{}
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
func (d *Dao) FawkesMoniPie(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []*statisticsmdl.FawkesMoni, err error) {
	var (
		db      = getFawkesTableName(matchOption.EventID)
		_arg    = `SELECT %s,%s FROM %s GROUP BY %s`
		column2 string
		options string
		args    []string
		sqls    []interface{}
	)
	column2, options = fawkesGetColumn(matchOption.EventID, matchOption.ClassType)
	if options != "" {
		args = append(args, options)
	}
	matchArgs, matchSqls := parseFawkesMatchOptions(matchOption)
	args = append(args, matchArgs...)
	sqls = append(sqls, matchSqls...)
	if len(args) > 0 {
		db += fmt.Sprintf(" WHERE %v", strings.Join(args, " AND "))
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_arg, matchOption.Column, column2, db, matchOption.Column), sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.FawkesMoni{}
		if err = rows.Scan(&re.Title, &re.Value); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func getFawkesInfoSQL(matchOption *statisticsmdl.FawkesMatchOption) (sqlStr string, args []interface{}) {
	var (
		db      = getFawkesTableName(matchOption.EventID)
		_arg    = `SELECT %s,%s FROM %s %s %s %s`
		column  string
		groupBy string
		orderBy string
		limit   string
	)
	// 默认可查询的列表
	defaultColumns := getDefaultColInfo(matchOption.EventID)
	// 前端请求需要查询的columns
	queryColumns := strings.Split(matchOption.QueryKeys, ",")
	// 构建SQL组
	var columnSQLs []string
	for _, column := range defaultColumns {
		contain := false
		for _, queryColumn := range queryColumns {
			if queryColumn == column {
				contain = true
				break
			}
		}
		if contain {
			realColumns, _ := fawkesGetColumn(matchOption.EventID, column)
			columnSQLs = append(columnSQLs, realColumns)
			// if options != "" {
			//	args = append(args, options)
			// }
		} else {
			columnSQLs = append(columnSQLs, "0")
		}
	}
	// WHERE 条件语句解析
	var (
		condition string
		matchArgs []interface{}
		err       error
	)
	if condition, matchArgs, err = parseFawkesMatchOptionsV2(matchOption); err != nil {
		log.Error("parseFawkesMatchOptionsV2 error %v", err)
		return
	}
	if condition != "" {
		db += fmt.Sprintf(" WHERE %v", condition)
	}
	args = append(args, matchArgs...)
	if matchOption.Column != "" {
		columns := strings.Split(matchOption.Column, ",")
		var groups, columnAlias []string
		for _, col := range columns {
			colConv, _ := fawkesGetColumn(matchOption.EventID, col)
			columnAlias = append(columnAlias, colConv)
			alias := strings.Split(colConv, "AS")
			groups = append(groups, alias[1])
		}
		column = strings.Join(columnAlias, ",")
		groupBy = fmt.Sprintf("GROUP BY %s", strings.Join(groups, ","))
	} else {
		column = "0"
	}
	if matchOption.OrderBy != "" {
		orderBy = "ORDER BY " + matchOption.OrderBy
	}
	if matchOption.Limit != 0 {
		limit = "LIMIT " + strconv.Itoa(matchOption.Limit)
	}
	sqlStr = fmt.Sprintf(_arg, column, strings.Join(columnSQLs, ","), db, groupBy, orderBy, limit)
	return
}

// CI 查询接口
func (d *Dao) CIInfoList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.CIBuildPackInfo{}
		if err = rows.Scan(&re.Command, &re.Count, &re.CIBuildFailedRate, &re.CIBuildSuccessRate, &re.CIBuildTimeAvg); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// CI JOB 查询接口
func (d *Dao) CIJobList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.CIJobPackInfo{}
		if err = rows.Scan(&re.Command, &re.Count, &re.CIJobFailedRate, &re.CIJobSuccessRate, &re.CIJobTimeAvg); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// laser 上报查询接口
func (d *Dao) SttLaserList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.LaserStt{}
		if err = rows.Scan(&re.Command, &re.Count, &re.LaserFailedRate, &re.LaserSuccessRate, &re.LaserAllSuccessRate); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// laser 上报查询接口
func (d *Dao) SttCICompileList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.CICompileStt{}
		if err = rows.Scan(&re.Command, &re.Count, &re.FailedRate, &re.SuccessRate, &re.TimeAvg, &re.StepsCount,
			&re.UptodateCount, &re.ExecutedCount, &re.CacheCount, &re.FastTotal, &re.FastRemote, &re.FastLocal, &re.AfterSyncTaskAvg, &re.BuildSourceLocalAvg, &re.BuildSourceRemoteAvg); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TechnologyInfoList 技术埋点数据统计接口
func (d *Dao) TechnologyInfoList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	log.Infoc(c, "TechnologyInfoList sql:\n %v\n args:\n %v", sqlStr, sqls)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.TechnologyInfo{}
		if err = rows.Scan(&re.Command, &re.Cnt, &re.Size); err != nil {
			log.Errorc(c, "rows.Scan error %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TechnologyQuantityInfoList 技术埋点检测数据统计接口
func (d *Dao) TechnologyQuantityInfoList(c context.Context, matchOption *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	sqlStr, sqls := getFawkesInfoSQL(matchOption)
	log.Infoc(c, "TechnologyInfoList sql:\n %v\n args:\n %v", sqlStr, sqls)
	rows, err := d.db.Query(c, sqlStr, sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &statisticsmdl.TechnologyQuantityInfo{}
		if err = rows.Scan(&re.DatacenterEventName, &re.DatacenterAppId, &re.DatacenterPlatform, &re.CntMax, &re.CntMin, &re.CntAvg); err != nil {
			log.Errorc(c, "rows.Scan error %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
