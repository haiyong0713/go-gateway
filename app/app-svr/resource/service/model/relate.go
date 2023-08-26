package model

import xtime "go-common/library/time"

// Version app version
type Version struct {
	Plat      int8   `json:"plat,omitempty"`
	Build     int    `json:"build,omitempty"`
	Condition string `json:"conditions,omitempty"`
}

// Relate relate card
type Relate struct {
	ID          int64               `json:"id,omitempty"`
	Param       int64               `json:"param,omitempty"`
	Goto        string              `json:"goto,omitempty"`
	Title       string              `json:"title,omitempty"`
	ResourceIDs string              `json:"resource_ids,omitempty"`
	TagIDs      string              `json:"tag_ids,omitempty"`
	ArchiveIDs  string              `json:"archive_ids,omitempty"`
	RecReason   string              `json:"rec_reason,omitempty"`
	Position    int32               `json:"position,omitempty"`
	STime       xtime.Time          `json:"stime,omitempty"`
	ETime       xtime.Time          `json:"etime,omitempty"`
	Versions    map[int8][]*Version `json:"versions,omitempty"`
	PgcIDs      string
}

// 投放平台版本
type PlatVer struct {
	Plat  int8   `json:"plat,omitempty"`
	Build int    `json:"build,omitempty"`
	Cond  string `json:"cond,omitempty"`
}

// app相关推荐
type AppRcmd struct {
	ID        int64               `json:"id"`
	CardType  string              `json:"card_type"`
	CardValue string              `json:"card_value"`
	Position  int32               `json:"position"`
	PlatVer   map[int8][]*PlatVer `json:"plat_ver"`
	RecReason string              `json:"rec_reason"`
	PgcIds    string              `json:"pgc_ids"`
}
