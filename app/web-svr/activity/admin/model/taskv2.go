package model

import "time"

type ActTask struct {
	ID            int64     `form:"id" json:"id" gorm:"column:id"`
	TaskName      string    `form:"task_name" json:"task_name" gorm:"column:task_name"`
	OrderID       int64     `form:"order_id" json:"order_id" gorm:"column:order_id"`
	Activity      string    `form:"activity" json:"activity" gorm:"column:activity"`
	Counter       string    `form:"counter" json:"counter" gorm:"column:counter"`
	Link          string    `form:"link" json:"link" gorm:"column:link"`
	FinishTimes   int       `form:"finish_times" json:"finish_times" gorm:"column:finish_times"`
	State         int       `form:"state" json:"state" gorm:"column:state"`
	LinkName      string    `form:"link_name" json:"link_name" gorm:"column:link_name"`
	TaskDesc      string    `form:"task_desc" json:"task_desc" gorm:"column:task_desc"`
	Ctime         time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime         time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	ActivityID    int64     `form:"activity_id" json:"activity_id" gorm:"column:activity_id"`
	RiskLevel     int64     `form:"risk_level" json:"risk_level" gorm:"column:risk_level"`
	RiskOperation int64     `form:"risk_operation" json:"risk_operation" gorm:"column:risk_operation"`
	IsFe          int64     `form:"is_fe" json:"is_fe" gorm:"column:is_fe"`
}
