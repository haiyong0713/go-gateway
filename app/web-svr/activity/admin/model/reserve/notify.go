package reserve

import "encoding/json"

const (
	ActSubjectNotifyStateNormal = iota + 1
	ActSubjectNotifyStateFrozen
	ActSubjectNotifyStateDelete
)

type ActSubjectNotify struct {
	ID         int64           `json:"id"`
	Sid        int64           `form:"sid" json:"sid" gorm:"column:sid"`
	RuleID     int64           `form:"rule_id" json:"rule_id" gorm:"column:rule_id"`
	NotifyID   string          `form:"notify_id" json:"notify_id" gorm:"column:notify_id" validate:"required,lt=32"`
	NotifyType uint8           `form:"notify_type" json:"notify_type" gorm:"column:notify_type" validate:"min=1"`
	Title      string          `form:"title" json:"title" gorm:"column:title" validate:"lt=32"`
	Receiver   string          `form:"receiver" json:"receiver" gorm:"column:receiver" validate:"required,lt=255"`
	Author     string          `form:"author" json:"author" gorm:"column:author" validate:"lt=16"`
	State      int             `form:"state" json:"state" gorm:"column:state" validate:"min=1"`
	Threshold  int64           `form:"threshold" json:"threshold" gorm:"column:threshold" validate:"min=1"`
	NotifyTime int             `gorm:"column:notify_time"`
	Ext        json.RawMessage `form:"ext" json:"ext" gorm:"column:ext"`
	TemplateID int             `form:"template_id" json:"template_id" gorm:"column:template_id" default:"1"`
}

func (ActSubjectNotify) TableName() string {
	return "act_subject_notify"
}
