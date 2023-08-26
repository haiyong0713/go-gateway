package dynamic

import (
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"

	"go-gateway/app/web-svr/native-page/interface/api"
)

// ParamActIndex .
type ParamActIndex struct {
	PageID  int64  `form:"page_id" validate:"min=1"`
	PType   int32  `form:"p_type" default:"0" validate:"min=0"`
	Offset  int64  `form:"offset" default:"0" validate:"min=0"`
	Ps      int64  `form:"ps" default:"50" validate:"min=1"`
	MobiApp string `form:"-"`
	Buvid   string `form:"-"`
}

// ParamMenuTab .
type ParamMenuTab struct {
	PageID  int64  `form:"page_id" validate:"min=1"`
	MobiApp string `form:"-"`
	Buvid   string `form:"-"`
}

// ParamActInline .
type ParamActInline struct {
	PageID        int64  `form:"page_id" validate:"min=1"`
	Offset        int64  `form:"offset" default:"0" validate:"min=0"`
	Ps            int64  `form:"ps" default:"50" validate:"min=1"`
	PrimaryPageID int64  `form:"primary_page_id"`
	MobiApp       string `form:"-"`
	Buvid         string `form:"-"`
}

// ParamNatPages .
type ParamNatPages struct {
	PageIDs []int64 `form:"page_ids,split" validate:"min=1,max=50,dive,min=1"`
}

// ParamActDynamic .
type ParamActDynamic struct {
	TopicID   int64  `form:"topic_id" validate:"min=1"`
	Ps        int64  `form:"ps" validate:"min=1"`
	Sortby    int32  `form:"sortby" validate:"min=0"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
	Types     string `form:"types"`
	Ukey      string `form:"ukey"`
	PageID    int64  `form:"page_id"`
}

// ParamResourceDyn .
type ParamResourceDyn struct {
	TopicID   int64  `form:"topic_id" validate:"min=1"`
	Ps        int64  `form:"ps" validate:"min=1"`
	Sortby    int32  `form:"sortby" validate:"min=0"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
	Types     string `form:"types"`
	Ukey      string `form:"ukey"`
	PageID    int64  `form:"page_id"`
}

// ParamVideoDyn .
type ParamVideoDyn struct {
	TopicID int64 `form:"topic_id" validate:"min=1"`
	Ps      int64 `form:"ps" default:"10" validate:"min=1"`
	Sortby  int32 `form:"sortby" validate:"min=0"`
}

// ParamNatModule .
type ParamNatModule struct {
	PageID int64  `form:"page_id" validate:"min=1"`
	Ukey   string `form:"ukey" validate:"min=1"`
}

type ParamAid struct {
	IDs string `form:"ids" validate:"min=1"`
}

type ParamSeasonIDs struct {
	IDs       string `form:"ids" validate:"min=1"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
	CardStyle int32  `form:"card_style"`
}

type ParamSeasonSource struct {
	FID       int64 `form:"fid" validate:"min=1"`
	Attribute int64 `form:"attribute" default:"0" validate:"min=0"`
	CardStyle int32 `form:"card_style"`
	Offset    int64 `form:"offset"`
	Ps        int64 `form:"ps" default:"15" validate:"min=1"`
}

type ParamResourceAid struct {
	IDs       string `form:"ids" validate:"min=1"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
}

type ParamLiveDyn struct {
	RoomIDs []int64 `form:"room_ids,split" validate:"min=1,max=50,dive,min=1"`
	IsHttps bool    `form:"is_https"`
}

type LiveDynRly struct {
	Cards map[int64]*livegrpc.LiveCardInfo `json:"cards"`
}

type ResourceAid struct {
	ID   int64  `json:"id" validate:"min=0"`
	Bvid string `json:"bvid"`
	Type int    `json:"type" validate:"min=0"`
	Fid  int64  `json:"fid"`
}

type ParamMinePages struct {
	Ps     int64 `form:"ps" default:"15" validate:"min=0,max=50"`
	Offset int64 `form:"offset" default:"0" validate:"min=0"`
}

type ParamTsPage struct {
	PID int64 `form:"pid" validate:"min=0"`
}

type ParamMinePageSave struct {
	PID          int64  `form:"pid"` //page id
	Title        string `form:"title"`
	BgColor      string `form:"bg_color"`
	VideoDisplay string `form:"video_display"`
	Modules      string `form:"modules"`
	UserSpace    string `form:"user_space"`
	ShareImage   string `form:"share_image"`
	Partitions   string `form:"partitions"`
	Dynamic      string `form:"dynamic"`
}

