package bubble

import (
	xtime "go-common/library/time"
)

type Bubble struct {
	ID        int64      `json:"id"`
	Position  []*Postion `json:"position"`
	Icon      string     `json:"icon"`
	Desc      string     `json:"desc"`
	URL       string     `json:"url"`
	STime     xtime.Time `json:"stime"`
	ETime     xtime.Time `json:"etime"`
	Operator  string     `json:"operator"`
	State     int        `json:"state"`
	WhiteList string     `json:"white_list"`
}

type Postion struct {
	Plat       int   `json:"plat"`
	PositionID int64 `json:"position_id"`
}
