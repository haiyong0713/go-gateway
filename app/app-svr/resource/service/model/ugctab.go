package model

import (
	xtime "go-common/library/time"
)

type UgcTabBase struct {
	ID      int64      `json:"id" gorm:"column:id"`
	CTime   xtime.Time `json:"ctime" gorm:"column:ctime"`
	MTime   xtime.Time `json:"mtime" gorm:"column:mtime"`
	Deleted int        `json:"deleted" gorm:"column:deleted"`
}

type UgcTabItem struct {
	UgcTabBase
	TabType      int32      `json:"tab_type" gorm:"column:tab_type"`
	Tab          string     `json:"tab" gorm:"column:tab"`
	LinkType     int32      `json:"link_type" gorm:"column:link_type"`
	Link         string     `json:"link" gorm:"column:link"`
	Bg           string     `json:"background" gorm:"column:background"`
	Selected     string     `json:"txt_selected" gorm:"column:selected_color"`
	Color        string     `json:"txt_color" gorm:"column:txt_color"`
	UgcType      int        `json:"ugc_type" gorm:"column:ugc_type"`
	Stime        xtime.Time `json:"stime" gorm:"column:stime"`
	Etime        xtime.Time `json:"etime" gorm:"column:etime"`
	Online       int        `json:"online" gorm:"column:online"`
	Builds       string     `json:"builds" gorm:"column:builds"`
	Arctype      string     `json:"arctype" gorm:"column:arctype"`
	Tagid        string     `json:"tagid" gorm:"column:tagid"`
	Upid         string     `json:"upid" gorm:"column:upid"`
	Avid         string     `json:"avid" gorm:"column:avid"`
	Username     string     `json:"username" gorm:"column:username"`
	AvidFile     string     `json:"avid_file" gorm:"column:avid_file"`
	UgcTabExtend `json:"-"`
}

type UgcTabExtend struct {
	AvidMap map[string]bool `json:"avid_map"`
}

type BuildLimit struct {
	Plat       int32  `json:"plat"`
	Build      int32  `json:"build"`
	Conditions string `json:"conditions"`
}
