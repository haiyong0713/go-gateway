package model

import (
	xtime "go-common/library/time"
)

const (
	//StatusWhite whitelist
	StatusWhite  = 0
	StatusValid  = 1
	StatusReady  = 2
	StatusFailed = 3
)

type WhitelistReq struct {
	Mids     []int64    `form:"mids,split" validate:"required"`
	Stime    xtime.Time `json:"stime" form:"stime"`
	Etime    xtime.Time `json:"etime" form:"etime"`
	Username string     `json:"username" form:"username"`
}

type WhitelistAdd struct {
	ID       int64      `json:"id" form:"id"`
	Mid      int64      `json:"mid" form:"mid"`
	MidName  string     `json:"mid_name" form:"mid_name"`
	State    int        `json:"state" form:"state"`
	Stime    xtime.Time `json:"stime" form:"stime"`
	Etime    xtime.Time `json:"etime" form:"etime"`
	Username string     `json:"username" form:"username"`
	Deleted  int        `json:"deleted" form:"deleted"`
}

type TopPhotoConf struct {
	Mid      int64  `json:"mid" form:"mid"`
	Bvid     string `json:"bvid" form:"bvid"`
	ImageUrl string `json:"image_url" form:"image_url"`
}

// Whitelist .
type Whitelist struct {
	WhitelistAdd
	MidConf TopPhotoConf `json:"mid_conf" form:"mid_conf"`
	Mtime   xtime.Time   `json:"mtime" form:"mtime"`
}

// WhitelistPager blacklist pager
type WhitelistPager struct {
	Item []*Whitelist
	Page Page
}

// TableName .
func (a Whitelist) TableName() string {
	return "whitelist"
}
