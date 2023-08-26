package model

import "go-gateway/app/app-svr/archive/service/api"

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
	TypeForNowRank = "nowRank"
	TypeForHisRank = "hisRank"
	TypeForDislike = "dislike"
	TypeForFollow  = "follow"
	TypeForLikeYLF = "like-ylf"
	TypeForLikeJD  = "like-jd"
)

// StatMsg stat info.
type StatMsg struct {
	Aid      int64         `json:"aid"`
	Click    int32         `json:"click"`
	DM       int32         `json:"dm"`
	Reply    int32         `json:"reply"`
	Fav      int32         `json:"fav"`
	Coin     int32         `json:"coin"`
	Share    int32         `json:"share"`
	NowRank  int32         `json:"now_rank"`
	HisRank  int32         `json:"his_rank"`
	Like     int32         `json:"like"`
	Follow   int32         `json:"follow"`
	Type     string        `json:"-"`
	Ts       int64         `json:"-"`
	Platform *ViewPlatform `json:"platform,omitempty"`
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

// StatCount is
type StatCount struct {
	Type      string        `json:"type"`
	Aid       int64         `json:"id"`
	Count     int32         `json:"count"`
	TimeStamp int64         `json:"timestamp"`
	Platform  *ViewPlatform `json:"platform,omitempty"`
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
	statMap[TypeForDislike] = int64(stat.DisLike)
	statMap[TypeForFav] = int64(stat.Fav)
	statMap[TypeForView] = int64(stat.View)
	statMap[TypeForFollow] = int64(stat.Follow)
	return
}

func MapToStat(resMap map[string]int64, stat *api.Stat) {
	stat.Aid = resMap["aid"]
	// 播放数
	stat.View = int32(resMap[TypeForView])
	// 弹幕数
	stat.Danmaku = int32(resMap[TypeForDm])
	// 评论数
	stat.Reply = int32(resMap[TypeForReply])
	// 收藏数
	stat.Fav = int32(resMap[TypeForFav])
	// 投币数
	stat.Coin = int32(resMap[TypeForCoin])
	// 分享数
	stat.Share = int32(resMap[TypeForShare])
	// 当前排名
	stat.NowRank = int32(resMap[TypeForNowRank])
	// 历史最高排名
	stat.HisRank = int32(resMap[TypeForHisRank])
	// 点赞数
	stat.Like = int32(resMap[TypeForLike])
	// 追番数
	stat.Follow = int32(resMap[TypeForFollow])
}
