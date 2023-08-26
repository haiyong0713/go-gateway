package apm

import (
	"strings"
)

// 告警入参模型
type AlertWebhookParams struct {
	Product          string            `json:"product"`           // 产品名称
	Team             string            `json:"team"`              // 团队名称
	Alertname        string            `json:"alertname"`         // 告警规则名称	- 对应: 事件主题Subject
	Identity         string            `json:"identity"`          // 告警对象 	- 对应: 事件关联对象Identity）
	Description      string            `json:"description"`       // 告警描述
	Status           string            `json:"status"`            // 事件状态
	StartAt          string            `json:"start_time"`        // 首次触发时间 - 如果没有，按收到时间补全
	TriggerAt        string            `json:"trigger_time"`      // 本次触发时间 - 如果没有，按收到时间补全
	EndAt            string            `json:"end_time"`          // 告警结束时间
	Labels           map[string]string `json:"labels"`            // 用于标识一条事件的属
	Annotations      map[string]string `json:"annotations"`       // 其它属性，不用于标识事件
	TriggerCondition string            `json:"trigger_condition"` // 触发条件 	- 取自Annotations
	TriggerValue     float64           `json:"trigger_value"`     // 触发值		- 取自Annotations
	RuleID           int               `json:"rule_id"`           // 告警规则ID 	- 取自Labels
	RuleType         string            `json:"rule_type"`         // 告警规则类型 - 取自Labels
	Receivers        []string          `json:"receivers"`         // 事件的接收人
	Channels         []string          `json:"channels"`          // 事件的接收渠道
}

type EventAlertAddReq struct {
	Title                  string   `json:"title"`
	EventId                int64    `json:"event_id"`
	DatacenterAppId        int64    `json:"datacenter_app_id"`
	Description            string   `json:"description"`
	Intervals              int64    `json:"intervals"`
	TimeField              string   `json:"time_field"`
	Cluster                int8     `json:"cluster" validate:"max=3,min=1"`
	Level                  int8     `json:"level" validate:"max=4,min=1"`
	TimeFrame              int64    `json:"time_frame"`
	AggType                int8     `json:"agg_type" validate:"max=8,min=1"`
	AggField               string   `json:"agg_field"`
	AggPercentile          int8     `json:"agg_percentile"`
	FilterQuery            string   `json:"filter_query"`
	DenominatorFilterQuery string   `json:"denominator_filter_query"`
	TriggerCondition       string   `json:"trigger_condition"`
	MinLogCount            int64    `json:"min_log_count"`
	GroupField             string   `json:"group_field"`
	IsLogDetail            int8     `json:"is_log_detail"`
	NotifyFields           string   `json:"notify_fields"`
	NotifyDuration         int64    `json:"notify_duration"`
	Channels               string   `json:"channels"`
	Targets                []string `json:"targets"`
	BotWebhook             string   `json:"bot_webhook"`
	Webhook                string   `json:"webhook"`
	MuteType               int8     `json:"mute_type" validate:"max=2,min=1"`
	MutePeriod             string   `json:"mute_period"`
	Creator                string   `json:"creator"`
	Operator               string   `json:"operator"`
}

type EventAlertQueryReq struct {
	EventId         int64  `json:"event_id" form:"event_id"`
	DatacenterAppId int64  `json:"datacenter_app_id" form:"datacenter_app_id"`
	EventName       string `json:"event_name" form:"event_name"`
	Title           string `json:"title" form:"title"`
	IsEnable        int8   `json:"is_enable" form:"is_enable"`
	Level           int8   `json:"level" form:"level"`
	Pn              int    `json:"pn" form:"pn" default:"1"`
	Ps              int    `json:"ps" form:"ps" default:"50"`
}

type EventAlertUpdateReq struct {
	Id                     int64    `json:"id"`
	DatacenterAppId        int64    `json:"datacenter_app_id"`
	Title                  string   `json:"title"`
	Description            string   `json:"description"`
	Version                int64    `json:"version"`
	Intervals              int64    `json:"intervals"`
	TimeField              string   `json:"time_field"`
	Cluster                int8     `json:"cluster" validate:"max=3,min=1"`
	Level                  int8     `json:"level" validate:"max=4,min=1"`
	TimeFrame              int64    `json:"time_frame"`
	AggType                int8     `json:"agg_type" validate:"max=8,min=1"`
	AggField               string   `json:"agg_field"`
	AggPercentile          int8     `json:"agg_percentile"`
	FilterQuery            string   `json:"filter_query"`
	DenominatorFilterQuery string   `json:"denominator_filter_query"`
	TriggerCondition       string   `json:"trigger_condition"`
	MinLogCount            int64    `json:"min_log_count"`
	GroupField             string   `json:"group_field"`
	IsLogDetail            int8     `json:"is_log_detail"`
	NotifyFields           string   `json:"notify_fields"`
	NotifyDuration         int64    `json:"notify_duration"`
	Channels               string   `json:"channels"`
	Targets                []string `json:"targets"`
	BotWebhook             string   `json:"bot_webhook"`
	Webhook                string   `json:"webhook"`
	MuteType               int8     `json:"mute_type" validate:"max=2,min=1"`
	MutePeriod             string   `json:"mute_period"`
	Operator               string   `json:"operator"`
}

