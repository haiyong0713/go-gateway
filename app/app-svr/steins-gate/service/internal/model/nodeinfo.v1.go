package model

import (
	"fmt"
	"math"
	"strconv"

	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	_screenshotFmt = "/bfs/steins-gate/%d_screenshot.jpg"
	_qteAhead      = 2
	_qteByTheEnd   = 1
	_seekAheadPos  = 2 // seek到选项前2秒
	_evaluationFmt = "%0.1f分"
	_tooShortToQte = 5 // <=5秒无法使用真qte
)

// NodeinfoPreReq is the request structure of the "nodeinfo" preview request
type NodeinfoPreReq struct {
	NodeID int64  `form:"node_id"`
	AID    int64  `form:"aid"`
	BvID   string `form:"bvid"`
	Portal int32  `form:"portal"  validate:"lte=1"`
	EdgeID int64  `form:"edge_id"`
	NodeInfoReqCursor
}

type NodeInfoReqCursor struct {
	CursorStr string `form:"cursor"`
	Cursor    int64  `form:"-"`
}

// ClientCompatibility 针对客户端兼容
func (v *NodeinfoPreReq) ClientCompatibility() {
	v.EdgeID = v.NodeID
}

func (v *NodeinfoPreReq) RootEdgeCompatibility() {
	v.EdgeID = RootEdge
}

func (v *NodeInfoReqCursor) HandlerCursor() {
	if v.CursorStr == "" {
		v.Cursor = IllegalCursor // 初始值为IllegalCursor，代表前台未传这个参数
		return
	}
	v.Cursor, _ = strconv.ParseInt(v.CursorStr, 10, 64)
}

func (v *NodeInfoParam) HandlerBvid() (err error) {
	if v.AID == 0 {
		if v.BvID == "" { // 如果aid和bvid都没有，直接返回
			return ecode.AidBvidNil
		}
		if v.AID, _ = GetAvID(v.BvID); v.AID == 0 {
			return ecode.BvidIllegal
		}
	}
	return
}

func (v *NodeinfoPreReq) HandlerBvid() (err error) {
	if v.AID == 0 {
		if v.BvID == "" { // 如果aid和bvid都没有，直接返回
			return fmt.Errorf("请输入aid/bvid！")
		}
		if v.AID, _ = GetAvID(v.BvID); v.AID == 0 {
			return fmt.Errorf("bvid非法！")
		}
	}
	return
}

// NodeInfoPreReply is the reply structure of the "nodeinfo" preview reply
type NodeInfoPreReply struct {
	*NodeInfo
	// HiddenVarsPreview用于编辑器展示
	HiddenVarsPreview []*HiddenVar `json:"hidden_vars_preview,omitempty"`
}

type NodeInfoParam struct {
	AID          int64  `form:"aid"`
	BvID         string `form:"bvid"`
	Portal       int32  `form:"portal"`
	NodeID       int64  `form:"node_id"`
	EdgeID       int64  `form:"edge_id"` // 对于edge图应该传edge_id,暂时兼容客户端逻辑使得edge_id=node_id
	GraphVersion int64  `form:"graph_version" validate:"gt=0,required"`
	Delay        int64  `form:"delay"`  // millisecond
	Screen       int    `form:"screen"` // from 0 to 8
	Buvid        string `form:"buvid"`
	MobiApp      string `form:"mobi_app"`
	Build        int64  `form:"build"`
	Channel      string `form:"channel"`
	Platform     string `form:"platform"`
	NodeInfoReqCursor
}

// 需要判断是不是Root Edge
func IsRootEdge(edgeID int64) (res bool) {
	res = edgeID == 0 || edgeID == 1
	return
}

// ClientCompatibility 针对客户端兼容
func (v *NodeInfoParam) ClientCompatibility() {
	v.EdgeID = v.NodeID
}

// 给root edge赋值
func (v *NodeInfoParam) RootEdgeCompatibility() {
	v.EdgeID = RootEdge
}

