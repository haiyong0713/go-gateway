package model

import (
	"strconv"
	"strings"

	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/playurl/service/api"
	steinsGrpc "go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	NodeOtypeNormal       int32 = 1
	NodeOtypePoint        int32 = 2
	NodeOtypeRandom       int32 = 3
	MsgActionPass               = "pass"
	ShowTimeDefault             = 0
	TextAlignUp                 = 1
	TextAlignRight              = 2
	TextAlignDown               = 3
	TextAlignLeft               = 4
	RegionalVarTypeNormal       = 1
	RegionalVarTypeRandom       = 2
	EdgeConditionTypeGe         = "ge"
	EdgeConditionTypeGt         = "gt"
	EdgeConditionTypeLe         = "le"
	EdgeConditionTypeLt         = "lt"
	EdgeConditionTypeEq         = "eq"
	EdgeAttrActionAdd           = "add"
	EdgeAttrActionSub           = "sub"
	EdgeAttrActionAssign        = "assign"
)

var (
	NodeOtypes = map[int32]struct{}{
		NodeOtypeNormal: {},
		NodeOtypePoint:  {},
		NodeOtypeRandom: {},
	}
	EdgeConditionTypes = map[string]struct{}{
		EdgeConditionTypeGe: {},
		EdgeConditionTypeGt: {},
		EdgeConditionTypeLe: {},
		EdgeConditionTypeLt: {},
		EdgeConditionTypeEq: {},
	}
	EdgeAttrActions = map[string]struct{}{
		EdgeAttrActionAdd:    {},
		EdgeAttrActionSub:    {},
		EdgeAttrActionAssign: {},
	}
	EdgeTextAligns = map[int]struct{}{
		TextAlignUp:    {},
		TextAlignRight: {},
		TextAlignDown:  {},
		TextAlignLeft:  {},
	}
)

// SaveGraphParam .
type SaveGraphParam struct {
	Graph *GraphParam `json:"graph"`
}

func (in *SaveGraphParam) FromAudit(graph *GraphAuditDB, edges []*steinsGrpc.GraphEdge, nodes []*steinsGrpc.GraphNode) {
	in.Graph = &GraphParam{
		ID:          graph.Id,
		State:       1,
		Aid:         graph.Aid,
		GlobalVars:  graph.GlobalVars,
		RegionalStr: graph.RegionalVars,
		Script:      graph.Script,
		SkinID:      graph.SkinId,
	}
	var nodeParams = make(map[int64]*NodeParam)
	for _, v := range nodes {
		nodeParams[v.Id] = &NodeParam{
			ID:       strconv.FormatInt(v.Id, 10),
			Name:     v.Name,
			Cid:      v.Cid,
			IsStart:  v.IsStart,
			Otype:    v.Otype,
			ShowTime: v.ShowTime,
			SkinID:   v.SkinId,
		}
	}
	for _, v := range edges {
		if node, ok := nodeParams[v.FromNode]; ok {
			node.Edges = append(node.Edges, &EdgeParam{
				Title:        v.Title,
				ToNodeID:     strconv.FormatInt(v.ToNode, 10),
				Weight:       v.Weight,
				TextAlign:    int(v.TextAlign),
				PosY:         v.PosY,
				PosX:         v.PosX,
				IsDefault:    int(v.IsDefault),
				Script:       v.Script,
				AttributeStr: v.Attribute,
				ConditionStr: v.Condition,
			})
		}
	}
	for _, v := range nodeParams {
		in.Graph.Nodes = append(in.Graph.Nodes, v)
	}
}

// Graph .
type Graph struct {
	ID     int64     `json:"id"`
	Aid    int64     `json:"aid"`
	Script string    `json:"script"`
	Ctime  time.Time `json:"ctime"`
}

// GraphCheck .
type GraphCheck struct {
	HasCid bool `json:"has_cid"`
}

// GraphShow .
type GraphShow struct {
	*Graph
	DisabledCids []int64 `json:"disabled_cids"`
}

// GraphParam .
type GraphParam struct {
	ID           int64          `json:"id"`
	State        int            `json:"state"`
	Script       string         `json:"script"`
	Aid          int64          `json:"aid"`
	RegionalVars []*RegionalVal `json:"regional_vars"`
	RegionalStr  string         `json:"-"`
	GlobalVars   string         `json:"global_vars"`
	SkinID       int64          `json:"skin_id"`
	Nodes        []*NodeParam   `json:"nodes"`
}

