package apm

import (
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/model"
)

// Prometheus const
const (
	PrometheusDel     = -1
	PrometheusAdd     = 1
	PrometheusModify  = 2
	PrometheusPublish = 3

	MetricStatusOn  = 1  // 指标生效
	MetricStatusOff = -1 // 指标未生效

	ActiveVersion   = 1 // 指标生效版本
	UnActiveVersion = 0 // 指标未生效版本
)

// PrometheusMetric struct
type PrometheusMetric struct {
	ID              int64      `json:"id" form:"id"`
	Metric          string     `json:"metric" form:"metric"`
	MetricType      string     `json:"metric_type" form:"metric_type"`
	ExecSQL         string     `json:"exec_sql" form:"exec_sql"`
	LabeledKeys     string     `json:"labeled_keys" form:"labeled_keys"`
	ValueKey        string     `json:"value_key" form:"value_key"`
	TimestampKey    string     `json:"timestamp_key" form:"timestamp_key"`
	Description     string     `json:"description" form:"description"`
	ApmDatabaseName string     `json:"apm_database_name" form:"apm_database_name"`
	ApmTableName    string     `json:"apm_table_name" form:"apm_table_name"`
	TimeFilter      int64      `json:"time_filter" form:"time_filter"`
	TimeOffset      int64      `json:"time_offset" form:"time_offset"`
	Operator        string     `json:"operator" form:"operator"`
	CTime           xtime.Time `json:"ctime" form:"ctime"`
	MTime           xtime.Time `json:"mtime" form:"mtime"`
	State           int8       `json:"state" form:"state"`
	Status          int8       `json:"status" form:"status" default:"-1"`
	URL             string     `json:"url" form:"url"`
	BusID           int64      `json:"bus_id" form:"bus_id"`
	BusName         string     `json:"bus_name" form:"bus_name"`
}

// PrometheusMetricListReq struct
type PrometheusMetricListReq struct {
	Metric          string `json:"metric" form:"metric"`
	ApmType         string `json:"metric_type" form:"metric_type"`
	ExecSQL         string `json:"exec_sql" form:"exec_sql"`
	LabeledKeys     string `json:"labeled_keys" form:"labeled_keys"`
	ValueKey        string `json:"value_key" form:"value_key"`
	TimestampKey    string `json:"timestamp_key" form:"timestamp_key"`
	Description     string `json:"description" form:"description"`
	ApmDatabaseName string `json:"apm_database_name" form:"apm_database_name"`
	ApmTableName    string `json:"apm_table_name" form:"apm_table_name"`
	TimeFilter      int64  `json:"time_filter" form:"time_filter"`
	TimeOffset      int64  `json:"time_offset" form:"time_offset"`
	Operator        string `json:"operator" form:"operator"`
	Pn              int    `json:"pn" form:"pn" default:"1" validate:"min=1"`
	Ps              int    `json:"ps" form:"ps" default:"20" validate:"min=1"`
	State           int8   `json:"state" form:"state"`
	Status          int8   `json:"status" form:"status"`
	BusID           int64  `json:"bus_id" form:"bus_id"`
}

// PrometheusMetricRes struct
type PrometheusMetricRes struct {
	PageInfo *model.PageInfo     `json:"page,omitempty"`
	Items    []*PrometheusMetric `json:"items,omitempty"`
	IsModify bool                `json:"is_modify"`
}

// PrometheusMetricPublishReq struct
type PrometheusMetricPublishReq struct {
	Operator    string `json:"operator" form:"operator"`
	Description string `json:"description" form:"description"`
}

// PrometheusMetricPublish struct
type PrometheusMetricPublish struct {
	ID              int64      `json:"id" form:"id"`
	MD5             string     `json:"md5" form:"md5"`
	LocalPath       string     `json:"local_path" form:"local_path"`
	LocalURL        string     `json:"local_url" form:"local_url"`
	Description     string     `json:"description" form:"description"`
	IsActiveVersion int8       `json:"is_active_version" form:"is_active_version"`
	Operator        string     `json:"operator" form:"operator"`
	CTime           xtime.Time `json:"ctime" form:"ctime"`
	MTime           xtime.Time `json:"mtime" form:"mtime"`
}

// PrometheusMetricPublishListReq struct
type PrometheusMetricPublishListReq struct {
	MD5         string     `json:"md5" form:"md5"`
	LocalPath   string     `json:"local_path" form:"local_path"`
	Description string     `json:"description" form:"description"`
	Operator    string     `json:"operator" form:"operator"`
	CTime       xtime.Time `json:"ctime" form:"ctime"`
	MTime       xtime.Time `json:"mtime" form:"mtime"`
	Pn          int        `json:"pn" form:"pn"`
	Ps          int        `json:"ps" form:"ps"`
}

// PrometheusMetricPublishListRes struct
type PrometheusMetricPublishListRes struct {
	PageInfo *model.PageInfo            `json:"page,omitempty"`
	Items    []*PrometheusMetricPublish `json:"items,omitempty"`
}

// PrometheusMetricPublishRollbackReq struct 告警指标回滚req
type PrometheusMetricPublishRollbackReq struct {
	Id int64 `json:"id" form:"id"`
}

// YmlMetricSelector struct
type YmlMetricSelector struct {
	Metric       string   `json:"metric" form:"metric" yaml:"metric"`
	ApmType      string   `json:"apm_type" form:"apm_type" yaml:"type"`
	Sql          string   `json:"sql" form:"sql" yaml:"sql"`
	LabeledKeys  []string `json:"labeled_keys" form:"labeled_keys" yaml:"labeled_keys"`
	ValueKey     string   `json:"value_key" form:"value_key" yaml:"value_key"`
	TimestampKey string   `json:"timestamp_key" form:"timestamp_key" yaml:"timestamp_key"`
	Help         string   `json:"help" form:"help" yaml:"help"`
}

// YmlMetricDatabase struct
type YmlMetricDatabase struct {
	Name      string               `json:"name" form:"name" yaml:"name"`
	Host      string               `json:"host" form:"host" yaml:"host"`
	Port      int64                `json:"port" form:"port" yaml:"port"`
	User      string               `json:"user" form:"user" yaml:"user"`
	Password  string               `json:"password" form:"password" yaml:"password"`
	Selectors []*YmlMetricSelector `json:"selectors" form:"selectors" yaml:"selectors"`
}

// YmlMetricConfig struct
type YmlMetricConfig struct {
	Databases []*YmlMetricDatabase `json:"databases" fom:"databases" yaml:"databases"`
}

type MetricDiff struct {
	CurVersion     string `json:"cur_version"`
	HistoryVersion string `json:"history_version"`
}
