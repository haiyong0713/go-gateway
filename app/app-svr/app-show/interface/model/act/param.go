package act

import (
	actapi "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/app-show/interface/model/dynamic"
)

// ParamActIndex .
type ParamActIndex struct {
	PageID       int64  `form:"page_id" validate:"min=1"`
	Device       string `form:"device"`
	VideoMeta    string `form:"video_meta"`
	MobiApp      string `form:"mobi_app"`
	Platform     string `form:"platform"`
	Build        int64  `form:"build"`
	Offset       int64  `form:"offset"  default:"0" validate:"min=0"`
	Ps           int64  `form:"ps" default:"41" validate:"min=1"`
	ActivityFrom string `form:"activity_from"`
	DynamicID    int64  `form:"dynamic_id"`
	Buvid        string `form:"-"`
	TfIsp        string `form:"-"`
	ShareOrigin  string `form:"share_origin"`
	TabID        int64  `form:"tab_id"`
	TabModuleID  int64  `form:"tab_module_id"`
	HttpsUrlReq  int    `form:"https_url_req"`
	FromSpmid    string `form:"from_spmid"` //动态列表和老视频卡动态组件
	CurrentTab   string `form:"current_tab"`
	UserAgent    string `form:"-"`
	Memory       int64  `form:"memory"`
}

// ParamInlineTab .
type ParamInlineTab struct {
	PageID        int64  `form:"page_id" validate:"min=1"`
	Device        string `form:"device"`
	VideoMeta     string `form:"video_meta"`
	MobiApp       string `form:"mobi_app"`
	Platform      string `form:"platform"`
	Build         int64  `form:"build"`
	Buvid         string `form:"-"`
	Offset        int64  `form:"offset"  default:"0" validate:"min=0"`
	Ps            int64  `form:"ps" default:"41" validate:"min=1"`
	TfIsp         string `form:"-"`
	HttpsUrlReq   int    `form:"https_url_req"`
	FromSpmid     string `form:"from_spmid"`      //动态列表和老视频卡动态组件
	PrimaryPageID int64  `form:"primary_page_id"` //一级页pageid 615及以上版本才会有
	UserAgent     string `form:"-"`
	Memory        int64  `form:"memory"`
}

// ParamMenuTab .
type ParamMenuTab struct {
	PageID      int64  `form:"page_id" validate:"min=1"`
	Device      string `form:"device"`
	VideoMeta   string `form:"video_meta"`
	MobiApp     string `form:"mobi_app"`
	Platform    string `form:"platform"`
	Build       int64  `form:"build"`
	Buvid       string `form:"-"`
	Offset      int64  `form:"offset"  default:"0" validate:"min=0"`
	Ps          int64  `form:"ps" default:"50" validate:"min=1"`
	TfIsp       string `form:"-"`
	HttpsUrlReq int    `form:"https_url_req"`
	FromSpmid   string `form:"from_spmid"` //动态列表和老视频卡动态组件
	UserAgent   string `form:"-"`
	Memory      int64  `form:"memory"`
	TabFrom     string `form:"tab_from"` //fix置顶发起活动
}

// ParamInlineTab .
type ParamFormatModule struct {
	PageID       int64  `json:"page_id"`
	Device       string `json:"device"`
	VideoMeta    string `json:"video_meta"` //动态列表接口使用
	MobiApp      string `json:"mobi_app"`
	Platform     string `json:"platform"`
	Build        int64  `json:"build"`
	Buvid        string `json:"buvid"`
	DynamicID    int64  `json:"dynamic_id"`
	ActivityFrom string `json:"activity_from"`
	TfIsp        string `json:"tf_isp"`
	HttpsUrlReq  int    `json:"https_url_req"`
	FromSpmid    string `json:"from_spmid"`  //动态列表和老视频卡动态组件
	CurrentTab   string `json:"current_tab"` //index 定位参数
	Mid          int64  `json:"mid"`
	//每周必看分享数据
	ShareOrigin string `json:"share_origin"`
	TabID       int64  `json:"tab_id"`
	TabModuleID int64  `json:"tab_module_id"`
	FromPage    string `json:"from_page"`
	//每周必看分享数据
	UserAgent string `json:"user_agent"`
	Memory    int64  `json:"memory"`
	TabFrom   string `json:"tab_from"`
}

// ParamActLike .
type ParamActLike struct {
	Sid   int64 `form:"sid" validate:"min=1"`
	Lid   int64 `form:"lid" validate:"min=1"`
	Score int64 `form:"score" default:"1" validate:"min=1,max=5"`
}

// ParamActFollow .
type ParamActFollow struct {
	Goto      string `form:"goto" validate:"min=1"`
	FID       int64  `form:"fid" validate:"min=1"`
	GroupID   int64  `form:"group_id"`
	ItemID    int64  `form:"item_id"`
	Type      int    `form:"type" validate:"min=0,max=1"`
	FromSpmid string `form:"from_spmid"`
	MobiApp   string `form:"mobi_app"`
	Platform  string `form:"platform"`
	Buvid     string `form:"-"`
	UserAgent string `form:"-"`
	Build     int64  `form:"build"`
}

type FollowRly struct {
	Num        int64 `json:"num,omitempty"`
	CanVoteNum int64 `json:"can_vote_num,omitempty"`
}

// ParamActDetail .
type ParamActDetail struct {
	ModuleID int64  `form:"module_id" validate:"min=1"`
	Build    int64  `form:"build"`
	MobiApp  string `form:"mobi_app"`
}

