package task

import (
	xtime "go-common/library/time"
)

// Attributes
const (
	BusinessAct        = 1
	AttrYes            = 1
	AttrNo             = 0
	AttrBitAutoReceive = uint(0)
	AttrBitIsCycle     = uint(1)
	AttrBitAward       = uint(2)
	AttrBitRule        = uint(3)
	AttrBitOut         = uint(4)
	AttrBitSpecial     = uint(5)
	AttrBitNoFinish    = uint(6)
	AttrBitNewTable    = uint(7)
	AttrBitDayCount    = uint(8)
	// AttrBitMultiSource 是否多数据源
	AttrBitMultiSource = uint(9)
	InitCount          = 0
	NotFinish          = 0
	HasFinish          = 1
	NotAward           = 0
	HasAward           = 1
	AwardTypeCurr      = 1
	AwardTypePend      = 2
	AwardTypeCoupon    = 3
	AwardTypeLottery   = 4
	// StateFinish 任务完成
	StateFinish = 1
	RiskLevelNo = 0
	RiskLevel1  = 1
	RiskLevel2  = 2
	RiskLevel3  = 3
	RiskLevel4  = 4
	RiskLevel5  = 5
	// RiskOperationRemark 提示用户
	RiskOperationRemark = 1
	// IsFeYes 前端上报
	IsFeYes = 1
)

// TaskItem .
type TaskItem struct {
	*Task
	Ctime          xtime.Time   `json:"ctime"`
	UserRound      int64        `json:"user_round"`
	UserCount      int64        `json:"user_count"`
	UserFinish     int64        `json:"user_finish"`
	UserAward      int64        `json:"user_award"`
	UserTotalCount int64        `json:"user_total_count"`
	UserRoundList  []xtime.Time `json:"user_round_list,omitempty"`
}

// TaskRule
type TaskRule struct {
	ID        int64      `json:"id"`
	TaskID    int64      `json:"task_id"`
	PreTask   string     `json:"pre_task"`
	Level     int64      `json:"level"`
	Object    int        `json:"object"`
	Count     int64      `json:"count"`
	CountType int        `json:"count_type"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
}

type TaskAll struct {
	AllFinish bool `json:"all_finish"`
}

// MidRule ...
type MidRule struct {
	Object int64 `json:"object"`
	MID    int64 `json:"mid"`
	State  int   `json:"state"`
	Count  int64 `json:"count"`
}

// Round get task round.
func (t *Task) Round(nowTs int64) (round int64) {
	if !t.IsCycle() || t.CycleDuration == 0 {
		return
	}
	if nowTs > t.Etime.Time().Unix() {
		nowTs = t.Etime.Time().Unix()
	}
	return (nowTs - t.Stime.Time().Unix()) / t.CycleDuration
}

// IsAutoReceive is cycle task.
func (t *Task) IsAutoReceive() bool {
	return t.attrVal(AttrBitAutoReceive) == AttrYes
}

// IsCycle is cycle task.
func (t *Task) IsCycle() bool {
	return t.attrVal(AttrBitIsCycle) == AttrYes
}

// HasAward has award.
func (t *Task) HasAward() bool {
	return t.attrVal(AttrBitAward) == AttrYes
}

// HasRule has rule.
func (t *Task) HasRule() bool {
	return t.attrVal(AttrBitRule) == AttrYes
}

// IsOut .
func (t *Task) IsOut() bool {
	return t.attrVal(AttrBitOut) == AttrYes
}

// IsSpecial .
func (t *Task) IsSpecial() bool {
	return t.attrVal(AttrBitSpecial) == AttrYes
}

func (t *Task) IsNoFinish() bool {
	return t.attrVal(AttrBitNoFinish) == AttrYes
}

func (t *Task) IsNewTable() bool {
	return t.attrVal(AttrBitNewTable) == AttrYes
}

func (t *Task) NeedDayCount() bool {
	return t.attrVal(AttrBitDayCount) == AttrYes
}

// IsMultiSource ...
func (t *Task) IsMultiSource() bool {
	return t.attrVal(AttrBitMultiSource) == AttrYes
}

// AttrVal get attr val by bit.
func (t *Task) attrVal(bit uint) int64 {
	return (t.Attribute >> bit) & 1
}

// AwardReply .
type AwardReply struct {
	Award int64 `json:"award,omitempty"`
}

//go:generate kratos t protoc --grpc task.proto

// Detail ...
type Detail struct {
	ID            int64  `json:"id"`
	TaskName      string `json:"task_name"`
	LinkName      string `json:"link_name"`
	OrderID       int64  `json:"order_id"`
	Activity      string `json:"activity"`
	ActivityID    int64  `json:"activity_id"`
	Counter       string `json:"counter"`
	Desc          string `json:"desc"`
	Link          string `json:"link"`
	FinishTimes   int64  `json:"finish_times"`
	State         int    `json:"state"`
	RiskLevel     int64  `json:"risk_level"`
	RiskOperation int64  `json:"risk_operation"`
	IsFe          int    `json:"is_fe"`
}

// TaskMember ...
type TaskMember struct {
	Counter string                 `json:"counter"`
	Count   int64                  `json:"count"`
	State   int                    `json:"state"`
	IsFe    int                    `json:"is_fe"`
	Params  map[string]interface{} `json:"params"`
}

// TaskReply ...
type TaskReply struct {
	List []*TaskDetail `json:"list"`
}

// TaskDetail 用户任务情况
type TaskDetail struct {
	Task   *SimpleTask `json:"task"`
	Member *TaskMember `json:"member"`
}

// SimpleTask ...
type SimpleTask struct {
	TaskName    string `json:"task_name"`
	LinkName    string `json:"link_name"`
	Desc        string `json:"desc"`
	Link        string `json:"link"`
	FinishTimes int64  `json:"finish_times"`
}