func (v *GraphParam) VarsNames() (res string) {
	for _, item := range v.RegionalVars {
		if item.IsShow == 1 {
			res += item.Name + "、"
		}
	}
	res = strings.TrimRight(res, `、`)
	return
}

// RegionalVal .
type RegionalVal struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	InitMin int    `json:"init_min"`
	InitMax int    `json:"init_max"`
	Type    int    `json:"type"`
	IsShow  int    `json:"is_show"`
}

// NodeParam .
type NodeParam struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Cid      int64        `json:"cid"`
	IsStart  int32        `json:"is_start"`
	Otype    int32        `json:"otype"`
	ShowTime int64        `json:"show_time"`
	SkinID   int64        `json:"skin_id"`
	Edges    []*EdgeParam `json:"edges"`
}

// EdgeParam .
type EdgeParam struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	ToNodeID     string           `json:"to_node_id"`
	Weight       int64            `json:"weight"`
	TextAlign    int              `json:"text_align"`
	PosX         int64            `json:"pos_x"`
	PosY         int64            `json:"pos_y"`
	IsDefault    int              `json:"is_default"`
	Script       string           `json:"script"`
	Attribute    []*EdgeAttribute `json:"attribute"`
	AttributeStr string           `json:"-"`
	Condition    []*EdgeCondition `json:"condition"`
	ConditionStr string           `json:"-"`
}

// EdgeAttribute .
type EdgeAttribute struct {
	VarID      string `json:"var_id"`
	Action     string `json:"action"`
	Value      int    `json:"value"`
	ActionType int    `json:"action_type"`
}

// EdgeCondition .
type EdgeCondition struct {
	VarID      string `json:"var_id"`
	Condition  string `json:"condition"`
	Value      int    `json:"value"`
	ActionType int    `json:"action_type"`
}

// PlayurlParam .
type PlayurlParam struct {
	Aid     int64  `form:"avid" validate:"min=1"`
	Cid     int64  `form:"cid" validate:"min=1"`
	Qn      int64  `form:"qn"`
	Fnver   int32  `form:"fnver"`
	Fnval   int32  `form:"fnval"`
	Session string `form:"session"`
	Buvid   string `form:"buvid"`
	Build   int32  `form:"build"`
	Type    string `form:"type"`
}

// VideoInfoParam .
type VideoInfoParam struct {
	Aid   int64  `form:"avid" validate:"min=1"`
	Cid   int64  `form:"cid" validate:"min=1"`
	Buvid string `form:"buvid"`
}

// GraphPubMsg .
type GraphPubMsg struct {
	Aid    int64  `json:"aid"`
	Action string `json:"action"`
}

// VideoInfo .
type VideoInfo struct {
	Playurl   *PlayurlRes `json:"playurl"`
	Dimension *Dimension  `json:"dimension"`
}

// PlayurlRes playurl res.
type PlayurlRes struct {
	From              string   `json:"from"`
	Result            string   `json:"result"`
	Message           string   `json:"message"`
	Quality           uint32   `json:"quality"`
	Format            string   `json:"format"`
	Timelength        uint64   `json:"timelength"`
	AcceptFormat      string   `json:"accept_format"`
	AcceptDescription []string `json:"accept_description"`
	AcceptQuality     []uint32 `json:"accept_quality"`
	VideoCodeCid      uint32   `json:"video_codecid"`
	SeekParam         string   `json:"seek_param"`
	SeekType          string   `json:"seek_type"`
	Abtid             int32    `json:"abtid,omitempty"`
	Durl              []*durl  `json:"durl,omitempty"`
	Dash              *dash    `json:"dash,omitempty"`
}

type durl struct {
	Order     uint32   `json:"order"`
	Length    uint64   `json:"length"`
	Size      uint64   `json:"size"`
	Ahead     string   `json:"ahead"`
	Vhead     string   `json:"vhead"`
	URL       string   `json:"url"`
	BackupURL []string `json:"backup_url"`
}

type dash struct {
	Duration       uint32      `json:"duration"`
	MinBufferTime  float32     `json:"minBufferTime"`
	MinBufferTime2 float32     `json:"min_buffer_time"`
	Video          []*dashItem `json:"video"`
	Audio          []*dashItem `json:"audio"`
}

