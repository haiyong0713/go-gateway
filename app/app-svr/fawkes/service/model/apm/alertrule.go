package apm

const (
	AlertRuleTypeMajor  = 0 // 主规则
	AlertRuleTypeAdjust = 1 // 微调规则
)

type AlertRuleListReq struct {
	HawkeyeId  int64  `json:"hawkeye_id" form:"hawkeye_id"`
	MetricName string `json:"metric_name" form:"metric_name"`
	Name       string `json:"name" form:"name"`
	Species    string `json:"species" form:"species"`
	Pn         int    `json:"pn" form:"pn" default:"1"`
	Ps         int    `json:"ps" form:"ps" default:"20"`
}

type AlertRuleSetReq struct {
	HawkeyeId        int64  `json:"hawkeye_id" form:"hawkeye_id"`
	HawkeyeAdjustId  int64  `json:"hawkeye_adjust_id" form:"hawkeye_adjust_id"`
	Name             string `json:"name" form:"name"`
	TriggerCondition string `json:"trigger_condition" form:"trigger_condition"`
	Species          string `json:"species" form:"species"`
	QueryExprs       string `json:"query_exprs" form:"query_exprs"`
	Markdown         string `json:"markdown" form:"markdown"`
	RuleType         int8   `json:"rule_type" form:"rule_type"`
	Operator         string `json:"operator" form:"operator"`
}

type AlertRule struct {
	Id               int64        `json:"id"`
	HawkeyeId        int64        `json:"hawkeye_id"`
	Name             string       `json:"name"`
	TriggerCondition string       `json:"trigger_condition"`
	Species          string       `json:"species"`
	QueryExprs       string       `json:"query_exprs"`
	Markdown         string       `json:"markdown"`
	RuleType         int8         `json:"rule_type"`
	AdjustRule       []*AlertRule `json:"adjust_rule,omitempty"`
	Operator         string       `json:"operator"`
	CTime            int64        `json:"ctime"`
	MTime            int64        `json:"mtime"`
}

type AlertRuleRel struct {
	Id           int64  `json:"id"`
	RuleId       int64  `json:"rule_id"`
	AdjustRuleId int64  `json:"adjust_rule_id"`
	Operator     string `json:"operator"`
	CTime        int64  `json:"ctime"`
	MTime        int64  `json:"mtime"`
}

type AlertRuleRes struct {
	PageInfo *Page        `json:"page,omitempty"`
	Items    []*AlertRule `json:"items,omitempty"`
}
