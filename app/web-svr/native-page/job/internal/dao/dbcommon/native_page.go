package dbcommon

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	actGRPC "go-gateway/app/web-svr/native-page/interface/api"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/job/internal/model"

	"github.com/pkg/errors"
)

const (
	_newNatPagesSQL = "SELECT `id`,`title`,`type`,`foreign_id`,`stime`,`creator`,`ctime`,`state`,`from_type`,`related_uid` FROM `native_page` WHERE id>? ORDER BY `id` ASC LIMIT ?"
	_upStateSQL     = "UPDATE `native_page` SET `state` = ? WHERE `id` = ? AND`state` = ?"
	_searchSQL      = "SELECT `id`,`foreign_id`,`type`,`stime` FROM `native_page` WHERE `stime` > ? AND `stime` <= ? AND `state` = ? LIMIT 50"
	_foreignSQL     = "SELECT `id` FROM `native_page` WHERE `foreign_id` = ? AND `type` = ? AND `state` = ?"
	_endPageSQL     = "SELECT `id`,`etime`,`related_uid`,`from_type` FROM `native_page` WHERE `etime` < ? AND `etime` > ? AND `state` = ? LIMIT 50"
	_offLineSQL     = "UPDATE `native_page` SET `state` = ?, `off_reason`=? WHERE `id` IN (%s) AND`state` = ?"
)

// ForeignFromIDs .
func (d *Dao) ForeignFromIDs(c context.Context, fid, ftype int64) (ids []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _foreignSQL, fid, ftype, model.PageOnLine); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &actGRPC.NativePage{}
		if err = rows.Scan(&a.ID); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		ids = append(ids, a.ID)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return

}

// EndList .
func (d *Dao) EndList(c context.Context) (list []*actGRPC.NativePage, err error) {
	var rows *xsql.Rows
	now := time.Now().Format("2006-01-02 15:04:05")
	if rows, err = d.db.Query(c, _endPageSQL, now, "0000-00-00 00:00:00", model.PageOnLine); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &actGRPC.NativePage{}
		if err = rows.Scan(&a.ID, &a.Etime, &a.RelatedUid, &a.FromType); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list = append(list, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

// OffLinePage .
func (d *Dao) OffLinePage(c context.Context, ids []int64, offReason string) (effID int64, err error) {
	var (
		res    sql.Result
		sqlStr []string
		params []interface{}
	)
	params = append(params, model.PageOffLine, offReason)
	for _, v := range ids {
		sqlStr = append(sqlStr, "?")
		params = append(params, v)
	}
	params = append(params, model.PageOnLine)
	if res, err = d.db.Exec(c, fmt.Sprintf(_offLineSQL, strings.Join(sqlStr, ",")), params...); err != nil {
		err = errors.Wrap(err, "OffLinePage d.db.Exec")
		return
	}
	return res.RowsAffected()
}

// SearchPage .
func (d *Dao) SearchPage(c context.Context) (list []*model.NatPage, err error) {
	var rows *xsql.Rows
	now := time.Now().Format("2006-01-02 15:04:05")
	if rows, err = d.db.Query(c, _searchSQL, "0000-00-00 00:00:00", now, model.PageWaitOnLine); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &model.NatPage{}
		if err = rows.Scan(&a.ID, &a.ForeignID, &a.Type, &a.Stime); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list = append(list, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

// UpPage .
func (d *Dao) UpPage(c context.Context, id, state int64) (effID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _upStateSQL, state, id, model.PageWaitOnLine); err != nil {
		err = errors.Wrap(err, "d.db.Exec")
		return
	}
	return res.RowsAffected()
}

func (d *Dao) pagingNewNatPages(c context.Context, id, limit int64) ([]*actGRPC.NativePage, error) {
	rows, err := d.db.Query(c, _newNatPagesSQL, id, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*actGRPC.NativePage{}, nil
		}
		log.Errorc(c, "Fail to query pagingNewNatPages, sql=%s id=%d error=%+v", _newNatPagesSQL, id, err)
		return nil, err
	}
	defer rows.Close()
	list := make([]*actGRPC.NativePage, 0)
	for rows.Next() {
		t := &actGRPC.NativePage{}
		err = rows.Scan(&t.ID, &t.Title, &t.Type, &t.ForeignID, &t.Stime, &t.Creator, &t.Ctime, &t.State, &t.FromType, &t.RelatedUid)
		if err != nil {
			log.Errorc(c, "Fail to scan nativePage row, error=%+v", err)
			continue
		}
		list = append(list, t)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get nativePages rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}
