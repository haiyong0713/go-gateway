package model

import "time"

// tv_xiaopeng_intervene 表结构
type TvXiaoPengInterveneModel struct {
	Id        uint64    `json:"id"`
	Type      int       `json:"key_type"`
	KeyWord   string    `json:"keyword"`
	CardType  int       `json:"card_type"`
	Aid       int64     `json:"aid"`
	Rank      uint      `json:"rank"`
	IsDeleted int       `json:"is_deleted"`
	Ctime     time.Time `json:"ctime"`
	Mtime     time.Time `json:"mtime"`
}

type XiaoPengRecShowList struct {
	Items []int64 `json:"items"`
}
