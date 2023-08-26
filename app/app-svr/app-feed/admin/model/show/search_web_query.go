package show

import (
	"fmt"
	"strings"
)

const (
	_queryInsertSQL = "INSERT INTO search_web_query(sid,value) VALUES %s"
	_queryEditSQL   = "UPDATE search_web_query SET value = CASE %s END WHERE id IN (%s)"
	_platInsertSQL  = "INSERT INTO search_web_plat(sid,plat,conditions,build) VALUES %s"
)

// SearchWebQuery search web query
type SearchWebQuery struct {
	ID      int64  `json:"id" form:"id"`
	SID     int64  `json:"sid" form:"sid" gorm:"column:sid"`
	Value   string `json:"value" form:"value"`
	Deleted int    `json:"deleted" form:"deleted"`
}

// TableName .
func (a SearchWebQuery) TableName() string {
	return "search_web_query"
}

// BatchAddQuerySQL .
func BatchAddQuerySQL(sID int64, data []*SearchWebQuery) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, sID, v.Value)
	}
	return fmt.Sprintf(_queryInsertSQL, strings.Join(rowStrings, ",")), param
}

// BatchEditQuerySQL .
func BatchEditQuerySQL(querys []*SearchWebQuery) (sql string, params []interface{}) {
	if len(querys) == 0 {
		return "", []interface{}{}
	}
	var (
		ids   []interface{}
		idSql []string
	)
	for _, query := range querys {
		sql = sql + " WHEN id = ? THEN ? "
		params = append(params, query.ID, query.Value)
		idSql = append(idSql, "?")
		ids = append(ids, query.ID)
	}
	params = append(params, ids...)
	return fmt.Sprintf(_queryEditSQL, sql, strings.Join(idSql, ",")), params
}

// BatchAddPlatSQL .
func BatchAddPlatSQL(sID int64, data []*SearchWebPlat) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?,?)")
		param = append(param, sID, v.Plat, v.Conditions, v.Build)
	}
	return fmt.Sprintf(_platInsertSQL, strings.Join(rowStrings, ",")), param
}
