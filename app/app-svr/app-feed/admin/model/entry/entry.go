package entry

import (
	"go-common/library/time"
)

type BaseModel struct {
	ID           int32     `json:"id" gorm:"column:id"`
	CTime        time.Time `json:"ctime" gorm:"column:ctime"`
	MTime        time.Time `json:"mtime" gorm:"column:mtime"`
	IsDeprecated int32     `json:"is_deprecated" gorm:"column:is_deprecated;default:0"`
}

type AppEntry struct {
	BaseModel
	EntryName    string    `json:"entry_name" gorm:"column:entry_name"`
	OnlineStatus int32     `json:"online_status" gorm:"column:online_status"`
	STime        time.Time `json:"stime" gorm:"column:stime"`
	ETime        time.Time `json:"etime" gorm:"column:etime"`
	CreatedBy    string    `json:"created_by" gorm:"column:created_by"`
	Platforms    string    `json:"platforms" gorm:"column:platforms"`
	TotalLoop    int32     `json:"total_loop" gorm:"column:total_loop"`
}

type AppEntryState struct {
	BaseModel
	StateName   string `json:"state_name" gorm:"column:state_name"`
	Url         string `json:"url" gorm:"column:url"`
	StaticIcon  string `json:"static_icon" gorm:"column:static_icon"`
	DynamicIcon string `json:"dynamic_icon" gorm:"column:dynamic_icon;"`
	EntryID     int32  `json:"entry_id" gorm:"entry_id"`
	LoopCount   int32  `json:"loop_count" gorm:"loop_count"`
}

type AppEntryTimeSetting struct {
	BaseModel
	EntryID   int32     `json:"entry_id" gorm:"entry_id"`
	StateID   int32     `json:"state_id" gorm:"state_id"`
	STime     time.Time `json:"stime" gorm:"column:stime"`
	PushTime  time.Time `json:"push_time" gorm:"column:push_time;default:'2009-12-31 23:59:59'"`
	CreatedBy string    `json:"created_by" gorm:"column:created_by"`
	SentLoop  int32     `json:"sent_loop" gorm:"column:sent_loop"`
}
