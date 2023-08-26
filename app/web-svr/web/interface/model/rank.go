package model

import arcmdl "go-gateway/app/app-svr/archive/service/api"

var RankV2Types = map[string]int64{
	"all":    1,
	"origin": 2,
	"rookie": 3,
}

type RankV2Cache struct {
	Note string        `json:"note"`
	List []*RankV2Item `json:"list"`
}

type RankV2Item struct {
	Aid    int64          `json:"aid"`
	Score  int64          `json:"score"`
	Play   int32          `json:"play"`
	Coin   int32          `json:"coin"`
	Danmu  int32          `json:"danmu"`
	Others []*RankV2Other `json:"others"`
}

type RankV2Other struct {
	Aid   int64 `json:"aid"`
	Score int64 `json:"score"`
}

type RankV2 struct {
	Note string       `json:"note"`
	List []*RankV2Arc `json:"list"`
}

type RankV2Arc struct {
	*arcmdl.Arc
	Bvid   string            `json:"bvid"`
	Score  int64             `json:"score"`
	Others []*RankV2OtherArc `json:"others,omitempty"`
}

type RankV2OtherArc struct {
	*arcmdl.Arc
	Bvid  string `json:"bvid"`
	Score int64  `json:"score"`
}
