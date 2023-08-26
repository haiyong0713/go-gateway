package bws

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_bwsFieldsSQL = "SELECT `id`,`name`,`area`,`ctime`,`mtime`,`bid`,`image`,`del` FROM `act_bws_fields` WHERE `bid` = ? AND `del` = 0"
)

// RawActFields .
func (d *Dao) RawActFields(c context.Context, bid int64) (rs *bwsmdl.ActFields, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _bwsFieldsSQL, bid); err != nil {
		log.Error("d.db.Query(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	rs = &bwsmdl.ActFields{}
	for rows.Next() {
		t := &bwsmdl.ActField{}
		if err = rows.Scan(&t.ID, &t.Name, &t.Area, &t.Ctime, &t.Mtime, &t.Bid, &t.Image, &t.Del); err != nil {
			log.Error("rows.Scan(%d) error(%v)", bid, err)
			return
		}
		rs.ActField = append(rs.ActField, t)
	}
	err = rows.Err()
	return
}
