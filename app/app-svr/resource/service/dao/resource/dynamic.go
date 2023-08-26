package resource

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_dySearSQL = "SELECT id,word,position FROM dynamic_search WHERE is_deleted = 0 ORDER BY position ASC"
)

// DySearch dynamic public search
func (d *Dao) DySearch(c context.Context) (res []*model.DySeach, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _dySearSQL); err != nil {
		log.Error("DySearch:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.DySeach)
		if err = rows.Scan(&r.ID, &r.Word, &r.Position); err != nil {
			log.Error("DySearch:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("DySearch.rows.Err() error(%v)", err)
	}
	return
}
