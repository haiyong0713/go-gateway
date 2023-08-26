package common

import "go-gateway/app/app-svr/app-car/interface/model"

const (
	ViewTypeUGC          = "ugc"
	ViewTypeUgcMulti     = "ugc_multi"
	ViewTypeUgcSingle    = "ugc_single"
	ViewTypeOGV          = "ogv"
	ViewTypeVideoSerial  = "video_serial"
	ViewTypeVideoChannel = "video_channel"
	ViewTypeFmSerial     = "fm_serial"
	ViewTypeFmChannel    = "fm_channel"
)

type ViewDetailReq struct {
	model.DeviceInfo
	Otype string `json:"otype" form:"otype"` // 与commonmdl.ItemType定义一致
	Oid   int64  `json:"oid" form:"oid"`
}

type ViewRcmdReq struct {
	model.DeviceInfo
	Otype      string `json:"otype" form:"otype"`
	Oid        int64  `json:"oid" form:"oid"`
	LoginEvent int64  `json:"login_event" form:"login_event"`
}

type ViewRcmdResp struct {
	Items []*Item `json:"items"`
}

type ViewV2SerialReq struct {
	model.DeviceInfo
	Otype    string `form:"type"`
	Oid      int64  `form:"id"`
	Mid      int64  `form:"-"`
	Buvid    string `form:"-"`
	PageNext string `form:"page_next"`
	PagePre  string `form:"page_previous"`
	Ps       int    `form:"ps"`
	Aid      int64  `form:"aid"`
}

type ViewV2SerialResp struct {
	Cards        []*Item              `json:"cards"`
	PageNext     *PageInfo            `json:"page_next"`
	PagePrevious *PageInfo            `json:"page_previous"`
	HasNext      bool                 `json:"has_next"`
	HasPrevious  bool                 `json:"has_previous"`
	History      *ViewV2SerialHistory `json:"history"`
}

type ViewV2SerialHistory struct {
	Aid      int64 `json:"aid"`
	Progress int64 `json:"progress"`
}

type TeslaMediaParseResp struct {
	Param    string `json:"param"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Cover    string `json:"cover"`
	Duration int64  `json:"duration"`
}