// NodeInfo def.
type NodeInfo struct {
	Title     string   `json:"title"`
	NodeID    int64    `json:"node_id"`
	EdgeID    int64    `json:"edge_id"`
	StoryList []*Story `json:"story_list,omitempty"`
	Edges     *Edge    `json:"edges,omitempty"`
	Buvid     string   `json:"buvid,omitempty"`
	Preload   *Preload `json:"preload,omitempty"`
	// HiddenVars 用于播放器展示
	HiddenVars []*HiddenVarInt `json:"hidden_vars,omitempty"`
}

func (v *NodeInfo) ClientCompatibility() {
	v.NodeID = v.EdgeID
}

// Preload def.
type Preload struct {
	Video []*PreVideo `json:"video"`
}

// PreVideo def.
type PreVideo struct {
	Aid int64 `json:"aid"`
	Cid int64 `json:"cid"`
}

// BuildPreVideo def.
func (v *Choice) BuildPreVideo(aid int64) (out *PreVideo) {
	out = new(PreVideo)
	out.Aid = aid
	out.Cid = v.CID
	return
}

// FromNode def.
func (pre *Preload) FromNode(nodeinfo *api.GraphNode, choices []*Choice, aid int64) {
	var (
		usedCids   = make(map[int64]struct{})
		tmpChoices []*Choice
	)
	if nodeinfo.ShowTime > 0 {
		for _, v := range choices { // QTE时置顶默认选项ci
			if v.IsDefault == 1 {
				tmpChoices = append([]*Choice{v}, tmpChoices...)
			} else {
				tmpChoices = append(tmpChoices, v)
			}
		}
	} else {
		tmpChoices = choices
	}
	for _, v := range tmpChoices { // 过滤重复cid
		if _, ok := usedCids[v.CID]; ok {
			continue
		}
		usedCids[v.CID] = struct{}{}
		pre.Video = append(pre.Video, v.BuildPreVideo(aid))
	}
}

// FromNodeV2 def.
func (pre *Preload) FromNodeV2(nodeinfo *api.GraphNode, questions []*Question, aid int64) {
	var (
		usedCids   = make(map[int64]struct{})
		tmpChoices []*ChoiceV2
	)
	for _, v := range questions { // QTE时置顶默认选项ci
		if v.Type == EdgeGroupDefault || v.Type == EdgeGroupNormal || v.Type == EdgeGroupPoint || v.Type == EdgeGroupPicture || v.Type == EdgeGroupBnj2020 {
			for _, item := range v.Choices {
				if item.IsDefault == 1 && nodeinfo.ShowTime > 0 {
					tmpChoices = append([]*ChoiceV2{item}, tmpChoices...)
				} else {
					tmpChoices = append(tmpChoices, item)
				}
			}
		}
	}
	for _, v := range tmpChoices { // 过滤重复cid
		if _, ok := usedCids[v.Cid]; ok {
			continue
		}
		usedCids[v.Cid] = struct{}{}
		pre.Video = append(pre.Video, v.BuildPreVideo(aid))
	}
}

