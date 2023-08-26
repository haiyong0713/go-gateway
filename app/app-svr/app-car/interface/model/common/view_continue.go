package common

import (
	"go-gateway/app/app-svr/app-car/interface/model"
)

type ViewContinueReq struct {
	model.DeviceInfo
	PageNext string `json:"page_next" form:"page_next"`
	Ps       int    `json:"ps" form:"ps"`
}

type ViewContinueResp struct {
	Items    []*Item               `json:"items"`
	PageNext *ViewContinuePageNext `json:"page_next"`
	Fm       *ViewContinueFm       `json:"fm"`
}

type ViewContinuePageNext struct {
	Business string `json:"business"`
	ViewAt   int64  `json:"view_at"`
	Ps       int    `json:"ps"`
}

type ViewContinueFm struct {
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	ViewAt int64  `json:"view_at"`
}

type HistoryTabReq struct {
	model.DeviceInfo
	Ps    int64 `form:"ps" default:"8" validate:"min=1"`
	IsWeb bool  `form:"-"`
}

type HistoryTabResp struct {
	SubTab []*HistoryTabSubTab `json:"sub_tab"`
}

type HistoryTabSubTab struct {
	TabName  string              `json:"tab_name"`
	TabType  string              `json:"tab_type"`
	ShowMore bool                `json:"show_more"`
	PageNext *HistoryTabPageNext `json:"page_next"`
	Items    []*HisItem          `json:"items"`
}

type HisItem struct {
	*Item
	PlayType string `json:"play_type"`
}

type HistoryTabPageNext struct {
	ViewAt       int64 `json:"view_at" form:"view_at"`
	Max          int64 `json:"max" form:"max"`
	SerialID     int64 `json:"serial_id" form:"serial_id"`
	SerialIDType int64 `json:"serial_id_type" form:"serial_id_type"`
}

type HistoryTabMoreReq struct {
	model.DeviceInfo
	Ps      int64  `form:"ps" default:"8" validate:"min=1,max=50"`
	TabType string `form:"tab_type" validate:"min=1"`
	HistoryTabPageNext
}

type HistoryTabMoreResp struct {
	PageNext *HistoryTabPageNext `json:"page_next"`
	Items    []*HisItem          `json:"items"`
	ShowMore bool                `json:"show_more"`
}

var (
	ItemTypeToSerialBusinessType = map[ItemType]int64{
		ItemTypeVideoSerial:  1,
		ItemTypeVideoChannel: 2,
		ItemTypeFmSerial:     3,
		ItemTypeFmChannel:    4,
	}
	SerialBusinessTypeToItemType = map[int64]ItemType{
		1: ItemTypeVideoSerial,
		2: ItemTypeVideoChannel,
		3: ItemTypeFmSerial,
		4: ItemTypeFmChannel,
	}
)

func FromHisSourceToPlayType(source string) string {
	if source == "car-audio" {
		return "audio"
	}
	return "video"
}
