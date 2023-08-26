package dao

import (
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
)

func NewDB() (*sql.DB, error) {
	var cfg struct {
		Client *sql.Config
	}
	if err := paladin.Get("db.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	db := sql.NewMySQL(cfg.Client)
	return db, nil
}