type EventAlertResp struct {
	Id                     int64    `json:"id"`
	EventId                int64    `json:"event_id"`
	DatacenterAppId        int64    `json:"datacenter_app_id"`
	EventName              string   `json:"event_name"`
	BillionId              int64    `json:"billion_id"`
	Title                  string   `json:"title"`
	Description            string   `json:"description"`
	Intervals              int64    `json:"intervals"`
	TimeField              string   `json:"time_field"`
	Cluster                int8     `json:"cluster"`
	Level                  int8     `json:"level"`
	TimeFrame              int64    `json:"time_frame"`
	AggType                int8     `json:"agg_type"`
	AggField               string   `json:"agg_field"`
	AggPercentile          int8     `json:"agg_percentile"`
	FilterQuery            string   `json:"filter_query"`
	DenominatorFilterQuery string   `json:"denominator_filter_query"`
	TriggerCondition       string   `json:"trigger_condition"`
	MinLogCount            int64    `json:"min_log_count"`
	GroupField             string   `json:"group_field"`
	IsLogDetail            int8     `json:"is_log_detail"`
	NotifyFields           string   `json:"notify_fields"`
	NotifyDuration         int64    `json:"notify_duration"`
	Channels               string   `json:"channels"`
	Targets                []string `json:"targets"`
	BotWebhook             string   `json:"bot_webhook"`
	Webhook                string   `json:"webhook"`
	MuteType               int8     `json:"mute_type"`
	MutePeriod             string   `json:"mute_period"`
	IsEnable               int8     `json:"is_enable"`
	Creator                string   `json:"creator"`
	Operator               string   `json:"operator"`
	CTime                  int64    `json:"ctime"`
	MTime                  int64    `json:"mtime"`
}

// EventAlertList struct.
type EventAlertList struct {
	PageInfo *Page             `json:"page"`
	Items    []*EventAlertResp `json:"items"`
}

