package view

import "go-gateway/app/app-svr/app-car/interface/model"

type LikeParam struct {
	model.DeviceInfo
	Aid  int64 `form:"aid"`
	Like int   `form:"like"`
}

type CommunityParam struct {
	model.DeviceInfo
	EpId int64 `form:"ep_id"`
}

type LikeWebParam struct {
	AppKey string `form:"appkey"`
	Oid    int64  `form:"oid"`
	Like   int    `form:"like"`
}

type CommunityWebParam struct {
	Cid int64 `form:"cid"`
}

type Community struct {
	Like int `json:"like"`
}
