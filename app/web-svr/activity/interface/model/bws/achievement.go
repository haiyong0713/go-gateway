package bws

import "go-common/library/time"

const (
	DefaultRank       = -1
	CompositeRankType = 1
)

// UserAchieve .
type UserAchieve struct {
	ID    int64     `json:"id"`
	Aid   int64     `json:"aid"`
	Award int64     `json:"award"`
	Ctime time.Time `json:"ctime"`
	Key   string    `json:"key"`
}

// UserAchieveDetail .
type UserAchieveDetail struct {
	*UserAchieve
	Name          string `json:"name"`
	Icon          string `json:"icon"`
	Dic           string `json:"dic"`
	LockType      int64  `json:"lockType"`
	Unlock        int64  `json:"unlock"`
	Bid           int64  `json:"bid"`
	IconBig       string `json:"icon_big"`
	IconActive    string `json:"icon_active"`
	IconActiveBig string `json:"icon_active_big"`
	SuitID        int64  `json:"suit_id"`
	AchievePoint  int64  `json:"achieve_point"`
	Level         int32  `json:"level"`
}

// CountAchieves count achieve
type CountAchieves struct {
	Aid   int64 `json:"aid"`
	Count int64 `json:"count"`
}

// RankAchieve .
type RankAchieve struct {
	Num   int64 `json:"num"`
	Ctime int64 `json:"ctime"`
}

type UserGrade struct {
	Pid    int64     `json:"pid"`
	Key    string    `json:"key"`
	Amount int64     `json:"amount"`
	Mtime  time.Time `json:"mtime"`
}

type RankUserGrade struct {
	Mid    int64   `json:"mid"`
	Amount float64 `json:"amount"`
}
