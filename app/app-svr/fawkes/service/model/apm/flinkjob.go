package apm

import (
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/model"
)

const (
	// FlinkJob状态
	FlinkJobDel     = -1
	FlinkJobAdd     = 1
	FlinkJobModify  = 2
	FlinkJobPublish = 3
)

// FlinkJobDB struct
type FlinkJobDB struct {
	ID          int64      `json:"id" form:"id"`
	LogID       string     `json:"log_id" form:"log_id"`
	Name        string     `json:"name" form:"name"`
	Description string     `json:"description" form:"description"`
	Owner       string     `json:"owner" form:"owner"`
	Operator    string     `json:"operator" form:"operator"`
	State       int        `json:"state" form:"state"`
	CTime       xtime.Time `json:"ctime" form:"ctime"`
	MTime       xtime.Time `json:"mtime" form:"mtime"`
	ModifyCount int64      `json:"modify_count" form:"modify_count"`
}

// FlinkJobReq struct
type FlinkJobReq struct {
	ID          int64      `json:"id" form:"id"`
	LogID       string     `json:"log_id" form:"log_id"`
	Name        string     `json:"name" form:"name"`
	Description string     `json:"description" form:"description"`
	Owner       string     `json:"owner" form:"owner"`
	Operator    string     `json:"operator" form:"operator"`
	State       int        `json:"state" form:"state"`
	StartTime   xtime.Time `json:"start_time" form:"start_time"`
	EndTime     xtime.Time `json:"end_time" form:"end_time"`
	Pn          int        `json:"pn" form:"pn"`
	Ps          int        `json:"ps" form:"ps"`
}

// FlinkJobRes struct
type FlinkJobRes struct {
	PageInfo *model.PageInfo `json:"page,omitempty"`
	Items    []*FlinkJobDB   `json:"items,omitempty"`
}

// EventFlinkRelReq	struct
type EventFlinkRelDB struct {
	ID          int64      `json:"id" form:"id"`
	EventID     int64      `json:"event_id" form:"event_id"`
	JobID       int64      `json:"flink_job_id" form:"flink_job_id"`
	Description string     `json:"description" form:"description"`
	Operator    string     `json:"operator" form:"operator"`
	CTime       xtime.Time `json:"ctime" form:"ctime"`
	MTime       xtime.Time `json:"mtime" form:"mtime"`
	State       int        `json:"state" form:"state"`
}

// EventFlinkRelReq	struct
type EventFlinkRelReq struct {
	ID          int64  `json:"id" form:"id"`
	EventID     int64  `json:"event_id" form:"event_id"`
	JobID       int64  `json:"flink_job_id" form:"flink_job_id"`
	Description string `json:"description" form:"description"`
	Operator    string `json:"operator" form:"operator"`
}

// EventFlinkPublishDiff struct
type EventFlinkPublishDiff struct {
	CurVersion     string `json:"cur_version"`
	HistoryVersion string `json:"history_version"`
}

// EventFieldSub	struct
type EventFieldSub struct {
	DefaultValue interface{} `json:"default_value" form:"default_value"`
	PropertyName string      `json:"property_name" form:"property_name"`
}

// EventSubRes	struct
type EventSubRes struct {
	DBName     string           `json:"database" form:"database"`
	TableName  string           `json:"table" form:"table"`
	SampleRate int              `json:"sample_rate,omitempty" form:"sample_rate"`
	Properties []*EventFieldSub `json:"properties" form:"properties"`
}

// EventFlinkRelPublish struct
type EventFlinkRelPublish struct {
	ID          int64      `json:"id" form:"id"`
	FlinkJobID  int64      `json:"flink_job_id" form:"flink_job_id"`
	MD5         string     `json:"md5" form:"md5"`
	LocalPath   string     `json:"local_path" form:"local_path"`
	LocalUrl    string     `json:"local_url" form:"local_url"`
	Description string     `json:"description" form:"description"`
	Operator    string     `json:"operator" form:"operator"`
	CTime       xtime.Time `json:"ctime" form:"ctime"`
	MTime       xtime.Time `json:"mtime" form:"mtime"`
}

// EventFlinkRelPublishListReq struct
type EventFlinkRelPublishListReq struct {
	FlinkJobID int64 `json:"flink_job_id" form:"flink_job_id"`
	Pn         int   `json:"pn" form:"pn" default:"1"`
	Ps         int   `json:"ps" form:"ps" default:"1"`
}

// EventFlinkRelPublishListRes struct
type EventFlinkRelPublishListRes struct {
	Items    []*EventFlinkRelPublish `json:"item" form:"item"`
	PageInfo *model.PageInfo         `json:"page" form:"page"`
}

// EventDuplicateRes struct
type EventDuplicateRes struct {
	EventIDS    []int64  `json:"event_ids" form:"event_ids"`
	DBName      string   `json:"db_name" form:"db_name"`
	TableName   string   `json:"table_name" form:"table_name"`
	Names       []string `json:"names" form:"names"`
	SampleRates []int    `json:"sample_rates" form:"sample_rates"`
	EventCount  int      `json:"event_count" form:"event_count"`
	MapName     string   `json:"map_name" form:"map_name"`
}

// JsonEvent struct
type JsonEvent struct {
	Mapping   map[string][]string     `json:"mapping" form:"mapping"`
	WideTable []string                `json:"wide_table" form:"wide_table"`
	Event     map[string]*EventSubRes `json:"event" form:"event"`
}
