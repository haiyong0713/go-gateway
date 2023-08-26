package model

import (
	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

// NodeInfo def.
type NodeInfo struct {
	Title  string `json:"title"`
	NodeID int64  `json:"node_id"`
	Edges  *Edge  `json:"edges,omitempty"`
}

type Edge struct {
	Type      int32            `json:"type"`
	ShowTime  int64            `json:"show_time"`
	Dimension arcApi.Dimension `json:"dimension,omitempty"`
	Version   int32            `json:"version"`
	Choices   []*Choice        `json:"choices,omitempty"`
}

type Choice struct {
	NodeID    int64  `json:"node_id"`
	CID       int64  `json:"cid"`
	Option    string `json:"option"`
	IsDefault int32  `json:"is_default,omitempty"`
	X         int64  `json:"x,omitempty"`
	Y         int64  `json:"y,omitempty"`
	TextAlign int32  `json:"text_align,omitempty"`
	Condition string `json:"-"`
}

func BuildChoices(in []*api.GraphEdge) (out []*Choice) {
	for k, v := range in {
		ch := new(Choice)
		ch.FromGraphEdge(k, v)
		out = append(out, ch)
	}
	return
}

func (v *Choice) FromGraphEdge(k int, in *api.GraphEdge) {
	v.NodeID = in.ToNode
	v.CID = in.ToNodeCid
	v.Option = FromOpt(k, in.Title)
	v.IsDefault = in.IsDefault
	v.X = in.PosX
	v.Y = in.PosY
	v.TextAlign = in.TextAlign
	v.Condition = in.Condition
}

func FromOpt(k int, opt string) string {
	var optPre = map[int]string{
		0: "A ",
		1: "B ",
		2: "C ",
		3: "D ",
	}
	return optPre[k] + opt
}

func BuildReply(nodeInfo *api.GraphNode, choices []*Choice) (res *NodeInfo) {
	res = new(NodeInfo)
	res.Title = nodeInfo.Name
	res.NodeID = nodeInfo.Id
	if len(choices) > 0 { // ending 节点不出edges
		res.Edges = &Edge{
			Type:     nodeInfo.Otype,
			ShowTime: nodeInfo.ShowTime,
			Dimension: arcApi.Dimension{ // 视频云rotate已处理
				Width:  nodeInfo.Width,
				Height: nodeInfo.Height,
				Rotate: 0,
			},
			Choices: choices,
			Version: 1,
		}
	}
	return

}
