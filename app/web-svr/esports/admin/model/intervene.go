package model

import (
	"fmt"
	"strings"
)

const (
	_dSearchInsertSQL = "INSERT INTO es_search_contest(mid,cid) VALUES %s"
)

// EsSearchCard search card main.
type EsSearchCard struct {
	ID        int64  `json:"id" form:"id"`
	QueryName string `json:"query_name" form:"query_name" validate:"required"`
	Stime     int64  `json:"stime" form:"stime" validate:"required"`
	Etime     int64  `json:"etime" form:"etime" validate:"required"`
	Status    int64  `json:"status" form:"status"`
	Mtime     string `json:"-" form:"mtime"`
	Detail    string `json:"-" form:"detail" gorm:"-" validate:"required"`
}

// EsSearchContest search contest.
type EsSearchContest struct {
	ID          int64  `json:"id"`
	Mid         int64  `json:"mid" form:"mid" validate:"required"`
	Cid         int64  `json:"cid" form:"cid" validate:"required"`
	IsDeleted   int    `json:"is_deleted" form:"is_deleted"`
	SeasonName  string `json:"season_name" form:"season_name" gorm:"-"`
	ContestName string `json:"contest_name" form:"contest_name" gorm:"-"`
}

// SearchInfo .
type SearchInfo struct {
	*EsSearchCard
	Detail []*EsSearchContest `json:"detail"`
}

// BatchAddDSearchSQL .
func BatchAddDSearchSQL(mainID int64, data []*EsSearchContest) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, mainID, v.Cid)
	}
	return fmt.Sprintf(_dSearchInsertSQL, strings.Join(rowStrings, ",")), param
}

// TableName .
func (s EsSearchCard) TableName() string {
	return "es_search_card"
}

// TableName .
func (s EsSearchContest) TableName() string {
	return "es_search_contest"
}
