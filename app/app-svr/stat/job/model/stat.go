package model

import (
	"go-gateway/app/app-svr/archive/service/api"
)

// is
const (
	TypeForView    = "view"
	TypeForDm      = "dm"
	TypeForReply   = "reply"
	TypeForFav     = "fav"
	TypeForCoin    = "coin"
	TypeForShare   = "share"
	TypeForRank    = "rank"
	TypeForLike    = "like"
	TypeForHisRank = "hisRank"
	TypeForNowRank = "nowRank"
	TypeForDislike = "dislike"
	TypeForFollow  = "follow"
	TypeForLikeYLF = "like-ylf"
	TypeForLikeJD  = "like-jd"
)

// StatMsg stat info.
type StatMsg struct {
	Aid      int64         `json:"aid"`
	Click    int           `json:"click"`
	DM       int           `json:"dm"`
	Reply    int           `json:"reply"`
	Fav      int           `json:"fav"`
	Coin     int           `json:"coin"`
	Share    int           `json:"share"`
	NowRank  int           `json:"now_rank"`
	HisRank  int           `json:"his_rank"`
	Like     int           `json:"like"`
	Follow   int           `json:"follow"`
	Type     string        `json:"-"`
	Ts       int64         `json:"-"`
	Platform *ViewPlatform `json:"platform,omitempty"`
}

// StatCount 单次消息接收结构体
type StatCount struct {
	Type      string        `json:"type"`
	Aid       int64         `json:"id"`
	Count     int           `json:"count"`
	TimeStamp int64         `json:"timestamp"`
	Platform  *ViewPlatform `json:"platform,omitempty"`
}

// ViewPlatform each platform view
type ViewPlatform struct {
	Web       int64 `json:"web"`
	H5        int64 `json:"h5"`
	Outer     int64 `json:"outer"`
	Ios       int64 `json:"ios"`
	Android   int64 `json:"android"`
	AndroidTV int64 `json:"androidTV"`
}

func ConvertStatToMap(stat *api.Stat) (statMap map[string]int64) {
	statMap = make(map[string]int64)
	statMap["aid"] = stat.Aid
	statMap[TypeForShare] = int64(stat.Share)
	statMap[TypeForReply] = int64(stat.Reply)
	statMap[TypeForHisRank] = int64(stat.HisRank)
	statMap[TypeForNowRank] = int64(stat.NowRank)
	statMap[TypeForCoin] = int64(stat.Coin)
	statMap[TypeForDm] = int64(stat.Danmaku)
	statMap[TypeForLike] = int64(stat.Like)
	statMap[TypeForFav] = int64(stat.Fav)
	statMap[TypeForView] = int64(stat.View)
	statMap[TypeForFollow] = int64(stat.Follow)
	return
}
