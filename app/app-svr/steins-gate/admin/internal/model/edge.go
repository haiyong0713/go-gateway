package model

import (
	"fmt"
	"sort"

	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	EdgeGroupDefault       = 0
	EdgeGroupNormal        = 1
	EdgeGroupPoint         = 2
	EdgeGroupPicture       = 3
	EdgeGroupIntervention  = 4
	_NodeJumpShowtime      = 0 // jumping node, show_time == 0
	NoDuration             = -1
	NoQTE                  = -1
	NoPause                = 0
	YesPause               = 1
	StartTimeLoad          = 300
	_jumpFormat            = "JUMP %d %d"
	InvisibleEdgesToNodeID = -1 // to_node=-1的edge不可见，用于展示空edge的edge_group

	// 	_screenshotFmt = "/bfs/steins-gate/%d_screenshot.jpg"
	_qteAhead    = 2
	_qteByTheEnd = 1

	// 	_seekAheadPos = 2 // seek到选项前2秒

	// 	_evaluationFmt = "%0.1f分"
	_tooShortToQte = 5 // <=5秒无法使用真qte
	RootEdge       = 1
)

type EdgeInfoV2 struct {
	Title  string  `json:"title"`
	EdgeID int64   `json:"edge_id"`
	Edges  *EdgeV2 `json:"edges,omitempty"`
}

func (v *EdgeV2) BuildEdge(questions []*Question, toNodeInfo *api.GraphNode) {
	v.Dimension = Dimension{ // 视频云rotate已处理
		Width:  toNodeInfo.Width,
		Height: toNodeInfo.Height,
		Rotate: 0,
	}
	v.Questions = questions
}

type EdgeV2 struct {
	Dimension Dimension   `json:"dimension,omitempty"`
	Questions []*Question `json:"questions,omitempty"`
}

type Question struct {
	ID         int64       `json:"id"`
	Type       int64       `json:"type"`
	TimePosDB  int64       `json:"-"`
	StartTime  int64       `json:"start_time,omitempty"`
	StartTimeR int64       `json:"start_time_r,omitempty"`
	Duration   int64       `json:"duration"`
	PauseVideo int32       `json:"pause_video"`
	Title      string      `json:"title"`
	Choices    []*ChoiceV2 `json:"choices,omitempty"`
}

// FromEdgeGroup def.
func (v *Question) FromEdgeGroup(edgeGroup *api.EdgeGroup) {
	v.ID = edgeGroup.Id
	v.Type = edgeGroup.Type
	v.TimePosDB = edgeGroup.StartTime
	v.Duration = edgeGroup.Duration
	v.PauseVideo = edgeGroup.PauseVideo
	v.Title = edgeGroup.Title
}

// ChoiceV2 中插选项
type ChoiceV2 struct {
	ID                int64  `json:"id"`
	PlatformAction    string `json:"platform_action"`
	NativeAction      string `json:"native_action"`
	Condition         string `json:"condition"`
	Cid               int64  `json:"cid"`
	X                 int64  `json:"x,omitempty"`
	Y                 int64  `json:"y,omitempty"`
	TextAlign         int32  `json:"text_align,omitempty"`
	Option            string `json:"option"`
	CustomImageUrl    string `json:"custom_image_url,omitempty"`
	CustomImageWidth  string `json:"custom_image_width,omitempty"`
	CustomImageHeight string `json:"custom_image_height,omitempty"`
	CustomImageRotate string `json:"custom_image_rotate,omitempty"`
	IsDefault         int32  `json:"is_default,omitempty"`
}

func IsRootEdge(edgeID int64) (res bool) {
	res = edgeID == 0 || edgeID == 1
	return
}

// 老的非中插树需要构建出group
func BuildEdgeGroup(node *api.GraphNode) (out map[int64]*api.EdgeGroup) {
	out = make(map[int64]*api.EdgeGroup)
	group := new(api.EdgeGroup)
	group.Type = int64(node.Otype)          // 将node的类型1=普通，2=定点位转化为新的edge_group的类型
	if node.ShowTime == _NodeJumpShowtime { // 跳转节点类型为0
		group.Type = EdgeGroupDefault
	}
	out[group.Id] = group
	return
}

