package stat

import (
	arcApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	TypeForView  = "view"
	TypeForDm    = "dm"
	TypeForReply = "reply"
	TypeForFav   = "fav"
	TypeForCoin  = "coin"
	TypeForShare = "share"
	TypeForRank  = "rank"
	TypeForLike  = "like"
)

// SeasonStat def.
type SeasonStat struct {
	LastTs int64 // last update time
	Stat   *api.Stat
	Aids   map[int64]struct{} // aids
}

// ArcToSn def.
type ArcToSn struct {
	Stat *arcApi.Stat // arc's stat which is used only for the season
	Sid  int64        // the season id of the archive
}

// SeasonMsg for JD
type SeasonMsg struct {
	SeasonID int64     `json:"season_id"`
	Action   string    `json:"action"` // update or delete
	Stat     *api.Stat `json:"stat"`
}

// Season structure modification message
type SeasonResult struct {
	SeasonID int64  `json:"season_id"`
	Action   string `json:"action"`
}

// Count is the core of the stat message
type Count struct {
	Type      string `json:"type"`
	Aid       int64  `json:"id"`
	Count     int    `json:"count"`
	DisLike   int    `json:"dislike_count"`
	TimeStamp int64  `json:"timestamp"`
}

// Msg stat info.
type Msg struct {
	Aid     int64  `json:"aid"`
	Click   int    `json:"click"`
	DM      int    `json:"dm"`
	Reply   int    `json:"reply"`
	Fav     int    `json:"fav"`
	Coin    int    `json:"coin"`
	Share   int    `json:"share"`
	NowRank int    `json:"now_rank"`
	HisRank int    `json:"his_rank"`
	Like    int    `json:"like"`
	DisLike int    `json:"dislike_count"`
	Type    string `json:"-"`
	Ts      int64  `json:"-"`
}

// Merge merge the message into the archive's stat
func Merge(m *Msg, s *arcApi.Stat) {
	if m.Click >= 0 && m.Type == TypeForView {
		s.View = int32(m.Click)
	}
	if m.Coin >= 0 && m.Type == TypeForCoin {
		s.Coin = int32(m.Coin)
	}
	if m.DM >= 0 && m.Type == TypeForDm {
		s.Danmaku = int32(m.DM)
	}
	if m.Fav >= 0 && m.Type == TypeForFav {
		s.Fav = int32(m.Fav)
	}
	if m.Reply >= 0 && m.Type == TypeForReply {
		s.Reply = int32(m.Reply)
	}
	if m.Share >= 0 && m.Type == TypeForShare && int32(m.Share) > s.Share {
		s.Share = int32(m.Share)
	}
	if m.NowRank >= 0 && m.Type == TypeForRank {
		s.NowRank = int32(m.NowRank)
	}
	if m.HisRank >= 0 && m.Type == TypeForRank {
		s.HisRank = int32(m.HisRank)
	}
	if m.Like >= 0 && m.Type == TypeForLike {
		s.Like = int32(m.Like)
	}
	if m.DisLike >= 0 && m.Type == TypeForLike {
		s.DisLike = int32(m.DisLike)
	}
}

// MergeArcStatToSn merge the
func MergeArcStat(snStat *api.Stat, arcStat *arcApi.Stat) {
	snStat.View = snStat.View + arcStat.View
	snStat.Reply = snStat.Reply + arcStat.Reply
	snStat.Coin = snStat.Coin + arcStat.Coin
	snStat.Danmaku = snStat.Danmaku + arcStat.Danmaku
	snStat.Like = snStat.Like + arcStat.Like
	snStat.Fav = snStat.Fav + arcStat.Fav
	snStat.Share = snStat.Share + arcStat.Share
}
