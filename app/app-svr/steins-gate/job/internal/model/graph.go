package model

import (
	"fmt"

	"go-gateway/app/app-svr/archive/service/api"
)

// all const
const (
	_graphPass               = 1
	SteinsRouteForStickVideo = "stick_video"
)

// Graph def
type Graph struct {
	ID    int64
	AID   int64
	State int
}

// Node def.
type Node struct {
	ID      int64
	CID     int64
	IsStart int
}

// ReqReturnGraph is the request structure to return the graph
type ReqReturnGraph struct {
	Arc     *api.Arc
	GraphID int64
}

// RetryOp def.
type RetryOp struct {
	Action   string
	Value    int64
	SubValue int64 // cid in case of sending aid & cid
}

// SteinsCid gives the real first cid of the steins-gate video
type SteinsCid struct {
	Aid   int64  `json:"aid"`
	Cid   int64  `json:"cid"`
	Route string `json:"route"`
}

// Key def.
func (v *SteinsCid) Key() string {
	return fmt.Sprintf("%d_%d", v.Aid, v.Cid)
}

// IsPass def.
func (v *Graph) IsPass() bool {
	return v.State == _graphPass
}

// EvaluationMsg def.
type EvaluationMsg struct {
	AID   int64 `json:"aid"`
	Score int64 `json:"score"`
	Time  int64 `json:"time"`
}
