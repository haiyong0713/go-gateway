package model

import (
	"fmt"
	"strings"
)

const _teamMapInsertSQL = "INSERT INTO es_teams_map(tid,aid) VALUES %s"

// TeamMap .
type TeamMap struct {
	ID        int64 `json:"id"`
	Tid       int64 `json:"tid"`
	Aid       int64 `json:"aid"`
	IsDeleted int   `json:"is_deleted"`
}

// TableName es_teams_map.
func (t TeamMap) TableName() string {
	return "es_teams_map"
}

// BatchAddTeamMapSQL .
func BatchAddTeamMapSQL(data []*TeamMap) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, v.Tid, v.Aid)
	}
	return fmt.Sprintf(_teamMapInsertSQL, strings.Join(rowStrings, ",")), param
}