type ParamModule struct {
	Meta      string                        `json:"meta"`
	Remark    string                        `json:"remark"`
	Category  int64                         `json:"category"`
	Width     int64                         `json:"width"`
	Length    int64                         `json:"length"`
	Rank      int64                         `json:"-"`
	Ukey      string                        `json:"ukey"`
	TopicID   int64                         `json:"topic_id"`
	DySort    int32                         `json:"dy_sort"`
	Attribute int64                         `json:"attribute"`
	Resources []*api.NativeTsModuleResource `json:"resources"`
}

type ParamResourceRole struct {
	RoleID    int32 `form:"role_id" validate:"required"`
	SeasonID  int32 `form:"season_id" validate:"required"`
	Ps        int   `form:"ps" default:"6" validate:"min=0"`
	Attribute int64 `form:"attribute" default:"0" validate:"min=0"`
}

// ParamTimelineSource .
type ParamTimelineSource struct {
	FID    int64 `form:"f_id" validate:"min=1"`
	Offset int32 `form:"offset"`
	Ps     int32 `form:"ps" default:"15" validate:"min=0,max=50"`
}

type NativeTsModuleExt struct {
	api.NativeTsModule
	Resources []*api.NativeTsModuleResource `json:"resources"`
}

func (tmp *NativeTsModuleExt) SetResourceNum(v *ParamModule, arcNumLeft *int) {
	resLen := len(v.Resources)
	if resLen >= *arcNumLeft {
		resLen = *arcNumLeft
	}
	tmp.Resources = v.Resources[:resLen]
	*arcNumLeft -= resLen
	tmp.Num = int64(resLen)
}

type ActArchiveListReq struct {
	Title  string `form:"title" validate:"required"`
	Ps     int64  `form:"ps"`
	Offset string `form:"offset"`
	Sort   int64  `form:"sort"`
}

type ParamResourceOrigin struct {
	Ps        int    `form:"ps" default:"10" validate:"min=1"`
	SourceID  string `form:"source_id" validate:"required"`
	RDBType   int    `form:"rdb_type" validate:"min=0"`
	Offset    int    `form:"offset" default:"0" validate:"min=0"`
	Attribute int64  `form:"attribute" default:"0" validate:"min=0"`
	SortType  int64  `form:"sort_type"`
}

type ProgressReq struct {
	PageID int64  `form:"page_id" validate:"required"`
	WebKey string `form:"web_key" validate:"required"`
	From   string `form:"from" validate:"required"`
}

type ProgressRly struct {
	Num int64 `json:"num"`
}

type ReserveProgressParam struct {
	Sid       int64                               `json:"sid"`
	GroupID   int64                               `json:"group_id"`
	Type      int64                               `json:"type"`
	DataType  int64                               `json:"data_type"`
	Dimension actGRPC.GetReserveProgressDimension `json:"dimension"`
	RuleID    int64                               `json:"rule_id"`
}

type ParamPlat struct {
	//进度条-任务统计-活动名
	Activity string `json:"activity,omitempty"`
	//进度条-任务统计-counter名
	Counter string `json:"counter,omitempty"`
	//进度条-任务统计：统计周期 单日:daily 累计：total
	StatPc string `json:"statPc,omitempty"`
}

type TsSpaceRly struct {
	*api.NativeUserSpace
	NativePage *struct {
		Title string `json:"title"`
	} `json:"native_page"`
}

type TsSpaceSaveReq struct {
	From         string `form:"-"`
	Title        string `form:"title" validate:"required"`
	DisplaySpace int64  `form:"display_space"`
	PageID       int64  `form:"page_id" validate:"min=1"`
}
type ParamEditorNewOrigin struct {
	Ps           int32 `form:"ps" default:"10" validate:"min=1"`
	ConfModuleID int64 `form:"conf_module_id" validate:"min=1"`
	Offset       int32 `form:"offset"`
}

type ParamEditorOrigin struct {
	Ps           int32 `form:"ps" default:"10" validate:"min=1"`
	ConfModuleID int64 `form:"conf_module_id" validate:"min=1"`
	Offset       int32 `form:"offset"`
}

type EdViewedArcsReq struct {
	Sid     int64  `form:"sid" validate:"min=1"`
	Counter string `form:"counter" validate:"required"`
}

type EdViewedArcsRly struct {
	Aids []int64 `json:"aids"`
}
