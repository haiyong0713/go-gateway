package search

import (
	xtime "go-common/library/time"
)

type BrandBlacklist struct {
	Id      int64      `gorm:"primary_key" json:"id" form:"id"`
	Desc    string     `json:"desc" form:"desc"`
	State   int32      `json:"state" form:"state"`
	CUname  string     `json:"c_uname" form:"c_uname"`
	MUname  string     `json:"m_uname" form:"m_uname"`
	Ctime   xtime.Time `json:"c_time" form:"ctime"`
	Mtime   xtime.Time `json:"m_time" form:"mtime"`
	Deleted int32      `json:"-" form:"-"`
}

func (l BrandBlacklist) TableName() string {
	return "search_brand_blacklist"
}

type BrandBlacklistQuery struct {
	Id          int64      `gorm:"primary_key" json:"id" form:"id"`
	BlacklistId int64      `json:"blacklist_id" form:"blacklist_id"`
	Query       string     `json:"query" form:"query"`
	State       int32      `json:"state" form:"state"`
	CUname      string     `json:"c_uname" form:"c_uname"`
	MUname      string     `json:"m_uname" form:"m_uname"`
	Ctime       xtime.Time `json:"c_time" form:"ctime"`
	Mtime       xtime.Time `json:"m_time" form:"mtime"`
	Deleted     int32      `json:"-" form:"-"`
}

func (q BrandBlacklistQuery) TableName() string {
	return "search_brand_blacklist_query"
}

type BrandBlacklistAddReq struct {
	Username  string `json:"username" form:"username"`
	Uid       int64  `json:"uid" form:"uid"`
	QueryList string `json:"query_list" form:"query_list" validate:"required"`
	Desc      string `json:"desc" form:"desc"`
}

type BrandBlacklistAddResp struct {
	BlacklistId  int64    `json:"blacklist_id" form:"blacklist_id"`
	EnabledQuery []string `json:"enabled_query" form:"enabled_query"`
}

type BrandBlacklistEditReq struct {
	Username    string `json:"username" form:"username"`
	Uid         int64  `json:"uid" form:"uid"`
	BlacklistId int64  `json:"blacklist_id" form:"blacklist_id" validate:"required"`
	QueryList   string `json:"query_list" form:"query_list" validate:"required"`
	Desc        string `json:"desc" form:"desc"`
}

type BrandBlacklistEditResp struct {
	EnabledQuery []string `json:"enabled_query" form:"enabled_query"`
}

type BrandBlacklistOptionReq struct {
	Username    string `json:"username" form:"username"`
	Uid         int64  `json:"uid" form:"uid"`
	BlacklistId int64  `json:"blacklist_id" form:"blacklist_id" validate:"required"`
	Option      int    `json:"option" form:"option" validate:"required"`
}

type BrandBlacklistOptionResp struct {
	EnabledQuery []string `json:"enabled_query" form:"enabled_query"`
}

type BrandBlacklistListReq struct {
	Username string `json:"username" form:"username"`
	Uid      int64  `json:"uid" form:"uid"`
	Pn       int    `json:"pn" form:"pn" default:"1"`
	Ps       int    `json:"ps" form:"ps" default:"20"`
	State    int    `json:"state" form:"state"`
	Keyword  string `json:"keyword" form:"keyword"`
	Order    int    `json:"order" form:"order"`
}

type BrandBlackListListResp struct {
	Page *Page                 `json:"page" form:"page"`
	List []*BrandBlacklistItem `json:"list" form:"list"`
}

type BrandBlacklistItem struct {
	BlacklistId int64      `json:"blacklist_id" form:"blacklist_id"`
	QueryList   []string   `json:"query_list" form:"query_list"`
	Desc        string     `json:"desc" form:"desc"`
	State       int32      `json:"state" form:"state"`
	CUname      string     `json:"c_uname" form:"c_uname"`
	MUname      string     `json:"m_uname" form:"m_uname"`
	Ctime       xtime.Time `json:"c_time" form:"ctime"`
	Mtime       xtime.Time `json:"m_time" form:"mtime"`
}

type Page struct {
	Pn    int `json:"pn" form:"pn"`
	Ps    int `json:"ps" form:"ps"`
	Total int `json:"total" form:"total"`
}
