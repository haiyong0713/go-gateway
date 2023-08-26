package anti_addiction

import (
	familymdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
	pushmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/push"
)

type RuleRly struct {
	Rules []*Rule `json:"rules"`
}

type Rule struct {
	Id         int64        `json:"id"`
	Version    string       `json:"version"`
	Frequency  int64        `json:"frequency"`
	Conditions []*Condition `json:"conditions"`
	Control    *Control     `json:"control"`
}

type Condition struct {
	Type   string           `json:"type"`
	Series *ConditionSeries `json:"series"`
}

// 连续使用时长
type ConditionSeries struct {
	MaxDuration int64 `json:"max_duration"`
	Interval    int64 `json:"interval"`
}

type Control struct {
	Type string       `json:"type"`
	Push *ControlPush `json:"push"`
}

// 应用内push
type ControlPush struct {
	// 不能使用 "git.bilibili.co/bapis/bapis-go/bilibili/broadcast/v1" 的PushMessageResp
	// 会导致 bilibili.rpc.Status 的覆盖注册
	Message *pushmdl.Message `json:"message"`
}

type AggregationStatusReq struct {
	BizTypes []string `form:"biz_types,split"`
}

type AggregationStatusRly struct {
	AntiAddiction  *RuleRly            `json:"anti_addiction,omitempty"`
	SleepRemind    *SleepRemindSetup   `json:"sleep_remind,omitempty"`
	FamilyTimelock *familymdl.Timelock `json:"family_timelock,omitempty"`
}

type SleepRemindSetup struct {
	Switch bool             `json:"switch"`
	Stime  string           `json:"stime"`
	Etime  string           `json:"etime"`
	Push   *pushmdl.Message `json:"push"`
}

type SetSleepRemindReq struct {
	Switch bool   `form:"switch"`
	Stime  string `form:"stime" validate:"required"`
	Etime  string `form:"etime" validate:"required"`
}
