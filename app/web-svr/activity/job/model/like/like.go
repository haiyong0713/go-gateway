package like

import (
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model"
)

const (
	LotteryArcType       = 5
	LotteryCustomizeType = 8
	LotteryVip           = 6
	LotteryAct           = 14
	LotteryActPoint      = 15
	CustomizeVip         = "vip"
	BuyVip               = "buyvip"
	LotteryOgvType       = 9
	LotteryActPointType  = 15
	Ogv                  = "ogv"
	LikeNormal           = 1
	LikeNotNormal        = 0
)

// Like like
type Like struct {
	ID       int64    `json:"id"`
	Wid      int64    `json:"wid"`
	Mid      int64    `json:"mid"`
	State    int      `json:"state"`
	StickTop int      `json:"stick_top"`
	Archive  *api.Arc `json:"archive,omitempty"`
	IsBlack  bool     `json:"is_black"`
}

// IsNormal 是否正常
func (l *Like) IsNormal() bool {
	if l.State == LikeNormal {
		return true
	}
	return false
}

// Item like item struct.
type Item struct {
	ID       int64         `json:"id"`
	Wid      int64         `json:"wid"`
	Ctime    model.StrTime `json:"ctime"`
	Sid      int64         `json:"sid"`
	Type     int           `json:"type"`
	Mid      int64         `json:"mid"`
	State    int           `json:"state"`
	StickTop int           `json:"stick_top"`
	Mtime    model.StrTime `json:"mtime"`
}

// ObjItem .
type ObjItem struct {
	ID    int64 `json:"id"`
	Wid   int64 `json:"wid"`
	Sid   int64 `json:"sid"`
	Type  int   `json:"type"`
	Mid   int64 `json:"mid"`
	State int   `json:"state"`
}

// Content  like_content.
type Content struct {
	ID      int64         `json:"id"`
	Message string        `json:"message"`
	IP      int64         `json:"ip"`
	Plat    int           `json:"plat"`
	Device  int           `json:"device"`
	Ctime   model.StrTime `json:"ctime"`
	Mtime   model.StrTime `json:"mtime"`
	Image   string        `json:"image"`
	Reply   string        `json:"reply"`
	Link    string        `json:"link"`
	ExName  string        `json:"ex_name"`
}

// WebData act web data.
type WebData struct {
	ID    int64         `json:"id"`
	Vid   int64         `json:"vid"`
	Data  string        `json:"data"`
	Name  string        `json:"name"`
	Stime model.StrTime `json:"stime"`
	Etime model.StrTime `json:"etime"`
	Ctime model.StrTime `json:"ctime"`
	Mtime model.StrTime `json:"mtime"`
	State int64         `json:"state"`
}

// Action like_action .
type Action struct {
	ID          int64         `json:"id"`
	Lid         int64         `json:"lid"`
	Mid         int64         `json:"mid"`
	Action      int64         `json:"action"`
	Ctime       model.StrTime `json:"ctime"`
	Mtime       model.StrTime `json:"mtime"`
	Sid         int64         `json:"sid"`
	IP          int64         `json:"ip"`
	ExtraAction int64         `json:"extra_action"`
}

// Archive .
type Archive struct {
	Aid       int64 `json:"aid"`
	MissionID int64 `json:"mission_id"`
	State     int   `json:"state"`
	Mid       int64 `json:"mid"`
	Attribute int32 `json:"attribute"`
}

// LotteryMsg .
type LotteryMsg struct {
	MissionID int64
	Mid       int64
	ObjID     int64
}

// ActPlatHistoryMsg ...
type ActPlatHistoryMsg struct {
	Activity  string `json:"activity"`
	Counter   string `json:"counter"`
	CounterID int64  `json:"counter_id"`
	MID       int64  `json:"mid"`
	TimeStamp int64  `json:"timestamp"`
	Diff      int64  `json:"diff"`
	Total     int64  `json:"total"`
}

// LotteryActPointInfo ...
type LotteryActPointInfo struct {
	SID     int64 `json:"sid"`
	GroupID int64 `json:"group_id"`
}

// ThresholdNotifyMsg ...
type ThresholdNotifyMsg struct {
	Activity  string `json:"activity"`
	Counter   string `json:"counter"`
	CounterID int64  `json:"counter_id"`
	MID       int64  `json:"mid"`
	TimeStamp int64  `json:"timestamp"`
	Diff      int64  `json:"diff"`
	Total     int64  `json:"total"`
}

// ActPlatHistoryVideoMsg ...
type ActPlatHistoryVideoMsg struct {
	Activity  string      `json:"activity"`
	Counter   string      `json:"counter"`
	CounterID int64       `json:"counter_id"`
	MID       int64       `json:"mid"`
	TimeStamp int64       `json:"timestamp"`
	Diff      int64       `json:"diff"`
	Total     int64       `json:"total"`
	Raw       []*VideoRaw `json:"raw"`
}

// VideoRaw ...
type VideoRaw struct {
	New *Video `json:"new"`
}

// Video ...
type Video struct {
	Aid int64 `json:"aid"`
}

// LotteryMsg .
type Msg struct {
	MissionID int64
	Mid       int64
	ObjID     int64
}

// Reply .
type Reply struct {
	Type     string `json:"type"`
	ID       int64  `json:"id"`
	ReplyMid int64  `json:"reply_mid"`
}

// LotteryAward .
type LotteryAward struct {
	Mid     int64 `json:"mid"`
	AwardID int64 `json:"award_id"`
	Lid     int64 `json:"lid"`
	Ctime   int64 `json:"ctime"`
}

