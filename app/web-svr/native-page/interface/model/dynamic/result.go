package dynamic

import (
	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	arccli "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/native-page/interface/api"
)

type InlineReply struct {
	PageID int64   `json:"page_id,omitempty"`
	Title  string  `json:"title,omitempty"`
	Items  []*Item `json:"cards,omitempty"`
}

// MenuReply .
type MenuReply struct {
	PageID int64   `json:"page_id,omitempty"`
	Title  string  `json:"title,omitempty"`
	Items  []*Item `json:"cards,omitempty"`
	//AttrBit    *AttrBit    `json:"attr_bit,omitempty"`
	BgColor    string      `json:"bg_color,omitempty"`
	Attentions *Attentions `json:"attentions,omitempty"`
}

// IndexReply .
type IndexReply struct {
	PageID       int64          `json:"page_id,omitempty"`
	Uid          int64          `json:"uid,omitempty"`
	BgColor      string         `json:"bg_color,omitempty"`
	FromType     int32          `json:"from_type,omitempty"`
	Title        string         `json:"title,omitempty"`
	ForeignID    int64          `json:"foreign_id,omitempty"`
	ForeignType  int64          `json:"foreign_type,omitempty"`
	ShareTitle   string         `json:"share_title,omitempty"`
	ShareImage   string         `json:"share_image,omitempty"`
	ShareURL     string         `json:"share_url,omitempty"`
	PcURL        string         `json:"pc_url,omitempty"`
	ShareCaption string         `json:"share_caption,omitempty"`
	PageURL      string         `json:"page_url,omitempty"`
	SkipURL      string         `json:"skip_url,omitempty"`
	Spmid        string         `json:"spmid"`
	Items        []*Item        `json:"cards,omitempty"`
	Attentions   *Attentions    `json:"attentions,omitempty"`
	Modules      []*ParamModule `json:"modules,omitempty"`
	Ver          string         `json:"ver,omitempty"`
	Bases        *Bases         `json:"bases,omitempty"`
}

type Bases struct {
	Head        *Item `json:"head,omitempty"`
	HoverButton *Item `json:"hover_button,omitempty"`
}

type ModulesReply struct {
	Card map[int64]*Item `json:"card"`
}

// Attentions .
type Attentions struct {
	Uids []int64 `json:"uids,omitempty"`
}

// NatPagesReply .
type NatReply struct {
	Items []*Item `json:"items,omitempty"`
}

// DynReply .
type DynReply struct {
	Display bool    `json:"display"`
	Items   []*Item `json:"items,omitempty"`
}

type NatModuleReply struct {
	MoreUrl   string     `json:"more_url"`
	MoreParam *MoreParam `json:"more_param"`
	Stime     int64      `json:"stime"`
	Etime     int64      `json:"etime"`
}

type IDsReply struct {
	List     []*Item `json:"list"`
	Offset   int64   `json:"offset,omitempty"`
	DyOffset string  `json:"dy_offset,omitempty"`
	HasMore  int32   `json:"has_more,omitempty"`
}

type TimelineSourceReply struct {
	List    []*Item `json:"list,omitempty"`
	Offset  int32   `json:"offset,omitempty"`
	HasMore int32   `json:"has_more,omitempty"`
}

type MoreParam struct {
	Offset   int64  `json:"offset"`
	DyOffset string `json:"dy_offset"`
	PageID   int64  `json:"page_id"`
}

// ModuleReply .
type ModuleReply struct {
	*api.NativePage
	Module *api.NativeModule `json:"module"`
}

// ModuleIDsReply .
type ModuleIDsReply struct {
	IDs   []int64 `json:"ids"`
	Count int64   `json:"count"`
}

// ModuleIDsReply .
type ModuleRankReply struct {
	IDs   []*RankInfo `json:"ids"`
	Count int64       `json:"count"`
}

type RankInfo struct {
	ID    int64 `json:"id"`
	Score int64 `json:"score"`
}

// SortRly .
type SortRly struct {
	ID   int64 `json:"id"`
	Rank int64 `json:"rank"`
}
type PageSaveRly struct {
	TopicID    int64  `json:"topic_id,omitempty"`
	PID        int64  `json:"pid,omitempty"`
	Title      string `json:"title,omitempty"`
	State      int64  `json:"state"`
	AuditState int64  `json:"audit_state"`
	ExpiryArcs bool   `json:"expiry_arcs"`
}

type TsPageRly struct {
	Title        string                          `json:"title,omitempty"`
	Pid          int64                           `json:"pid,omitempty"`
	ForeignID    int64                           `json:"foreign_id,omitempty"`
	BgColor      string                          `json:"bg_color,omitempty"`
	State        int64                           `json:"state,omitempty"`
	AuditState   int64                           `json:"audit_state,omitempty"`
	Uid          int64                           `json:"-"`
	TsID         int64                           `json:"-"`
	Attribute    int64                           `json:"-"`
	Modules      []*NativeTsModuleExt            `json:"modules,omitempty"`
	IsAdmin      bool                            `json:"is_admin,omitempty"`
	VideoDisplay string                          `json:"video_display"`
	AuditTime    int64                           `json:"audit_time"`
	AuditType    string                          `json:"audit_type"`
	ShareImage   string                          `json:"share_image"`
	UpShareImage string                          `json:"-"`
	Template     string                          `json:"template,omitempty"`
	PageSources  map[int64]*api.NativePageSource `json:"-"`
	Partitions   string                          `json:"partitions,omitempty"`
	Dynamic      string                          `json:"dynamic,omitempty"`
}

