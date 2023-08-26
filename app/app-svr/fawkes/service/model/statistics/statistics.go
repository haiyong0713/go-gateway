package statistics

import (
	"encoding/json"
	"errors"

	"go-gateway/app/app-svr/fawkes/service/model/apm"
)

const (
	CIBUILD                        = 1 // CI构建
	CIJOB                          = 2 // Pipline Job
	LASER                          = 3 // laser日志统计
	CICOMPILE                      = 4 // CI编译耗时
	TECHNOLOGYSTORAGE              = 5 // 技术埋点存储
	TECHNOLOGYSTORAGE_UNREGISTERED = 6 // 未注册的技术埋点存储
	TECHNOLOGYQUANTITY             = 7 // 技术埋点写入监测
)

type FawkesMatchOption struct {
	EventID   int64  `form:"event_id"`
	ClassType string `form:"class_type"`
	Column    string `form:"column"`
	FilterStr string `form:"filters"`
	Filters   []*apm.Filter

	// 查询算子关键字C
	QueryKeys string `form:"query_keys"`

	// 基础字段
	AppKey                 string `form:"app_key"`
	StartTime              int64  `form:"start_time"`
	EndTime                int64  `form:"end_time"`
	In                     string `form:"in"`
	OrderBy                string `form:"order_by"`
	Limit                  int    `form:"limit"`
	PkgType                int    `form:"pkg_type"`
	Status                 int    `form:"status"`
	CIEnv                  string `form:"ci_env"`
	JobName                string `form:"job_name"`
	PipelineID             int64  `form:"pipeline_id"`
	JobStage               string `form:"stage"`
	JobStatus              int    `form:"job_status"`
	JobID                  int64  `form:"job_id"`
	JobURL                 string `form:"job_url"`
	JobTagList             string `form:"tag_list"`
	JobRunnerInfo          string `form:"runner_info"`
	OptimizeLevel          int    `form:"optimize_level"`
	LaserSilenceStatus     int    `form:"silence_status"`
	LaserParseStatus       int    `form:"parse_status"`
	Operator               string `form:"operator"`
	BuildEnv               int    `form:"build_env"`
	BuildLogURL            string `form:"build_log_url"`
	BusId                  int64  `form:"bus_id"`
	CreatedDatacenterAppId int64  `form:"created_datacenter_app_id"`
	DatacenterAppId        int64  `form:"datacenter_app_id"`
	TechnologyOwner        string `form:"owner"`
	TechnologyName         string `form:"event_name"`
	TechnologyTopic        string `form:"kafka_topic"`
	TechnologyHiveTable    string `form:"datacenter_dwd_table_name"`
}

func (matchOption *FawkesMatchOption) Check() (err error) {
	// event_id 为必传字段
	if matchOption.FilterStr != "" {
		if err = json.Unmarshal([]byte(matchOption.FilterStr), &matchOption.Filters); err != nil {
			return err
		}
	}
	if matchOption.EventID == 0 {
		err = errors.New("event_id 不能为空")
		return
	}
	// 开始时间不能为空
	if matchOption.StartTime == 0 {
		err = errors.New("start_time异常")
		return
	}
	return
}

// ci构建统计模型
type FawkesMoni struct {
	Title string  `json:"title"`
	Value float64 `json:"value"`
}

// 网络基础数据
type CIBuildPackInfo struct {
	Command            string  `json:"command"`
	Count              int64   `json:"count"`
	CIBuildFailedRate  float64 `json:"ci_build_failed_rate"`
	CIBuildSuccessRate float64 `json:"ci_build_success_rate"`
	CIBuildTimeAvg     float64 `json:"ci_build_time_avg"`
}

type CIJobPackInfo struct {
	Command          string  `json:"command"`
	Count            int64   `json:"count"`
	CIJobFailedRate  float64 `json:"ci_job_failed_rate"`
	CIJobSuccessRate float64 `json:"ci_job_success_rate"`
	CIJobTimeAvg     float64 `json:"ci_job_time_avg"`
}

type LaserStt struct {
	Command             string  `json:"command"`
	Count               int64   `json:"count"`
	LaserFailedRate     float64 `json:"laser_failed_rate"`
	LaserSuccessRate    float64 `json:"laser_success_rate"`
	LaserAllSuccessRate float64 `json:"laser_all_success_rate"`
}

type CICompileStt struct {
	Command              string  `json:"command"`
	Count                int64   `json:"count"`
	FailedRate           float64 `json:"ci_compile_failed_rate"`
	SuccessRate          float64 `json:"ci_compile_success_rate"`
	TimeAvg              float64 `json:"ci_compile_time_avg"`
	StepsCount           int64   `json:"sum_steps_count"`
	UptodateCount        int64   `json:"sum_uptodate_count"`
	ExecutedCount        int64   `json:"sum_executed_count"`
	CacheCount           int64   `json:"sum_cache_count"`
	FastTotal            int64   `json:"sum_fast_total"`
	FastRemote           int64   `json:"sum_fast_remote"`
	FastLocal            int64   `json:"sum_fast_local"`
	AfterSyncTaskAvg     float64 `json:"avg_after_sync_task"`
	BuildSourceLocalAvg  float64 `json:"avg_build_source_local"`
	BuildSourceRemoteAvg float64 `json:"avg_build_source_remote"`
}

type TechnologyInfo struct {
	Command string `json:"command"`
	Cnt     int64  `json:"sum_cnt"`
	Size    int64  `json:"sum_part_real_size"`
}

type TechnologyQuantityInfo struct {
	DatacenterEventName string  `json:"datacenter_event_name"`
	DatacenterAppId     string  `json:"datacenter_app_id"`
	DatacenterPlatform  string  `json:"datacenter_platform"`
	CntMax              int64   `json:"max_cnt"`
	CntMin              int64   `json:"min_cnt"`
	CntAvg              float64 `json:"avg_cnt"`
}
