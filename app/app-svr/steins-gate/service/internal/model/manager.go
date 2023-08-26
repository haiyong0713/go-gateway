package model

import "go-gateway/app/app-svr/archive/service/api"

// ManagerGraph .
type ManagerGraph struct {
	Aid         int64    `json:"aid"`
	NodeNameArr []string `json:"-"`
	NodeNames   string   `json:"node_names"`
	EdgeNameArr []string `json:"-"`
	EdgeNames   string   `json:"edge_names"`
}

// RecentArcs .
type RecentArcs struct {
	Arcs []*api.Arc `json:"arcs"`
}

// RecentArcReq .
type RecentArcReq struct {
	Stime int64 `form:"stime"`
	Etime int64 `form:"etime"`
}
