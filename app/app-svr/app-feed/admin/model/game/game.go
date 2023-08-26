package game

import (
	xtime "go-common/library/time"
)

// Game .
type Game struct {
	ID    int64
	Title string
	Image string
}

// GameInfo game info
type GameInfo struct {
	Code      int    `json:"code"`
	RequestID string `json:"request_id"`
	Ts        int64  `json:"ts"`
	Data      *struct {
		GameName       string  `json:"game_name"`
		GameIcon       string  `json:"game_icon"`
		GameBookStatus int     `json:"game_book_status"`
		GameLink       string  `json:"game_link"`
		Grade          float64 `json:"grade"`
		BookNum        int     `json:"book_num"`
	} `json:"data"`
}

// Info str
type Info struct {
	IsOnline  bool       `json:"is_online"`
	BaseID    int64      `json:"game_base_id"`
	Name      string     `json:"game_name"`
	Icon      string     `json:"game_icon"`
	Link      string     `json:"game_link"`
	Status    int        `json:"game_status"`
	BeginDate xtime.Time `json:"begin_date"`
	GameTags  string     `json:"game_tags"`
}

// EntryInfo Game Entry Info
type EntryInfo struct {
	BaseID int64  `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"iosLink"`
}
