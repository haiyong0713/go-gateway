package bnj

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

const (
	_upAmountSQL = "UPDATE currency_user SET amount = ? WHERE mid = ?"
	_amountSQL   = "SELECT amount FROM currency_user WHERE mid = ?"
)

// RawCurrencyAmount get currency user data.
func (d *Dao) RawCurrencyAmount(c context.Context, mid int64) (amount int64, err error) {
	row := d.db.QueryRow(c, _amountSQL, mid)
	if err = row.Scan(&amount); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawCurrencyUser:QueryRow")
		}
	}
	return
}

// UpCurrencyAmount get currency user data.
func (d *Dao) UpCurrencyAmount(c context.Context, amount, mid int64) (err error) {
	if _, err = d.db.Exec(c, _upAmountSQL, amount, mid); err != nil {
		err = errors.Wrap(err, "UpCurrencyAmount:Exec")
	}
	return
}
