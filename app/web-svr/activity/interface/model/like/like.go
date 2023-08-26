package like

import (
	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	xtime "go-common/library/time"
	garcmdl "go-gateway/app/app-svr/archive/service/api"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
)

const (
	// 处于线上状态的数据
	ActRelationSubjectStatusNormal = 0
	// 动作 删除缓存 && 更新缓存 && 失效
	Update = 1
	Delete = 2
	Unable = 3
	// 更新缓存通用方法动作id
	ActRelationFlushItemInfo2Cache = 1
	ActSubjectFlushItemInfo2Cache  = 2
)

// Like struct
type Like struct {
	*Item
	Archive *garcmdl.Arc `json:"archive,omitempty"`
}

// Item like item struct.
type Item struct {
	ID       int64      `json:"id"`
	Wid      int64      `json:"wid"`
	Ctime    xtime.Time `json:"act_ctime"`
	Sid      int64      `json:"sid"`
	Type     int64      `json:"type"`
	Mid      int64      `json:"mid"`
	State    int64      `json:"state"`
	StickTop int64      `json:"stick_top"`
	Mtime    xtime.Time `json:"mtime"`
	Referer  string     `json:"referer"`
}

// GroupItem .
type GroupItem struct {
	ID      int64  `json:"id"`
	Sid     int64  `json:"sid"`
	State   int    `json:"state"`
	Type    int    `json:"type"`
	Mid     int64  `json:"mid"`
	Wid     int64  `json:"wid"`
	Ctime   string `json:"ctime"`
	Likes   int    `json:"likes"`
	Liked   int    `json:"liked"`
	Message string `json:"message"`
	Device  string `json:"device"`
	Image   string `json:"image"`
	Plat    string `json:"plat"`
	Reply   string `json:"reply"`
	Link    string `json:"link"`
}

// List .
type List struct {
	*Item
	Object   interface{}  `json:"object"`
	Like     int64        `json:"like"`
	Liked    int64        `json:"liked"`
	Likes    int64        `json:"likes"`
	HasLikes int32        `json:"has_likes"`
	Click    int64        `json:"click"`
	Coin     int64        `json:"coin"`
	Share    int64        `json:"share"`
	Reply    int64        `json:"reply"`
	Dm       int64        `json:"dm"`
	Fav      int64        `json:"fav"`
	List     []*SimpleArc `json:"list"`
}

// ItemObj .
type ItemObj struct {
	*Item
	Score    int64 `json:"score"`
	HasLiked int64 `json:"has_liked"`
}

// ActReply .
type ActReply struct {
	Lid   int64 `json:"lid"`
	ActID int64 `json:"act_id"`
	Score int64 `json:"score"`
}

// Slider .
type Slider struct {
	List []*List `json:"list"`
}

// ListInfo .
type ListInfo struct {
	List []*List `json:"list"`
	*Page
	ShowVote bool `json:"-"`
}

// LikeActList
type LikeActList struct {
	ShowVote bool    `json:"show_vote"`
	LikeList []*List `json:"like_list"`
	List     []*List `json:"list"`
}

// LidLikeRes .
type LidLikeRes struct {
	Score int64
	Lid   int64
}

// Extend like_extend .
type Extend struct {
	ID    int64      `json:"id"`
	Lid   int64      `json:"lid"`
	Like  int64      `json:"like"`
	Ctime xtime.Time `json:"ctime"`
	Mtime xtime.Time `json:"mtime"`
}

// Tag .
type Tag struct {
	ID   int64  `json:"tag_id,omitempty"`
	Name string `json:"tag_name,omitempty"`
}

// ArgTag .
type ArgTag struct {
	Archive *garcmdl.Arc `json:"archive,omitempty"`
	Tags    []string     `json:"tags,omitempty"`
	Bvid    string       `json:"bvid,omitempty"`
}

// ArticleTag .
type ArticleTag struct {
	Meta    *artmdl.Meta `json:"meta"`
	HasLike int32        `json:"has_like"`
}

// ContentTag .
type ContentTag struct {
	Cont *LikeContent `json:"cont"`
	Act  *accapi.Info `json:"act"`
}

// EsLikesReply .
type EsLikesReply struct {
	Lids  []int64
	Count int64
}

type ExtraTimesDetail struct {
	ID    int64      `json:"id"`
	Sid   int64      `json:"sid"`
	Mid   int64      `json:"mid"`
	Num   int        `json:"num"`
	Ctime xtime.Time `json:"ctime"`
}

type ExtendTokenDetail struct {
	ID    int64      `json:"id"`
	Sid   int64      `json:"sid"`
	Mid   int64      `json:"mid"`
	Token string     `json:"token"`
	Max   int64      `json:"max"`
	Ctime xtime.Time `json:"ctime"`
}

