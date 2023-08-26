package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go-common/library/log"

	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	EdgeGroupDefault       = 0
	EdgeGroupNormal        = 1
	EdgeGroupPoint         = 2
	EdgeGroupPicture       = 3
	EdgeGroupIntervention  = 4
	EdgeGroupBnj2020       = 127
	_NodeJumpShowtime      = 0 // jumping node, show_time == 0
	NoDuration             = -1
	NoQTE                  = -1
	NoPause                = 0
	YesPause               = 1
	StartTimeLoad          = 300
	_jumpFormat            = "JUMP %d %d"
	_UpdateCurrentTime     = "UPDATE_CURRENT_TIME %.2f"
	InvisibleEdgesToNodeID = -1 // to_node=-1的edge不可见，用于展示空edge的edge_group
	_toTypeToTime          = 1
)

// EdgeInfoV2 参数 仅多一个choicesIDs
type EdgeInfoV2Param struct {
	NodeInfoParam
	ChoiceIDs []int64 `form:"choices,split" validate:"dive,gt=0"`
}

// ClientCompatibility 针对客户端兼容
func (v *EdgeInfoV2Param) ClientCompatibility() {
	v.NodeID = v.EdgeID
}

// 前台不会传跳转节点过来，需要加入非root的跳转节点
func (v *EdgeInfoV2Param) ChoiceHandler() {
	if !IsRootEdge(v.EdgeID) {
		v.ChoiceIDs = append(v.ChoiceIDs, v.EdgeID)
	}
}

// NodeInfo def.
type EdgeInfoV2 struct {
	Title          string       `json:"title"`
	EdgeID         int64        `json:"edge_id"`
	StoryList      []*Story     `json:"story_list,omitempty"`
	Edges          *EdgeV2      `json:"edges,omitempty"`
	Buvid          string       `json:"buvid,omitempty"`
	Preload        *Preload     `json:"preload,omitempty"`
	HiddenVars     []*HiddenVar `json:"hidden_vars,omitempty"` // 用于播放器展示隐藏变量
	IsLeaf         int64        `json:"is_leaf"`               // 是否叶子节点，1=叶子节点即不是结尾；0=不是叶子即结尾节点
	NoTutorial     int32        `json:"no_tutorial,omitempty"`
	NoBacktracking int32        `json:"no_backtracking,omitempty"`
	NoEvaluation   int32        `json:"no_evaluation,omitempty"`
}

type EdgeV2 struct {
	Dimension Dimension   `json:"dimension,omitempty"`
	Questions []*Question `json:"questions,omitempty"`
	Skin      *api.Skin   `json:"skin,omitempty"`
}

// Question 对应选项页，即edge group
type Question struct {
	ID          int64       `json:"id"`
	Type        int64       `json:"type"`
	TimePosDB   int64       `json:"-"`
	StartTime   int64       `json:"start_time,omitempty"`
	StartTimeR  int64       `json:"start_time_r,omitempty"`
	Duration    int64       `json:"duration"`
	PauseVideo  int32       `json:"pause_video"`
	Title       string      `json:"title"`
	Choices     []*ChoiceV2 `json:"choices,omitempty"`
	FadeInTime  int32       `json:"fade_in_time,omitempty"`
	FadeOutTime int32       `json:"fade_out_time,omitempty"`
}

// FromEdgeGroup def.
func (v *Question) FromEdgeGroup(edgeGroup *api.EdgeGroup) {
	v.ID = edgeGroup.Id
	v.Type = edgeGroup.Type
	v.TimePosDB = edgeGroup.StartTime
	v.Duration = edgeGroup.Duration
	v.PauseVideo = edgeGroup.PauseVideo
	v.Title = edgeGroup.Title
	v.FadeInTime = edgeGroup.FadeInTime
	v.FadeOutTime = edgeGroup.FadeOutTime
}

// ChoiceV2 中插选项
type ChoiceV2 struct {
	ID                int64                   `json:"id"`
	PlatformAction    string                  `json:"platform_action"`
	NativeAction      string                  `json:"native_action"`
	Condition         string                  `json:"condition"`
	Cid               int64                   `json:"cid"`
	X                 int64                   `json:"x,omitempty"`
	Y                 int64                   `json:"y,omitempty"`
	TextAlign         int32                   `json:"text_align,omitempty"`
	Option            string                  `json:"option"`
	CustomImageUrl    string                  `json:"custom_image_url,omitempty"`
	CustomImageWidth  string                  `json:"custom_image_width,omitempty"`
	CustomImageHeight string                  `json:"custom_image_height,omitempty"`
	CustomImageRotate string                  `json:"custom_image_rotate,omitempty"`
	IsDefault         int32                   `json:"is_default,omitempty"`
	Selected          *api.EdgeFrameAnimation `json:"selected,omitempty"`
	Submited          *api.EdgeFrameAnimation `json:"submited,omitempty"`
	IsHidden          int32                   `json:"is_hidden,omitempty"`
	Width             int32                   `json:"width,omitempty"`
	Height            int32                   `json:"height,omitempty"`
}