// FromNode 因为story实际上和node是强绑定的，所以保持为FromNode
func (story *Story) FromNode(node *api.GraphNode, arcPage map[int64]*arcApi.Page, bfsHost string) {
	story.NodeID = node.Id
	story.Title = node.Name
	story.Cid = node.Cid
	story.Cover = fmt.Sprintf(bfsHost+_screenshotFmt, node.Cid)
	if len(arcPage) == 0 { // 预览传入nil，暂时无seek逻辑
		return
	}
	if page, ok := arcPage[node.Cid]; ok && page.Duration > 0 {
		if node.ShowTime > 0 && page.Duration > _tooShortToQte { // 仅当qteStyle为2 && 开启了倒计时 && 视频长度 > 5 => 才倒计时
			//nolint:gomnd
			story.StartPos = (page.Duration - node.ShowTime - _seekAheadPos) * 1000
		} else {
			//nolint:gomnd
			story.StartPos = (page.Duration - _seekAheadPos) * 1000
		}
		if story.StartPos < 0 {
			story.StartPos = 0
		}
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

func (v *NodeInfo) BuildReply(nodeInfo *api.GraphNode, choices []*Choice, arcView *arcApi.SteinsGateViewReply, hiddenVars []*HiddenVar) {
	arcPage := PickPagesMap(arcView)
	v.Title = nodeInfo.Name
	v.NodeID = nodeInfo.Id
	qteStyle := _qteAhead
	if page, ok := arcPage[nodeInfo.Cid]; ok && page.Duration <= _tooShortToQte { // 视频长度小于等于5秒时候强制下发qte_style=1
		qteStyle = _qteByTheEnd
	}
	if len(choices) > 0 { // ending 节点不出edges
		v.Edges = new(Edge)
		v.Edges.BuildEdge(choices, nodeInfo, qteStyle)
	}
	if len(hiddenVars) > 0 {
		for _, val := range hiddenVars {
			if val.IsDisplay() {
				newV := new(HiddenVarInt)
				newV.FromHVar(val)
				v.HiddenVars = append(v.HiddenVars, newV)
			}
		}
	}
}

type Story struct {
	NodeID    int64  `json:"node_id"`
	EdgeID    int64  `json:"edge_id"`
	Title     string `json:"title"`
	Cid       int64  `json:"cid"`
	StartPos  int64  `json:"start_pos"`
	Cover     string `json:"cover"`
	IsCurrent int    `json:"is_current,omitempty"`
	Cursor    int64  `json:"cursor"`
}

func (story *Story) ClientCompatibility() {
	story.NodeID = story.EdgeID
}

type Edge struct {
	Type      int32     `json:"type"`
	ShowTime  int64     `json:"show_time"`
	QteStyle  int       `json:"qte_style"`
	Dimension Dimension `json:"dimension,omitempty"`
	Version   int32     `json:"version"`
	Choices   []*Choice `json:"choices,omitempty"`
}

type Choice struct {
	NodeID    int64     `json:"node_id"`
	EdgeID    int64     `json:"edge_id"`
	CID       int64     `json:"cid"`
	Option    string    `json:"option"`
	IsDefault int32     `json:"is_default,omitempty"`
	X         int64     `json:"x,omitempty"`
	Y         int64     `json:"y,omitempty"`
	TextAlign int32     `json:"text_align,omitempty"`
	Condition string    `json:"-"`
	Skin      *api.Skin `json:"skin,omitempty"`
}

func (v *Choice) ForNode(k int, in *api.GraphEdge, skinInfo *api.Skin) {
	v.NodeID = in.ToNode
	v.CID = in.ToNodeCid
	v.Option = FromOpt(k, in.Title)
	v.IsDefault = in.IsDefault
	v.X = in.PosX
	v.Y = in.PosY
	v.TextAlign = in.TextAlign
	v.Condition = in.Condition
	v.Skin = skinInfo
}

func (v *Choice) ForEdge(k int, in *api.GraphEdge, skinInfo *api.Skin) {
	v.EdgeID = in.Id
	v.ForNode(k, in, skinInfo)
}

func (v *Choice) ClientCompatibility() {
	v.NodeID = v.EdgeID
}

type InfocNode struct {
	FromNID int64
	ToNID   int64
	AID     int64
	MID     int64
	NodeExtended
	Build    int64
	Channel  string
	MobiApp  string
	Platform string
}

type InfocMark struct {
	AID          int64
	MID          int64
	GraphVersion int64
	Mark         int64
	LogTime      int64
}

type NodeExtended struct {
	Type   int32 `json:"type"`
	Delay  int64 `json:"delay"`
	Screen int   `json:"screen"`
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

// EdgeAttr def.
type EdgeAttr struct {
	ID        int64
	FromNID   int64
	ToNID     int64
	Attribute string
}

// HvarReq 对edge和node透明
type HvarReq struct {
	CurrentID, CurrentCursorID int64 // 读隐藏变量时，所以改名叫currentID
	Choices, CursorChoices     string
}

// FromRecord 对edge和node透明
func (v *HvarReq) fromRecordCore(lastRec *api.GameRecords, portal int32, firstID, pickID, currentCursorID int64, inChoices, inCursorChoices string) {
	if lastRec == nil || portal == 2 { // 用户无存档时，查初始隐藏变量存档
		v.CurrentID = firstID // edge图传0，node图传根节点id
		v.CurrentCursorID = 0
		v.Choices = fmt.Sprintf("%d", firstID)
		v.CursorChoices = fmt.Sprintf("%d", 0)
		return
	}
	if portal == 0 { // 正常选择需要查上一个节点的隐藏变量存档
		v.Choices = lastRec.Choices
		v.CursorChoices = lastRec.CursorChoice
		v.CurrentCursorID = lastRec.CurrentCursor
	} else if portal == 1 { // 回溯需要查当前节点的隐藏变量存档
		v.CurrentID = pickID
		v.CurrentCursorID = currentCursorID
		v.Choices = inChoices
		v.CursorChoices = inCursorChoices
	}
}

type EdgeFromCache struct {
	IsEnd  bool    `json:"is_end"`
	ToEIDs []int64 `json:"to_eids"`
}

// EdgeAttrsCache def.
type EdgeAttrsCache struct {
	EdgeAttrs []*EdgeAttr `json:"attrs"`
	HasAttrs  bool        `json:"has_attrs"`
}

type RecordReq struct {
	GraphWithAID map[int64]int64 // key=graph_id, value=aid
	MID          int64
	Buvid        string
}

// PickGraphIDs picks the graph ids from the request
func (v *RecordReq) PickGraphIDs() (graphIDs []int64) {
	for gid := range v.GraphWithAID {
		graphIDs = append(graphIDs, gid)
	}
	return
}

// Evaluation
// 925/10=92.5 92.5+0.5=93 floor(93)=93 93/10=9.3
// 924/10=92.4 92.4+0.5=92.9 floor(92.9)=92 92/10=9.2
func Evaluation(val int64) string {
	return fmt.Sprintf(_evaluationFmt, math.Floor(float64(val)/10+0.5)/10)
}

func SteinsPage(in *arcApi.Page) (out *api.Page) {
	out = new(api.Page)
	out.Cid = in.Cid
	out.Page = in.Page
	out.From = in.From
	out.Part = in.Part
	out.Duration = in.Duration
	out.Vid = in.Vid
	out.Desc = in.Desc
	out.WebLink = in.WebLink
	out.Dimension = api.Dimension{
		Width:  in.Dimension.Width,
		Height: in.Dimension.Height,
		Rotate: in.Dimension.Rotate,
	}
	return
}

func GetFirstEdge(graph *api.GraphInfo) (edge *api.GraphEdge) {
	edge = &api.GraphEdge{
		Id:        RootEdge,
		GraphId:   graph.Id,
		ToNode:    graph.FirstNid,
		ToNodeCid: graph.FirstCid,
	}
	return
}

func (v *Edge) BuildEdge(choices []*Choice, toNodeInfo *api.GraphNode, qteStyle int) {
	v.Type = toNodeInfo.Otype
	v.ShowTime = toNodeInfo.ShowTime
	v.QteStyle = qteStyle
	v.Dimension = Dimension{ // 视频云rotate已处理
		Width:  toNodeInfo.Width,
		Height: toNodeInfo.Height,
		Rotate: 0,
		Sar:    toNodeInfo.Sar,
	}
	v.Choices = choices
	v.Version = 1
}

// ErrInfo def.
type ErrInfo struct {
	ErrType int64  `json:"err_type"`
	ErrId   string `json:"err_id"`
}

// BuildErrInfo def.
func (v *ErrInfo) BuildErrInfo(errType int64, errId string) {
	v.ErrType = errType
	v.ErrId = errId
	//nolint:gosimple
	return

}
