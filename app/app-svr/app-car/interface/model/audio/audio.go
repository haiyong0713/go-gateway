package audio

import "go-gateway/app/app-svr/app-car/interface/model"

type ShowAudioParam struct {
	model.DeviceInfo
}

type ChannelAudioParam struct {
	model.DeviceInfo
	Pn        int   `form:"pn"`
	ChannelID int64 `form:"channel_id"`
}

type ReportPlayParam struct {
	model.DeviceInfo
	Aid    int64 `form:"aid"`
	Cid    int64 `form:"cid"`
	Detail int64 `form:"detail"`
}
