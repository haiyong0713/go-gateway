package show

import (
	"fmt"
	"strings"

	"go-gateway/app/app-svr/app-feed/admin/model/common"

	xtime "go-common/library/time"
)

const (
	_ogvQuerySQL    = "INSERT INTO search_ogv_query(sid,value) VALUES %s"
	_ogvMoreShowSQL = "INSERT INTO search_ogv_moreshow(sid,word,`type`,value) VALUES %s"
)

// SearchOgv .
type SearchOgv struct {
	ID             int64       `gorm:"column:id" json:"id"`
	Color          int64       `gorm:"column:color" json:"color"`
	ColorStr       string      `gorm:"column:-" json:"color_str"`
	Stime          xtime.Time  `gorm:"column:stime" json:"stime"`
	Plat           string      `gorm:"column:plat" json:"plat"`
	HdCover        string      `gorm:"column:hd_cover" json:"hd_cover"`
	HdBg           string      `gorm:"column:hd_bg" json:"hd_bg"`
	HdTitle        string      `gorm:"column:hd_title" json:"hd_title"`
	HdSubtitle     string      `gorm:"column:hd_subtitle" json:"hd_subtitle"`
	GameStatus     int64       `gorm:"column:game_status" json:"game_status"`
	GamePos        int64       `gorm:"column:game_pos" json:"game_pos"`
	GameValue      string      `gorm:"column:game_value" json:"game_value"`
	PgcPos         int64       `gorm:"column:pgc_pos" json:"pgc_pos"`
	PgcIds         string      `gorm:"column:pgc_ids" json:"pgc_ids"`
	PgcMoreURL     string      `gorm:"column:pgc_more_url" json:"pgc_more_url"`
	PgcMoreType    int64       `gorm:"column:pgc_more_type" json:"pgc_more_type"`
	MoreshowStatus int64       `gorm:"column:moreshow_status" json:"moreshow_status"`
	MoreshowPos    int64       `gorm:"column:moreshow_pos" json:"moreshow_pos"`
	Check          int64       `gorm:"column:check" json:"check"`
	Person         string      `gorm:"column:person" json:"person"`
	Query          interface{} `gorm:"column:-" json:"query"`
	QueryStr       string      `gorm:"column:-" json:"-"`
	MoreshowValue  interface{} `gorm:"column:-" json:"moreshow_value"`
	PgcMediaID     []int32     `gorm:"column:-" json:"pgc_media_id"`
	ColorValue     string      `form:"color_value" json:"color_value"`
	Mtime          xtime.Time  `gorm:"column:mtime" json:"mtime"`
	Ctime          xtime.Time  `gorm:"column:ctime" json:"ctime"`
}

// SearchOgvMoreshow .
type SearchOgvMoreshow struct {
	ID      int64      `gorm:"column:id" json:"id"`
	Sid     int64      `gorm:"column:sid" json:"sid"`
	Word    string     `gorm:"column:word" json:"word"`
	Type    int        `gorm:"column:type" json:"type"`
	Value   string     `gorm:"column:value" json:"value"`
	Deleted int        `gorm:"column:deleted" json:"deleted"`
	Ctime   xtime.Time `gorm:"column:ctime" json:"ctime"`
	Mtime   xtime.Time `gorm:"column:mtime" json:"mtime"`
}

// SearchOgvQuery search web query
type SearchOgvQuery struct {
	ID      int64  `json:"id" form:"id"`
	SID     int64  `json:"sid" form:"sid" gorm:"column:sid"`
	Value   string `json:"value" form:"value"`
	Deleted int    `json:"deleted" form:"deleted"`
}

// TableName .
func (a SearchOgv) TableName() string {
	return "search_ogv"
}

/*
---------------------------
 struct param
---------------------------
*/

// SearchOgvAP add param
type SearchOgvAP struct {
	ID             int64      `gorm:"column:id" json:"id"`
	Query          string     `form:"query" gorm:"-" json:"query" validate:"required"`
	Color          int64      `form:"color" gorm:"column:color" json:"color" validate:"required"`
	Stime          xtime.Time `form:"stime" gorm:"column:stime" json:"stime" validate:"required"`
	Plat           string     `form:"plat" gorm:"column:plat" json:"plat" validate:"required"`
	HdCover        string     `form:"hd_cover" gorm:"column:hd_cover" json:"hd_cover" validate:"required"`
	HdBg           string     `form:"hd_bg" gorm:"column:hd_bg" json:"hd_bg"`
	HdTitle        string     `form:"hd_title" gorm:"column:hd_title" json:"hd_title" validate:"required"`
	HdSubtitle     string     `form:"hd_subtitle" gorm:"column:hd_subtitle" json:"hd_subtitle"`
	GameStatus     int64      `form:"game_status" gorm:"column:game_status" json:"game_status" validate:"required"`
	GamePos        int64      `form:"game_pos" gorm:"column:game_pos" json:"game_pos"`
	GameValue      string     `form:"game_value" gorm:"column:game_value" json:"game_value"`
	PgcPos         int64      `form:"pgc_pos" gorm:"column:pgc_pos" json:"pgc_pos" validate:"required"`
	PgcIds         string     `form:"pgc_ids" gorm:"column:pgc_ids" json:"pgc_ids" validate:"required"`
	PgcMoreURL     string     `form:"pgc_more_url" gorm:"column:pgc_more_url" json:"pgc_more_url"`
	PgcMoreType    int64      `form:"pgc_more_type" gorm:"column:pgc_more_type" json:"pgc_more_type"`
	MoreshowStatus int64      `form:"moreshow_status" gorm:"column:moreshow_status" json:"moreshow_status" validate:"required"`
	MoreshowPos    int64      `form:"moreshow_pos" gorm:"column:moreshow_pos" json:"moreshow_pos"`
	MoreshowValue  string     `form:"moreshow_value" gorm:"-" json:"moreshow_value"`
	ColorValue     string     `form:"color_value" json:"color_value"`
	Person         string     `gorm:"column:person" json:"person"`
	Ctime          xtime.Time `gorm:"column:ctime" json:"ctime"`
	Mtime          xtime.Time `gorm:"column:mtime" json:"mtime"`
}

