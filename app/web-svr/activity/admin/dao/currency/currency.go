package currency

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/currency"
)

const (
	_saveCurrSQL      = "UPDATE currency SET `name` = ?,unit = ?,state = ? WHERE id = ?"
	_upCurrRelaSQL    = "UPDATE currency_relation SET `is_deleted` = ? WHERE id = ?"
	_userCreateSQL    = "CREATE TABLE IF NOT EXISTS currency_user_%d LIKE currency_user"
	_userLogCreateSQL = "CREATE TABLE IF NOT EXISTS currency_user_log_%d LIKE currency_user_log"
)

// SaveCurrency save currency.
func (d *Dao) SaveCurrency(c context.Context, arg *currency.SaveArg) (err error) {
	if err = d.DB.Model(&currency.Currency{}).Exec(_saveCurrSQL, arg.Name, arg.Unit, arg.State, arg.ID).Error; err != nil {
		log.Error("SaveCurrency Update(%+v) error(%v)", arg, err)
	}
	return
}

// SaveCurrRelation save currency relation.
func (d *Dao) SaveCurrRelation(c context.Context, id int64, isDeleted int) (err error) {
	if err = d.DB.Model(&currency.Relation{}).Exec(_upCurrRelaSQL, isDeleted, id).Error; err != nil {
		log.Error("SaveCurrRelation id(%d) isDeleted(%d) error(%v)", id, isDeleted, err)
	}
	return
}

// UserCreate create user table.
func (d *Dao) UserCreate(c context.Context, id int64) (err error) {
	if err = d.DB.Exec(fmt.Sprintf(_userCreateSQL, id)).Error; err != nil {
		log.Error("UserCreate id(%d) error(%v)", id, err)
	}
	return
}

// UserLogCreate create user log table.
func (d *Dao) UserLogCreate(c context.Context, id int64) (err error) {
	if err = d.DB.Exec(fmt.Sprintf(_userLogCreateSQL, id)).Error; err != nil {
		log.Error("UserLogCreate id(%d) error(%v)", id, err)
	}
	return
}
