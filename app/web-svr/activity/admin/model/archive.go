package model

import "go-gateway/app/app-svr/archive/service/api"

// ArchiveParam .
type ArchiveParam struct {
	Aids  []int64 `json:"aids" form:"aids,split"`
	Bvids string  `form:"bvids"`
}

type BvArc struct {
	*api.Arc
	Bvid string `json:"bvid"`
}