// VipLottery .
type VipLottery struct {
	Mid      int64  `json:"mid"`
	Ctime    int64  `json:"ctime"`
	ActToken string `json:"act_token"`
}

// Extend .
type Extend struct {
	ID    int64         `json:"id"`
	Lid   int64         `json:"lid"`
	Like  int64         `json:"like"`
	Ctime model.StrTime `json:"ctime"`
	Mtime model.StrTime `json:"mtime"`
}

// LotteryAddTimesMsg ...
type LotteryAddTimesMsg struct {
	MID        int64  `json:"mid"`
	SID        string `json:"sid"`
	CID        int64  `json:"cid"`
	ActionType int    `json:"action_type"`
	OrderNo    string `json:"order_no"`
}
type Lottery struct {
	ID   int64  `json:"id"`
	Sid  string `json:"sid"`
	Info string `json:"info"`
}

type LotteryDetail struct {
	ID  int64  `json:"id"`
	Sid string `json:"sid"`
}

type OttVipLottery struct {
	ActivityPlatformID int64  `json:"activity_platform_id"`
	OrderNo            string `json:"order_no"`
	Mid                int64  `json:"mid"`
	Ctime              int64  `json:"ctime"`
}

type CustomizeLottery struct {
	Mid      int64  `json:"mid"`
	ActToken string `json:"act_token"`
	Ctime    int64  `json:"ctime"`
	OrderNo  string `json:"order_no"`
	Type     string `json:"type"`
}

type ArticleDay struct {
	YesterdayPeople int64 `json:"yesterday_people"`
	BeforePeople    int64 `json:"before_people"`
}

type ContributionUser struct {
	Mid          int64  `json:"mid"`
	UpArchives   int64  `json:"up_archives"`
	Likes        int32  `json:"likes"`
	Views        int32  `json:"views"`
	LightVideos  uint32 `json:"light_videos"`
	Bcuts        int32  `json:"bcuts"`
	SnUpArchives int64  `json:"sn_up_archives"`
	SnLikes      int32  `json:"sn_likes"`
}

type ActContributions struct {
	ID           int64 `json:"id"`
	Mid          int64 `json:"mid"`
	UpArchives   int64 `json:"up_archives"`
	Likes        int64 `json:"likes"`
	Views        int64 `json:"views"`
	LightVideos  int64 `json:"light_videos"`
	Bcuts        int64 `json:"bcuts"`
	SnUpArchives int64 `json:"sn_up_archives"`
	SnLikes      int64 `json:"sn_likes"`
}

type ContriAward struct {
	ID           int64 `json:"id"`
	AwardType    int64 `json:"award_type"`
	UpArchives   int64 `json:"up_archives"`
	Likes        int64 `json:"likes"`
	Views        int64 `json:"views"`
	LightVideos  int64 `json:"light_videos"`
	Bcuts        int64 `json:"bcuts"`
	SnUpArchives int64 `json:"sn_up_archives"`
	SnLikes      int64 `json:"sn_likes"`
}

type ArcScore struct {
	Aid   int64
	Score int64
}

type ProductRoleDB struct {
	ID           int64  `json:"id"`
	CategoryID   int64  `json:"category_id"`
	CategoryType int64  `json:"category_type"`
	Role         string `json:"role"`
	Product      string `json:"product"`
	Tags         string `json:"tags"`
	TagsType     int64  `json:"tags_type"`
	VoteNum      int64  `json:"vote_num"`
}

type ProductRoleArc struct {
	Aid     int64 `json:"aid"`
	PubDate int64 `json:"pub_date"`
}

type ProductRoleHot struct {
	Aid     int64 `json:"aid"`
	PubDate int64 `json:"pub_date"`
	HotNum  int64 `json:"hot_num"`
}

type SelCategory struct {
	CategoryID   int64
	CategoryName string
}

type YgVote struct {
	YellowVote int64 `json:"yellow_vote"`
	GreenVote  int64 `json:"green_vote"`
}

type YellowGreenPeriod struct {
	YingYuanView      int32
	YingYuanVote      int64
	GreenYingYuanSid  int64
	YellowYingYuanSid int64
}

type GaiaRisk struct {
	SceneName   string           `json:"scene_name"`
	DecisionCtx *RiskDecisionCtx `json:"decision_ctx"`
	Decision    string           `json:"decision"`
}

type RiskDecisionCtx struct {
	Mid        interface{} `json:"mid"`
	ID         interface{} `json:"id"`
	CategoryID interface{} `json:"category_id"`
	Status     interface{} `json:"status"`
	VoteDate   string      `json:"vote_date"`
}

type LIDWithVote struct {
	ID    int64 `json:"id"`
	Wid   int64 `json:"wid"`
	Vote  int64 `json:"vote"`
	Order int64 `json:"order"`
}

type ArticleDayEventCtx struct {
	Action      string `json:"action"`
	Mid         int64  `json:"mid"`
	ActivityUid string `json:"activity_uid"`
	ID          int64  `json:"id"`
	UserAction  int64  `json:"user_action"`
	Sid         int64  `json:"sid"`
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

type LIDWithVotes []*LIDWithVote

func (ps LIDWithVotes) Len() int {
	return len(ps)
}

func (ps LIDWithVotes) Less(i, j int) bool {
	return ps[i].Vote > ps[j].Vote
}

func (ps LIDWithVotes) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
