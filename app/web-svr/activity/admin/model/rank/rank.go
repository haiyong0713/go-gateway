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
	// AidSource 视频数据源
	AidSource = 1
)

// CreateReq 创建请求
type CreateReq struct {
	SID       int64      `form:"sid" validate:"min=1"`
	SIDSource int        `form:"sid_source" validate:"min=1"`
	Stime     xtime.Time `form:"stime"`
	Etime     xtime.Time `form:"etime"`
}

// DetailReq 创建请求
type DetailReq struct {
	SID       int64 `form:"sid" validate:"min=1"`
	SIDSource int   `form:"sid_source" validate:"min=1"`
}

// OfflineReq 下线请求
type OfflineReq struct {
	ID int64 `form:"id" validate:"min=1"`
}

// GetInterventionReq 获取黑白名单
type GetInterventionReq struct {
	ID               int64 `form:"id" json:"id" validate:"min=1"`
	InterventionType int   `form:"intervention_type" json:"intervention_type" validate:"min=1"`
	ObjectType       int   `form:"object_type" json:"object_type" validate:"min=1"`
	Pn               int   `form:"pn" json:"pn" default:"1"`
	Ps               int   `form:"ps" json:"ps" default:"10"`
}

// UpdateInterventionReq 编辑黑白名单
type UpdateInterventionReq struct {
	ID   int64  `form:"id" json:"id" validate:"min=1"`
	List string `form:"list" json:"list"`
}

// ResultReq 结果
type ResultReq struct {
	ID            int64 `form:"id" json:"id" validate:"min=1"`
	AttributeType int   `form:"rank_attribute" json:"rank_attribute"`
	Pn            int   `form:"pn" json:"pn" default:"1"`
	Ps            int   `form:"ps" json:"ps" default:"10"`
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
	IsAuto         int64      `form:"is_auto" json:"is_auto"`
	IsShowScore    int64      `form:"is_show_score" json:"is_show_score"`
	State          int64      `form:"state" json:"state"`
	StatisticsTime string     `form:"statistics_time" json:"statistics_time"`
	Stime          xtime.Time `form:"stime" json:"stime"`
	Etime          xtime.Time `form:"etime" json:"etime"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
}

// Log 日志
type Log struct {
	ID            int64      `form:"id" json:"id" validate:"min=1"`
	RankID        int64      `form:"rank_id" json:"rank_id"`
	Batch         int64      `form:"batch" json:"batch"`
	RankAttribute int64      `form:"rank_attribute" json:"rank_attribute"`
	State         int64      `form:"state" json:"state"`
	Ctime         xtime.Time `json:"ctime"`
	Mtime         xtime.Time `json:"mtime"`
}

// Intervention 黑白名单
type Intervention struct {
	ID               int64      ` json:"id"`
	OID              int64      `json:"oid"`
	Score            int64      `json:"score"`
	State            int        `json:"state"`
	InterventionType int        `json:"intervention_type"`
	ObjectType       int        `json:"object_type"`
	Ctime            xtime.Time `json:"ctime"`
	Mtime            xtime.Time `json:"mtime"`
}

// InterventionReply 列表
type InterventionReply struct {
	List []*Intervention ` json:"list"`
	Page *Page           ` json:"page"`
}

// Page ...
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// OidResult 排行榜结果
type OidResult struct {
	ID            int64      `json:"id"`
	OID           int64      `json:"oid"`
	Rank          int64      `json:"rank"`
	Score         int64      `json:"score"`
	RankAttribute int        `json:"rank_attribute"`
	State         int        `json:"state"`
	Batch         int        `json:"batch"`
	Remark        string     `json:"remark"`
	RemarkStruct  *Remark    `json:"_"`
	Ctime         xtime.Time `json:"ctime"`
	Mtime         xtime.Time `json:"mtime"`
}

// Remark ...
type Remark struct {
	Aids []int64 `json:"aids"`
}

// Account 账号信息
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// Result 用户积分信息
type Result struct {
	ID      int64    `json:"id"`
	Account *Account `json:"account"`
	State   int      `json:"state"`
	Rank    int64    `json:"rank"`
	Score   int64    `json:"score"`
	Videos  []*Video `json:"video"`
}

// ExportReq ...
type ExportReq struct {
	ID            int64 `form:"id" json:"id"  validate:"required"`
	AttributeType int   `form:"attribute_type" json:"attribute_type" `
}

// Video ...
type Video struct {
	ID       int64      `json:"id"`
	Mid      int64      `json:"mid"`
	Aid      int64      `json:"aid"`
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
	State    int        `json:"state"`
	Rank     int64      `json:"rank"`
	Bvid     string     `json:"bvid"`
	PubDate  xtime.Time `json:"ctime"`
}

// ResultEditReq ...
type ResultEditReq struct {
	ID           int64         `json:"id" form:"id" validate:"min=1"`
	Result       string        `json:"rank" form:"rank"`
	ResultStruct []*ResultEdit `json:"_"`
}

// PublishReq ...
type PublishReq struct {
	ID            int64 `json:"id" form:"id" validate:"min=1"`
	AttributeType int   `json:"rank_attribute" form:"rank_attribute" validate:"min=1"`
	Batch         int64 `json:"batch" form:"batch" validate:"min=1"`
}

// ResultEdit 排行榜结果变更
type ResultEdit struct {
	ID      int64              `json:"id"`
	Account *Account           `json:"account"`
	State   int                `json:"state"`
	Rank    int64              `json:"rank"`
	Score   int64              `json:"score"`
	Video   []*ResultVideoEdit `json:"video"`
}

// ResultVideoEdit ...
type ResultVideoEdit struct {
	ID    int64 `json:"id"`
	State int   `json:"state"`
	AID   int64 `json:"aid"`
	Score int64 `json:"score"`
	Rank  int64 `json:"rank"`
}

// ResultReply ...
type ResultReply struct {
	Rank  []*Result `json:"rank"`
	Page  *Page     `json:"page"`
	Batch int64     `json:"batch"`
}

// Snapshot 快照
type Snapshot struct {
	ID            int64      `json:"id"`
	MID           int64      `json:"mid"`
	AID           int64      `json:"aid"`
	TID           int64      `json:"tid"`
	View          int64      `json:"view"`
	Danmaku       int64      `json:"danmaku"`
	Reply         int64      `json:"reply"`
	Fav           int64      `json:"fav"`
	Coin          int64      `json:"coin"`
	Share         int64      `json:"share"`
	Like          int64      `json:"like"`
	Videos        int64      `json:"videos"`
	Rank          int64      `json:"rank"`
	RankAttribute int        `json:"rank_attribute"`
	Score         int64      `json:"score"`
	State         int        `json:"state"`
	Batch         int        `json:"batch"`
	Remark        string     `json:"remark"`
	Ctime         xtime.Time `json:"ctime"`
	Mtime         xtime.Time `json:"mtime"`
}

// AidScore ...
type AidScore struct {
	Aid   int64 `json:"aid"`
	Score int64 `json:"score"`
}

// MidRank mid rank
type MidRank struct {
	Mid   int64       `json:"mid"`
	Score int64       `json:"score"`
	Rank  int64       `json:"rank"`
	Aids  []*AidScore `json:"aids"`
}
