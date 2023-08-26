package currency

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/currency"

	"github.com/pkg/errors"
)

const (
	_currencySQL     = "SELECT id,`name`,unit FROM currency WHERE id = ? AND state = 1"
	_relationSQL     = "SELECT id,currency_id,business_id,foreign_id FROM currency_relation WHERE business_id = ? AND foreign_id = ? AND is_deleted = 0"
	_userSQL         = "SELECT id,mid,amount FROM currency_user_%d WHERE mid = ?"
	_currencySumSQL  = "SELECT sum(`amount`) from currency_user_%d "
	_userLogSQL      = "SELECT id,from_mid,to_mid,change_amount,remark,ctime FROM currency_user_log_%d WHERE to_mid = ? ORDER BY mtime DESC LIMIT 50"
	_userLogAddSQL   = "INSERT INTO currency_user_log_%d (from_mid,to_mid,change_amount,remark) VALUES (?,?,?,?)"
	_reduceAmountSQL = "INSERT INTO currency_user_%d (mid,amount) VALUES(?,?) ON DUPLICATE KEY UPDATE amount = amount-?"
	_addAmountSQL    = "INSERT INTO currency_user_%d (mid,amount) VALUES(?,?) ON DUPLICATE KEY UPDATE amount = amount+?"
)

// RawCurrency get currency data form database.
func (d *Dao) RawCurrency(c context.Context, id int64) (data *currency.Currency, err error) {
	data = new(currency.Currency)
	row := d.db.QueryRow(c, _currencySQL, id)
	if err = row.Scan(&data.ID, &data.Name, &data.Unit); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawCurrency:QueryRow")
		}
	}
	return
}

// RawRelation get relation currency ids.
func (d *Dao) RawRelation(c context.Context, businessID, foreignID int64) (data *currency.CurrencyRelation, err error) {
	data = new(currency.CurrencyRelation)
	row := d.db.QueryRow(c, _relationSQL, businessID, foreignID)
	if err = row.Scan(&data.ID, &data.CurrencyID, &data.BusinessID, &data.ForeignID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawRelation:QueryRow")
		}
	}
	return
}

// RawCurrencyUser get currency user data.
func (d *Dao) RawCurrencyUser(c context.Context, mid, id int64) (data *currency.CurrencyUser, err error) {
	data = new(currency.CurrencyUser)
	row := d.db.QueryRow(c, fmt.Sprintf(_userSQL, id), mid)
	if err = row.Scan(&data.ID, &data.Mid, &data.Amount); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawCurrencyUser:QueryRow")
		}
	}
	return
}

// CurrencySum .
func (d *Dao) CurrencySum(c context.Context, id int64) (amount int64, err error) {
	var data sql.NullInt64
	row := d.db.QueryRow(c, fmt.Sprintf(_currencySumSQL, id))
	if err = row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CurrencySum:QueryRow")
		}
		return
	}
	amount = data.Int64
	return
}

// RawCurrencyUserLog get currency user log.
func (d *Dao) RawCurrencyUserLog(c context.Context, mid, id int64) (list []*currency.CurrencyUserLog, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_userLogSQL, id), mid)
	if err != nil {
		log.Error("RawCurrencyUserLog:d.db.Query(%d,%d) error(%v)", mid, id, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(currency.CurrencyUserLog)
		if err = rows.Scan(&n.ID, &n.FromMid, &n.ToMid, &n.ChangeAmount, &n.Remark, &n.Ctime); err != nil {
			log.Error("RawCurrencyUserLog:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawCurrencyUserLog:rows.Err() error(%v)", err)
	}
	return
}

// UpUserAmount update user amount.
func (d *Dao) UpUserAmount(c context.Context, id, fromMid, toMid, amount int64, remark string) (err error) {
	var tx *xsql.Tx
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("UpUserAmount d.db.Begin error(%v)", err)
		return
	}
	if fromMid > 0 {
		if _, err = tx.Exec(fmt.Sprintf(_reduceAmountSQL, id), fromMid, -amount, amount); err != nil {
			log.Error("UpUserAmount tx.Exec reduce amount(%d) fromMid(%d) error(%v)", amount, fromMid, err)
			tx.Rollback()
			return
		}
	}
	if toMid > 0 {
		if _, err = tx.Exec(fmt.Sprintf(_addAmountSQL, id), toMid, amount, amount); err != nil {
			log.Error("UpUserAmount tx.Exec add amount(%d) toMid(%d) error(%v)", amount, toMid, err)
			tx.Rollback()
			return
		}
	}
	// add log
	if _, err = tx.Exec(fmt.Sprintf(_userLogAddSQL, id), fromMid, toMid, amount, remark); err != nil {
		log.Error("UpUserAmount tx.Exec log fromMid(%d) toMid(%d) amount(%d) remark(%s) error(%v)", fromMid, toMid, amount, remark, err)
		tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("UpUserAmount tx.Exec add amount(%d) toMid(%d) error(%v)", amount, toMid, err)
	}
	return
}
