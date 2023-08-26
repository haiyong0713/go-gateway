package model

import archiveapi "go-gateway/app/app-svr/archive/service/api"

type RankListType int64

const (
	RankListTypeNone   = RankListType(0)
	RankListTypeAll    = RankListType(1)
	RankListTypeOrigin = RankListType(2)
	RankListTypeRookie = RankListType(3)

	PlatH5  = 15
	PlatXcx = 16
)

var OriType = map[int64]string{
	0: "",
	1: "_origin",
}

type RankAid struct {
	Aid   int64 `json:"aid"`
	Score int64 `json:"score"`
}

type OnlineAid struct {
	Aid   int64 `json:"aid"`
	Count int64 `json:"count"`
}

type BvArc struct {
	*archiveapi.Arc
	Bvid string `json:"bvid"`
}

type RankList struct {
	Note string     `json:"note"`
	List []*RankArc `json:"list"`
}

type RankArc struct {
	Aid    int64      `json:"aid"`
	Score  int64      `json:"score"`
	Play   int64      `json:"play"`
	Coin   int64      `json:"coin"`
	Danmu  int64      `json:"danmu"`
	Others []*RankAid `json:"others,omitempty"`
}