type LikeListItem struct {
	ActTime xtime.Time `json:"act_time"`
	Action  int64      `json:"action"`
	*LikeContent
}

type Invite struct {
	HasInvited int64 `json:"has_invited"`
	Total      int64 `json:"total"`
}

type ArcLists struct {
	Default []*ArcBvInfo `json:"default,omitempty"`
	Special []*ArcBvInfo `json:"special,omitempty"`
}

type ArcInfo struct {
	*garcmdl.Arc
	Bvid     string `json:"bvid"`
	IsFollow int64  `json:"is_follow"`
}

type ArcBvInfo struct {
	*garcmdl.Arc
	Bvid string `json:"bvid"`
}

type GaiaReport struct {
	Scene    string `json:"scene"`
	TraceId  string `json:"trace_id"`
	EventTs  int64  `json:"event_ts"`
	EventCtx string `json:"event_ctx"`
}

type ActRelationInfo struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	NativeIDs      string `json:"native_ids"`
	H5IDs          string `json:"h5_ids"`
	WebIDs         string `json:"web_ids"`
	LotteryIDs     string `json:"lottery_ids"`
	ReserveIDs     string `json:"reserve_ids"`
	VideoSourceIDs string `json:"video_source_ids"`
	FollowIDs      string `json:"follow_ids"`
	SeasonIDs      string `json:"season_ids"`
	ReserveConfig  string `json:"reserve_config"`
	FollowConfig   string `json:"follow_config"`
	SeasonConfig   string `json:"season_config"`
	FavoriteInfo   string `json:"favorite_info"`
	FavoriteConfig string `json:"favorite_config"`
	MallIDs        string `json:"mall_ids"`
	MallConfig     string `json:"mall_config"`
	TopicIDs       string `json:"topic_ids"`
	TopicConfig    string `json:"topic_config"`
}

type ActRelationInfoReply struct {
	Name           string                       `json:"name,omitempty"`
	NativeIDs      []int64                      `json:"native_ids,omitempty"`
	H5IDs          []int64                      `json:"h5_ids,omitempty"`
	WebIDs         []int64                      `json:"web_ids,omitempty"`
	LotteryIDs     []string                     `json:"lottery_ids,omitempty"`
	ReserveIDs     []int64                      `json:"reserve_ids,omitempty"`
	VideoSourceIDs []int64                      `json:"video_source_ids,omitempty"`
	NativeID       int64                        `json:"native_id,omitempty"`
	ReserveID      int64                        `json:"reserve_id,omitempty"`
	ReserveItem    *ActRelationInfoReserveItem  `json:"reserve_item,omitempty"`
	ReserveItems   *ActRelationInfoReserveItems `json:"reserve_items,omitempty"`
}