func (v *ChoiceV2) BuildPreVideo(aid int64) (out *PreVideo) {
	out = new(PreVideo)
	out.Aid = aid
	out.Cid = v.Cid
	return
}

// 将语法树结构和表达式结构转化为中插选项的native_action表达式字段
func (v *ChoiceV2) genNativeAction(edge *api.GraphEdge) {
	var (
		attributes  []*EdgeAttribute
		expressions []string
	)
	if err := json.Unmarshal([]byte(edge.Attribute), &attributes); err != nil {
		log.Error("Edge %+v, Json Unmarshal Attributes Err %v", edge, err)
		return
	}
	for _, item := range attributes {
		switch item.ActionType {
		case attributeIsSyntax: // 语法树转为表达式后输出
			newItem := new(EdgeAttribute)
			newItem.FromSyntax(item)
			expressions = append(expressions, newItem.Action)
		case attributeIsExpr: // 表达式直出action
			expressions = append(expressions, item.Action)
		default:
			log.Error("Edge %+v, Invalid Type %d", edge, item.ActionType)
		}
	}
	if len(expressions) == 0 {
		return
	}
	v.NativeAction = strings.Join(expressions, split)
}

// 将语法树结构和表达式结构转化为中插选项的condition表达式字段
func (v *ChoiceV2) genCondition(edge *api.GraphEdge) {
	var (
		conds       []*EdgeCondition
		expressions []string
	)
	if err := json.Unmarshal([]byte(edge.Condition), &conds); err != nil {
		log.Error("Edge %+v, Json Unmarshal Condition Err %v", edge, err)
		return
	}
	for _, item := range conds {
		switch item.ActionType {
		case attributeIsSyntax: // 语法树转为表达式后输出
			newItem := new(EdgeCondition)
			newItem.FromSyntax(item)
			expressions = append(expressions, newItem.Condition)
		case attributeIsExpr: // 表达式直出action
			expressions = append(expressions, item.Condition)
		default:
			log.Error("Edge %+v, Invalid Type %d", edge, item.ActionType)
		}
	}
	if len(expressions) == 0 {
		return
	}
	v.Condition = strings.Join(expressions, conditionSeparator)
}

func (v *ChoiceV2) FromEdge(edge *api.GraphEdge, groupType int64, graph *api.GraphInfo, index int, anms map[string]*api.EdgeFrameAnimation) {
	if IsInterventionGraph(graph) || IsEdgeGraph(graph) {
		v.ID = edge.Id
	} else {
		v.ID = edge.ToNode
	}
	v.Cid = edge.ToNodeCid
	v.IsDefault = edge.IsDefault
	v.Option = edge.Title
	if groupType == EdgeGroupDefault || groupType == EdgeGroupNormal ||
		groupType == EdgeGroupPoint || groupType == EdgeGroupPicture ||
		groupType == EdgeGroupBnj2020 {
		v.PlatformAction = fmt.Sprintf(_jumpFormat, v.ID, v.Cid)
		v.Option = FromOpt(index, edge.Title)
	}
	if groupType == EdgeGroupPoint {
		v.X = edge.PosX
		v.Y = edge.PosY
		v.TextAlign = edge.TextAlign
	}
	if groupType == EdgeGroupBnj2020 {
		v.Selected = anms["selected"]
		v.Submited = anms["submited"]
		v.IsHidden = edge.IsHidden
		v.X = edge.PosX
		v.Y = edge.PosY
		v.Width = edge.Width
		v.Height = edge.Height
	}
	if groupType == EdgeGroupIntervention && edge.ToType == _toTypeToTime { // 中插跳进度功能
		iTime := edge.ToTime
		//nolint:gomnd
		toTime := float64(iTime) * 0.001
		v.PlatformAction = fmt.Sprintf(_UpdateCurrentTime, toTime)
	}
	if edge.Attribute != "" { // 展示表达式
		v.genNativeAction(edge)
	}
	if edge.Condition != "" { // 展示条件表达式
		v.genCondition(edge)
	}
}

