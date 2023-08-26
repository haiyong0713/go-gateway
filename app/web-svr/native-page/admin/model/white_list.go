package model

import (
	xtime "go-common/library/time"
)

const (
	FromActivity = "activity"
)

type AddWhiteListReq struct {
	Mid int64 `form:"mid" validate:"required,min=1"`
}

type BatchAddWhiteListReq struct {
	Mids []int64 `form:"mids,split" validate:"min=1,max=50,dive,min=1"`
}

type AddWhiteListOuterReq struct {
	Mid   int64  `form:"mid" validate:"required,min=1"`
	From  string `form:"from" validate:"required"`
	Uid   int64  `form:"uid" validate:"required"`
	Uname string `form:"uname" validate:"required"`
}

type DeleteWhiteListReq struct {
	ID int `form:"id" validate:"required,min=1"`
}

type GetWhiteListReq struct {
	Mid int64 `form:"mid"`
	Pn  int64 `form:"pn" default:"1"`
	Ps  int64 `form:"ps" default:"20"`
}

type GetWhiteListRly struct {
	Total int64       `json:"total"`
	List  []*ListItem `json:"list"`
}

type ListItem struct {
	*WhiteListRecord
	UserName string `json:"user_name"`
}

type WhiteListRecord struct {
	ID          int        `gorm:"AUTO_INCREMENT;column:id;type:INT;primary_key" json:"id"`
	Mid         int64      `gorm:"column:mid;type:BIGINT;default:0;" json:"mid"`
	Creator     string     `gorm:"column:creator;type:VARCHAR;size:32;" json:"creator"`
	CreatorUID  int        `gorm:"column:creator_uid;type:INT;default:0;" json:"creator_uid"`
	Modifier    string     `gorm:"column:modifier;type:VARCHAR;size:32;" json:"modifier"`
	ModifierUID int        `gorm:"column:modifier_uid;type:INT;default:0;" json:"modifier_uid"`
	FromType    string     `gorm:"column:from_type;type:VARCHAR;size:32;" json:"from_type"`
	State       int        `gorm:"column:state;type:TINYINT;default:0;" json:"state"`
	Ctime       xtime.Time `gorm:"column:ctime;type:DATETIME;default:CURRENT_TIMESTAMP;" json:"ctime"`
	Mtime       xtime.Time `gorm:"column:mtime;type:DATETIME;default:CURRENT_TIMESTAMP;" json:"mtime"`
}
