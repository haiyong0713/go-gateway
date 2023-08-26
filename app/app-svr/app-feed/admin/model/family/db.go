package family

import (
	xtime "go-common/library/time"
)

const (
	// family_relation.state
	StateUnbind = 0
	StateBind   = 1
)

type FamilyRelation struct {
	ID        int64      `json:"id" gorm:"id"`
	ParentMid int64      `json:"parent_mid" gorm:"parent_mid"`
	ChildMid  int64      `json:"child_mid" gorm:"child_mid"`
	State     int64      `json:"state" gorm:"state"`
	Ctime     xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime     xtime.Time `json:"mtime" gorm:"mtime"`
}

func (t *FamilyRelation) TableName() string {
	return "family_relation"
}

type FamilyLog struct {
	ID       int64      `json:"id" gorm:"id"`
	Mid      int64      `json:"mid" gorm:"mid"`
	Operator string     `json:"operator" gorm:"operator"`
	Content  string     `json:"content" gorm:"content"`
	Ctime    xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime    xtime.Time `json:"mtime" gorm:"mtime"`
}

func (t *FamilyLog) TableName() string {
	return "family_log"
}
