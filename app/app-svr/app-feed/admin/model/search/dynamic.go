package search

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// DySeach dynamic public search
type DySeach struct {
	ID        int64      `json:"id"`
	Word      string     `json:"word"`
	Position  int64      `json:"position"`
	Uname     string     `json:"uname"`
	Uid       int64      `json:"uid"`
	IsDeleted int64      `json:"is_deleted"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
}

// DynamicSearchPager  dynamic public search
type DySeaPager struct {
	Item []*DySeach  `json:"item"`
	Page common.Page `json:"page"`
}

// DySeach dynamic public search
func (a DySeach) TableName() string {
	return "dynamic_search"
}

/*
---------------------------
 struct param
---------------------------
*/

// DySeachAP add param
type DySeachAP struct {
	ID       int64  `form:"id"`
	Word     string `form:"word" validate:"required"`
	Position int64  `form:"position" validate:"required"`
	Uname    string `form:"uname"`
	Uid      int64  `form:"uid"`
}

// DySeachUP update param
type DySeachUP struct {
	ID       int64 `form:"id" validate:"required"`
	Position int64 `form:"position" validate:"required"`
}

// DySeachLP list param
type DySeachLP struct {
	Word string `form:"word"`
	Ps   int    `form:"ps" default:"20"`
	Pn   int    `form:"pn" default:"1"`
}

// DySeachDel del param
type DySeachDel struct {
	ID int64 `form:"id" validate:"required"`
}

// TableName .
func (a DySeachAP) TableName() string {
	return "dynamic_search"
}

// TableName .
func (a DySeachUP) TableName() string {
	return "dynamic_search"
}
