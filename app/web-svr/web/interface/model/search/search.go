package search

import (
	"encoding/json"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	watchedmdl "git.bilibili.co/bapis/bapis-go/live/xroom-gate/common"
)

// WxSearchType .
const (
	WxSearchType         = "wechat"
	SearchDftGotoSearch  = 0
	SearchDftGotoArchive = 1
	SearchDftGotoArticle = 2
	SearchDftGotoBangumi = 3
	SearchDftGotoURL     = 4

	HotTypeArchive = 1
	HotTypeArticle = 2
	HotTypePGC     = 3
	HotTypeURL     = 4
)

// SearchAllCommon search common.
type SearchAllCommon struct {
	Code           int             `json:"code,omitempty"`
	SeID           string          `json:"seid,omitempty"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"pagesize,omitempty"`
	Total          int             `json:"total,omitempty"`
	NumResults     int             `json:"numResults"`
	NumPages       int             `json:"numPages"`
	SuggestKeyword string          `json:"suggest_keyword"`
	RqtType        string          `json:"rqt_type,omitempty"`
	CostTime       json.RawMessage `json:"cost_time,omitempty"`
	ExpList        json.RawMessage `json:"exp_list,omitempty"`
	EggHit         int             `json:"egg_hit"`
	PageInfo       json.RawMessage `json:"pageinfo,omitempty"`
	TopTList       json.RawMessage `json:"top_tlist,omitempty"`
	EggInfo        *struct {
		ID     int64 `json:"id,omitempty"`
		Source int64 `json:"source,omitempty"`
	} `json:"egg_info,omitempty"`
	ShowColumn     int      `json:"show_column"`
	ShowModuleList []string `json:"show_module_list"`
}

// Search all search.
type Search struct {
	*SearchAllCommon
	Result *Result `json:"result,omitempty"`
}

// SearchAll search v2.
type SearchAll struct {
	*SearchAllCommon
	InBlackKey int8          `json:"in_black_key"`
	InWhiteKey int8          `json:"in_white_key"`
	Result     []*ResultInfo `json:"result,omitempty"`
}

// ResultInfo result info.
type ResultInfo struct {
	ResultType string      `json:"result_type"`
	Data       interface{} `json:"data"`
}

// Result search all result.
type Result struct {
	Activity      json.RawMessage    `json:"activity"`
	Article       json.RawMessage    `json:"article"`
	BiliUser      []*SearchUser      `json:"bili_user"`
	Card          []*SearchVideoCard `json:"card"`
	Comic         json.RawMessage    `json:"comic"`
	LiveRoom      json.RawMessage    `json:"live_room"`
	MediaBangumi  []*SearchSeason    `json:"media_bangumi"`
	MediaFt       []*SearchSeason    `json:"media_ft"`
	OperationCard json.RawMessage    `json:"operation_card"`
	Special       json.RawMessage    `json:"special"`
	Star          json.RawMessage    `json:"star"`
	Tag           json.RawMessage    `json:"tag"`
	Topic         json.RawMessage    `json:"topic"`
	Tv            json.RawMessage    `json:"tv"`
	Twitter       json.RawMessage    `json:"twitter"`
	User          []*SearchUser      `json:"user"`
	Video         []*SearchVideo     `json:"video"`
	WebGame       []*SearchGame      `json:"web_game"`
	Tips          []*SearchTip       `json:"tips"`
	Esports       []*SearchEsport    `json:"esports"`
}

type SearchEsport struct {
	ID        int64        `json:"id"`
	Title     string       `json:"title"`
	MatchList []*MatchList `json:"match_list"`
}

type MatchList struct {
	ID          int64 `json:"id"`
	HomeTeamId  int   `json:"home_team_id"`
	GuestTeamId int   `json:"guest_team_id"`
}

type SearchTip struct {
	Linktype string           `json:"linktype"`
	Position int64            `json:"position"`
	Trackid  string           `json:"trackid"`
	Type     string           `json:"type"`
	Value    *SearchTipsValue `json:"value"`
}

type SearchTipsValue struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type SearchNewRecTag struct {
	TagName  string `json:"tag_name"`
	TagStyle int    `json:"tag_style"`
}

