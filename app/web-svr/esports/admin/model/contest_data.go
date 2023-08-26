package model

import (
	"fmt"
	"strings"
)

const (
	_cDataInsertSQL = "INSERT INTO es_contests_data(cid,url,point_data, av_cid) VALUES %s"
	_cDataEditSQL   = "UPDATE es_contests_data SET url = CASE %s END,point_data = CASE %s END, av_cid = CASE %s END WHERE id IN (%s)"
)

// ContestData .
type ContestData struct {
	ID        int64  `json:"id"`
	CID       int64  `json:"cid" gorm:"column:cid"`
	URL       string `json:"url"`
	PointData int64  `json:"point_data"`
	IsDeleted int    `json:"is_deleted"`
	AvCID     int64  `json:"av_cid" gorm:"column:av_cid"`
}

// TableName es_contests_data.
func (t ContestData) TableName() string {
	return "es_contests_data"
}

// BatchAddCDataSQL .
func BatchAddCDataSQL(cID int64, data []*ContestData) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?, ?)")
		param = append(param, cID, v.URL, v.PointData, v.AvCID)
	}
	return fmt.Sprintf(_cDataInsertSQL, strings.Join(rowStrings, ",")), param
}

// BatchEditCDataSQL .
func BatchEditCDataSQL(cDatas []*ContestData) (sql string, param []interface{}) {
	if len(cDatas) == 0 {
		return "", []interface{}{}
	}
	var (
		urlStr, pDataStr, avCIDStr string
		ids                        []interface{}
		idSql                      []string
	)
	for _, module := range cDatas {
		urlStr += " WHEN id = ? THEN ?"
		param = append(param, module.ID, module.URL)
		idSql = append(idSql, "?")
		ids = append(ids, module.ID)
	}
	for _, module := range cDatas {
		pDataStr += " WHEN id = ? THEN ?"
		param = append(param, module.ID, module.PointData)
	}
	for _, data := range cDatas {
		avCIDStr += " WHEN id = ? THEN ?"
		param = append(param, data.ID, data.AvCID)
	}
	param = append(param, ids...)
	return fmt.Sprintf(_cDataEditSQL, urlStr, pDataStr, avCIDStr, strings.Join(idSql, ",")), param
}
