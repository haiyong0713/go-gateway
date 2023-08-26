package show

import (
	"fmt"
	"strings"

	xtime "go-common/library/time"
)

const (
	_shieldInsertSQL = "INSERT INTO search_shield_query(sid,value) VALUES %s"
	_shieldEditSQL   = "UPDATE search_shield_query SET value = CASE %s END WHERE id IN (%s)"
)

// SearchShieldQuery search web query
type SearchShieldQuery struct {
	ID      int64      `json:"id" form:"id"`
	SID     int64      `json:"sid" form:"sid" gorm:"column:sid"`
	Value   string     `json:"value" form:"value"`
	Deleted int        `json:"deleted" form:"deleted"`
	Mtime   xtime.Time `json:"mtime"`
}

// TableName .
func (a SearchShieldQuery) TableName() string {
	return "search_shield_query"
}

// BatchAddShieldSQL .
func BatchAddShieldSQL(sID int64, data []*SearchShieldQuery) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, sID, v.Value)
	}
	return fmt.Sprintf(_shieldInsertSQL, strings.Join(rowStrings, ",")), param
}

// BatchEditShieldSQL .
func BatchEditShieldSQL(querys []*SearchShieldQuery) (sql string, params []interface{}) {
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
	return fmt.Sprintf(_shieldEditSQL, sql, strings.Join(idSql, ",")), params
}
