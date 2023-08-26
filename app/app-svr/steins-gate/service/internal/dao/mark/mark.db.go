package mark

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/pkg/errors"
)

const (
	_rawMarkSQL        = "SELECT mark FROM %s WHERE aid=? and mid=?"
	_rawMarkMSQL       = "SELECT aid,mark FROM %s WHERE mid=? AND aid IN (%s)"
	_addMarkSQL        = "INSERT INTO %s(aid,mid,mark) VALUE(?,?,?) ON DUPLICATE KEY UPDATE mark=VALUES(mark)"
	_rawEvaluationSQL  = "SELECT evaluation FROM arc_evaluation WHERE aid=?"
	_rawEvaluationsSQL = "SELECT aid,evaluation FROM arc_evaluation WHERE aid IN (%s)"
)

func markTableName(aid int64) string {
	return fmt.Sprintf("game_evaluation_%02d", aid%50)
}

func markTableMName(mid int64) string {
	return fmt.Sprintf("game_evaluation_m_%02d", mid%50)
}

// RawMark get mark by aid and mid.
func (d *Dao) rawMark(c context.Context, aid, mid int64) (res int64, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(_rawMarkSQL, markTableName(aid)), aid, mid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	return
}

// RawMark get mark by aid and mid.
func (d *Dao) rawMarksM(c context.Context, aids []int64, mid int64) (res map[int64]int64, err error) {
	query := fmt.Sprintf(_rawMarkMSQL, markTableMName(mid), xstr.JoinInts(aids))
	rows, err := d.db.Query(c, query, mid)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", _rawMarkMSQL, err)
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		var (
			mark    int64
			aidTemp int64
		)
		if err = rows.Scan(&mark, &aidTemp); err != nil {
			err = errors.Wrapf(err, "aids %v", aids)
			return
		}
		res[aidTemp] = mark
	}
	err = rows.Err()
	return
}

// AddMark adds a new mark
func (d *Dao) addMark(c context.Context, aid, mid, mark int64) (err error) {
	var tx *sql.Tx
	if tx, err = d.db.Begin(c); err != nil || tx == nil {
		log.Error("db: begintran BeginTran d.db.Begin error(%v) aid = %d, mid = %d", err, aid, mid)
		return
	}
	if _, err = tx.Exec(fmt.Sprintf(_addMarkSQL, markTableName(aid)), aid, mid, mark); err != nil {
		//nolint:errcheck
		tx.Rollback()
		err = errors.Wrapf(err, "d.db.Exec(%s) error(%v) aid = %d, mid = %d", _addMarkSQL, err, aid, mid)
		return
	}
	if _, err = tx.Exec(fmt.Sprintf(_addMarkSQL, markTableMName(mid)), aid, mid, mark); err != nil {
		//nolint:errcheck
		tx.Rollback()
		err = errors.Wrapf(err, "d.db.ExecM(%s) error(%v) aid = %d, mid = %d", _addMarkSQL, err, aid, mid)
		return
	}
	err = tx.Commit()
	return
}

// RawMark get mark by aid and mid.
func (d *Dao) rawEvaluation(c context.Context, aid int64) (res int64, err error) {
	row := d.db.QueryRow(c, _rawEvaluationSQL, aid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "aid %d", aid)
		}
	}
	return
}

func (d *Dao) rawEvaluations(c context.Context, ids []int64) (res map[int64]int64, err error) {
	query := fmt.Sprintf(_rawEvaluationsSQL, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, query)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", query, err)
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		var (
			aid        int64
			evaluation int64
		)
		if err = rows.Scan(&aid, &evaluation); err != nil {
			err = errors.Wrapf(err, "edgeIDs %v", ids)
			return
		}
		res[aid] = evaluation
	}
	err = rows.Err()
	return

}
