package exporttask

import (
	"encoding/json"
	xtime "go-common/library/time"
)

const (
	TaskStateAdd = iota + 1
	TaskStateDoing
	TaskStateFinish
	TaskStateFail
	TaskStateCancel
)

type ExportTask struct {
	ID        int64           `json:"id" form:"id" gorm:"column:id"`
	Author    string          `json:"author" form:"author" gorm:"column:author"`
	SID       string          `json:"sid" form:"sid" gorm:"column:sid"`
	TaskType  uint8           `json:"task_type" form:"task_type" gorm:"column:task_type"`
	State     uint8           `json:"state" form:"state" gorm:"column:state"`
	StartTime xtime.Time      `json:"start_time" form:"start_time" gorm:"column:start_time"`
	EndTime   xtime.Time      `json:"end_time" form:"end_time" gorm:"column:end_time"`
	DownURL   string          `json:"down_url" form:"down_url" gorm:"column:down_url"`
	TimeCost  int64           `json:"timecost" form:"timecost" gorm:"column:timecost"`
	Machine   string          `json:"machine" form:"machine" gorm:"column:machine"`
	StartAt   xtime.Time      `json:"start_at" form:"start_at" gorm:"column:start_at"`
	EndAt     xtime.Time      `json:"end_at" form:"end_at" gorm:"column:end_at"`
	Ext       json.RawMessage `json:"ext" form:"ext" gorm:"column:ext"`
	CTime     xtime.Time      `json:"ctime" form:"ctime" gorm:"column:ctime"`
	MTime     xtime.Time      `json:"mtime" form:"mtime" gorm:"column:mtime"`
}

func (ExportTask) TableName() string {
	return "act_export_task"
}
