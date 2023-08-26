package bnj

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

const (
	_amountSQL = "SELECT amount FROM currency_user WHERE mid = ?"
)

// RawCurrencyAmount get currency user data.
func (d *Dao) RawCurrencyAmount(c context.Context, mid int64) (amount int64, err error) {
	row := d.db.QueryRow(c, _amountSQL, mid)
	if err = row.Scan(&amount); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawCurrencyAmount:QueryRow")
		}
	}
	return
}
