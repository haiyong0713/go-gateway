package dao

import (
	"go-common/library/conf/paladin"
	"go-common/library/database/gorm"
	"go-common/library/database/sql"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("DB.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Archive").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

func NewORM() (orm *gorm.DB, cf func(), err error) {
	var (
		cfg gorm.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("DB.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Archive").UnmarshalTOML(&cfg); err != nil {
		return
	}
	orm, err = gorm.Open(&cfg)
	cf = func() {}
	return
}
