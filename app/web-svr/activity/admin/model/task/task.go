package task

import (
	"fmt"
	"go-common/library/time"
)

const (
	AttrYes          = 1
	AttrBitIsAutoGet = 0
	AttrBitIsCycle   = 1
	AttrBitHasAward  = 2
	AttrBitHasRule   = 3
	AttrBitOut       = 4
	AttrBitSpecial   = 5
	AttrBitNoFinish  = 6
	AttrBitNewTable  = 7
	AttrBitDayCount  = 8
	TaskOffline      = 0
	TaskOnline       = 1
)

// AddArg task add arg.
type AddArg struct {
	Name          string    `form:"name" validate:"min=1"`
	BusinessID    int64     `form:"business_id" validate:"min=1"`
	ForeignID     int64     `form:"foreign_id" validate:"min=1"`
	Rank          int64     `form:"rank"`
	FinishCount   int64     `form:"finish_count" validate:"min=1"`
	Attribute     int64     `form:"attribute"`
	CycleDuration int64     `form:"cycle_duration"`
	AwardType     int64     `form:"award_type"`
	AwardID       int64     `form:"award_id"`
	AwardCount    int64     `form:"award_count"`
	State         int64     `form:"state"`
	Stime         time.Time `form:"stime" validate:"min=1"`
	Etime         time.Time `form:"etime" validate:"min=1"`
	PreTask       string    `form:"pre_task"`
	Level         int       `form:"level"`
}

// AddArgV2 v2
type AddArgV2 struct {
	Name          string     `form:"name" json:"name" validate:"required"`
	BusinessID    int64      `form:"business_id" json:"business_id" validate:"min=1"`
	ForeignID     int64      `form:"foreign_id" json:"foreign_id" validate:"min=1"`
	Rank          int64      `form:"rank" json:"rank"`
	FinishCount   int64      `form:"finish_count" json:"finish_count"`
	Attribute     int64      `form:"attribute" json:"attribute"`
	CycleDuration int64      `form:"cycle_duration" json:"cycle_duration"`
	AwardType     int64      `form:"award_type" json:"award_type"`
	AwardID       int64      `form:"award_id" json:"award_id"`
	AwardCount    int64      `form:"award_count" json:"award_count"`
	State         int64      `form:"state" json:"state"`
	Stime         time.Time  `form:"stime" json:"stime" validate:"min=1"`
	Etime         time.Time  `form:"etime" json:"etime" validate:"min=1"`
	TaskRule      []*AddRule `form:"rule" json:"rule"`
}

// SaveArg task save arg.
type SaveArg struct {
	ID int64 `form:"id" validate:"min=1"`
	AddArg
}

// AddAward .
type AddAward struct {
	TaskID int64 `form:"task_id" validate:"min=1"`
	Award  int64 `form:"award" validate:"min=1"`
}

// Item task show item.
type Item struct {
	*Task
	*Rule
}

// Task .
type Task struct {
	ID            int64     `json:"id" gorm:"column:id"`
	Name          string    `json:"name" gorm:"column:name"`
	BusinessID    int64     `json:"business_id" gorm:"column:business_id"`
	ForeignID     int64     `json:"foreign_id" gorm:"column:foreign_id"`
	Rank          int64     `json:"rank" gorm:"column:rank"`
	FinishCount   int64     `json:"finish_count" gorm:"column:finish_count"`
	Attribute     int64     `json:"attribute" gorm:"column:attribute"`
	CycleDuration int64     `json:"cycle_duration" gorm:"column:cycle_duration"`
	AwardType     int64     `json:"award_type" gorm:"column:award_type"`
	AwardID       int64     `json:"award_id" gorm:"column:award_id"`
	AwardCount    int64     `json:"award_count" gorm:"column:award_count"`
	AwardExpire   int64     `json:"award_expire" gorm:"column:award_expire"`
	State         int64     `json:"state" gorm:"column:state"`
	Stime         time.Time `json:"stime" gorm:"column:stime" time_format:"2006-01-02 15:04:05"`
	Etime         time.Time `json:"etime" gorm:"column:etime" time_format:"2006-01-02 15:04:05"`
	Ctime         time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime         time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// Rule task rule struct.
type Rule struct {
	ID        int64     `json:"task_rule_id" gorm:"column:id"`
	TaskID    int64     `json:"task_id" gorm:"column:task_id"`
	PreTask   string    `json:"pre_task" gorm:"column:pre_task"`
	Level     int       `json:"level" gorm:"column:level"`
	Object    int       `form:"object" json:"object"`
	Count     int       `form:"count" json:"count"`
	CountType int       `form:"count_type" json:"count_type"`
	Ctime     time.Time `json:"rule_ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"rule_mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// AddRule task rule struct.
type AddRule struct {
	PreTask   string `json:"pre_task" gorm:"column:pre_task"`
	Level     int    `json:"level" gorm:"column:level"`
	Object    int    `form:"object" json:"object"`
	Count     int    `form:"count" json:"count"`
	CountType int    `form:"count_type" json:"count_type"`
}

// TableName task def.
func (Task) TableName() string {
	return "task"
}

func (t *Task) HasRule() bool {
	return t.AttrVal(AttrBitHasRule) == AttrYes
}

// AttrVal get attr val by bit.
func (t *Task) AttrVal(bit uint) int64 {
	return (t.Attribute >> bit) & 1
}

// TableName task rule def.
func (Rule) TableName() string {
	return "task_rule"
}

type TaskUserState struct {
	ID         int64 `json:"id" gorm:"column:id"`
	Mid        int64 `json:"mid" gorm:"column:mid"`
	BusinessID int64 `json:"business_id" gorm:"column:business_id"`
	TaskID     int64 `json:"task_id" gorm:"column:task_id"`
	ForeignID  int64 `json:"foreign_id" gorm:"column:foreign_id"`
	Round      int64 `json:"round" gorm:"column:round"`
	Count      int64 `json:"count" gorm:"column:cnt"`
	Finish     int64 `json:"finish" gorm:"column:finish"`
	Award      int64 `json:"award" gorm:"column:award"`
	RoundCount int64 `json:"round_count" gorm:"column:round_count"`
}

func (TaskUserState) TableName() string {
	return "task_user_state"
}

func TaskUserStateTable(sid int64) string {
	return fmt.Sprintf("task_user_state_%02d", sid%100)
}
