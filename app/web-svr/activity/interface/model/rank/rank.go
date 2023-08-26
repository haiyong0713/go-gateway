package rank

import (
	xtime "go-common/library/time"
)

const (
	// RankStateOnline 有效
	RankStateOnline = 1
	// RankStateOffline 无效
	RankStateOffline = 0
	// RankTypeUp  up主维度
	RankTypeUp = 1
	// RankTypeArchive  稿件维度
	RankTypeArchive = 2
	// InterventionTypeWhite 白名单
	InterventionTypeWhite = 1
	// InterventionTypeBlack 黑名单
	InterventionTypeBlack = 2
	// InterventionObjectUp 干预对象up主
	InterventionObjectUp = 1
	// InterventionObjectArchive 干预对象稿件
	InterventionObjectArchive = 2
	// IsShowScore 是否展示分数
	IsShowScore = 1
	// IsNotShowScore 不展示分数
	IsNotShowScore = 0
	// AidSource 视频数据源
	AidSource = 1
)

// Redis rank redis struct
type Redis struct {
	Mid   int64   `json:"mid"`
	Rank  int     `json:"rank"`
	Score int64   `json:"score"`
	Aids  []int64 `json:"aids"`
	Diff  int64   `json:"diff"`
}

// MidRank mid rank
type MidRank struct {
	Mid   int64       `json:"mid"`
	Score int64       `json:"score"`
	Rank  int64       `json:"rank"`
	Aids  []*AidScore `json:"aids"`
}

// AidScore ...
type AidScore struct {
	Aid   int64 `json:"aid"`
	Score int64 `json:"score"`
}

// Result 用户积分信息
type Result struct {
	Account *Account `json:"account"`
	Score   int64    `json:"score"`
	Videos  []*Video `json:"video"`
}

// MidRankReply ...
type MidRankReply struct {
	Rank  int64    `json:"rank"`
	Video []*Video `json:"video"`
	Score int64    `json:"score"`
}

// Rank 排行榜
type Rank struct {
	ID             int64      `form:"id" json:"id" validate:"min=1"`
	SID            int64      `form:"sid" json:"sid"`
	SIDSource      int        `form:"sid_source" json:"sid_source"`
	Ratio          string     `form:"ratio" json:"ratio" validate:"required"`
	RankType       int        `form:"rank_type" json:"rank_type" validate:"min=1"`
	RankAttribute  int64      `form:"rank_attribute" json:"rank_attribute"`
	Top            int64      `form:"top" json:"top" validate:"min=1"`
	IsAuto         int64      `form:"is_auto" json:"is_auto" validate:"required"`
	IsShowScore    int64      `form:"is_show_score" json:"is_show_score"`
	State          int64      `form:"state" json:"state"`
	StatisticsTime string     `form:"statistics_time" json:"statistics_time"`
	Stime          xtime.Time `form:"stime" json:"stime"`
	Etime          xtime.Time `form:"etime" json:"etime"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
}

// Account 账号信息
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// ResultReply 排行榜返回结构
type ResultReply struct {
	Rank []*Result `json:"rank"`
	Page *Page     `json:"page"`
}

// Page ...
type Page struct {
	Num     int  `json:"num"`
	Size    int  `json:"size"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// Video ...
type Video struct {
	Mid      int64      `json:"mid"`
	Bvid     string     `json:"bvid"`
	TypeName string     `json:"tname"`
	Title    string     `json:"title"`
	Desc     string     `json:"desc"`
	Duration int64      `json:"duration"`
	Pic      string     `json:"pic"`
	View     int32      `json:"view"`
	Like     int32      `json:"like"`
	Danmaku  int32      `json:"danmaku"`
	Reply    int32      `json:"reply"`
	Fav      int32      `json:"fav"`
	Coin     int32      `json:"coin"`
	Share    int32      `json:"share"`
	Score    int64      `json:"score"`
	PubDate  xtime.Time `json:"ctime"`
}
