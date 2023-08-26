package teen_manual

import (
	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	db *gorm.DB
}

func NewDao(cfg *conf.Config) *Dao {
	db := orm.NewMySQL(cfg.ORM)
	db.LogMode(true)
	return &Dao{db: db}
}
