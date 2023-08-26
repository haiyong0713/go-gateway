package dao

import (
	"go-common/library/database/sql"

	"go-common/library/conf/paladin.v2"
)

type DB struct {
	fawkesDB *sql.DB
}

func NewDB() (db *DB, cf func(), err error) {
	var (
		fawkes sql.Config
		ct     paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		panic(err)
	}
	if err = ct.Get("fawkes").UnmarshalTOML(&fawkes); err != nil {
		panic(err)
	}
	db = &DB{
		fawkesDB: sql.NewMySQL(&fawkes),
	}
	cf = func() {
		db.fawkesDB.Close()
	}
	return
}
