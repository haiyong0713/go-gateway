package vip

import "go-gateway/app/app-svr/app-car/interface/model"

const (
	StateUserIsReceived  int8 = 1
	StateUserNotReceived int8 = 2
)

type VipParam struct {
	model.DeviceInfo
	BatchToken string `form:"batch_token"`
}

type CodeOpenParam struct {
	model.DeviceInfo
	Code string `form:"code" validate:"min=16"`
}

type VipReceived struct {
	MID        int64  `json:"mid"`
	Buvid      string `json:"buvid"`
	Channel    string `json:"channel"`
	BatchToken string `json:"batch_token"`
	OrderNo    string `json:"order_no"`
	State      int8   `json:"state"`
}
