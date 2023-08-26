package common

import (
	"go-gateway/app/app-svr/app-car/interface/model"
)

const (
	DynTypeVideo          = 8
	DynTypeBangumi        = 512
	DynTypePGCBangumi     = 4097
	DynTypePGCMovie       = 4098
	DynTypePGCGuoChuang   = 4100
	DynTypePGCDocumentary = 4101
)

func DynamicIsUGC(dynaType int64) bool {
	return dynaType == DynTypeVideo
}

func DynamicIsOGV(dynaType int64) bool {
	return (dynaType == DynTypeBangumi) || (dynaType == DynTypePGCBangumi) ||
		(dynaType == DynTypePGCMovie) || (dynaType == DynTypePGCGuoChuang) ||
		(dynaType == DynTypePGCDocumentary)
}

type DynamicReq struct {
	model.DeviceInfo
	PageNext   string `json:"page_next" form:"page_next"`
	NeedUplist bool   `json:"need_uplist" form:"need_uplist"`
	Vmid       int64  `json:"vmid" form:"vmid"`
}

type DynamicResp struct {
	Uplist      []*Uplist        `json:"up_list"`
	DynamicList []*Item          `json:"dynamic_list"`
	PageNext    *DynamicPageNext `json:"page_next"`
	HasNext     bool             `json:"has_next"`
}

type Uplist struct {
	HasUpdate bool   `json:"has_update"`
	Mid       int64  `json:"mid"`
	Name      string `json:"name"`
	Face      string `json:"face"`
}

type DynamicPageNext struct {
	//UpdateNum      int64  `json:"update_num"`
	HistoryOffset string `json:"history_offset"`
	//UpdateBaseline string `json:"update_baseline"`
	Page int64 `json:"page" default:"1" validate:"min=1"`
}
