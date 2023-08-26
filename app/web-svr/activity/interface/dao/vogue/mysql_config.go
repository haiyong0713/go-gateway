package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"
)

const (
	_confSQL = "SELECT config FROM act_vogue WHERE name = ?"
)

func (d *Dao) RawConfig(c context.Context, name string) (config string, err error) {
	row := d.db.QueryRow(c, _confSQL, name)
	if err = row.Scan(&config); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		log.Error("RawConfig(%s) error(%v)", name, err)
	}
	return
}
