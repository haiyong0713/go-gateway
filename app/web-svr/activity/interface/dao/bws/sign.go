package bws

import (
	"context"
	"fmt"
	"go-common/library/xstr"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_signsSQL        = "SELECT `id` FROM act_bws_point_sign WHERE is_delete = 0 AND pid = ? ORDER BY ID"
	_bwsSignsSQL     = "SELECT id,bid,pid,stime,etime,state,points,provide_points,sign_points,is_delete,ctime,mtime FROM act_bws_point_sign WHERE id in (%s) AND  is_delete = 0"
	_bwsIncrPointSQL = "UPDATE `act_bws_point_sign` SET `provide_points` = `provide_points` + `sign_points` WHERE `id` = ? AND `points` >= `provide_points`"
)

// RawSigns .
func (d *Dao) RawSigns(c context.Context, pid int64) (ids []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _signsSQL, pid); err != nil {
		log.Error("RawSigns: db.Exec(%d) error(%v)", pid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.PointSign)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("RawSigns:row.Scan() error(%v)", err)
			return
		}
		ids = append(ids, r.ID)
	}
	err = rows.Err()
	return
}

// RawBwsSign .
func (d *Dao) RawBwsSign(c context.Context, ids []int64) (rs map[int64]*bwsmdl.PointSign, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bwsSignsSQL, xstr.JoinInts(ids))); err != nil {
		log.Error("RawBwsSign: db.Exec(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	rs = make(map[int64]*bwsmdl.PointSign, len(ids))
	for rows.Next() {
		r := new(bwsmdl.PointSign)
		if err = rows.Scan(&r.ID, &r.Bid, &r.Pid, &r.Stime, &r.Etime, &r.State, &r.Points, &r.ProvidePoints, &r.SignPoints, &r.IsDelete, &r.Ctime, &r.Mtime); err != nil {
			log.Error("RawBwsSign:row.Scan() error(%v)", err)
			return
		}
		rs[r.ID] = r
	}
	err = rows.Err()
	return
}

// IncrSignPoint .
func (d *Dao) IncrSignPoint(c context.Context, id int64) (err error) {
	if _, err = d.db.Exec(c, _bwsIncrPointSQL, id); err != nil {
		log.Error("IncrSignPoint error d.db.Exec(%d) error(%v)", id, err)
	}
	return
}
