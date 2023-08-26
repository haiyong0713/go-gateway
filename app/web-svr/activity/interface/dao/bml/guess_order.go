package bml

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bml"
	"strings"
)

const (
	_AddGuessOrderRecord    = "INSERT INTO act_bml_2021_guess_order (mid,guess_type,reward_id,order_no,guess_answer ,state) VALUES(?,?,?,?,?,?)"
	_GuessOrderListByMid    = "SELECT id, mid , guess_type , reward_id, state,order_no , guess_answer , ctime , mtime FROM act_bml_2021_guess_order WHERE mid = ? "
	_UpdateGuessOrderRecord = "UPDATE act_bml_2021_guess_order SET state=1 WHERE id=? "
)

func (d *Dao) AddGuessOrderRecord(ctx context.Context, record *bml.GuessOrderRecord) (int64, error) {
	row, err := d.db.Exec(ctx, _AddGuessOrderRecord, record.Mid, record.GuessType, record.RewardId, record.OrderNo, record.GuessAnswer, record.State)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, ecode.BMLGuessRepeateDrawError
		}
		return 0, errors.Wrap(err, "AddGuessOrderRecord err")
	}
	return row.LastInsertId()
}

func (d *Dao) RawGuessOrderListByMid(ctx context.Context, mid int64) (res []*bml.GuessOrderRecord, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, _GuessOrderListByMid, mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawGuessOrderListByMid Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bml.GuessOrderRecord)
		if err = rows.Scan(&r.Id, &r.Mid, &r.GuessType, &r.RewardId, &r.State, &r.OrderNo, &r.GuessAnswer, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawGuessOrderListByMid Scan err")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawGuessOrderListByMid rows")
	}
	return res, nil
}

func (d *Dao) UpdateGuessOrderById(ctx context.Context, id int64) (int64, error) {
	row, err := d.db.Exec(ctx, _UpdateGuessOrderRecord, id)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserAward")
	}
	return row.RowsAffected()
}