// ParamBaseDetail .
type ParamBaseDetail struct {
	PageID int64 `form:"page_id" validate:"min=1"`
}

// ParamLike .
type ParamLike struct {
	Sid           int64  `form:"sid"`
	SortType      int32  `form:"sort_type" default:"1"`
	Pn            int32  `form:"pn" default:"1" validate:"min=1"`
	Ps            int32  `form:"ps" default:"10" validate:"min=1,max=50"`
	VideoMeta     string `form:"video_meta"`
	MobiApp       string `form:"mobi_app"`
	Device        string `form:"device"`
	Platform      string `form:"platform"`
	Build         int64  `form:"build"`
	Offset        int64  `form:"offset" default:"-1"`
	DyOffset      string `form:"dy_offset"`
	ModuleID      int64  `form:"module_id"`
	TopicID       int64  `form:"topic_id"`
	Goto          string `form:"goto" default:"video"`
	AvSort        int64  `form:"av_sort"`
	Buvid         string `form:"-"`
	TfIsp         string `form:"-"`
	RemoteFrom    string `form:"remote_from"`
	Attr          int64  `form:"attr"`
	DyType        string `form:"dy_type"`
	FromSpmid     string `form:"from_spmid"`
	ConfModuleID  int64  `form:"conf_module_id"`
	SourceID      string `form:"source_id"`
	PrimaryPageID int64  `form:"primary_page_id"` //一级页pageid 618及以上版本才会有
}

// ParamSupernatant .
type ParamSupernatant struct {
	Ps           int64  `form:"ps" default:"10" validate:"min=1,max=30"`
	Offset       int64  `form:"offset"`
	ConfModuleID int64  `form:"conf_module_id" validate:"min=1"`
	LastIndex    int64  `form:"last_index" default:"-1"`
	PageID       int64  `form:"page_id" validate:"min=1"`
	MobiApp      string `form:"mobi_app"`
	Device       string `form:"device"`
}

// AvidReq .
type AvidReq struct {
	ModuleID   int64
	AvSort     int64
	Offset     int64
	Ps         int64
	VideoMeta  string
	MobiApp    string
	Buvid      string
	Platform   string
	Build      int64
	RemoteFrom string
	Device     string
	TfIsp      string
	FromSpmid  string
}

// NewAvidReq .
type NewAvidReq struct {
	ModuleID int64
	Offset   int64
	Ps       int64
	MobiApp  string
	Buvid    string
	Platform string
	Build    int64
	Device   string
	TfIsp    string
	NetType  int32
	TfType   int32
}

// VideoReply 视频卡 .
type VideoReply struct {
	Total   int64
	HasMore int32
	Offset  int64
	DyReply []*dynamic.DyCard
}

// VideoReply 视频卡 .
type NewVideoReply struct {
	Total    int64
	HasMore  int32
	Offset   int64
	Item     []*Item
	DyOffset string
}

// VideoActReq .
type VideoActReq struct {
	Sid        int64
	SortType   int32
	AvSort     int64
	Offset     int64
	Ps         int64
	VideoMeta  string
	MobiApp    string
	Platform   string
	Build      int64
	Buvid      string
	RemoteFrom string
	Device     string
	TfIsp      string
	FromSpmid  string
}

// NewVideoActReq .
type NewVideoActReq struct {
	Sid      int64
	SortType int32
	AvSort   int64
	Offset   int64
	Ps       int64
	MobiApp  string
	Platform string
	Build    int64
	Buvid    string
	Device   string
	NetType  int32
	TfType   int32
}

// NewDynReq .
type NewDynReq struct {
	TopicID  int64
	Types    string
	PageSize int64
	DyOffset string
	Mid      int64
	MobiApp  string
	Buvid    string
	Build    int64
	Platform string
	NetType  int32
	TfType   int32
}

// ResourceActReq .
type ResourceActReq struct {
	Sid      int64
	SortType int32
	AvSort   int64
	Offset   int64
	Ps       int64
	Mid      int64
	MobiApp  string
	Device   string
}

type ResourceOriginReq struct {
	SourceID    string
	RdbType     int64
	Offset      int64
	Ps          int64
	Mid         int64
	MobiApp     string
	Device      string
	Platform    string
	Build       int64
	FID         int64
	SortType    int64
	MustseeType int32
	Buvid       string
}

type ParamActTab struct {
	TabID       int64  `form:"tab_id" validate:"min=1"`
	TabModuleID int64  `form:"tab_module_id" validate:"min=1"`
	PageID      int64  `form:"page_id"`
	MobiApp     string `form:"mobi_app"`
	Build       int64  `form:"build"`
}

// ParamReceive.
type ParamReceive struct {
	Goto      string `form:"goto" validate:"min=1"`
	FID       int64  `form:"fid" validate:"min=1"`
	State     int    `form:"state"`
	FromSpmid string `form:"from_spmid"`
}

type ReserveProgressParam struct {
	Sid       int64                              `json:"sid"`
	Type      int64                              `json:"type"`
	DataType  int64                              `json:"data_type"`
	Dimension actapi.GetReserveProgressDimension `json:"dimension"`
	RuleID    int64                              `json:"rule_id"`
}

type ParamPlat struct {
	//进度条-任务统计-活动名
	Activity string `json:"activity,omitempty"`
	//进度条-任务统计-counter名
	Counter string `json:"counter,omitempty"`
	//进度条-任务统计：统计周期 单日:daily 累计：total
	StatPc string `json:"statPc,omitempty"`
}

type ParamLive struct {
	IDs    []int64
	IsLive int64
}
