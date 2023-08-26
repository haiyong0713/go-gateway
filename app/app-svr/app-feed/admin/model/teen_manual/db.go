package teen_manual

import (
	xtime "go-common/library/time"
)

type TeenagerManualLog struct {
	ID       int64      `json:"id" gorm:"id"`
	Mid      int64      `json:"mid" gorm:"mid"`
	Operator string     `json:"operator" gorm:"operator"`
	Content  string     `json:"content" gorm:"content"`
	Remark   string     `json:"remark" gorm:"remark"`
	Ctime    xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime    xtime.Time `json:"mtime" gorm:"mtime"`
}

func (t *TeenagerManualLog) TableName() string {
	return "teenager_manual_log"
}
