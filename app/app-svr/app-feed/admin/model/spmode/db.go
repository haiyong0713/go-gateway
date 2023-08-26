package spmode

import (
	xtime "go-common/library/time"
)

const (
	// state
	StateQuit = 0
	StateOpen = 1
	// teenager_users.operation
	OperationQuitManager     = 13 //退出-后台
	OperationQuitFyMgrUnbind = 16 //退出-亲子平台后台解绑
	// manual_force
	ManualForceQuit = 0
	ManualForceOpen = 1
	// model
	ModelTeenager = 0
	ModelLesson   = 1
)

type TeenagerUsers struct {
	ID          int64      `json:"id" gorm:"id"`
	Mid         int64      `json:"mid" gorm:"mid"`
	Password    string     `json:"password" gorm:"password"`
	State       int64      `json:"state" gorm:"state"`
	Ctime       xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime       xtime.Time `json:"mtime" gorm:"mtime"`
	Model       int64      `json:"model" gorm:"model"`
	Operation   int64      `json:"operation" gorm:"operation"`
	QuitTime    xtime.Time `json:"quit_time" gorm:"quit_time"`
	PwdType     int64      `json:"pwd_type" gorm:"pwd_type"`
	ManualForce int64      `json:"manual_force" gorm:"manual_force"`
	MfOperator  string     `json:"mf_operator" gorm:"mf_operator"`
	MfTime      xtime.Time `json:"mf_time" gorm:"mf_time"`
}

func (t *TeenagerUsers) TableName() string {
	return "teenager_users"
}

type DeviceUserModel struct {
	ID          int64      `json:"id" gorm:"id"`
	MobiApp     string     `json:"mobi_app" gorm:"mobi_app"`
	DeviceToken string     `json:"device_token" gorm:"device_token"`
	Password    string     `json:"password" gorm:"password"`
	State       int64      `json:"state" gorm:"state"`
	Ctime       xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime       xtime.Time `json:"mtime" gorm:"mtime"`
	Model       int64      `json:"model" gorm:"model"`
	Operation   int64      `json:"operation" gorm:"operation"`
	QuitTime    xtime.Time `json:"quit_time" gorm:"quit_time"`
	PwdType     int64      `json:"pwd_type" gorm:"pwd_type"`
}

func (t *DeviceUserModel) TableName() string {
	return "device_user_model"
}

type SpecialModeLog struct {
	ID          int64      `json:"id" gorm:"id"`
	RelatedKey  string     `json:"related_key" gorm:"related_key"`
	OperatorUid int64      `json:"operator_uid" gorm:"operator_uid"`
	Operator    string     `json:"operator" gorm:"operator"`
	Content     string     `json:"content" gorm:"content"`
	Ctime       xtime.Time `json:"ctime" gorm:"ctime"`
	Mtime       xtime.Time `json:"mtime" gorm:"mtime"`
}

func (t *SpecialModeLog) TableName() string {
	return "special_mode_log"
}
