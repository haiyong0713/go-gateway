package invite

import (
	xtime "go-common/library/time"
)

// InviteRelation .
type InviteRelation struct {
	ID               int64      `json:"id"`
	Mid              int64      `json:"mid"`
	ActivityUID      string     `json:"activity_uid"`
	Tel              string     `json:"tel"`
	TelHash          string     `json:"tel_hash"`
	Token            string     `json:"token"`
	Buvid            string     `json:"buvid"`
	IP               string     `json:"ip"`
	InvitedTime      int64      `json:"invited_time"`
	ExpireTime       int64      `json:"expire_time"`
	InvitedMid       int64      `json:"invited_mid"`
	InvitedLoginTime int64      `json:"invited_login_time"`
	Rank             int        `json:"rank"`
	IsNew            int        `json:"is_new"`
	IsBlocked        int        `json:"is_blocked"`
	CTime            xtime.Time `json:"ctime"`
	MTime            xtime.Time `json:"mtime"`
}

// Account ...
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// InviterReply ...
type InviterReply struct {
	Account *Account `json:"account"`
}