type dashItem struct {
	ID            uint32        `json:"id"`
	BaseURL       string        `json:"baseUrl"`
	BaseURL2      string        `json:"base_url"`
	BackupURL     []string      `json:"backupUrl"`
	BackupURL2    []string      `json:"backup_url"`
	Bandwidth     uint32        `json:"bandwidth"`
	MimeType      string        `json:"mimeType"`
	MimeType2     string        `json:"mime_type"`
	Codecs        string        `json:"codecs"`
	Width         uint32        `json:"width"`
	Height        uint32        `json:"height"`
	FrameRate     string        `json:"frameRate"`
	FrameRate2    string        `json:"frame_rate"`
	Sar           string        `json:"sar"`
	StartWithSAP  uint32        `json:"startWithSap"`
	StartWithSAP2 uint32        `json:"start_with_sap"`
	SegmentBase   *segmentBase  `json:"SegmentBase"`
	SegmentBase2  *segmentBase2 `json:"segment_base"`
	Codecid       uint32        `json:"codecid"`
}

type segmentBase struct {
	Initialization string `json:"Initialization"`
	IndexRange     string `json:"indexRange"`
}

type segmentBase2 struct {
	Initialization string `json:"initialization"`
	IndexRange     string `json:"index_range"`
}

// FromPlayurl from playurl data.
func (p *PlayurlRes) FromPlayurl(reply *api.SteinsPreviewReply) {
	p.From = reply.Playurl.From
	p.Result = reply.Playurl.Result
	//p.Message = reply
	p.Quality = reply.Playurl.Quality
	p.Format = reply.Playurl.Format
	p.Timelength = reply.Playurl.Timelength
	p.AcceptFormat = reply.Playurl.AcceptFormat
	p.AcceptDescription = reply.Playurl.AcceptDescription
	p.AcceptQuality = reply.Playurl.AcceptQuality
	p.VideoCodeCid = reply.Playurl.VideoCodecid
	p.SeekParam = reply.Playurl.SeekParam
	p.SeekType = reply.Playurl.SeekType
	p.Abtid = reply.Playurl.Abtid
	for _, v := range reply.Playurl.Durl {
		if v == nil {
			continue
		}
		durlItem := new(durl)
		durlItem.FromDurl(v)
		p.Durl = append(p.Durl, durlItem)
	}
	if reply.Playurl.Dash != nil {
		pDash := new(dash)
		pDash.FromDash(reply.Playurl.Dash)
		p.Dash = pDash
	}
}

// FromDurl from durl data.
func (d *durl) FromDurl(item *api.Durl) {
	d.Order = item.Order
	d.Length = item.Length
	d.Size = item.Size_
	d.Ahead = item.Ahead
	d.Vhead = item.Vhead
	d.URL = item.Url
	d.BackupURL = item.BackupUrl
}

// FromDash from dash data.
func (d *dash) FromDash(item *api.Dash) {
	d.Duration = item.Duration
	d.MinBufferTime = item.MinBufferTime
	d.MinBufferTime2 = item.MinBufferTime
	for _, v := range item.Video {
		if v == nil {
			continue
		}
		videoItem := new(dashItem)
		videoItem.FromDashItem(v)
		d.Video = append(d.Video, videoItem)
	}
	for _, v := range item.Audio {
		if v == nil {
			continue
		}
		audioItem := new(dashItem)
		audioItem.FromDashItem(v)
		d.Audio = append(d.Audio, audioItem)
	}
}

// FromDashItem from dash item.
func (d *dashItem) FromDashItem(item *api.DashItem) {
	d.ID = item.Id
	d.BaseURL = item.BaseUrl
	d.BaseURL2 = item.BaseUrl
	d.BackupURL = item.BackupUrl
	d.BackupURL2 = item.BackupUrl
	d.Bandwidth = item.Bandwidth
	d.MimeType = item.MimeType
	d.MimeType2 = item.MimeType
	d.Codecs = item.Codecs
	d.Width = item.Width
	d.Height = item.Height
	d.FrameRate = item.FrameRate
	d.FrameRate2 = item.FrameRate
	d.Sar = item.Sar
	d.StartWithSAP = item.StartWithSAP
	d.StartWithSAP2 = item.StartWithSAP
	if item.SegmentBase != nil {
		d.SegmentBase = &segmentBase{
			Initialization: item.SegmentBase.GetInitialization(),
			IndexRange:     item.SegmentBase.GetIndexRange(),
		}
		d.SegmentBase2 = &segmentBase2{
			Initialization: item.SegmentBase.GetInitialization(),
			IndexRange:     item.SegmentBase.GetIndexRange(),
		}
	}
	d.Codecid = item.Codecid
}

func HasHvar(graph *steinsGrpc.GraphInfo) (res bool) {
	variables, err := GetVarsMap(graph)
	if err != nil {
		log.Error("GraphID %d GetVarsMap Err %v", graph.Id, err)
		return
	}
	return len(variables) > 0

}
