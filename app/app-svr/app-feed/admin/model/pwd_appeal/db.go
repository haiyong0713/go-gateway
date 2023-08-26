package pwd_appeal

import (
	xtime "go-common/library/time"
)

const (
	// pwd_appeal.state
	StatePending = 1
	StatePass    = 2
	StateReject  = 3
	// pwd_appeal.mode
	ModeTeenager = 1
	ModeClass    = 2
)

var (
	ModeMap = map[int64]string{
		ModeTeenager: "青少年模式找回密码",
		ModeClass:    "课堂模式找回密码",
	}
	StateMap = map[int64]string{
		StatePending: "待处理",
		StatePass:    "申诉通过",
		StateReject:  "申诉驳回",
	}
)

type PwdAppeal struct {
	ID           int64      `json:"id" gorm:"id"`
	Mid          int64      `json:"mid" gorm:"mid"`
	DeviceToken  string     `json:"device_token" gorm:"device_token"`
	Mobile       int64      `json:"mobile" gorm:"mobile"`
	Mode         int64      `json:"mode" gorm:"mode"`
	State        int64      `json:"state" gorm:"state"`
	RejectReason string     `json:"reject_reason" gorm:"reject_reason"`
	Pwd          string     `json:"pwd" gorm:"pwd"`
	UploadKey    string     `json:"upload_key" gorm:"upload_key"`
	Ctime        xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime        xtime.Time `json:"mtime" gorm:"mtime"`
	Operator     string     `json:"operator" gorm:"operator"`
}

func (t *PwdAppeal) TableName() string {
	return "pwd_appeal"
}

type PwdAppealLog struct {
	ID          int64      `json:"id" gorm:"id"`
	AppealID    int64      `json:"appeal_id" gorm:"appeal_id"`
	OperatorUid int64      `json:"operator_uid" gorm:"operator_uid"`
	Operator    string     `json:"operator" gorm:"operator"`
	Content     string     `json:"content" gorm:"content"`
	Ctime       xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime       xtime.Time `json:"mtime" gorm:"mtime"`
}

func (t *PwdAppealLog) TableName() string {
	return "pwd_appeal_log"
}
