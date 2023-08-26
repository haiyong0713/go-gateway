package model

import (
	"encoding/json"

	"go-gateway/app/app-svr/steins-gate/service/api"
)

type NodeShow struct {
	NodeCore
	Choices []*EdgeShow `json:"choices"`
}

type NodeCore struct {
	NodeID     int64  `json:"node_id"`
	Name       string `json:"name"`
	AID        int64  `json:"aid"`
	CID        int64  `json:"cid"`
	ShowTime   int64  `json:"show_time"`
	VideoTitle string `json:"video_title"`
}

func (v *NodeCore) FromProto(node *api.GraphNode, video *Video, aid int64) {
	v.NodeID = node.Id
	v.Name = node.Name
	v.CID = node.Cid
	v.ShowTime = node.ShowTime
	v.VideoTitle = video.Title
	v.AID = aid
}

type EdgeShow struct {
	Title      string    `json:"title"`
	Attributes string    `json:"attributes"`
	ToNode     *NodeCore `json:"to_node"`
}

type EdgeAttribute struct {
	TextAlign int32  `json:"text_align"`
	PosX      int64  `json:"pos_x"`
	PosY      int64  `json:"pos_y"`
	Attribute string `json:"attribute"`
	Condition string `json:"condition"`
}

func (v *EdgeAttribute) FromProto(edge *api.GraphEdge) {
	v.TextAlign = edge.TextAlign
	v.PosX = edge.PosX
	v.PosY = edge.PosY
	v.Attribute = edge.Attribute
	v.Condition = edge.Condition
}

func (v *EdgeAttribute) ToMsg() string {
	str, _ := json.Marshal(v)
	return string(str)
}

func (v *EdgeShow) FromProto(edge *api.GraphEdge, toNode *NodeCore) {
	v.Title = edge.Title
	v.ToNode = toNode
	edgeAttr := new(EdgeAttribute)
	edgeAttr.FromProto(edge)
	v.Attributes = edgeAttr.ToMsg()
}

// GraphDB is graph struct in DB
type GraphDB struct {
	api.GraphInfo
	State int
}
