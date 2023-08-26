package exporttask

import xtime "go-common/library/time"

const (
	ExportTypeBfs = iota
	ExportTypeCsv
	ExportTypeJson
)

type ReqExportTaskAdd struct {
	Author      string     `json:"_author" form:"_author" validate:"required,lt=32"`
	SID         string     `json:"_sid" form:"_sid" gorm:"_sid" validate:"required,lt=64"`
	TaskType    uint8      `json:"_task_type" form:"_task_type" validate:"required,min=1"`
	StartTime   xtime.Time `json:"_start_time" form:"_start_time" time_format:"2006-01-02 15:04:05" validate:"required"`
	EndTime     xtime.Time `json:"_end_time" form:"_end_time" time_format:"2006-01-02 15:04:05" validate:"required"`
	Synchronize bool       `json:"_synchronize" form:"_synchronize"`
	ExportType  uint8      `json:"_export_type" form:"_export_type"`
}

type ReqExportTaskState struct {
	TaskID int64 `json:"task_id" form:"task_id" validate:"required,min=1"`
}

type ReqExportTaskList struct {
	SID      string `json:"sid" form:"sid" validate:"required,lt=64"`
	Author   string `json:"author" form:"author" validate:"required,lt=32"`
	Type     int    `json:"type" form:"type"`
	Page     int    `form:"page" default:"1" validate:"min=1"`
	PageSize int    `form:"pagesize" default:"15" validate:"min=1"`
}