func BuildQuestions(edges []*api.GraphEdge, edgeGroups map[int64]*api.EdgeGroup,
	arcView *arcApi.SteinsGateViewReply, toNodeInfo *api.GraphNode) (out []*Question, err error) {
	questionMap := make(map[int64]*Question)
	for _, item := range edges { // 将edges根据其各自的group_id放入map中
		_, ok := questionMap[item.GroupId]
		if !ok {
			group, ok := edgeGroups[item.GroupId]
			if !ok {
				continue
			}
			question := new(Question)
			question.FromEdgeGroup(group)
			questionMap[question.ID] = question
		}
		if item.ToNode == InvisibleEdgesToNodeID { // 不可见的edge仅加入其edgeGroup, 不加入choice
			continue
		}
		choice := new(ChoiceV2)
		choice.FromEdge(item, questionMap[item.GroupId].Type, len(questionMap[item.GroupId].Choices))
		questionMap[item.GroupId].Choices = append(questionMap[item.GroupId].Choices, choice)
	}
	var noIntervention []*Question
	for _, v := range questionMap {
		if v.Type == EdgeGroupIntervention {
			out = append(out, v)
			continue
		}
		noIntervention = append(noIntervention, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TimePosDB < out[j].TimePosDB })
	out = append(out, noIntervention...)
	timeHandle(out, toNodeInfo, arcView)
	return
}

// 处理对应时间相关
func timeHandle(questions []*Question, toNodeInfo *api.GraphNode, arcView *arcApi.SteinsGateViewReply) {
	arcPage := PickPagesMap(arcView)
	qteStyle := _qteAhead
	if page, ok := arcPage[toNodeInfo.Cid]; ok && page.Duration <= _tooShortToQte { // 视频长度小于等于5秒时候强制下发qte_style=1
		qteStyle = _qteByTheEnd
	}
	for _, question := range questions {
		switch question.Type {
		case EdgeGroupDefault: //
			question.Duration = NoDuration
			question.StartTimeR = 0
			question.PauseVideo = NoPause
		case EdgeGroupNormal, EdgeGroupPoint, EdgeGroupPicture:
			if toNodeInfo.ShowTime == NoQTE { // no qte
				question.Duration = NoDuration
				question.StartTimeR = StartTimeLoad
				question.PauseVideo = YesPause
			}
			if toNodeInfo.ShowTime > 0 {
				if qteStyle == _qteByTheEnd {
					//nolint:gomnd
					question.Duration = toNodeInfo.ShowTime * 1000
					question.StartTimeR = StartTimeLoad
					question.PauseVideo = YesPause
				}
				if qteStyle == _qteAhead {
					//nolint:gomnd
					question.Duration = toNodeInfo.ShowTime * 1000
					//nolint:gomnd
					question.StartTimeR = toNodeInfo.ShowTime * 1000
					question.PauseVideo = NoPause
				}
			}
		case EdgeGroupIntervention:
			question.StartTime = question.TimePosDB
			// duration, pauseVideo pick db
		}
	}
}

func (v *ChoiceV2) FromEdge(edge *api.GraphEdge, groupType int64, index int) {
	v.ID = edge.Id
	v.Cid = edge.ToNodeCid
	v.IsDefault = edge.IsDefault
	v.Option = edge.Title
	if groupType == EdgeGroupDefault || groupType == EdgeGroupNormal ||
		groupType == EdgeGroupPoint || groupType == EdgeGroupPicture {
		v.PlatformAction = fmt.Sprintf(_jumpFormat, v.ID, v.Cid)
		v.Option = FromOpt(index, edge.Title)
	}
	if groupType == EdgeGroupPoint {
		v.X = edge.PosX
		v.Y = edge.PosY
		v.TextAlign = edge.TextAlign
	}
}

func PickPagesMap(arcView *arcApi.SteinsGateViewReply) (arcPage map[int64]*arcApi.Page) {
	arcPage = make(map[int64]*arcApi.Page)
	if arcView == nil { // 预览传入nil，暂时无seek逻辑
		return
	}
	for _, v := range arcView.Pages {
		arcPage[v.Cid] = v
	}
	return

}
