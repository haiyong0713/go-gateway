package bws

import (
	"go-common/library/time"
)

// UserPoint .
type UserPoint struct {
	ID      int64     `json:"id"`
	Pid     int64     `json:"pid"`
	Points  int64     `json:"points"`
	Ctime   time.Time `json:"ctime"`
	IsPoint bool      `json:"is_point"`
}

// UserLockPointReply .
type UserLockPointReply struct {
	*UserPoint
	Info *Point `json:"info"`
}

// UserPointDetail .
type UserPointDetail struct {
	*UserPoint
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Fid          int64  `json:"fid"`
	Image        string `json:"image"`
	Unlocked     int64  `json:"unlocked"`
	LoseUnlocked int64  `json:"lose_unlocked"`
	LockType     int64  `json:"lockType"`
	Dic          string `json:"dic"`
	Rule         string `json:"rule"`
	Bid          int64  `json:"bid"`
	ID           int64  `json:"id"`
	Owner        int64  `json:"_"`
}

// RechargeAward .
type RechargeAward struct {
	*PointsLevel
	Awards []*PointsAward `json:"awards"`
}

// PointReply .
type PointReply struct {
	*Point
	UnlockTotal int64            `json:"unlock_total"`
	Sign        *SignInfoReply   `json:"sign,omitempty"`
	Recharge    []*RechargeAward `json:"recharge,omitempty"`
}

// RechargeAwardReply  .
type RechargeAwardReply struct {
	Recharge []*Unlocks `json:"recharge"`
}

// Unlocks .
type Unlocks struct {
	*Point
	Unlock    []*RechargeReply `json:"unlock"`
	NotUnlock []*RechargeReply `json:"not_unlock"`
}

// RechargeReply .
type RechargeReply struct {
	Level  int32  `json:"level"`
	Icon   string `json:"icon"`
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
	ID     int64  `json:"id"`
}

// FieldsReply .
type FieldsReply struct {
	Fields map[int64]*ActField `json:"fields"`
}
