package model

import (
	"net/http"

	bcmdl "git.bilibili.co/bapis/bapis-go/infra/service/broadcast"
)

// Coin add type.
const (
	CoinAddArcType  = 1
	CoinAddArtType  = 2
	CoinArcBusiness = "archive"
	CoinArtBusiness = "article"
	GameArchive     = 1
	GameCustom      = 2
)

var (
	RankTypeAll    = 1
	RankTypeOrigin = 2
	RankTypeRookie = 3
	// RankType rank type params
	RankType = map[int]string{
		RankTypeAll:    "all",
		RankTypeOrigin: "origin",
		RankTypeRookie: "rookie",
	}
	// DayType day params
	DayType = map[int]int{
		1:  1,
		3:  3,
		7:  7,
		30: 30,
	}
	RegionDayType = map[int]int{
		3: 3,
		7: 7,
	}
	RegionDayAll = map[int]int{
		1: 1,
		3: 3,
		7: 7,
	}
	RegionDayOne = map[int]int{
		1: 1,
		3: 3,
	}
	RankOriginalType = []int{0, 1}
	// ArcType arc params type all:0 and recent:1
	ArcType = map[int]int{
		0: 0,
		1: 1,
	}
	// IndexDayType rank index day type
	IndexDayType = []int{
		1,
		3,
		7,
	}
	// OriType original or not
	OriType = []string{
		0: "",
		1: "_origin",
	}
	// AllType all or origin type
	AllType = []string{
		0: "all",
		1: "origin",
	}
	// TagIDs feedback tag ids
	TagIDs = []int64{
		300, //播放卡顿
		301, //进度条君无法调戏
		302, //画音不同步
		303, //弹幕无法加载/弹幕延迟
		304, //出现浮窗广告
		305, //无限小电视
		306, //黑屏
		307, //其他
		354, //校园网无法访问
		553, //跳过片头片尾时间有误
	}
	// LimitTypeIDMap view limit type id 32 完结动画  33 连载动画
	LimitTypeIDMap = map[int32]struct{}{32: {}, 33: {}}
	// RecSpecTypeName recommend data special type name
	RecSpecTypeName = map[int32]string{
		28: "原创",
		30: "V家",
		31: "翻唱",
		59: "演奏",
	}
	// LikeType thumbup like type
	LikeType = map[int8]string{
		1: "like",
		2: "like_cancel",
		3: "dislike",
		4: "dislike_cancel",
	}
	// NewListRid new list need more rids
	NewListRid = map[int32]int32{
		177: 37,
		23:  147,
		11:  185,
	}
	// DefaultServer  broadcst servers default value.
	DefaultServer = &bcmdl.ServerListReply{
		Domain:    "broadcast.chat.bilibili.com",
		TcpPort:   7821,
		WsPort:    7822,
		WssPort:   7823,
		Heartbeat: 30,
		Nodes:     []string{"broadcast.chat.bilibili.com"},
		Backoff: &bcmdl.Backoff{
			MaxDelay:  300,
			BaseDelay: 3,
			Factor:    1.8,
			Jitter:    0.3,
		},
		HeartbeatMax: 3,
	}
)

// IndexSet index setting.
type IndexSet struct {
	Sort string `json:"sort"`
	Len  int64  `json:"len"`
}

// IndexM index message.
type IndexM struct {
	Mid      int64
	Cookies  []*http.Cookie
	Settings string
}

// CheckFeedTag check if tagID in TagIDs
func CheckFeedTag(tagID int64) bool {
	check := false
	for _, id := range TagIDs {
		if tagID == id {
			check = true
			break
		}
	}
	return check
}

type GamePromote struct {
	Num     int64      `json:"num"`
	List    []GameInfo `json:"list"`
	Name    string     `json:"name"`
	Subname string     `json:"subname"`
}

type GameInfo struct {
	Tp    int64  `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Img   string `json:"img"`
	Aid   *int64 `json:"aid"`
	Bvid  string `json:"bvid"`
}

type LikeCheckEvent struct {
	Mid        int64  `json:"mid"`
	Buvid      string `json:"buvid"`
	IP         string `json:"ip"`
	Platform   string `json:"platform"`
	Ctime      string `json:"ctime"`
	Action     string `json:"action"`
	Api        string `json:"api"`
	Origin     string `json:"origin"`
	Referer    string `json:"referer"`
	UserAgent  string `json:"user_agent"`
	Build      string `json:"build"`
	Avid       int64  `json:"avid"`
	UpMid      int64  `json:"up_mid"`
	Pubtime    string `json:"pubtime"`
	LikeSource string `json:"like_source"`
}

type ArcCheckEvent struct {
	Mid         int64  `json:"mid"`
	Buvid       string `json:"buvid"`
	IP          string `json:"ip"`
	Platform    string `json:"platform"`
	Ctime       string `json:"ctime"`
	Action      string `json:"action"`
	Api         string `json:"api"`
	Origin      string `json:"origin"`
	Referer     string `json:"referer"`
	UserAgent   string `json:"user_agent"`
	Build       string `json:"build"`
	Avid        int64  `json:"avid"`
	UpMid       int64  `json:"up_mid"`
	Pubtime     string `json:"pubtime"`
	LikeSource  string `json:"like_source,omitempty"`
	ItemType    string `json:"item_type"`
	ShareSource string `json:"share_source,omitempty"`
	CoinNum     int64  `json:"coin_num,omitempty"`
	Title       string `json:"title"`
	PlayNum     int32  `json:"play_num"`
	Token       string `json:"token,omitempty"`
	EabX        int8   `json:"eab_x"`
	Ramval      int64  `json:"ramval"`
	Gaia        int    `json:"ga"`
}

type UserActInfoc struct {
	Buvid    string
	Build    string
	Client   string
	Ip       string
	Uid      int64
	Aid      int64
	Mid      int64
	Sid      string
	Refer    string
	Url      string
	From     string
	ItemID   string
	ItemType string
	Action   string
	ActionID string
	Ua       string
	Ts       string
	Extra    string
	IsRisk   string
}

type RcmdInfoc struct {
	API         string
	IP          string
	Mid         int64
	Buvid       string
	Ptype       int
	Time        int64
	FreshType   int
	IsRec       int
	Trackid     string
	ReturnCode  int
	UserFeature string
	Showlist    string
	IsFeed      int64
	FreshIdx    int64
	FreshIdx1h  int64
	FeedVersion string
	YNum        int64
}