// SearchOgvUP update param
type SearchOgvUP struct {
	ID             int64      `form:"id" validate:"required"`
	Query          string     `form:"query" gorm:"-" json:"query" validate:"required"`
	Color          int64      `form:"color" gorm:"column:color" json:"color" validate:"required"`
	Stime          xtime.Time `form:"stime" gorm:"column:stime" json:"stime" validate:"required"`
	Plat           string     `form:"plat" gorm:"column:plat" json:"plat" validate:"required"`
	HdCover        string     `form:"hd_cover" gorm:"column:hd_cover" json:"hd_cover" validate:"required"`
	HdBg           string     `form:"hd_bg" gorm:"column:hd_bg" json:"hd_bg"`
	HdTitle        string     `form:"hd_title" gorm:"column:hd_title" json:"hd_title" validate:"required"`
	HdSubtitle     string     `form:"hd_subtitle" gorm:"column:hd_subtitle" json:"hd_subtitle"`
	GameStatus     int64      `form:"game_status" gorm:"column:game_status" json:"game_status" validate:"required"`
	GamePos        int64      `form:"game_pos" gorm:"column:game_pos" json:"game_pos"`
	GameValue      string     `form:"game_value" gorm:"column:game_value" json:"game_value"`
	PgcPos         int64      `form:"pgc_pos" gorm:"column:pgc_pos" json:"pgc_pos" validate:"required"`
	PgcIds         string     `form:"pgc_ids" gorm:"column:pgc_ids" json:"pgc_ids" validate:"required"`
	PgcMoreURL     string     `form:"pgc_more_url" gorm:"column:pgc_more_url" json:"pgc_more_url"`
	PgcMoreType    int64      `form:"pgc_more_type" gorm:"column:pgc_more_type" json:"pgc_more_type"`
	MoreshowStatus int64      `form:"moreshow_status" gorm:"column:moreshow_status" json:"moreshow_status" validate:"required"`
	MoreshowPos    int64      `form:"moreshow_pos" gorm:"column:moreshow_pos" json:"moreshow_pos"`
	MoreshowValue  string     `form:"moreshow_value" gorm:"-" json:"moreshow_value"`
	ColorValue     string     `form:"color_value" json:"color_value"`
	Person         string     `gorm:"column:person" json:"person"`
	Check          int64      `gorm:"column:check" json:"check"`
}

// SearchOgvLP list param
type SearchOgvLP struct {
	ID     int    `form:"id"`
	Check  int    `form:"check"`
	Person string `form:"person"`
	Query  string `form:"query"`
	Title  string `form:"title"`
	Ps     int    `form:"ps" default:"20"`
	Pn     int    `form:"pn" default:"1"`
	Ts     int64  `form:"ts"`    //时间戳
	Stime  int64  `form:"stime"` //开始时间
	Etime  int64  `form:"etime"` //结束时间
}

// SearchOgvPager .
type SearchOgvPager struct {
	Item []*SearchOgv `json:"item"`
	Page common.Page  `json:"page"`
}

// SearchOgvOption option web card (online,hidden,pass,reject)
type SearchOgvOption struct {
	ID     int64  `form:"id" validate:"required"`
	Check  int    `form:"check" validate:"required"`
	Person string `gorm:"column:-" json:"person"`
}

// TableName .
func (a SearchOgvQuery) TableName() string {
	return "search_ogv_query"
}

// TableName .
func (a SearchOgvOption) TableName() string {
	return "search_ogv"
}

// TableName .
func (a SearchOgvAP) TableName() string {
	return "search_ogv"
}

// TableName .
func (a SearchOgvMoreshow) TableName() string {
	return "search_ogv_moreshow"
}

// TableName .
func (a SearchOgvUP) TableName() string {
	return "search_ogv"
}

// BatchAddOgvQuerySQL .
func BatchAddOgvQuerySQL(sID int64, data []*SearchOgvQuery) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, sID, v.Value)
	}
	return fmt.Sprintf(_ogvQuerySQL, strings.Join(rowStrings, ",")), param
}

// BatchAddOgvMoreShowSQL .
func BatchAddOgvMoreShowSQL(sID int64, data []*SearchOgvMoreshow) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?,?)")
		param = append(param, sID, v.Word, v.Type, v.Value)
	}
	return fmt.Sprintf(_ogvMoreShowSQL, strings.Join(rowStrings, ",")), param
}