type EventAlertDB struct {
	Id                     int64  `json:"id"`
	EventId                int64  `json:"event_id"`
	DatacenterAppId        int64  `json:"datacenter_app_id"`
	EventName              string `json:"event_name"`
	BillionId              int64  `json:"billion_id"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	Intervals              int64  `json:"intervals"`
	TimeField              string `json:"time_field"`
	Cluster                int8   `json:"cluster"`
	Level                  int8   `json:"level"`
	TimeFrame              int64  `json:"time_frame"`
	AggType                int8   `json:"agg_type"`
	AggField               string `json:"agg_field"`
	AggPercentile          int8   `json:"agg_percentile"`
	FilterQuery            string `json:"filter_query"`
	DenominatorFilterQuery string `json:"denominator_filter_query"`
	TriggerCondition       string `json:"trigger_condition"`
	MinLogCount            int64  `json:"min_log_count"`
	GroupField             string `json:"group_field"`
	IsLogDetail            int8   `json:"is_log_detail"`
	NotifyFields           string `json:"notify_fields"`
	NotifyDuration         int64  `json:"notify_duration"`
	Channels               string `json:"channels"`
	Targets                string `json:"targets"`
	BotWebhook             string `json:"bot_webhook"`
	Webhook                string `json:"webhook"`
	MuteType               int8   `json:"mute_type"`
	MutePeriod             string `json:"mute_period"`
	IsEnable               int8   `json:"is_enable"`
	Creator                string `json:"creator"`
	Operator               string `json:"operator"`
	Version                int64  `json:"version"`
	CTime                  int64  `json:"ctime"`
	MTime                  int64  `json:"mtime"`
}

type EventAlertBillionsReq struct {
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Schedule      *BillionsAlertSchedule `json:"schedule"`
	TimeField     string                 `json:"timeField"`
	AppId         string                 `json:"appId"`
	EsCluster     string                 `json:"esCluster"`
	Severity      string                 `json:"severity"`
	Search        *BillionsAlertSearch   `json:"search"`
	Trigger       *BillionsAlertTrigger  `json:"trigger"`
	Action        *BillionsAlertAction   `json:"action"`
	SchemaVersion int64                  `json:"schemaVersion,omitempty"`
}

type BillionsAlertSchedule struct {
	Interval string `json:"interval"`
}

type BillionsAlertSearch struct {
	SearchTimeframe    string              `json:"searchTimeframe"`
	AggregationType    string              `json:"aggregationType"`
	FieldToAggregateOn string              `json:"fieldToAggregateOn"`
	Percentile         int8                `json:"percentile,omitempty"`
	Query              *BillionsAlertQuery `json:"queryDefinition"`
	DenominatorQuery   *BillionsAlertQuery `json:"denominatorQueryDefinition,omitempty"`
	GroupBy            []string            `json:"groupBy,omitempty"`
}

type BillionsAlertQuery struct {
	Query string ` json:"query"`
}

type BillionsAlertTrigger struct {
	Operator        string `json:"operator"`
	Value           int64  `json:"value,omitempty"`
	RangeStart      int64  `json:"rangeStart,omitempty"`
	RangeEnd        int64  `json:"rangeEnd,omitempty"`
	AtLeastLogCount int64  `json:"atLeastLogCount,omitempty"`
}

type BillionsAlertAction struct {
	Alert *BillionsAlert `json:"alert"`
}

type BillionsAlert struct {
	CheckedIncludeLog      bool                   `json:"checkedIncludeLog"`
	ShouldIncludeAllFields bool                   `json:"shouldIncludeAllFields"`
	IncludeLogFields       []string               `json:"includeLogFields"`
	SuppressAlertDuration  string                 `json:"suppressAlertDuration"`
	AlertChannels          []string               `json:"alertChannels"`
	Targets                []*BillionsAlertTarget `json:"targets"`
	WechatRobotWebhookUrl  string                 `json:"wechatRobotWebhookUrl,omitempty"`
	WebhookUrl             string                 `json:"webhookUrl,omitempty"`
	MutePeriod             *BillionsAlertMute     `json:"mutePeriod"`
}
type BillionsAlertTarget struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type BillionsAlertMute struct {
	Type   string                   `json:"type"`
	Period *BillionsAlertMutePeriod `json:"period"`
}

type BillionsAlertMutePeriod struct {
	FromDay  int64  `json:"fromDay"`
	ToDay    int64  `json:"toDay"`
	FromTime string `json:"fromTime"`
	ToTime   string `json:"toTime"`
}

type BillionsAlertOpt struct {
	RuleId    []string `json:"ruleIdList"`
	Operation string   `json:"operation"`
}

func (db *EventAlertDB) Convert2Resp() (resp *EventAlertResp) {
	return &EventAlertResp{
		Id:                     db.Id,
		EventId:                db.EventId,
		DatacenterAppId:        db.DatacenterAppId,
		EventName:              db.EventName,
		BillionId:              db.BillionId,
		Title:                  db.Title,
		Level:                  db.Level,
		Description:            db.Description,
		Intervals:              db.Intervals,
		TimeFrame:              db.TimeFrame,
		TimeField:              db.TimeField,
		Cluster:                db.Cluster,
		AggType:                db.AggType,
		AggField:               db.AggField,
		AggPercentile:          db.AggPercentile,
		FilterQuery:            db.FilterQuery,
		DenominatorFilterQuery: db.DenominatorFilterQuery,
		TriggerCondition:       db.TriggerCondition,
		MinLogCount:            db.MinLogCount,
		GroupField:             db.GroupField,
		IsLogDetail:            db.IsLogDetail,
		NotifyFields:           db.NotifyFields,
		NotifyDuration:         db.NotifyDuration,
		Channels:               db.Channels,
		Targets:                strings.Split(db.Targets, "|"),
		BotWebhook:             db.BotWebhook,
		Webhook:                db.Webhook,
		MuteType:               db.MuteType,
		MutePeriod:             db.MutePeriod,
		IsEnable:               db.IsEnable,
		Creator:                db.Creator,
		Operator:               db.Operator,
		CTime:                  db.CTime,
		MTime:                  db.MTime,
	}
}

const (
	// billions alert trigger operator
	AlertTrigOptGte      = "gte"
	AlertTrigOptLt       = "lt"
	AlertTrigOptInRange  = "in_range"
	AlertTrigOptNotRange = "not_in_range"
	AlertTrigOptUp       = "relative_up"
	AlertTrigOptDown     = "relative_down"
	// billions alert switch
	EventAlertOn  = 1
	EventAlertOff = 0
	// billions alert init version
	AlertInitVersion = 0
	// billions alert operation
	AlertOptEnable  = "enable"
	AlertOptDisable = "disable"
	AlertOptDelete  = "delete"
)
