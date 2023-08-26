package space

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	sch "go-gateway/app/app-svr/app-car/interface/model/search"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

type SpaceParamV2 struct {
	model.DeviceInfo
	Mid      int64  `form:"-"`
	Buvid    string `form:"buvid"`
	UpMid    int64  `form:"up_mid" validate:"required"`
	PageNext string `form:"page_next"`
	Ps       int    `form:"ps" default:"20"`
}

type SpaceRespV2 struct {
	Account  *AccountInfo   `json:"account"`
	ArcItems []*common.Item `json:"items"`
	PageNext *sch.PageInfo  `json:"page_next"`
	HasNext  bool           `json:"has_next"`
}

type AccountInfo struct {
	Mid        int64                `json:"mid"`
	Name       string               `json:"name"`
	Face       string               `json:"face"`
	FansCount  int64                `json:"fans_count"`
	VideoCount int64                `json:"video_count"`
	Relation   *model.Relation      `json:"relation"`
	VipInfo    *accountgrpc.VipInfo `json:"vip"`
}

type ArcIdsRes struct {
	Aids     []int64
	Total    int64
	PageNext *sch.PageInfo
	HasNext  bool
}
