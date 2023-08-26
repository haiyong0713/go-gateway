package tips

import xtime "go-common/library/time"

type SearchTipDB struct {
	ID          int64      `gorm:"column:id"`
	Title       string     `gorm:"column:title"`
	SubTitle    string     `gorm:"column:sub_title"`
	IsImmediate int        `gorm:"column:is_immediate"`
	STime       xtime.Time `gorm:"column:stime"`
	ETime       xtime.Time `gorm:"column:etime"`
	CUser       string     `gorm:"column:cuser"`
	Status      int        `gorm:"column:online_status"`
	HasBgImg    int        `gorm:"column:has_bg_img"`
	JumpUrl     string     `gorm:"column:jump_url"`
	Plat        int        `gorm:"column:plat"`
}

func (*SearchTipDB) TableName() string {
	return "search_tips"
}

type SearchTipQueryDB struct {
	ID          int64  `gorm:"column:id" json:"-" form:"-"`
	SearchWord  string `gorm:"column:search_word" json:"query" form:"query"`
	SearchTipID int64  `gorm:"column:s_t_id" json:"-" form:"-"`
	Deleted     int    `gorm:"column:deleted" json:"-" form:"-"`
}

func (*SearchTipQueryDB) TableName() string {
	return "search_tips_query"
}

// 接口响应的 plat 字段
type PlatRes struct {
	PlatType int    `json:"plat_type" form:"plat_type"`
	PlatName string `json:"plat_name" form:"plat_name"`
}

// 接口响应
type SearchTipRes struct {
	ID          int64              `json:"id" form:"id"`
	Title       string             `json:"title" form:"title"`
	SubTitle    string             `json:"sub_title" form:"sub_title"`
	IsImmediate int                `json:"is_immediate" form:"is_immediate"`
	SearchWord  []SearchTipQueryDB `json:"search_word" form:"search_word"`
	STime       xtime.Time         `json:"stime" form:"stime"`
	ETime       xtime.Time         `json:"etime" form:"etime"`
	CUser       string             `json:"cuser" form:"cuser"`
	Status      int                `json:"status" form:"status"`
	HasBgImg    int                `json:"has_bg_img" form:"has_bg_img"`
	JumpUrl     string             `json:"jump_url" form:"jump_url"`
	Plat        []PlatRes          `json:"plat" form:"plat"`
}
