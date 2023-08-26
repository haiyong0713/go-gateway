package dubbing

import (
	"go-gateway/app/web-svr/activity/interface/model/task"
)

// MidScore 用户积分情况
type MidScore struct {
	Aids  []int64 `json:"aids"`
	Rank  int64   `json:"rank"`
	Score int64   `json:"score"`
}

// MemberActivityInfoReply 用户活动状态信息
type MemberActivityInfoReply struct {
	Task  []*task.MidRule         `json:"task"`
	Rank  *Rank                   `json:"rank"`
	Score map[int64]*RedisDubbing `json:"score"`
}

// RankMember 用户积分信息
type RankMember struct {
	Account *Account `json:"account"`
	Score   int64    `json:"score"`
	Videos  []*Video `json:"video"`
}

// RankReply 排行榜返回结构
type RankReply struct {
	Rank []*RankMember `json:"rank"`
}

// Rank ...
type Rank struct {
	Rank  int      `json:"rank"`
	Video []*Video `json:"video"`
	Score int64    `json:"score"`
	Diff  int64    `json:"diff"`
}

// Video ...
type Video struct {
	Bvid     string `json:"bvid"`
	TypeName string `json:"tname"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	Duration int64  `json:"duration"`
	Pic      string `json:"pic"`
	View     int32  `json:"view"`
}

// Account 账号信息
type Account struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Sex  string `json:"sex"`
}

// MoneyCountReply ..
type MoneyCountReply struct {
	MoneyCount *MoneyCount `json:"money_count"`
}

// MoneyCount ...
type MoneyCount struct {
	Money int64 `json:"money"`
	Count int64 `json:"count"`
}

// RedisDubbing rank redis struct
type RedisDubbing struct {
	Mid   int64 `json:"mid"`
	Score int64 `json:"score"`
	Diff  int64 `json:"diff"`
	Rank  int   `json:"rank"`
}

// MapMidDubbingScore ...
type MapMidDubbingScore struct {
	Mid   int64                   `json:"mid"`
	Score map[int64]*RedisDubbing `json:"score"`
}
