package bubble

import (
	xtime "go-common/library/time"
)

const (
	StateOnline  = 1
	StateOffline = -1

	BubbleNoExist = -1
	BubblePushing = 0
	BubblePushed  = 1
)

type Param struct {
	ID        int64      `form:"id"`
	Position  string     `form:"position"`
	Icon      string     `form:"icon"`
	Desc      string     `form:"desc"`
	URL       string     `form:"url"`
	STime     xtime.Time `form:"stime"`
	ETime     xtime.Time `form:"etime"`
	Operator  string     `form:"operator"`
	State     int        `form:"state"`
	WhiteList string     `form:"white_list"`
}

type ParamPostion struct {
	Plat       int   `json:"plat"`
	PositionID int64 `json:"position_id"`
}

type List struct {
	Page  *Page     `json:"page"`
	Items []*Bubble `json:"items"`
}

type Page struct {
	Total    int `json:"total"`
	PageNum  int `json:"num"`
	PageSize int `json:"size"`
}

type Bubble struct {
	ID        int64           `json:"id"`
	Position  []*ParamPostion `json:"position"`
	Icon      string          `json:"icon"`
	Desc      string          `json:"desc"`
	URL       string          `json:"url"`
	STime     xtime.Time      `json:"stime"`
	ETime     xtime.Time      `json:"etime"`
	Operator  string          `json:"operator"`
	State     int             `json:"state"`
	WhiteList string          `json:"white_list"`
}

type Sidebar struct {
	ID           int64  `json:"id"`
	Plat         int    `json:"plat"`
	Logo         string `json:"logo"`
	LogoSelected string `json:"logo_selected"`
	Name         string `json:"name"`
	Limit        string `json:"limit"`
}

type SidebarLimit struct {
	ID         int64  `json:"-"`
	Conditions string `json:"conditions"`
	Build      int64  `json:"build"`
}