// 支持中插的结构返回
func (res *EdgeInfoV2) BuildEdgeV2Reply(toNodeInfo *api.GraphNode, edgeInfo *api.GraphEdge, questions []*Question, skin *api.Skin) {
	res.Title = toNodeInfo.Name
	res.EdgeID = edgeInfo.Id
	res.Edges = new(EdgeV2)
	res.Edges.BuildEdge(questions, toNodeInfo, skin)
}

func (v *EdgeV2) BuildEdge(questions []*Question, toNodeInfo *api.GraphNode, skin *api.Skin) {
	v.Dimension = Dimension{ // 视频云rotate已处理
		Width:  toNodeInfo.Width,
		Height: toNodeInfo.Height,
		Rotate: 0,
		Sar:    toNodeInfo.Sar,
	}
	v.Questions = questions
	v.Skin = skin
}

// 根据edge和edge group构建question
func BuildQuestions(edges []*api.GraphEdge, edgeGroups map[int64]*api.EdgeGroup,
	qtePicker QTEStylePicker, toNodeInfo *api.GraphNode,
	graph *api.GraphInfo, edgeAnms map[int64]*api.EdgeFrameAnimations) (out []*Question, isLeaf int64, err error) {
	getAnms := func(edgeID int64) map[string]*api.EdgeFrameAnimation {
		eAnms, ok := edgeAnms[edgeID]
		if !ok {
			return nil
		}
		return eAnms.Animations
	}
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
		choice.FromEdge(item, questionMap[item.GroupId].Type, graph, len(questionMap[item.GroupId].Choices), getAnms(item.Id))
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
	if len(noIntervention) == 0 { // 无type=0/1/2/3的枝杈选项，所以为末梢的叶子节点
		isLeaf = 1
	}
	out = append(out, noIntervention...)
	timeHandle(out, toNodeInfo, qtePicker)
	return
}

type QTEStylePicker func(cid int64) int

func ArcViewPicker(arcView *arcApi.SteinsGateViewReply) QTEStylePicker {
	arcPage := PickPagesMap(arcView)
	return func(cid int64) int {
		qteStyle := _qteAhead
		if page, ok := arcPage[cid]; ok && page.Duration <= _tooShortToQte { // 视频长度小于等于5秒时候强制下发qte_style=1
			qteStyle = _qteByTheEnd
		}
		return qteStyle
	}
}

func VideoUpViewPicker(upView *VideoUpView) QTEStylePicker {
	arcPage := make(map[int64]*Video, len(upView.Videos))
	for _, v := range upView.Videos {
		arcPage[v.Cid] = v
	}
	return func(cid int64) int {
		qteStyle := _qteAhead
		if page, ok := arcPage[cid]; ok && page.Duration <= _tooShortToQte { // 视频长度小于等于5秒时候强制下发qte_style=1
			qteStyle = _qteByTheEnd
		}
		return qteStyle
	}
}

// 处理对应时间相关
func timeHandle(questions []*Question, toNodeInfo *api.GraphNode, qtePicker QTEStylePicker) {
	qteStyle := _qteAhead
	if qtePicker != nil {
		qteStyle = qtePicker(toNodeInfo.Cid)
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
		case EdgeGroupBnj2020:
			question.StartTime = question.TimePosDB
			// duration, pauseVideo pick db
		}
	}
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

// GenEdgeStoryList 用于生成edgeInfo的storyList
func GenEdgeStoryList(arcPage map[int64]*arcApi.Page, bfsHost string, currentCursor int64,
	edgeArr, cursorArr []int64, nodes map[int64]*api.GraphNode, edges map[int64]*api.GraphEdge) (storyList []*Story) {
	for i := 0; i < len(edgeArr); i++ {
		tmpe, ok := edges[edgeArr[i]]
		if !ok {
			continue
		}
		tmpn, ok := nodes[tmpe.ToNode]
		if !ok {
			continue
		}
		story := &Story{EdgeID: edgeArr[i], Cursor: cursorArr[i]} // edge图需要单独赋值edgeid
		story.FromNode(tmpn, arcPage, bfsHost)
		story.ClientCompatibility()
		if currentCursor == cursorArr[i] { // 判断是不是亮
			story.IsCurrent = 1
		}
		storyList = append(storyList, story)
	}
	return

}
