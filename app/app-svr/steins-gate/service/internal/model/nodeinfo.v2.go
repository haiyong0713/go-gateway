package model

import (
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

func (v *EdgeInfoV2) BuildReplyV2(nodeInfo *api.GraphNode, questions []*Question, hiddenVars []*HiddenVar, skin *api.Skin) {
	v.Title = nodeInfo.Name
	v.EdgeID = nodeInfo.Id
	v.Edges = new(EdgeV2)
	v.Edges.BuildEdge(questions, nodeInfo, skin)
	v.HiddenVars = hiddenVars // v2全部透出，无需过滤
}

// GenStoryList 用于生成nodeInfo的storyList
func GenStoryList(arcPage map[int64]*arcgrpc.Page, bfsHost string, currentCursor int64,
	nodeArr, cursorArr []int64, nodes map[int64]*api.GraphNode) (storyList []*Story) {
	for i := 0; i < len(nodeArr); i++ {
		tmpn, ok := nodes[nodeArr[i]]
		if !ok {
			continue
		}
		story := &Story{EdgeID: nodeArr[i], Cursor: cursorArr[i]} // node图也需要单独赋值edgeid
		story.FromNode(tmpn, arcPage, bfsHost)
		if currentCursor == cursorArr[i] { // 当前点亮节点
			story.IsCurrent = 1
		}
		storyList = append(storyList, story)
	}
	return
}

type EdgeInfoV2PreReq struct {
	AID       int64   `form:"aid" validate:"gt=0,required"`
	Portal    int32   `form:"portal"  validate:"lte=1"`
	EdgeID    int64   `form:"edge_id"`
	ChoiceIDs []int64 `form:"choices,split" validate:"dive,gt=0"`
	NodeInfoReqCursor
}

func (v *EdgeInfoV2PreReq) ChoiceHandler() {
	if !IsRootEdge(v.EdgeID) {
		v.ChoiceIDs = append(v.ChoiceIDs, v.EdgeID)
	}
}

func (v *EdgeInfoV2PreReq) RootEdgeCompatibility() {
	v.EdgeID = RootEdge
}

type EdgeInfoV2PreReply struct {
	*EdgeInfoV2
	// HiddenVarsPreview用于编辑器展示
	HiddenVarsPreview []*HiddenVar `json:"hidden_vars_preview,omitempty"`
}
