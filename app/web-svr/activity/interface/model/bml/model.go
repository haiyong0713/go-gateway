package bml

import "go-common/library/time"

const (
	GuessTypeCommon = 1
	GuessTypeJoker  = 2
)

const (
	GuessOrderRecordStateInit     = 0
	GuessOrderRecordStateComplete = 1
	GuessOrderRecordStateDelete   = 2
)

type GuessOrderRecord struct {
	Id          int64     `json:"id"`
	Mid         int64     `json:"mid"`
	GuessType   int       `json:"guess_type"`
	RewardId    int64     `json:"reward_id"`
	State       int       `json:"state"`
	OrderNo     string    `json:"order_no"`
	GuessAnswer string    `json:"guess_answer"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

type RewardConf struct {
	RewardId      int64 `json:"reward_id"`
	RewardVersion int64 `json:"reward_version"`
	StockLimit    int   `json:"stock_limit"`
}

type GuessResult struct {
	GuessType int  `json:"guess_type"`
	IsRight   bool `json:"is_right"`
}

type GuessRecordItem struct {
	GuessType int   `json:"guess_type"`
	GuessTime int64 `json:"guess_time"`
	RewardId  int64 `json:"reward_id"`
}
