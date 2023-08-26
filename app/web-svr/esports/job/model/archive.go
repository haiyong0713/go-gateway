package model

import (
	"encoding/json"
	"sync"

	arcClient "git.bilibili.co/bapis/bapis-go/archive/service"
)

type ArcMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
}

// Archive model
type Archive struct {
	Aid       int64  `json:"aid"`
	Mid       int64  `json:"mid"`
	TypeID    int16  `json:"typeid"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Attribute int32  `json:"attribute"`
	Copyright int8   `json:"copyright"`
	State     int    `json:"state"`
	Access    int    `json:"access"`
	PubTime   string `json:"pubtime"`
}

type SyncAutoTag struct {
	sync.Mutex
	RunningNum int64
	MaxNum     int64
}

func (auto *SyncAutoTag) Release() {
	auto.Lock()
	if auto.RunningNum >= 1 {
		auto.RunningNum--
	}
	auto.Unlock()
}

func (auto *SyncAutoTag) LockWithCheck() bool {
	auto.Lock()
	defer func() {
		auto.Unlock()
	}()
	if auto.RunningNum < auto.MaxNum {
		auto.RunningNum++
		return true
	}
	return false
}

// Message databus
type Message struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

// AutoRule
type AutoRule struct {
	Mids     map[int64]*RuleRs
	Tags     map[string]*RuleRs
	Keywords map[string]*RuleRs
}

// RuleRs
type RuleRs struct {
	ID       int64
	Name     string
	GameIDs  string
	MatchIDs string
}

type ArchiveStats struct {
	AID int64 `json:"aid"`

	CoinBefore14 int32 `json:"coin_before_14"`
	CoinIn14     int32 `json:"coin_in_14"`

	DanmakuBefore14 int32 `json:"danmaku_before_14"`
	DanmakuIn14     int32 `json:"danmaku_in_14"`

	FavoriteBefore14 int32 `json:"favorite_before_14"`
	FavoriteIn14     int32 `json:"favorite_in_14"`

	LikeBefore14 int32 `json:"like_before_14"`
	LikeIn14     int32 `json:"like_in_14"`

	ReplyBefore14 int32 `json:"reply_before_14"`
	ReplyIn14     int32 `json:"reply_in_14"`

	ShareBefore14 int32 `json:"share_before_14"`
	ShareIn14     int32 `json:"share_in_14"`

	ViewBefore14 int32 `json:"view_before_14"`
	ViewIn14     int32 `json:"view_in_14"`
}

func (stats *ArchiveStats) Rebuild(statsFromArcClient *arcClient.Stat) {
	stats.CoinIn14 = statsFromArcClient.Coin - stats.CoinBefore14
	stats.DanmakuIn14 = statsFromArcClient.Danmaku - stats.DanmakuBefore14
	stats.FavoriteIn14 = statsFromArcClient.Fav - stats.FavoriteBefore14
	stats.LikeIn14 = statsFromArcClient.Like - stats.LikeBefore14
	stats.ReplyIn14 = statsFromArcClient.Reply - stats.ReplyBefore14
	stats.ShareIn14 = statsFromArcClient.Share - stats.ShareBefore14
	stats.ViewIn14 = statsFromArcClient.View - stats.ViewBefore14
}

func (stats ArchiveStats) CalculateScore(statsType int) (score float64) {
	switch statsType {
	case ArchiveStatsTypeOfBefore14:
		score = float64(stats.CoinBefore14)*0.4 + float64(stats.FavoriteBefore14)*0.3 +
			float64(stats.DanmakuBefore14)*0.4 + float64(stats.ReplyBefore14)*0.4 +
			float64(stats.ViewBefore14)*0.25 + float64(stats.LikeBefore14)*0.4 +
			float64(stats.ShareBefore14)*0.6
	case ArchiveStatsTypeOfIn14:
		score = float64(stats.CoinIn14)*0.4 + float64(stats.FavoriteIn14)*0.3 +
			float64(stats.DanmakuIn14)*0.4 + float64(stats.ReplyIn14)*0.4 +
			float64(stats.ViewIn14)*0.25 + float64(stats.LikeIn14)*0.4 +
			float64(stats.ShareIn14)*0.6
	}

	return
}

const (
	ArchiveStatsTypeOfBefore14 = iota
	ArchiveStatsTypeOfIn14
)
