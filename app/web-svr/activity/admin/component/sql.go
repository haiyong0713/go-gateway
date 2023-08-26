package component

import (
	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"
)

var (
	GlobalOrm *gorm.DB
)

func InitByCfg(cfg *orm.Config) error {
	GlobalOrm = orm.NewMySQL(cfg)
	return Ping()
}

func Ping() error {
	return nil
}
