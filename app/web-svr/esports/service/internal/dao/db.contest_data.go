package dao

import (
	"fmt"
	"go-gateway/app/web-svr/esports/service/internal/model"
	"strings"
)

const (
	/** 枚举 **/
	/** ** 软删状态 ** **/
	ContestDataRecordDeleted    = 1
	ContestDataRecordNotDeleted = 0

	/** sql **/
	_cDataInsertSQL = "INSERT INTO es_contests_data(cid,url,point_data, av_cid) VALUES %s"

	contestDataTableName = "es_contests_data"
)

func (d *dao) batchAddCDataSQL(cID int64, data []*model.ContestDataModel) (sql string, param []interface{}) {
	if len(data) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, "(?,?,?, ?)")
		param = append(param, cID, v.Url, v.PointData, v.AvCid)
	}
	return fmt.Sprintf(_cDataInsertSQL, strings.Join(rowStrings, ",")), param
}
