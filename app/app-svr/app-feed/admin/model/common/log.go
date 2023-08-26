package common

import "encoding/json"

const (
	//ActionAdd .
	ActionAdd = "add"
	//ActionEdit .
	ActionEdit = "edit"
	//ActionUpdate .
	ActionUpdate = "update"
	//ActionDelete .
	ActionDelete = "delete"
	//ActionOpt .
	ActionOpt = "opt"
	//ActionOnline .
	ActionOnline = "online"
	//ActionOffline .
	ActionOffline = "offline"
	//BusinessID action log business ID
	BusinessID = 204
	// tab menu business
	TabBusinessID = 207
	// skin menu bsiness
	SkinBusinessID = 208
	//LogPopularStars popular new start card log
	LogPopularStars = 0
	//LogChannelTab channel tab log
	LogChannelTab = 1
	//LogEventTopic popular event topic log
	LogEventTopic = 2
	//LogSWEBCard search web card log
	LogSWEBCard = 3
	//LogSWEB search web log
	LogSWEB = 4
	//LogPopRcmd popular recommend
	LogPopRcmd = 8
	//LogWebRcmdCard web recommand card log
	LogWebRcmdCard = 5
	//LogWebRcmd web recommand log
	LogWebRcmd = 6
	//LogSearchEgg search egg
	LogSearchEgg = 7
	//LogSelectedSerie .
	LogSelectedSerie = 9
	//LogSelectedResource .
	LogSelectedResource = 10
	//LogDynSear dynamic public search
	LogDynSear = 11
	//LogSeashShield log search shield
	LogSeashShield = 12
	// LogAggregation
	LogAggregation = 13
	// LogEntrance
	LogEntrance = 14
	// LogWebSerModule
	LogWebSerModule = 15
	// LogOgvModule
	LogOgvModule = 16
	// LogResourceCustomConfig
	LogResourceCustomConfig = 17
	// LogBubble
	LogBubble = 18
	// LogEntranceHidden
	LogEntranceHidden = 460
	// LogLiveCard
	LogLiveCard = 19
	// LogArticleCard
	LogArticleCard = 20
	// LogInformationRecommendCard
	LogInformationRecommendCard = 30
	// LogMngIcon
	LogMngIcon = 530
	// LogSidebar
	LogSidebar = 531
	// LogFeatureBuild
	LogFeatureBuild = 21
	// 搜索提示
	LogSearchTips = 22
	// 搜索结果白名单
	LogSearchWhiteList = 23
	// LogFeatureBusinessConfig
	LogFeatureBusinessConfig = 24
	// LogFeatureABTest
	LogFeatureABTest = 25
)

// LogManager .
type LogManager struct {
	ID        int    `json:"id"`
	OID       int    `json:"oid"`
	Uname     string `json:"uname"`
	UID       int    `json:"uid"`
	Type      int    `json:"module"`
	ExtraData string `json:"content"`
	Action    string `json:"action"`
	CTime     string `json:"ctime"`
	ActionEn  string `json:"action_english"`
	Str_0     string `json:"str_0"`
	Str_1     string `json:"str_1"`
}

// LogSearch .
type LogSearch struct {
	ID        int    `json:"id"`
	OID       int    `json:"oid"`
	Uname     string `json:"uname"`
	UID       int    `json:"uid"`
	Type      int    `json:"type"`
	ExtraData string `json:"extra_data"`
	Action    string `json:"action"`
	ActionEn  string `json:"action_english"`
	CTime     string `json:"ctime"`
	Str_0     string `json:"str_0"`
	Str_1     string `json:"str_1"`
}

type LogES struct {
	Code int           `json:"code"`
	Data *SearchResult `json:"data"`
}

// Page .
type SearchPage struct {
	Pn    int   `json:"num"`
	Ps    int   `json:"size"`
	Total int64 `json:"total"`
}

// SearchResult search result (deprecated).
type SearchResult struct {
	Order  string            `json:"order"`
	Sort   string            `json:"sort"`
	Result []json.RawMessage `json:"result"`
	Debug  string            `json:"debug"`
	Page   *SearchPage       `json:"page"`
}

// ManagerPage .
type ManagerPage struct {
	CurrentPage int `json:"current_page"`
	TotalItems  int `json:"total_items"`
	PageSize    int `json:"page_size"`
}

// LogManagers .
type LogManagers struct {
	Item []*LogManager `json:"item"`
	Page ManagerPage   `json:"pager"`
}

type Log struct {
	Type      int64  `form:"module" validate:"required"`
	Uname     string `form:"uname"`
	Starttime string `form:"starttime"`
	Endtime   string `form:"endtime"`
	Ps        int64  `form:"pagesize" default:"20"`
	Pn        int64  `form:"page" default:"1"`
	ID        int64  `form:"id"`
	Query     string `form:"query"`
	Action    string `form:"action"`
	Title     string `form:"title"`
}
