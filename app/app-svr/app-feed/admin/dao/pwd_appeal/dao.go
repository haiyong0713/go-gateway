package pwd_appeal

import (
	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	db *gorm.DB
}

func NewDao(cfg *conf.Config) *Dao {
	db := orm.NewMySQL(cfg.ORMManager)
	db.LogMode(true)
	return &Dao{db: db}
}
