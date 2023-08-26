package invite

import (
	xtime "go-common/library/time"
)

const (
	// UserIsBlocked 。。。
	UserIsBlocked = 1
)

// BaseInfo .
type BaseInfo struct {
	Buvid   string `form:"buvid"`
	Origin  string `form:"appkey"`
	UA      string `form:"ua"`
	Referer string
	IP      string
}

// BindReply .
type BindReply struct {
}

// TokenUserResp 获取token信息返回结果
type TokenUserResp struct {
	Mid        int64 `json:"mid"`
	ExpireTime int64 `json:"expire_time"`
	Tp         int64 `json:"tp"`
}

// FiToken ...
type FiToken struct {
	ID          int64  `json:"id,omitempty"`
	Mid         int64  `json:"mid,omitempty"`
	Token       string `json:"token,omitempty"`
	ExpireTime  int64  `json:"expire_time,omitempty"`
	Tp          int64  `json:"tp,omitempty"`
	ActivityUID string `json:"activity_uid,omitempty"`
	Source      int64  `json:"source"`
}

// TokenResp 生成token返回结果
type TokenResp struct {
	Token string `json:"token"`
}

// BindReq ...
type BindReq struct {
	ActivityID string `json:"activity_uid" form:"activity_uid" validate:"required"`
	Token      string `json:"token" form:"token" validate:"required"`
	Tel        string `json:"tel" form:"tel" validate:"required"`
	*BaseInfo
}

// AllInviteLog ...
type AllInviteLog struct {
	Mid          int64
	ActivityUID  string
	Tel          string
	Token        string
	InvitedTime  int64
	Source       int64
	InviteStatus int64
}

// UserShareLog .
type UserShareLog struct {
	ID             int64      `json:"id,omitempty"`
	Mid            int64      `json:"mid,omitempty"`
	ActivityUID    string     `json:"activity_uid,omitempty"`
	FirstShareTime int64      `json:"first_share_time,omitempty"`
	LastShareTime  int64      `json:"last_share_time,omitempty"`
	IsShareExpire  int        `json:"is_share_expire,omitempty"`
	FirstEnterTime int64      `json:"first_enter_time,omitempty"`
	LastEnterTime  int64      `json:"last_enter_time,omitempty"`
	CTime          xtime.Time `json:"ctime,omitempty"`
	MTime          xtime.Time `json:"mtime,omitempty"`
}
