package model

import (
	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

func (res *NodeInfo) BuildEdgeReply(toNodeInfo *api.GraphNode, edgeInfo *api.GraphEdge, choices []*Choice,
	arcView *arcApi.SteinsGateViewReply, hiddenVars []*HiddenVar) {
	res.Title = toNodeInfo.Name
	res.EdgeID = edgeInfo.Id
	res.ClientCompatibility()
	arcPage := PickPagesMap(arcView)
	qteStyle := _qteAhead
	if page, ok := arcPage[edgeInfo.ToNodeCid]; ok && page.Duration <= _tooShortToQte { // 视频长度小于等于5秒时候强制下发qte_style=1
		qteStyle = _qteByTheEnd
	}
	if len(choices) > 0 { // ending 节点不出edges
		for i := 0; i < len(choices); i++ {
			choices[i].ClientCompatibility()
		}
		res.Edges = new(Edge)
		res.Edges.BuildEdge(choices, toNodeInfo, qteStyle)
	}
	if len(hiddenVars) > 0 {
		for _, v := range hiddenVars {
			if v.IsDisplay() {
				newV := new(HiddenVarInt)
				newV.FromHVar(v)
				res.HiddenVars = append(res.HiddenVars, newV) // 将float强转为int输出
			}
		}
	}

}
