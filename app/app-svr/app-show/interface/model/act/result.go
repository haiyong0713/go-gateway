package act

import (
	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	"go-gateway/app/app-svr/app-show/interface/model/dynamic"
	arccli "go-gateway/app/app-svr/archive/service/api"
)

// LikeListRely .
type LikeListRely struct {
	Attentions *Attentions `json:"attentions,omitempty"`
	Cards      []*Item     `json:"cards,omitempty"`
	Page       *Page       `json:"page,omitempty"`
	Offset     int64       `json:"offset"`
	HasMore    int32       `json:"has_more"`
	DyOffset   string      `json:"dy_offset"`
	Color      *Color      `json:"color,omitempty"`
	AttrBit    *AttrBit    `json:"attr_bit,omitempty"`
}

type SupernatantReply struct {
	Color   *Color   `json:"color,omitempty"`
	AttrBit *AttrBit `json:"attr_bit,omitempty"`
	*Supernatant
}

type Supernatant struct {
	Cards     []*Item `json:"cards,omitempty"`
	HasMore   int32   `json:"has_more"`
	UrlExt    *UrlExt `json:"url_ext,omitempty"`
	LastIndex int64   `json:"last_index"`
	TitleConf *Item   `json:"title_conf,omitempty"` //浮层标题相关配置
}

// Page .
type Page struct {
	Pn    int32 `json:"pn,omitempty"`
	Ps    int32 `json:"ps,omitempty"`
	Total int64 `json:"total,omitempty"`
}

// IndexReply .
type IndexReply struct {
	PageID       int64        `json:"page_id,omitempty"`
	Title        string       `json:"title,omitempty"`
	ForeignID    int64        `json:"foreign_id,omitempty"`
	ForeignType  int64        `json:"foreign_type,omitempty"`
	ShareTitle   string       `json:"share_title,omitempty"`
	ShareImage   string       `json:"share_image,omitempty"`
	ShareURL     string       `json:"share_url,omitempty"`
	ShareCaption string       `json:"share_caption,omitempty"`
	ShareType    int32        `json:"share_type,omitempty"`
	PageURL      string       `json:"page_url,omitempty"`
	DynamicInfo  *DynamicInfo `json:"dynamic_info,omitempty"`
	Items        []*Item      `json:"cards,omitempty"`
	Attentions   *Attentions  `json:"attentions,omitempty"`
	Offset       int64        `json:"offset"`
	HasMore      int32        `json:"has_more"`
	VersionMsg   string       `json:"version_msg,omitempty"`
	Bases        *Bases       `json:"bases,omitempty"`
	Color        *Color       `json:"color,omitempty"`
	AttrBit      *AttrBit     `json:"attr_bit,omitempty"`
	UpSpace      *UpSpace     `json:"up_space,omitempty"`
	IsUpSponsor  bool         `json:"is_up_sponsor,omitempty"`
	FromType     int32        `json:"from_type,omitempty"`
}

type UpSpace struct {
	SpacePageURL     string `json:"space_page_url,omitempty"`
	ExclusivePageURL string `json:"exclusive_page_url,omitempty"`
}

type AttrBit struct {
	NotNight bool `json:"not_night"` //适配夜间模式 true:不需要适配夜间模式 false:需要适配页面模式
}

// InlineReply .
type InlineReply struct {
	PageID     int64       `json:"page_id,omitempty"`
	Title      string      `json:"title,omitempty"`
	Items      []*Item     `json:"cards,omitempty"`
	Offset     int64       `json:"offset"`
	HasMore    int32       `json:"has_more"`
	VersionMsg string      `json:"version_msg,omitempty"`
	Attentions *Attentions `json:"attentions,omitempty"`
}

// MenuReply .
type MenuReply struct {
	PageID     int64       `json:"page_id,omitempty"`
	Title      string      `json:"title,omitempty"`
	Items      []*Item     `json:"cards,omitempty"`
	Offset     int64       `json:"offset"`
	HasMore    int32       `json:"has_more"`
	VersionMsg string      `json:"version_msg,omitempty"`
	AttrBit    *AttrBit    `json:"attr_bit,omitempty"`
	Color      *Color      `json:"color,omitempty"`
	Attentions *Attentions `json:"attentions,omitempty"`
	TabConf    *TabConf    `json:"tab_conf,omitempty"` //首页tab配置
	Bases      *Bases      `json:"bases,omitempty"`
}

type TabConf struct {
	TabTopColor    string `json:"tab_top_color,omitempty"`
	TabMiddleColor string `json:"tab_middle_color,omitempty"`
	TabBottomColor string `json:"tab_bottom_color,omitempty"`
	FontColor      string `json:"font_color,omitempty"`
	BarType        int32  `json:"bar_type,omitempty"`
	BgImage1       string `json:"bg_image_1,omitempty"`
	BgImage2       string `json:"bg_image_2,omitempty"`
}

// BaseReply .
type BaseReply struct {
	PageID int64  `json:"page_id,omitempty"`
	Bases  *Bases `json:"bases,omitempty"`
}

// 参与组件
type Bases struct {
	Participation *Item      `json:"participation"`
	Head          *Item      `json:"head,omitempty"`
	SingleDynamic *SingleDyn `json:"single-dynamic,omitempty"`
	HoverButton   *Item      `json:"hover_button,omitempty"`
	BottomButton  *Item      `json:"bottom_button,omitempty"`
}

type SingleDyn struct {
	Title  string          `json:"title,omitempty"`
	DyCard *dynamic.DyCard `json:"dy_card,omitempty"`
}

// Attentions .
type Attentions struct {
	Uids []int64 `json:"uids,omitempty"`
}

// DynamicInfo .
type DynamicInfo struct {
	ViewCount           *int64 `json:"view_count,omitempty"`
	DiscussCount        *int64 `json:"discuss_count,omitempty"`
	IsFollowed          bool   `json:"is_followed"`
	DisplaySubscribeBtn bool   `json:"display_subscribe_btn,omitempty"`
	DisplayViewNum      bool   `json:"display_view_num,omitempty"`
}

// LikedReply .
type LikedReply struct {
	Score int64  `json:"score,omitempty"`
	Toast string `json:"toast,omitempty"`
}

// DetailReply .
type DetailReply struct {
	PageID      int64      `json:"page_id,omitempty"`
	Title       string     `json:"title,omitempty"`
	Name        string     `json:"name,omitempty"`
	ForeignID   int64      `json:"foreign_id,omitempty"`
	ForeignType int64      `json:"foreign_type,omitempty"`
	Sid         int64      `json:"sid,omitempty"`
	Cards       []*Item    `json:"cards,omitempty"`
	Param       *PartParam `json:"param,omitempty"`
	Setting     *Setting   `json:"setting,omitempty"`
}

// PartParam .
type PartParam struct {
	TopicID int64  `json:"topic_id,omitempty"`
	Goto    string `json:"goto"`
	AvSort  int64  `json:"av_sort,omitempty"`
	Attr    int64  `json:"attr,omitempty"`
	DyType  string `json:"dy_type,omitempty"`
}

// ResourceReply .
type ResourceReply struct {
	List     []*Item `json:"list"`
	Offset   int64   `json:"offset,omitempty"`
	DyOffset string  `json:"dy_offset,omitempty"`
	HasMore  int32   `json:"has_more,omitempty"`
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
