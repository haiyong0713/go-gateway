package apm

import "time"

type Env string

const (
	EnvTest = Env("test")
	EnvProd = Env("prod")
)

const (
	AlertTypeUncategorized = -1
	AlertTypeReal          = 1
	AlertTypeFalse         = 2
	AlertTypeError         = 3
	AlertTypeTest          = 4

	AlertStatusUnresolved = -1
	AlertStatusResolved   = 1
)

func (v Env) Valid() bool {
	return v == EnvTest || v == EnvProd
}

type Alert struct {
	Id           int64      `json:"id"`
	RuleId       int64      `json:"rule_id"`
	Env          Env        `json:"env"`
	Rule         *AlertRule `json:"rule"`
	AlertMd5     string     `json:"alert_md5"`
	AppKey       string     `json:"app_key"`
	Type         int8       `json:"alert_type"`
	Status       int8       `json:"alert_status"`
	Description  string     `json:"description"`
	Duration     int64      `json:"duration"`
	Labels       string     `json:"labels"`
	TriggerValue string     `json:"trigger_value"`
	Operator     string     `json:"operator"`
	StartTime    time.Time  `json:"start_time"`
	CTime        time.Time  `json:"ctime"`
	MTime        time.Time  `json:"mtime"`
}

type AlertAddReq struct {
	RuleId       int64       `json:"rule_id" form:"rule_id" validate:"required"`
	AlertMd5     string      `json:"alert_md5" form:"alert_md5" validate:"required"`
	Env          Env         `json:"env" form:"env"`
	Status       int8        `json:"status" form:"status"`
	Duration     int64       `json:"duration" form:"duration"`
	Labels       string      `json:"labels" form:"labels"`
	TriggerValue interface{} `json:"trigger_value" form:"trigger_value"`
	Operator     string      `json:"operator" form:"operator"`
	StartTime    time.Time   `json:"start_time" form:"start_time"`
}

type AlertIndicatorReq struct {
	AppKey    string `json:"app_key" form:"app_key"`
	Env       Env    `json:"env" form:"env"`
	StartTime int64  `json:"start_time" form:"start_time"`
	EndTime   int64  `json:"end_time" form:"end_time"`
	RuleId    int64  `json:"rule_id" form:"rule_id"`
	Status    int8   `json:"status" form:"status"`
	Type      int8   `json:"alert_type" form:"alert_type"`
}

type AlertListReq struct {
	Env       Env    `json:"env" form:"env"`
	StartTime int64  `json:"start_time" form:"start_time"`
	EndTime   int64  `json:"end_time" form:"end_time"`
	AppKey    string `json:"app_key" form:"app_key"`
	AlertMd5  string `json:"alert_md5" form:"alert_md5"`
	Type      int8   `json:"type" form:"type"`
	RuleId    int64  `json:"rule_id" form:"rule_id"`
	Status    int8   `json:"status" form:"status"`
	Pn        int    `json:"pn" form:"pn" default:"1"`
	Ps        int    `json:"ps" form:"ps" default:"20"`
}

type AlertUpdateReq struct {
	Id          int64  `json:"id" form:"id" validate:"required"`
	Type        int8   `json:"type" form:"type"`
	Status      int8   `json:"status" form:"status"`
	Description string `json:"description" form:"description"`
	Operator    string `json:"operator" form:"operator"`
}

type AlertRes struct {
	PageInfo *Page    `json:"page,omitempty"`
	Items    []*Alert `json:"items,omitempty"`
}

type AlertIndicator struct {
	ErrorAlertRate         float64 `json:"error_alert_rate"`
	FalseAlertRate         float64 `json:"false_alert_rate"`
	UncategorizedAlertRate float64 `json:"uncategorized_alert_rate"`
	Quality                float64 `json:"quality"`
}

// AlertReasonConfig apm_alert_reason_config apm告警根因配置表
type AlertReasonConfig struct {
	Id                 int64     `json:"id"`
	RuleId             int64     `json:"rule_id" validate:"required"`
	EventId            int64     `json:"event_id"`
	QueryType          string    `json:"query_type"`
	QuerySql           string    `json:"query_sql"`
	QueryCondition     string    `json:"query_condition"`
	ImpactFactorFields string    `json:"impact_factor_fields"`
	Description        string    `json:"description"`
	Operator           string    `json:"operator"`
	CTime              time.Time `json:"ctime"`
	MTime              time.Time `json:"mtime"`
}

// AlertReasonConfigReq 告警根因配置查询 req
type AlertReasonConfigReq struct {
	RuleId int64 `json:"rule_id" form:"rule_id"`
}

type AlertReasonField struct {
	FieldKey   string `json:"field_key"`
	FieldName  string `json:"field_name"`
	IsRequired int8   `json:"is_required"`
	FieldType  int8   `json:"field_type"`
}

// AlertReasonConfigResp 告警根因配置查询 resp
type AlertReasonConfigResp struct {
	Items []*AlertReasonConfigItem `json:"items"`
}

type AlertReasonConfigItem struct {
	Id                   int64               `json:"id"`
	RuleId               int64               `json:"rule_id" validate:"required"`
	EventId              int64               `json:"event_id"`
	EventName            string              `json:"event_name"`
	Databases            string              `json:"db_name"`
	DistributedTableName string              `json:"distributed_table_name"`
	QueryType            string              `json:"query_type"`
	QuerySql             string              `json:"query_sql"`
	QueryCondition       string              `json:"query_condition"`
	ImpactFactorFields   []*AlertReasonField `json:"impact_factor_fields"`
	Description          string              `json:"description"`
	Operator             string              `json:"operator"`
	CTime                time.Time           `json:"ctime"`
	MTime                time.Time           `json:"mtime"`
}

type AlertReasonConfigSet struct {
	EventId            int64  `json:"event_id"`
	QueryType          string `json:"query_type"`
	QueryCondition     string `json:"query_condition"`
	ImpactFactorFields string `json:"impact_factor_fields"`
}

// AlertReasonConfigSetReq 告警根因配置Set req
type AlertReasonConfigSetReq struct {
	RuleId   int64                   `json:"rule_id" validate:"required"`
	Operator string                  `json:"operator"`
	Configs  []*AlertReasonConfigSet `json:"configs"`
}

// AlertReasonConfigAddReq 告警根因配置add req
type AlertReasonConfigAddReq struct {
	RuleId             int64  `json:"rule_id"`
	EventId            int64  `json:"event_id"`
	QueryType          string `json:"query_type"`
	CustomQuerySql     string `json:"custom_query_sql"`
	QueryCondition     string `json:"query_condition"`
	ImpactFactorFields string `json:"impact_factor_fields"`
	Description        string `json:"description"`
	Operator           string `json:"operator"`
}

// AlertReasonConfigUpdateReq 告警根因配置update req
type AlertReasonConfigUpdateReq struct {
	Id                 int64  `json:"id"`
	EventId            int64  `json:"event_id"`
	QueryType          string `json:"query_type"`
	CustomQuerySql     string `json:"custom_query_sql"`
	QueryCondition     string `json:"query_condition"`
	ImpactFactorFields string `json:"impact_factor_fields"`
	Description        string `json:"description"`
	Operator           string `json:"operator"`
}

// AlertReasonConfigDeleteReq 告警根因配置update req
type AlertReasonConfigDeleteReq struct {
	Id int64 `json:"id"`
}
