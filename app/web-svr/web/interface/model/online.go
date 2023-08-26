package model

import v1 "go-gateway/app/app-svr/archive/service/api"

// Online struct of online api response
type Online struct {
	RegionCount map[int32]int64 `json:"region_count"`
}

type OnlineTotal struct {
	BuvidCount int64 `json:"buvid_count"`
	ConnCount  int64 `json:"conn_count"`
	IPCount    int64 `json:"ip_count"`
}

// OnlineAid online aids and count
type OnlineAid struct {
	Aid   int64 `json:"aid"`
	Count int64 `json:"count"`
}

// OnlineArc archive whit online count
type OnlineArc struct {
	*v1.Arc
	Bvid        string `json:"bvid"`
	OnlineCount int64  `json:"online_count"`
}

// StrongGuide.
type StrongGuide struct {
	IsShow bool   `json:"is_show"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}
