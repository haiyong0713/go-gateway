package dao

import (
	"go-common/library/conf/paladin"
	"go-common/library/database/orm"

	"github.com/jinzhu/gorm"
)

func NewDB() (db *gorm.DB, err error) {
	var cfg struct {
		Client *orm.Config
	}
	if err = paladin.Get("db.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = orm.NewMySQL(cfg.Client)
	return
}

func (d *dao) DB() *gorm.DB {
	return d.db
}
