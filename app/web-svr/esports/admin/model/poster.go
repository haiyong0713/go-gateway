package model

import "time"

type Poster struct {
	ID            int64     `json:"id" gorm:"column:id"`
	BgImage       string    `json:"bg_image" gorm:"column:bg_image"`
	ContestID     int64     `json:"contest_id" gorm:"column:contest_id"`
	OnlineStatus  int32     `json:"online_status" gorm:"column:online_status"`
	IsCenteral    int32     `json:"is_centeral" gorm:"column:is_centeral"`
	PositionOrder int32     `json:"position_order" gorm:"column:position_order"`
	CreatedBy     string    `json:"created_by" gorm:"column:created_by"`
	CTime         time.Time `json:"ctime" gorm:"column:ctime"`
	MTime         time.Time `json:"mtime" gorm:"column:mtime"`
	IsDeprecated  int32     `json:"is_deprecated" gorm:"column:is_deprecated"`
}
