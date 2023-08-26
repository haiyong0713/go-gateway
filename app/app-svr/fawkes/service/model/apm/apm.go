package apm

import (
	xtime "go-common/library/time"
)

// Page struct.
type Page struct {
	Total    int `json:"total"`
	PageNum  int `json:"pn"`
	PageSize int `json:"ps"`
}

// 业务组
type Bus struct {
	ID                      int64  `json:"id"`
	AppKeys                 string `json:"app_keys"`
	Name                    string `json:"name"`
	Description             string `json:"description"`
	Owner                   string `json:"owner"`
	DatacenterBusKey        string `json:"datacenter_bus_key"`
	DatacenterDwdTableNames string `json:"datacenter_dwd_table_names"`
	Shared                  int8   `json:"shared"`
	Operator                string `json:"operator"`
	Ctime                   int64  `json:"ctime"`
	Mtime                   int64  `json:"mtime"`
}

// Command - 数据模型
type Command struct {
	ID       int64  `json:"id"`
	AppKey   string `json:"app_key"`
	GroupId  int64  `json:"group_id"`
	Command  string `json:"command"`
	Operator string `json:"operator"`
	Ctime    int64  `json:"ctime"`
	Mtime    int64  `json:"mtime"`
}

// Command事件组
type CommandGroup struct {
	ID          int64      `json:"id"`
	AppKey      string     `json:"app_key"`
	BusID       int64      `json:"bus_id"`
	EventID     int64      `json:"event_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Commands    []*Command `json:"commands"`
	Operator    string     `json:"operator"`
	BusName     string     `json:"bus_name"`
	State       int8       `json:"state"`
	Ctime       int64      `json:"ctime"`
	Mtime       int64      `json:"mtime"`
}

// Command事件组 - 高级选项
type CommandGroupAdvanced struct {
	ID          int64  `json:"id"`
	AppKey      string `json:"app_key"`
	EventID     int64  `json:"event_id"`
	FieldName   string `json:"field_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DisplayType int64  `json:"display_type"`
	Mapping     string `json:"mapping"`
	QueryType   string `json:"query_type"`
	Operator    string `json:"operator"`
	Ctime       int64  `json:"ctime"`
	Mtime       int64  `json:"mtime"`
}

// 监控事件veda配置项
type EventVedaConfig struct {
	ID             int64  `json:"id" form:"id"`
	EventID        int64  `json:"event_id"`
	EventName      string `json:"event_name"`
	VedaDBName     string `json:"veda_db_name"`
	VedaIndexTable string `json:"veda_index_table"`
	VedaStackTable string `json:"veda_stack_table"`
	HashColumn     string `json:"hash_column"`
	Ctime          int64  `json:"ctime"`
	Mtime          int64  `json:"mtime"`
}

/// Result Struct

// ResultBusList struct.
type ResultBusList struct {
	PageInfo *Page  `json:"page,omitempty"`
	Items    []*Bus `json:"items,omitempty"`
}

// ResultCommandGroupList struct.
type ResultCommandGroupList struct {
	PageInfo *Page           `json:"page,omitempty"`
	Items    []*CommandGroup `json:"items,omitempty"`
}

// ApmEventSetting
type ApmEventSetting struct {
	ID              int64  `json:"id"`
	AppKey          string `json:"app_key"`
	EventID         int64  `json:"event_id"`
	Name            string `json:"name"`
	SampleDesc      string `json:"sample_desc"`
	SampleConfigKey string `json:"sample_conf_key"`
	Ctime           int64  `json:"ctime"`
	Mtime           int64  `json:"mtime"`
}

// ApmDetailSetup
type ApmDetailSetup struct {
	TimeISO       int64  `json:"time_iso"`
	Mid           string `json:"mid"`
	Buvid         string `json:"buvid"`
	Model         string `json:"model"`
	Brand         string `json:"brand"`
	Osver         string `json:"osver"`
	AppKey        string `json:"app_key"`
	Version       string `json:"version"`
	VersionCode   string `json:"version_code"`
	Province      string `json:"province"`
	ISP           string `json:"isp"`
	FFVersion     string `json:"ff_version"`
	ConfigVersion string `json:"config_version"`
}

// CrashRule struct
type CrashRule struct {
	ID           int64      `json:"id" form:"id"`
	AppKeys      string     `json:"app_keys" form:"app_keys"`
	BusID        int64      `json:"bus_id" form:"bus_id"`
	BusName      string     `json:"bus_name" form:"bus_name"`
	RuleName     string     `json:"rule_name" form:"rule_name"`
	KeyWords     string     `json:"key_words" form:"key_words"`
	Operator     string     `json:"operator" form:"operator"`
	Ctime        xtime.Time `json:"ctime" form:"ctime"`
	Mtime        xtime.Time `json:"mtime" form:"mtime"`
	PageKeyWords string     `json:"page_key_words" form:"page_key_words"`
	Description  string     `json:"description" form:"description"`
}

// CrashRuleReq struct
type CrashRuleReq struct {
	ID           int64      `json:"id" form:"id"`
	AppKeys      string     `json:"app_keys" form:"app_keys"`
	BusID        int64      `json:"bus_id" form:"bus_id"`
	RuleName     string     `json:"rule_name" form:"rule_name"`
	KeyWords     string     `json:"key_words" form:"key_words"`
	Operator     string     `json:"operator" form:"operator"`
	Ctime        xtime.Time `json:"ctime" form:"ctime"`
	Mtime        xtime.Time `json:"mtime" form:"mtime"`
	Pn           int        `json:"pn" form:"pn"`
	Ps           int        `json:"ps" form:"ps"`
	PageKeyWords string     `json:"page_key_words" form:"page_key_words"`
	Description  string     `json:"description" form:"description"`
}

type CrashRuleRes struct {
	Items    []*CrashRule `json:"items"`
	PageInfo *Page        `json:"page"`
}
