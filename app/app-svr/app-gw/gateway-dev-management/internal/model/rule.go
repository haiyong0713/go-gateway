package model

type ConfigRuleReq struct {
	Team      string `form:"team" validate:"required"`
	Type      string `form:"type" validate:"required"`
	Interface string `form:"interface" validate:"required"`
	Code      string `form:"code" validate:"required"`
	Threshold int64  `form:"threshold" validate:"required"`
}

type CodeRule struct {
	Id     int64
	Team   string
	Type   string
	Method string
	Code   string
	RuleId int64
}

type FindRuleDetailReply struct {
	Code    int64           `json:"code"`
	Data    *FindRuleDetail `json:"data"`
	Message string          `json:"message"`
	TTL     int64           `json:"ttl"`
}

type FindRuleIdReply struct {
	Code int64 `json:"code"`
	Data struct {
		Items []*FindRuleDetail `json:"items"`
	} `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type FindRuleDetail struct {
	AlertRuleTemplateID int64  `json:"alert_rule_template_id"`
	Category            string `json:"category"`
	Channels            struct {
		Channels     []string `json:"channels"`
		WechatConfig struct {
			Mtype    string `json:"mtype"`
			Pipeline string `json:"pipeline"`
		} `json:"wechat_config"`
	} `json:"channels"`
	CreatedBy         string       `json:"created_by"`
	Ctime             string       `json:"ctime"`
	DeletedTime       string       `json:"deleted_time"`
	EditedBy          string       `json:"edited_by"`
	Enabled           bool         `json:"enabled"`
	ForDuration       int64        `json:"for_duration"`
	ID                int64        `json:"id"`
	IsDefaultTemplate bool         `json:"is_default_template"`
	Level             string       `json:"level"`
	MaxRepeatTimes    int64        `json:"max_repeat_times"`
	Mtime             string       `json:"mtime"`
	Name              string       `json:"name"`
	NotifyFormatName  string       `json:"notify_format_name"`
	Product           string       `json:"product"`
	Querys            []*RuleQuery `json:"querys"`
	Receivers         []string     `json:"receivers"`
	RepeatInterval    int64        `json:"repeat_interval"`
	Species           string       `json:"species"`
	SyncName          string       `json:"sync_name"`
	Team              string       `json:"team"`
	TriggerCondition  string       `json:"trigger_condition"`
	TriggerSummary    string       `json:"trigger_summary"`
	TriggerType       string       `json:"trigger_type"`
}

type RuleQuery struct {
	Desc     string `json:"desc"`
	Expr     string `json:"expr"`
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Scopes   []struct {
		EditableLevel int64 `json:"editable_level"`
		LabelMatchers []struct {
			MatcherType string   `json:"matcher_type"`
			Values      []string `json:"values"`
		} `json:"label_matchers"`
		LabelName   string `json:"label_name"`
		MainKey     bool   `json:"main_key"`
		PreviewExpr string `json:"preview_expr"`
	} `json:"scopes"`
	Threshold int64 `json:"threshold"`
}

type InsertRuleReq struct {
	Category string `json:"category"`
	Species  string `json:"species"`
	Channels struct {
		Channels     []string `json:"channels"`
		WechatConfig struct {
			Mtype    string `json:"mtype"`
			Pipeline string `json:"pipeline"`
		} `json:"wechat_config"`
	} `json:"channels"`
	Enabled          bool         `json:"enabled"`
	ForDuration      int64        `json:"for_duration"`
	Level            string       `json:"level"`
	Name             string       `json:"name"`
	NotifyFormatName string       `json:"notify_format_name"`
	Product          string       `json:"product"`
	Querys           []*RuleQuery `json:"querys"`
	Receivers        []string     `json:"receivers"`
	RepeatInterval   int64        `json:"repeat_interval"`
	Team             string       `json:"team"`
	TriggerSummary   string       `json:"trigger_summary"`
	TriggerType      string       `json:"trigger_type"`
}

type FindRuleReply struct {
	AlertRuleTemplateID interface{} `json:"alert_rule_template_id"`
	Category            string      `json:"category"`
	Channels            struct {
		Channels     []string `json:"channels"`
		WechatConfig struct {
			Mtype    string `json:"mtype"`
			Pipeline string `json:"pipeline"`
		} `json:"wechat_config"`
	} `json:"channels"`
	CreatedBy         string `json:"created_by"`
	Ctime             string `json:"ctime"`
	DeletedTime       string `json:"deleted_time"`
	EditedBy          string `json:"edited_by"`
	Enabled           bool   `json:"enabled"`
	ForDuration       int64  `json:"for_duration"`
	ID                int64  `json:"id"`
	IsDefaultTemplate bool   `json:"is_default_template"`
	Level             string `json:"level"`
	MaxRepeatTimes    int64  `json:"max_repeat_times"`
	Mtime             string `json:"mtime"`
	Name              string `json:"name"`
	NotifyFormatName  string `json:"notify_format_name"`
	Product           string `json:"product"`
	Querys            []struct {
		Desc     string `json:"desc"`
		Expr     string `json:"expr"`
		Name     string `json:"name"`
		Operator string `json:"operator"`
		Scopes   []struct {
			EditableLevel int64 `json:"editable_level"`
			LabelMatchers []struct {
				MatcherType string   `json:"matcher_type"`
				Values      []string `json:"values"`
			} `json:"label_matchers"`
			LabelName   string `json:"label_name"`
			MainKey     bool   `json:"main_key"`
			PreviewExpr string `json:"preview_expr"`
		} `json:"scopes"`
		Threshold int64 `json:"threshold"`
	} `json:"querys"`
	RepeatInterval   int64  `json:"repeat_interval"`
	Species          string `json:"species"`
	SyncName         string `json:"sync_name"`
	Team             string `json:"team"`
	TriggerCondition string `json:"trigger_condition"`
	TriggerSummary   string `json:"trigger_summary"`
	TriggerType      string `json:"trigger_type"`
}

type UpdateRuleReq struct {
	Category string `json:"category"`
	Channels struct {
		Channels     []string `json:"channels"`
		WechatConfig struct {
			Mtype    string `json:"mtype"`
			Pipeline string `json:"pipeline"`
		} `json:"wechat_config"`
	} `json:"channels"`
	Enabled          bool         `json:"enabled"`
	ForDuration      int64        `json:"for_duration"`
	Level            string       `json:"level"`
	Name             string       `json:"name"`
	NotifyFormatName string       `json:"notify_format_name"`
	Product          string       `json:"product"`
	Querys           []*RuleQuery `json:"querys"`
	Receivers        []string     `json:"receivers"`
	RepeatInterval   int64        `json:"repeat_interval"`
	Species          string       `json:"species"`
	Team             string       `json:"team"`
	TriggerSummary   string       `json:"trigger_summary"`
	TriggerType      string       `json:"trigger_type"`
}

type CustomizedTeamRuleReply struct {
	ID        int64  `json:"id"`
	Team      string `json:"team"`
	Type      string `json:"type"`
	Method    string `json:"method"`
	Code      int64  `json:"code"`
	Threshold int64  `json:"threshold"`
}

type RuleReply struct {
	Code    int64  `json:"code"`
	Data    string `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type ReceiverGroupReply struct {
	Code int64 `json:"code"`
	Data struct {
		Items []struct {
			ID     int64  `json:"id"`
			Name   string `json:"name"`
			Role   string `json:"role"`
			Status string `json:"status"`
			Team   string `json:"team"`
			Type   string `json:"type"`
		} `json:"items"`
		Total int64 `json:"total"`
	} `json:"data"`
	Message string `json:"message"`
	TTL     int64  `json:"ttl"`
}

type PutReceiverGroupReq struct {
	GroupIds []int64 `json:"group_ids"`
	Status   string  `json:"status"`
}

type MyServiceReply struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}
