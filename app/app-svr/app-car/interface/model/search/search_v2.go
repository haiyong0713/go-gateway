package search

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
)

type SearchParamV2 struct {
	model.DeviceInfo
	Mid      int64  `form:"-"`
	Buvid    string `form:"buvid"`
	Keyword  string `form:"keyword" validate:"required"`
	PageNext string `form:"page_next"`
	Ps       int    `form:"ps" default:"20"`
}

type SearchRespV2 struct {
	ArcItems []*common.Item `json:"items"`
	UpItems  []*UpItemV2    `json:"up_items"`
	PageNext *PageInfo      `json:"page_next"`
	HasNext  bool           `json:"has_next"`
}

type UpItemV2 struct {
	Mid        int64  `json:"mid"`
	Name       string `json:"name"`        // up主昵称
	Face       string `json:"face"`        // 头像url（搜索返回的url有误，需调用账号接口二次查询）
	FansCount  int    `json:"fans_count"`  // 粉丝数
	VideoCount int    `json:"video_count"` // 视频数
	Desc       string `json:"desc"`        // 个人简介
}

type PageInfo struct {
	Pn int `json:"pn"`
	Ps int `json:"ps"`
}

type ArcIdsRes struct {
	Aids     []int64
	Sids     []int32
	PageNext *PageInfo
	HasNext  bool
}

type ChannelArcIdsRes struct {
	Aids       []int64
	PageNext   *PageInfo
	NextOffset string
	HasNext    bool
	Arcs       interface{}
}

type MainArcIdsRes struct {
	Aids     []int64
	Sids     []int32
	Arcs     interface{}
	Seams    interface{}
	PageNext *PageInfo
	HasNext  bool
}
