package search

import (
	"go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

var (
	// NotDelete not delete
	NotDelete uint8
	// Delete delete
	Delete uint8 = 1
	// Business log business ID
	Business = 202
	// ActionAddBlack log action
	ActionAddBlack = "ActionAddBlack"
	// ActionDelBlack log action
	ActionDelBlack = "ActionDelBlack"
	// ActionAddInter log action
	ActionAddInter = "ActionAddInter"
	// ActionUpdateInter log action
	ActionUpdateInter = "ActionUpdateInter"
	// ActionUpdateSearch log action
	ActionUpdateSearch = "ActionUpdateSearch"
	// ActionPublishHot log action
	ActionPublishHot = "ActionPublishHot"
	// ActionPublishDark log action
	ActionPublishDark = "ActionPublishDark"
	// ActionOpenAddHot log action
	ActionOpenAddHot = "ActionOpenAddHot"
	// ActionDeleteHot delete hot word
	ActionDeleteHot = "ActionDeleteHot"
	// ActionOpenAddDark log action
	ActionOpenAddDark = "ActionOpenAddDark"
	// ActionDeleteDark action delete darkword
	ActionDeleteDark = "ActionDeleteDark"
	// hot list 重新排序
	HotSort = "HotSort"
	// HotAI hot word from AI
	HotAI uint8 = 1
	// HotOpe hot word from operate
	HotOpe uint8 = 2
)

// Hot search history from ai and search words
type Hot struct {
	ID         uint   `json:"-"`
	Searchword string `json:"searchword"`
	PV         int64  `json:"pv"`
	UV         int64  `json:"uv"`
	Click      int64  `json:"click"`
	Atime      string
}

// OpenHot open api for searhc add hot every day
type OpenHot struct {
	Date   string `json:"date"`
	Values []Hot  `json:"values"`
}

// Dark search dark list
type Dark struct {
	ID         uint   `json:"id"`
	Searchword string `json:"searchword" form:"searchword"`
	PV         int64  `json:"pv" form:"pv"`
	UV         int64  `json:"uv" form:"uv"`
	Click      int64  `json:"click"`
	Atime      string `json:"atime"`
	Deleted    uint8  `json:"deleted"`
}

// OpenDark open api for search add dark word every day
type OpenDark struct {
	Date   string `json:"date"`
	Values []Dark `json:"values"`
}

// Black search Black
type Black struct {
	Searchword string    `json:"searchword" form:"searchword"  validate:"required"`
	ID         int       `json:"id"`
	Deleted    uint8     `json:"deleted"`
	Mtime      time.Time `json:"mtime"`
}

// AddBlack add search Black
type AddBlack struct {
	Searchword string `json:"searchword" form:"searchword"  validate:"required"`
}

// InterHisListPager .
type InterHisListPager struct {
	Item []*Intervene `json:"item"`
	Page common.Page  `json:"page"`
}

// InterHisParam Intervene list param
type InterHisParam struct {
	Searchword string `form:"searchword"`
	Date       string `form:"date"`
	Ps         int    `form:"ps" default:"20"` // 分页大小
	Pn         int    `form:"pn" default:"1"`  // 第几个分页
}

// Intervene search intervene word
type Intervene struct {
	ID             int       `json:"id" form:"id"`
	Searchword     string    `json:"searchword" form:"searchword"`
	Rank           int       `json:"position" form:"position"`
	Pv             int       `json:"pv"`
	Uv             int       `json:"uv"`
	Click          int       `json:"click"`
	Tag            string    `json:"tag" form:"tag"`
	Stime          time.Time `json:"stime" form:"stime"`
	Etime          time.Time `json:"etime" form:"etime"`
	Deleted        uint8     `json:"deleted"`
	Type           int       `json:"type" form:"type" validate:"required"`
	Image          string    `json:"image" form:"image"`
	GotoType       int       `json:"goto_type" form:"goto_type"`
	GotoValue      string    `json:"goto_value" form:"goto_value"`
	BvID           string    `json:"bvid" gorm:"-"`
	ShowWord       string    `json:"show_word" form:"show_word"`
	Intervene      int       `json:"intervene" form:"intervene"` // 是否是干预类型
	OnlinePosition int       `json:"online_pos" form:"online_pos"`
	ResourceId     string    `json:"resource_id" form:"resource_id"`
}

// BlackListParam black list param
type BlackListParam struct {
	Searchword string `form:"searchword"`
	Ps         int    `form:"ps" default:"20"` // 分页大小
	Pn         int    `form:"pn" default:"1"`  // 第几个分页
}

// BlackListPager .
type BlackListPager struct {
	Item []*Black    `json:"item"`
	Page common.Page `json:"page"`
}

// InterveneAdd add search intervene word
type InterveneAdd struct {
	ID         int       `json:"id" form:"id"`
	Searchword string    `json:"searchword" form:"searchword"`
	Rank       int       `json:"position" form:"position"`
	OldRank    int       `json:"-" form:"-" gorm:"-"`
	Tag        string    `json:"tag" form:"tag"`
	Stime      time.Time `json:"stime" form:"stime"`
	Etime      time.Time `json:"etime" form:"etime"`
	Type       int       `json:"type" form:"type" validate:"required"`
	Image      string    `json:"image" form:"image"`
	GotoType   int       `json:"goto_type" form:"goto_type"`
	GotoValue  string    `json:"goto_value" form:"goto_value"`
	ShowWord   string    `form:"show_word"`
	ResourceId string    `json:"resource_id" form:"resource_id"`
}

// HotwordOut hotword out put with publish state
type HotwordOut struct {
	Hotword        []Intervene `json:"hotword"`
	State          uint8       `json:"state"`
	LastSyncTime   int64       `json:"last_sync_time"`
	LastOnlineTime int64       `json:"last_online_time"`
}

// DarkwordOut hotword out put with publish state
type DarkwordOut struct {
	Darkword []Dark `json:"darkword"`
	State    uint8  `json:"state"`
}

// History search History
type History struct {
	ID         int    `json:"id" form:"id"`
	Searchword string `json:"searchword"`
	Pv         int    `json:"pv"`
	Uv         int    `json:"vu"`
	Click      int    `json:"click"`
	Position   int    `json:"position"`
	Atime      string `json:"atime"`
	Tag        string `json:"tag"`
	Deleted    uint8  `json:"deleted"`
}

// PublishState hot word publish state
type PublishState struct {
	Date  string
	State bool
}

// HotPubLog hotword publish log
type HotPubLog struct {
	ID         int       `json:"id" form:"id"`
	Searchword string    `json:"searchword" form:"searchword"`
	Position   int       `json:"position" form:"position"`
	Pv         int       `json:"pv"`
	Tag        string    `json:"tag" form:"tag"`
	Stime      time.Time `json:"stime" form:"stime"`
	Etime      time.Time `json:"etime" form:"etime"`
	Atime      string    `json:"atime"`
	Groupid    int64     `json:"groupid"`
}

// DarkPubLog dark publish log
type DarkPubLog struct {
	ID         uint   `json:"id"`
	Searchword string `json:"searchword" form:"searchword"`
	Pv         int64  `json:"pv" form:"pv"`
	Atime      string `json:"atime"`
	Groupid    int64  `json:"groupid"`
}

// 重新排序时候的参数结构
type SortConfigItem struct {
	Id         int    `json:"id"`
	Position   int    `json:"position"`
	Searchword string `json:"searchword"`
	Intervene  int    `json:"intervene"`
	ResourceId string `json:"resource_id"`
}

// 上次同步的时间
type LastTimeSyncItem struct {
	LastTime int64 `json:"last_time"`
}

type StaticticsListItem struct {
	LogDate    string `json:"log_date"`
	Rank       int    `json:"position"`
	Pv         int    `json:"pv"`
	Uv         int    `json:"uv"`
	ClickRate  string `json:"click_rate"`
	SearchWord string `json:"search_word"`
	STime      string `json:"stime"`
	ETime      string `json:"etime"`
}

type StaticticsLiveListItem struct {
	SearchWord string `json:"search_word"`
	Pos        int    `json:"position"`
	Pv         int    `json:"pv"`
	Click      int    `json:"click"`
	Clickrate  string `json:"click_rate"`
}

// TableName search box history
func (a Hot) TableName() string {
	return "search_histories"
}

// TableName search_blacklist
func (a Black) TableName() string {
	return "search_blacklist"
}

// TableName search_darkword
func (a Dark) TableName() string {
	return "search_darkword"
}

// TableName search_histories
func (a History) TableName() string {
	return "search_histories"
}

// TableName search_blacklist
func (a AddBlack) TableName() string {
	return "search_blacklist"
}

// TableName search_intervene
func (a Intervene) TableName() string {
	return "search_intervene"
}

// TableName InterveneAdd search_intervene
func (a InterveneAdd) TableName() string {
	return "search_intervene"
}

// TableName DarkPubLog dark word publish log
func (a DarkPubLog) TableName() string {
	return "search_darkword_log"
}

// TableName DarkPubLog dark word publish log
func (a HotPubLog) TableName() string {
	return "search_hotword_log"
}