type ActRelationInfoReserveItem struct {
	Sid       int64  `json:"sid"`
	Name      string `json:"name"`
	Total     int64  `json:"total"`
	State     int64  `json:"state"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	ActStatus int64  `json:"act_status"`
}

type ActRelationInfoReserveItems struct {
	State       int64                         `json:"state"`
	Total       int64                         `json:"total"`
	ReserveList []*ActRelationInfoReserveItem `json:"reserve_list"`
}

type YgVote struct {
	YellowVote int64 `json:"yellow_vote"`
	GreenVote  int64 `json:"green_vote"`
}

type YellowGreenPeriod struct {
	YellowSid         int64
	GreenSid          int64
	GreenYingYuanSid  int64
	YellowYingYuanSid int64
	LotterySid        string
	Cid               int64
}

type UpVoteEventCtx struct {
	Action       string `json:"action"`
	Mid          int64  `json:"mid"`
	ActivityUid  string `json:"activity_uid"`
	UpMid        int64  `json:"up_mid"`
	Content      string `json:"content"`
	UpCategoryID int64  `json:"up_category_id"`
	CategoryName string `json:"up_category_name"`
	Buvid        string `json:"buvid"`
	Ip           string `json:"ip"`
	Platform     string `json:"platform"`
	Ctime        string `json:"ctime"`
	Api          string `json:"api"`
	Origin       string `json:"origin"`
	UserAgent    string `json:"user_agent"`
	Build        string `json:"build"`
	MobiApp      string `json:"mobi_app"`
	Referer      string `json:"referer"`
	Level        int32  `json:"level"`
}

type VideoVoteEventCtx struct {
	Action      string `json:"action"`
	Mid         int64  `json:"mid"`
	ActivityUid string `json:"activity_uid"`
	TargetID    int64  `json:"target_id"`
	ID          int64  `json:"id"`
	Score       int64  `json:"score"`
	Buvid       string `json:"buvid"`
	Ip          string `json:"ip"`
	Platform    string `json:"platform"`
	Ctime       string `json:"ctime"`
	Api         string `json:"api"`
	Origin      string `json:"origin"`
	UserAgent   string `json:"user_agent"`
	Build       string `json:"build"`
	MobiApp     string `json:"mobi_app"`
	Referer     string `json:"referer"`
}

type ParamFilter struct {
	Area    string   `form:"area" default:"activity"`
	Keys    []string `form:"keys,split" validate:"min=1,max=5,dive,min=1"`
	Message string   `form:"message" validate:"required"`
	Level   int32    `form:"level" validate:"min=1" default:"20"`
}

type ActSensitive struct {
	IsSensitive bool `json:"is_sensitive"`
}

type ReservesTime []ReserveTime

type ReserveTime struct {
	Sid   int64
	Mtime int64
}

func (rt ReservesTime) Len() int {
	return len(rt)
}

func (rt ReservesTime) Less(i, j int) bool {
	return rt[i].Mtime > rt[j].Mtime
}

func (rt ReservesTime) Swap(i, j int) {
	rt[i], rt[j] = rt[j], rt[i]
}

type ParamViewData struct {
	Sid  int64 `form:"sid" validate:"min=1"`
	Type int64 `form:"type"`
	Pn   int   `form:"pn" default:"1" validate:"min=1"`
	Ps   int   `form:"ps" validate:"min=1,max=50" default:"10"`
}

// WebData act web data.
type WebData struct {
	ID      int64                  `json:"id"`
	Vid     int64                  `json:"vid"`
	Data    string                 `json:"data"`
	OutData map[string]interface{} `json:"out_data"`
	Bvid    string                 `json:"bvid"`
	Aid     int64                  `json:"aid"`
	Arc     *garcmdl.Arc           `json:"arc"`
	Name    string                 `json:"name"`
	Stime   string                 `json:"stime"`
	Etime   string                 `json:"etime"`
	Ctime   string                 `json:"ctime"`
	Mtime   string                 `json:"mtime"`
}

// WebDataRes .
type WebDataRes struct {
	ID      int64                  `json:"id"`
	Vid     int64                  `json:"vid"`
	Data    map[string]interface{} `json:"data"`
	Archive *garcmdl.Arc           `json:"archive,omitempty"`
	Name    string                 `json:"name"`
	Stime   string                 `json:"stime"`
	Etime   string                 `json:"etime"`
	Ctime   string                 `json:"ctime"`
	Mtime   string                 `json:"mtime"`
}

// WebDataArc .
type WebDataArc struct {
	ID      int64                  `json:"id"`
	Vid     int64                  `json:"vid"`
	Data    string                 `json:"data"`
	OutData map[string]interface{} `json:"out_data"`
	Bvid    string                 `json:"bvid"`
	Arc     *garcmdl.Arc           `json:"arc"`
	Name    string                 `json:"name"`
	Stime   string                 `json:"stime"`
	Etime   string                 `json:"etime"`
	Ctime   string                 `json:"ctime"`
	Mtime   string                 `json:"mtime"`
}

type UpActReserveRelationInfo struct {
	ID                int64      `json:"id"`
	Sid               int64      `json:"sid"`
	Mid               int64      `json:"mid"`
	Oid               string     `json:"oid"`
	Type              int64      `json:"type"`
	State             int64      `json:"state"`
	Ctime             xtime.Time `json:"ctime"`
	Mtime             xtime.Time `json:"mtime"`
	LivePlanStartTime xtime.Time `json:"live_plan_start_time"`
	Audit             int64      `json:"audit"`
	AuditChannel      int64      `json:"audit_channel"`
	DynamicID         string     `json:"dynamic_id"`
	DynamicAudit      int64      `json:"dynamic_audit"`
	LotteryType       int64      `json:"lottery_type"`
	LotteryID         string     `json:"lottery_id"`
	LotteryAudit      int64      `json:"lottery_audit"`
}

type CacheData struct {
	Type int64 `form:"type" validate:"min=1"`
	Sid  int64 `form:"sid"`
	Mid  int64 `form:"mid"`
}

type UpActReserveRelationBind struct {
	ID    int64  `json:"id"`
	Sid   int64  `json:"sid"`
	Oid   string `json:"oid"`
	OType int64  `json:"o_type"`
	Rid   string `json:"rid"`
	RType int64  `json:"r_type"`
}

type TagConvertItem struct {
	TagID   int64  `json:"tag_id"`
	TagName string `json:"tag_name"`
}
