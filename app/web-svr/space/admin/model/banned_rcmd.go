package model

import (
	"time"
)

type UpRcmdBlackListItem struct {
	ID        int64     `json:"id" form:"id" gorm:"column:id"`
	Mid       int64     `json:"mid" form:"mid" gorm:"column:mid"`
	IsDeleted int32     `json:"is_deleted" form:"is_deleted" default:"0" gorm:"column:is_deleted"`
	CTime     time.Time `json:"ctime" form:"ctime" gorm:"column:ctime"`
	MTime     time.Time `json:"mtime" form:"mtime" gorm:"column:mtime"`
}

type UpRcmdBlackListCreateReq struct {
	Mids []int64 `json:"mids" form:"mids"`
}

type UpRcmdBlackListCreateRep struct {
	FailMids []int64 `json:"fail_mids" form:"fail_mids"`
}

type UpRcmdBlackListDeleteReq struct {
	Mid int64 `json:"mid" form:"mid"`
}

type UpRcmdBlackListSearchReq struct {
	Ps  int64 `json:"ps" form:"ps"`
	Pn  int64 `json:"pn" form:"pn"`
	Mid int64 `json:"mid" form:"mid"`
}

type UserInfo struct {
	Mid   int64     `json:"mid" form:"mid"`
	UName string    `json:"uname" form:"uname"`
	MTime time.Time `json:"mtime" form:"mtime"`
	//Avatar string `json:"avatar" form:"avatar"`
}

type UpRcmdBlackListSearchRep struct {
	Page  Page        `json:"page" form:"page"`
	Items []*UserInfo `json:"items" form:"items"`
}

type UserInfoSearchReq struct {
	Mids []int64 `json:"mids" form:"mids"`
}

type UserInfoSearchRep struct {
	Items []*UserInfo `json:"items" form:"items"`
}