type SearchGame struct {
	Status    int64  `json:"status"`
	Author    string `json:"author"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	Pos       int64  `json:"pos"`
	CardType  int64  `json:"card_type"`
	State     int64  `json:"state"`
	Corner    string `json:"corner"`
	CardValue string `json:"card_value"`
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	Desc      string `json:"desc"`
}

// SearchTypeRes search type res.
type SearchTypeRes struct {
	Code           int             `json:"code,omitempty"`
	SeID           string          `json:"seid,omitempty"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"pagesize,omitempty"`
	Total          int             `json:"total,omitempty"`
	NumResults     int             `json:"numResults"`
	NumPages       int             `json:"numPages"`
	SuggestKeyword string          `json:"suggest_keyword"`
	RqtType        string          `json:"rqt_type,omitempty"`
	CostTime       json.RawMessage `json:"cost_time,omitempty"`
	ExpList        json.RawMessage `json:"exp_list,omitempty"`
	EggHit         int             `json:"egg_hit"`
	PageInfo       json.RawMessage `json:"pageinfo,omitempty"`
	Result         json.RawMessage `json:"result,omitempty"`
	ShowColumn     int             `json:"show_column"`
	InBlackKey     int8            `json:"in_black_key"`
	InWhiteKey     int8            `json:"in_white_key"`
}

// SearchRec search recommend.
type SearchRec struct {
	Code           int             `json:"code,omitempty"`
	SeID           string          `json:"seid,omitempty"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"pagesize,omitempty"`
	Total          int             `json:"total,omitempty"`
	NumResults     int             `json:"numResults"`
	NumPages       int             `json:"numPages"`
	SuggestKeyword string          `json:"suggest_keyword"`
	RqtType        string          `json:"rqt_type,omitempty"`
	CostTime       json.RawMessage `json:"cost_time,omitempty"`
	ExpList        json.RawMessage `json:"exp_list,omitempty"`
	EggHit         int             `json:"egg_hit"`
	Result         json.RawMessage `json:"result,omitempty"`
}

type SquareArg struct {
	Limit    int    `form:"limit" validate:"min=1,max=50"`
	IsInner  int64  `form:"is_inner" validate:"min=0,max=1"`
	Platform string `form:"platform"`
}

type SquareResult struct {
	Trending *SquareTrending `json:"trending"`
}

type SquareTrending struct {
	Title   string        `json:"title"`
	Trackid string        `json:"trackid"`
	List    []*SquareList `json:"list"`
}

type SquareList struct {
	Keyword  string `json:"keyword"`
	ShowName string `json:"show_name"`
	Icon     string `json:"icon"`
	URI      string `json:"uri"`
	Goto     string `json:"goto"`
}

// Hot struct
type Hot struct {
	Code    int    `json:"code"`
	SeID    string `json:"seid"`
	TrackID string `json:"trackid"`
	List    []*struct {
		Keyword   string `json:"keyword"`
		ShowName  string `json:"show_name"`
		Icon      string `json:"icon"`
		URI       string `json:"uri"`
		GotoType  int    `json:"goto_type"`
		GotoValue string `json:"goto_value"`
		WordType  int    `json:"word_type"`
	} `json:"list"`
	ExpStr string `json:"exp_str"`
}

// SearchAllArg search all api arguments.
type SearchAllArg struct {
	Pn            int    `form:"page"`
	Keyword       string `form:"keyword" validate:"required"`
	Rid           int    `form:"tids"`
	Duration      int    `form:"duration" validate:"gte=0,lte=4"`
	FromSource    string `form:"from_source"`
	Highlight     int    `form:"highlight"`
	FromSpmid     string `form:"from_spmid"`
	Platform      string `form:"platform"`
	SingleColumn  int    `form:"-"`
	DynamicOffset int64  `form:"dynamic_offset"`
	IsInner       int64  `form:"is_inner" validate:"min=0,max=1"`
	PageSize      int64  `form:"page_size" validate:"min=0,max=50"`
	MobiApp       string `form:"mobi_app"`
}

// SearchTypeArg search type api arguments.
type SearchTypeArg struct {
	Pn         int    `form:"page" validate:"min=1" default:"1"`
	SearchType string `form:"search_type" validate:"required"`
	Keyword    string `form:"keyword" validate:"required"`
	Order      string `form:"order"`
	Rid        int64  `form:"tids"`
	FromSource string `form:"from_source"`
	Platform   string `form:"platform"`
	Duration   int    `form:"duration" validate:"min=0,max=4"`
	// article
	CategoryID int64 `form:"category_id"`
	// special
	VpNum int `form:"vp_num"`
	// bili user
	BiliUserVl    int    `form:"bili_user_vl" default:"3"`
	UserType      int    `form:"user_type" validate:"min=0,max=3"`
	OrderSort     int    `form:"order_sort"`
	Highlight     int    `form:"highlight"`
	SingleColumn  int    `form:"-"`
	FromSpmid     string `form:"from_spmid"`
	IsInner       int64  `form:"is_inner" validate:"min=0,max=1"`
	DynamicOffset int64  `form:"dynamic_offset"`
	PageSize      int64  `form:"page_size" validate:"min=0,max=50"`
	MobiApp       string `form:"mobi_app"`
}

// SearchDefault search default
type SearchDefault struct {
	Trackid   string `json:"seid"`
	ID        int64  `json:"id"`
	Type      int    `json:"type"`
	ShowName  string `json:"show_name"`
	Name      string `json:"name"`
	GotoType  int    `json:"goto_type"`
	GotoValue string `json:"goto_value"`
	URL       string `json:"url"`
}

// SearchUpRecArg search up rec arg.
type SearchUpRecArg struct {
	ServiceArea string  `form:"service_area" validate:"required"`
	Platform    string  `form:"platform" validate:"required"`
	ContextID   int64   `form:"context_id"`
	MainTids    []int64 `form:"main_tids,split"`
	SubTids     []int64 `form:"sub_tids,split"`
	MobiApp     string  `form:"mobi_app"`
	Device      string  `form:"device"`
	Build       int64   `form:"build"`
	Ps          int     `form:"ps" default:"5" validate:"min=1,max=15"`
	Buvid       string  `form:"buvid"`
}

// SearchUpRecRes .
type SearchUpRecRes struct {
	UpID      int64  `json:"up_id"`
	RecReason string `json:"rec_reason"`
	Tid       int16  `json:"tid"`
	SecondTid int16  `json:"second_tid"`
}

// UpRecInfo .
type UpRecInfo struct {
	Mid      int64               `json:"mid"`
	Name     string              `json:"name"`
	Face     string              `json:"face"`
	Official accmdl.OfficialInfo `json:"official"`
	Follower int64               `json:"follower"`
	Vip      struct {
		Type   int32 `json:"type"`
		Status int32 `json:"status"`
	} `json:"vip"`
	RecReason   string `json:"rec_reason"`
	Tid         int16  `json:"tid"`
	Tname       string `json:"tname"`
	SecondTid   int16  `json:"second_tid"`
	SecondTname string `json:"second_tname"`
	Sign        string `json:"sign"`
}

// UpRecData .
type UpRecData struct {
	TrackID string       `json:"track_id"`
	List    []*UpRecInfo `json:"list"`
}

// SearchEgg .
type SearchEgg struct {
	Plat map[int64][]*struct {
		EggID int64  `json:"egg_id"`
		Plat  int    `json:"plat"`
		URL   string `json:"url"`
		MD5   string `json:"md5"`
		Size  int64  `json:"size"`
	} `json:"plat"`
	ShowCount int `json:"show_count"`
}

// SearchEggRes .
type SearchEggRes struct {
	EggID     int64              `json:"egg_id"`
	ShowCount int                `json:"show_count"`
	Source    []*SearchEggSource `json:"source"`
}

// SearchEggSource .
type SearchEggSource struct {
	URL  string `json:"url"`
	MD5  string `json:"md5"`
	Size int64  `json:"size"`
}

// SearchType search types
const (
	SearchTypeAll      = "all"
	SearchTypeVideo    = "video"
	SearchTypeBangumi  = "media_bangumi"
	SearchTypeMovie    = "media_ft"
	SearchTypeLive     = "live"
	SearchTypeLiveRoom = "live_room"
	SearchTypeLiveUser = "live_user"
	SearchTypeArticle  = "article"
	SearchTypeSpecial  = "special"
	SearchTypeTopic    = "topic"
	SearchTypeUser     = "bili_user"
	SearchTypePhoto    = "photo"
	WxSearchTypeAll    = "wx_all"
)

// SearchDefaultArg search default params.
var SearchDefaultArg = map[string]map[string]int{
	SearchTypeAll: {
		"highlight":         1,
		"video_num":         20,
		"media_bangumi_num": 3,
		"media_ft_num":      3,
		"is_new_pgc":        1,
		"live_room_num":     1,
		"card_num":          1,
		"activity":          1,
		"bili_user_num":     1,
		"bili_user_vl":      3,
		"user_num":          1,
		"user_video_limit":  3,
		"is_star":           1,
	},
	SearchTypeVideo: {
		"highlight":  1,
		"pagesize":   20,
		"is_new_pgc": 1,
	},
	SearchTypeBangumi: {
		"highlight": 1,
		"pagesize":  20,
	},
	SearchTypeMovie: {
		"highlight": 1,
		"pagesize":  20,
	},
	SearchTypeLive: {
		"highlight":     1,
		"live_user_num": 6,
		"live_room_num": 40,
	},
	SearchTypeLiveRoom: {
		"highlight": 1,
		"pagesize":  40,
	},
	SearchTypeLiveUser: {
		"highlight": 1,
		"pagesize":  30,
	},
	SearchTypeArticle: {
		"highlight": 1,
		"pagesize":  20,
	},
	SearchTypeSpecial: {
		"pagesize": 20,
	},
	SearchTypeTopic: {
		"pagesize": 20,
	},
	SearchTypeUser: {
		"highlight": 1,
		"pagesize":  20,
	},
	SearchTypePhoto: {
		"pagesize": 20,
	},
	WxSearchTypeAll: {
		"video_num":         20,
		"media_bangumi_num": 3,
		"media_ft_num":      3,
		"is_new_pgc":        1,
	},
}

// SearchLiveRoom
type SearchLiveRoom struct {
	// 以下为搜索透传字段
	Area             int64    `json:"area"`
	Attentions       int64    `json:"attentions"`
	CateName         string   `json:"cate_name"`
	Cover            string   `json:"cover"`
	HitColumns       []string `json:"hit_columns"`
	IsLiveRoomInline int64    `json:"is_live_room_inline"`
	LiveStatus       int64    `json:"live_status"`
	LiveTime         string   `json:"live_time"`
	Online           int64    `json:"online"`
	RankIndex        int64    `json:"rank_index"`
	RankOffset       int64    `json:"rank_offset"`
	RankScore        int64    `json:"rank_score"`
	Roomid           int64    `json:"roomid"`
	ShortID          int64    `json:"short_id"`
	Tags             string   `json:"tags"`
	Title            string   `json:"title"`
	Type             string   `json:"type"`
	Uface            string   `json:"uface"`
	UID              int64    `json:"uid"`
	Uname            string   `json:"uname"`
	UserCover        string   `json:"user_cover"`
	// 补充字段,直播人气改为看过
	WatchedShow *watchedmdl.WatchedShow `json:"watched_show"`
}

type SearchLiveUser struct {
	Area       int64         `json:"area"`
	AreaV2ID   int64         `json:"area_v2_id"`
	Attentions int64         `json:"attentions"`
	CateName   string        `json:"cate_name"`
	HitColumns []interface{} `json:"hit_columns"`
	IsLive     bool          `json:"is_live"`
	LiveStatus int64         `json:"live_status"`
	LiveTime   string        `json:"live_time"`
	RankIndex  int64         `json:"rank_index"`
	RankOffset int64         `json:"rank_offset"`
	RankScore  int64         `json:"rank_score"`
	Roomid     int64         `json:"roomid"`
	Tags       string        `json:"tags"`
	Type       string        `json:"type"`
	Uface      string        `json:"uface"`
	UID        int64         `json:"uid"`
	Uname      string        `json:"uname"`
}
