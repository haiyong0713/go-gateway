package task

import (
	xtime "go-common/library/time"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"
)

// Attributes
const (
	BusinessAct = 1
	AttrYes     = 1
	AttrNo      = 0
	// AttrBitAutoReceive 是否自动领取
	AttrBitAutoReceive = uint(0)
	// AttrBitIsCycle 是否周期任务
	AttrBitIsCycle = uint(1)
	// AttrBitAward 是否有奖励
	AttrBitAward = uint(2)
	// AttrBitRule 是否有前置条件
	AttrBitRule = uint(3)
	// AttrBitOut 是否外部
	AttrBitOut = uint(4)
	// AttrBitSpecial 是否特殊处理
	AttrBitSpecial = uint(5)
	// AttrBitSpecial 是否不需要完成
	AttrBitNoFinish = uint(6)
	// AttrBitNewTable 是否使用新的分表策略
	AttrBitNewTable = uint(7)
	// AttrBitDayCount 是否每日统计
	AttrBitDayCount = uint(8)
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
	// TaskObjectArchive 任务对象稿件
	TaskObjectArchive = 1
	// TaskObjectLike 任务对象点赞个数
	TaskObjectLike = 2
	// TaskObjectLCoin 任务对象投币个数
	TaskObjectLCoin = 3
	// TaskObjectLView 任务对象播放量个数
	TaskObjectLView = 4
	// TaskObjectLFav 任务对象收藏个数
	TaskObjectLFav = 5
	// TaskObjectLScore 任务对象总得分
	TaskObjectLScore = 6
	CountTypeTotal   = 1
	CountTypeSingle  = 2
	// IsFinish 已经完成了
	IsFinish = 1
	// IsNotFinish 未完成
	IsNotFinish = 0
)

// Task ...
type Task struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	BusinessID    int64      `json:"business_id"`
	ForeignID     int64      `json:"foreign_id"`
	Rank          int64      `json:"rank"`
	FinishCount   int64      `json:"finish_count"`
	Attribute     int64      `json:"attribute"`
	CycleDuration int64      `json:"cycle_duration"`
	Stime         xtime.Time `json:"stime"`
	Etime         xtime.Time `json:"etime"`
	AwardType     int64      `json:"award_type"`
	AwardID       int64      `json:"award_id"`
	AwardCount    int64      `json:"award_count"`
	PreTask       string     `json:"pre_task"`
	Level         int64      `json:"level"`
	AwardExpire   int64      `json:"award_expire"`
}

// Item ...
type Item struct {
	*Task
	Ctime          xtime.Time   `json:"ctime"`
	UserRound      int64        `json:"user_round"`
	UserCount      int64        `json:"user_count"`
	UserFinish     int64        `json:"user_finish"`
	UserAward      int64        `json:"user_award"`
	UserTotalCount int64        `json:"user_total_count"`
	UserRoundList  []xtime.Time `json:"user_round_list,omitempty"`
}

// Rule ...
type Rule struct {
	ID        int64      `json:"id"`
	TaskID    int64      `json:"task_id"`
	PreTask   string     `json:"pre_task"`
	Object    int        `json:"object"`
	Count     int64      `json:"count"`
	CountType int        `json:"count_type"`
	Level     int64      `json:"level"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
}

// UserState 用户任务记录
type UserState struct {
	ID         int64      `json:"id"`
	MID        int64      `json:"mid"`
	BusinessID int64      `json:"business_id"`
	ForeignID  int64      `json:"foreign_id"`
	TaskID     int64      `json:"task_id"`
	Round      int        `json:"round"`
	Count      int        `json:"count"`
	Finish     int        `json:"finish"`
	Award      int        `json:"award"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
}

// State 任务统计结构
type State struct {
	ID         int64 `json:"id"`
	BusinessID int64 `json:"business_id"`
	TaskID     int64 `json:"task_id"`
	ForeignID  int64 `json:"foreign_id"`
	Num        int64 `json:"num"`
}

// MidRule ...
type MidRule struct {
	Object int   `json:"object"`
	MID    int64 `json:"mid"`
	State  int   `json:"state"`
	Count  int64 `json:"count"`
}

// MidRuleBatch ...
type MidRuleBatch struct {
	Mid     int64      `json:"mid"`
	MidRule []*MidRule `json:"mid_rule"`
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

// IsNoFinish ...
func (t *Task) IsNoFinish() bool {
	return t.attrVal(AttrBitNoFinish) == AttrYes
}

// IsNewTable ...
func (t *Task) IsNewTable() bool {
	return t.attrVal(AttrBitNewTable) == AttrYes
}

// NeedDayCount ...
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

// CountArchive 统计稿件个数是否满足
func (t *Rule) CountArchive(arcs []*mdlRank.ArchiveStat) (int64, bool) {
	return int64(len(arcs)), int64(len(arcs)) >= t.Count
}

// CountLike 统计稿件点赞数是否满足
func (t *Rule) CountLike(arcs []*mdlRank.ArchiveStat) (int64, bool) {
	var count int64
	var state bool

	for _, v := range arcs {
		if t.IsSingleCount() && int64(v.Like) >= t.Count {
			state = true
		}
		count += int64(v.Like)
	}
	if count >= t.Count {
		state = true

	}
	return count, state
}

// IsSingleCount 是否单个稿件统计
func (t *Rule) IsSingleCount() bool {
	return t.CountType == CountTypeSingle
}

// ChildTaskFunc 返回任务验证函数
func (t *Rule) ChildTaskFunc(arcs []*mdlRank.ArchiveStat) (int64, bool) {
	if arcs == nil {
		return 0, false
	}
	switch t.Object {
	case TaskObjectArchive:
		return t.CountArchive(arcs)
	case TaskObjectLike:
		return t.CountLike(arcs)
	case TaskObjectLScore:
		return t.CountScore(arcs)
	}
	return 0, false
}

// CountScore 统计稿件积分否满足
func (t *Rule) CountScore(arcs []*mdlRank.ArchiveStat) (int64, bool) {
	var count int64
	var state bool
	for _, v := range arcs {
		if t.IsSingleCount() && int64(v.Score) >= t.Count {
			state = true
		}
		count += int64(v.Score)
	}
	if count >= t.Count {
		state = true
	}
	return count, state
}
