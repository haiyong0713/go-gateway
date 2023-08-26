package like

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_upSubjectStateSQL    = "UPDATE `subject_stat` SET num = num + ? WHERE sid = ? LIMIT 1"
	_addSubjectStatSQL    = "INSERT INTO `subject_stat`(sid,num) VALUES(?,?) ON DUPLICATE KEY UPDATE num = num + ?"
	_querySubjectStateSQL = "SELECT `num`,`sid` FROM `subject_stat` WHERE `sid` IN (%s)"
)

// RawReservesTotal .
func (d *Dao) RawReservesTotal(c context.Context, sids []int64) (rly map[int64]int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(_querySubjectStateSQL, xstr.JoinInts(sids))); err != nil {
		log.Error("RawSubjectStats:d.db.Query(%v) error(%v)", sids, err)
		return
	}
	defer rows.Close()
	rly = make(map[int64]int64)
	for _, sid := range sids {
		rly[sid] = 0
	}
	for rows.Next() {
		r := &lmdl.SubStat{}
		if err = rows.Scan(&r.Num, &r.Sid); err != nil {
			log.Error("RawSubjectStats:d.db.Scan() error(%v)", err)
			return
		}
		rly[r.Sid] = r.Num
	}
	err = rows.Err()
	return
}

// IncrSubjectStat .
func (d *Dao) IncrSubjectStat(c context.Context, sid int64, num int32) (err error) {
	var (
		res      sql.Result
		affected int64
	)
	if res, err = d.db.Exec(c, _upSubjectStateSQL, num, sid); err != nil {
		log.Errorc(c, "IncrSubjectStat _upSubjectStateSQL %d,%d error(%v)", num, sid, err)
		return
	}
	if affected, err = res.RowsAffected(); err != nil {
		log.Errorc(c, "IncrSubjectStat res.RowsAffected() %d,%d error(%v)", num, sid, err)
		return
	}
	if affected > 0 {
		return
	}
	// 第一次插入数据
	_, err = d.db.Exec(c, _addSubjectStatSQL, sid, num, num)
	if err != nil {
		log.Errorc(c, "IncrSubjectStat _addSubjectStatSQL %d,%d error(%v)", num, sid, err)
	}
	return
}
