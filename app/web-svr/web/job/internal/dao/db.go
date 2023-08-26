package dao

import (
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
)

func NewDB() (r *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.Map
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("show").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = sql.NewMySQL(&cfg)
	cf = func() { r.Close() }
	return
}