func (tp *TsPageRly) Trans2TsPageResourceRly() *TsPageResourceRly {
	rly := &TsPageResourceRly{}
	rly.TsPageRly = *tp
	rly.Modules = make([]*ModuleExt, 0, len(tp.Modules))
	for _, m := range tp.Modules {
		me := &ModuleExt{}
		me.NativeTsModuleExt = *m
		me.Resources = make([]*ResourceDetail, 0, len(m.Resources))
		for _, r := range m.Resources {
			rt := &ResourceDetail{}
			rt.NativeTsModuleResource = *r
			me.Resources = append(me.Resources, rt)
		}
		rly.Modules = append(rly.Modules, me)
	}
	return rly
}

type TsPageResourceRly struct {
	TsPageRly
	Modules            []*ModuleExt         `json:"modules"`
	SpaceButton        string               `json:"space_button,omitempty"`
	UserSpace          *api.NativeUserSpace `json:"user_space,omitempty"`
	IsPartitionChanged bool                 `json:"is_partition_changed,omitempty"`
}

type ModuleExt struct {
	NativeTsModuleExt
	Resources []*ResourceDetail `json:"resources"`
}

type ResourceDetail struct {
	api.NativeTsModuleResource
	Title string       `json:"title,omitempty"`
	Arc   *ResourceArc `json:"arc,omitempty"`
}

type ResourceArc struct {
	Title   string `json:"title,omitempty"`
	Pic     string `json:"pic,omitempty"`
	Danmuku int32  `json:"danmuku,omitempty"`
	View    int32  `json:"view,omitempty"`
}

type MinePagesRly struct {
	SpaceButton string      `json:"space_button"`
	List        []*MinePage `json:"list"`
	Offset      int64       `json:"offset"`
	HasMore     int32       `json:"has_more"`
}

type MinePage struct {
	Title      string `json:"title,omitempty"`
	Pid        int64  `json:"pid,omitempty"`
	ForeignID  int64  `json:"foreign_id,omitempty"`
	State      int64  `json:"state,omitempty"` //page state
	AuditState int64  `json:"audit_state"`
}

type PageMsg struct {
	ID         int64  `json:"id"`
	RelatedUid int64  `json:"related_uid"`
	State      int32  `json:"state"`
	FromType   int32  `json:"from_type"`
	Mtime      string `json:"mtime"`
	Type       int64  `json:"type"`
	Title      string `json:"title"`
	ActType    int64  `json:"act_type"`
	OffReason  string `json:"off_reason"`
}

type TsWhiteRly struct {
	Status int `json:"status"`
}

type TsSendReq struct {
	Title        string
	Pid          int64
	BgColor      string
	State        int64
	Uid          int64
	TsID         int64
	Url          string
	Modules      []*NativeTsModuleExt
	AuditTime    int64
	ShareImage   string
	Partitions   string
	Template     string
	AuditContent AuditContent
	Dynamic      string
	IsFirstAudit bool
}

type SendModule struct {
	Meta         string
	Remark       string
	BgColor      string
	Url          string
	ShareImage   string
	Partitions   string
	AuditContent int64
	Dynamic      string
}

type UpActPagesReply struct {
	Offset  int64         `json:"offset,omitempty"`
	HasMore int32         `json:"has_more,omitempty"`
	List    []*UpActPages `json:"list,omitempty"`
}

type UpActPages struct {
	PID       int64  `json:"pid,omitempty"`
	Title     string `json:"title,omitempty"`
	ForeignID int64  `json:"foreign_id,omitempty"`
}

type ResourceRoleReply struct {
	List []*Item `json:"list"`
}

type MyArchiveListRly struct {
	Total   int64          `json:"total"`
	HasMore bool           `json:"has_more"`
	List    []*ArchiveItem `json:"list"`
}

type ActArchiveListRly struct {
	HasMore bool           `json:"has_more"`
	Offset  string         `json:"offset"`
	List    []*ArchiveItem `json:"list"`
}

type ArchiveItem struct {
	Aid      int64  `json:"aid"`
	Title    string `json:"title"`
	Pubdate  string `json:"pubdate"`
	Duration string `json:"duration"`
	Pic      string `json:"pic"`
	View     string `json:"view"`
	Danmaku  string `json:"danmaku"`
}

type TsSettingRly struct {
	SpaceButton string `json:"space_button"`
}

type ReserveRly struct {
	ChangeType  int64                             `json:"change_type"`  // 1:类型A 2:类型C 4:类型CD
	DisplayType int64                             `json:"display_type"` // 1:类型A 2:类型C 3:类型D
	Item        *actGRPC.UpActReserveRelationInfo `json:"item"`         //预约信息
	Arc         *arccli.Arc                       //稿件信息
	Live        *LiveInfos                        //直播信息
	Account     *accgrpc.Card                     //账号信息
}

type LiveInfos struct {
	// 房间id 长号
	RoomId int64
	// 主播id
	Uid int64
	// 房间标题
	Title string
	// key为 entry_from，value为跳转房间地址，一个entry_from对应一个地址 此地址应该完全透传
	// 如果entry_from传了NONE，那key也会是NONE
	// https://info.bilibili.co/pages/viewpage.action?pageId=164638733
	JumpUrl map[string]string
	// key:直播场次ID即liveId; value:场次数据
	SessionInfoPerLive *roomgategrpc.SessionInfoPerLive
}

type ResourceExt struct {
	ImgUrl string `json:"img_url"`
	Length int64  `json:"length"`
	Width  int64  `json:"width"`
}

type PartitionV2Rly struct {
	Partitions []*Partition `json:"partitions"`
}

type Partition struct {
	ID       int64        `json:"id,omitempty"`
	Name     string       `json:"name,omitempty"`
	Children []*Partition `json:"children,omitempty"`
}
