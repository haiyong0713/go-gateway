package fm_v2

import (
	fmRec "git.bilibili.co/bapis/bapis-go/ott-recommend/automotive-channel"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
)

const (
	// Deprecated: 最近播放卡片已下架
	AudioHistory  = FmType("audio_history")
	AudioFeed     = FmType("audio_feed")
	AudioVertical = FmType("audio_vertical")
	AudioUp       = FmType("audio_up")
	AudioRelate   = FmType("audio_relate")
	AudioSeason   = FmType("audio_season")
	AudioSeasonUp = FmType("audio_season_up")
	// AudioHome 算法侧将首页推荐卡片在一个接口下发，包含了audio_vertical、audio_season、audio_season_up，仅用于首页tab卡片请求
	AudioHome = FmType("audio_home")
	// AudioHomeV2 v2.3版本及以后使用的首页推荐接口，包含了audio_vertical、audio_season，仅用于首页tab卡片请求
	AudioHomeV2 = FmType("audio_home_v2")

	DefaultStyle   = TabStyle(0)
	CircleStyle    = TabStyle(1)
	RectangleStyle = TabStyle(2)
)

type FmType string

type TabStyle int // FM卡片样式

type FmShowParam struct {
	model.DeviceInfo
	Mid           int64  `form:"-"`
	Buvid         string `form:"buvid"`
	BootFmType    FmType `form:"boot_fm_type"`
	BootFmId      int64  `form:"boot_fm_id"`
	Ps            int    `form:"ps"`
	PageNext      string `form:"page_next"`
	ManualRefresh int    `form:"manual_refresh"` // 刷新/分页时，是否手动触发
}

type FmShowResp struct {
	TabItems []*TabItem `json:"tab_items"`
	PageNext *PageInfo  `json:"page_next"` // 下一页分页参数
	HasNext  bool       `json:"has_next"`  // 是否到底
	BootInfo *BootInfo  `json:"boot_info"` // 冷启携带的信息（type和id），用于分页时去重
}

type ShowV2Param struct {
	model.DeviceInfo
	Mid           int64  `form:"-"`
	Buvid         string `form:"buvid"`
	GuestId       int64  `form:"guest_id"`
	PinPs         int    `form:"pin_ps"`         // 金刚位的页面大小，默认10
	RecPs         int    `form:"rec_ps"`         // 精选推荐页面大小，默认10
	RecPageNext   string `form:"rec_page_next"`  // 向下翻页信息json
	ManualRefresh int    `form:"manual_refresh"` //0 非手动刷新（刚进入APP） 1 手动刷新（点击一级tab刷新 / 上拉刷新 / 下拉刷新）
	HasPin        bool   `form:"has_pin"`        // 是否展示金刚位
	PinMore       bool   `form:"pin_more"`       // 金刚位是否展示更多
	Mode          int    `form:"mode"`           // 0默认，开启个性化   1关闭个性化推荐
}

type ShowV2Resp struct {
	PinItems    []*common.Item `json:"pin_items"`
	RecItems    []*common.Item `json:"rec_items"`
	RecPageNext *PageInfo      `json:"rec_page_next"`
	RecHasNext  bool           `json:"rec_has_next"`
	HasPin      bool           `json:"has_pin"`  // 是否展示金刚位
	PinMore     bool           `json:"pin_more"` // 金刚位是否展示更多
}

type PinPageParam struct {
	model.DeviceInfo
	Mid         int64  `form:"-"`
	Buvid       string `form:"buvid"`
	GuestId     int64  `form:"guest_id"`
	PinPs       int    `form:"pin_ps"`
	PinPageNext string `form:"pin_page_next"`
}

type PinPageResp struct {
	PinItems    []*common.Item `json:"pin_items"`     // 金刚位卡片
	TopText     string         `json:"top_text"`      // 金刚位页面的顶部文案
	PinPageNext *PageInfo      `json:"pin_page_next"` // 金刚位下一页
	PinHasNext  bool           `json:"pin_has_next"`  // 金刚位 存在下一页
}

type TabItem struct {
	FmType        FmType   `json:"fm_type"`
	FmId          int64    `json:"fm_id,omitempty"`
	Title         string   `json:"title"`
	SubTitle      string   `json:"sub_title"` // 副标题，垂类/频道会携带
	Cover         string   `json:"cover"`
	Style         TabStyle `json:"style"`
	IsBoot        int      `json:"is_boot,omitempty"`         // 是否为冷启播单 0:否 1:是
	ServerInfo    string   `json:"server_info,omitempty"`     // 推荐算法测透传的数据
	ServerExtra   string   `json:"server_extra"`              // 播单的拓展信息，将作为请求播单列表时的参数
	FirstArcTitle string   `json:"first_arc_title,omitempty"` // 首个稿件的标题
}

type BootInfo struct {
	BootFmType FmType `json:"boot_fm_type,omitempty"`
	BootFmId   int64  `json:"boot_fm_id,omitempty"`
}

type ServerExtra struct {
	FmTitle string `json:"fm_title"`
}

