package model

import xtime "go-common/library/time"

const (
	//StatusWhite whitelist
	StatusWhite  = 0
	StatusValid  = 1
	StatusReady  = 2
	StatusFailed = 3
)

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
