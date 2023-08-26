package common

import "go-gateway/app/app-svr/app-car/interface/model"

type PayInfoReq struct {
	model.DeviceInfo
	Ptype      int   `json:"p_type" form:"p_type"`
	SeasonId   int64 `json:"season_id" form:"season_id"`
	Epid       int64 `json:"epid" form:"epid"`
	DeviceType int   `json:"device_type" form:"device_type"`
}

type PayInfoResp struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Url   string `json:"url"`
}

type PayStateReq struct {
	model.DeviceInfo
	Ptype      int   `json:"p_type" form:"p_type"`
	SeasonId   int64 `json:"season_id" form:"season_id"`
	Epid       int32 `json:"epid" form:"epid"`
	DeviceType int   `json:"device_type" form:"device_type"`
}

type PayStateResp struct {
	IsSuccess bool `json:"is_success"`
}