type FmListParam struct {
	model.DeviceInfo
	Mid         int64  `form:"-"`
	Buvid       string `form:"buvid"`
	FmType      FmType `form:"fm_type" validate:"required"`
	FmId        int64  `form:"fm_id"`
	BootOid     int64  `form:"oid"`
	BootCid     int64  `form:"cid"`
	ServerExtra string `form:"server_extra"`
	PageNext    string `form:"page_next"`
	PagePre     string `form:"page_previous"`
	Ps          int    `form:"ps"`
}

type FmListResp struct {
	FmItems      []*common.FmItem `json:"fm_items"`
	PageNext     *PageInfo        `json:"page_next"`
	PagePrevious *PageInfo        `json:"page_previous"`
	HasNext      bool             `json:"has_next"`
	HasPrevious  bool             `json:"has_previous"`
}

type PageInfo struct {
	Pn       int64  `json:"pn,omitempty"`
	Ps       int    `json:"ps,omitempty"`
	Business string `json:"business,omitempty"`
	ViewAt   int64  `json:"view_at,omitempty"`
	Max      int64  `json:"max,omitempty"`
	Oid      int64  `json:"oid,omitempty"`
}

type HandlerResp struct {
	PageResp
	OidParam *common.Params // 供物料提取公共方法使用
	OidList  []int64        // 供返回列表排序使用
}

// OidsWithUpwardReq 合集/播单oid查找请求（支持游标前序翻页）
type OidsWithUpwardReq struct {
	model.DeviceInfo
	FmType      FmType
	FmId        int64
	Cursor      int64 // 游标，从哪个oid开始查
	Upward      bool  // 是否向上查找
	WithCurrent bool  // 返回结果中，是否包含当前游标
	Ps          int
}

type HandleTabItemsReq struct {
	model.DeviceInfo
	Mid     int64
	Buvid   string
	FmType  FmType
	FmId    int64
	PageReq *PageReq
}

type HandleTabItemsResp struct {
	PageResp PageResp
	TabItems []*TabItem
}

type TabItemsAll struct {
	BootTab  *TabItem   // 冷启播单（一个）
	FeedTab  *TabItem   // 为你推荐（一个）
	HomeTab  []*TabItem // 算法排序播单（多个）
	PageResp PageResp
}

// PageReq 请求中的分页数据解析结果
type PageReq struct {
	PageNext      *PageInfo // 下一页
	PagePre       *PageInfo // 上一页
	NextEmpty     bool      // 未携带下一页信息
	PreEmpty      bool      // 未携带上一页信息
	PageSize      int       // 分页大小
	ManualRefresh int       // 是否手动刷新
}

// PageResp 返回的分页数据
type PageResp struct {
	PageNext    *PageInfo // 下一页
	PagePre     *PageInfo // 上一页
	HasNext     bool      // 是否到底
	HasPrevious bool      // 是否到顶
}

type HistoryReportFm struct {
	Source        string `json:"source"` // 1."Series"(合集)  2."Channel"（频道）
	Mid           int64  `json:"mid"`
	Buvid         string `json:"buvid"`
	FmID          int64  `json:"fm_id"`
	FmType        string `json:"fm_type"`    // FM类型 1.合集:"audio_season"或"audio_season_up" 2.频道:"audio_vertical"
	PlayTime      string `json:"play_time"`  // 格式："2006-01-02 15:04:05"
	PlayEvent     int    `json:"play_event"` // 上报请求对应的用户行为 1.自动上报（间隔25s） 2.其他行为（包含切集、完播、暂停）
	Aid           int64  `json:"aid"`
	ArchivesCount int    `json:"archives_count,omitempty"` // 合集稿件数（source为频道则置空）
}

// RecResp 推荐返回的卡片
type RecResp struct {
	PageResp PageResp
	Items    []*fmRec.FMRecItemInfo
}

// FmLikeParam FM点赞
type FmLikeParam struct {
	model.DeviceInfo
	Mid    int64  `form:"-"`
	Buvid  string `form:"buvid"`
	UpMid  int64  `form:"up_mid"`
	Oid    int64  `form:"oid"`
	Action int    `form:"action"`
	Path   string `form:"-"`
	UA     string `form:"-"`
}

// AICardIds 算法返回的物料ID
type AICardIds struct {
	Aids          []int64            // ugc
	FmSerialIds   []int64            // FM合集ID
	FmChannelIds  []int64            // FM频道ID
	FmChannelShow map[int64]int64    // 内容透出卡的透出稿件 key: 频道ID val: 透出稿件ID
	Order         []string           // CardMap 的 key的排序
	CardMap       map[string]*AICard // key: ${item_type}:${item_id}或${oid}  val: 卡片
}

type AICard struct {
	*fmRec.Card
	Index int // 位次，从0开始
}

// ChannelInfoAI 算法侧提供的频道信息
type ChannelInfoAI struct {
	HeatScore    int64   `json:"heat_score"`    // 热度值
	ArchiveCount int64   `json:"archive_count"` // 稿件数量
	Cover        string  `json:"cover"`         // 频道卡片封面（优先级高于配置）
	SubTitle     string  `json:"sub_title"`     // 副标题（优先级高于配置，低于aids里的稿件标题）
	Avids        []int64 `json:"avids"`         // 内容透出稿件列表（工期原因v2.3仅携带一个稿件）
}
