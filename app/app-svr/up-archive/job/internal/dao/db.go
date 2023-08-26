package dao

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
)

type DB struct {
	resultDB *sql.DB
	tempDB   *sql.DB
}

func NewDB() (db *DB, cf func(), err error) {
	var (
		resultCfg, tempCfg sql.Config
		ct                 paladin.TOML
	)
	db = &DB{}
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Result").UnmarshalTOML(&resultCfg); err != nil {
		return
	}
	if err = ct.Get("Temp").UnmarshalTOML(&tempCfg); err != nil {
		return
	}
	db.resultDB = sql.NewMySQL(&resultCfg)
	db.tempDB = sql.NewMySQL(&tempCfg)
	cf = func() {
		db.resultDB.Close()
		db.tempDB.Close()
	}
	return
}
