package spmode

import (
	"github.com/jinzhu/gorm"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	db    *gorm.DB
	redis *redis.Redis
}

func NewDao(cfg *conf.Config) *Dao {
	db := orm.NewMySQL(cfg.ORM)
	db.LogMode(true)
	return &Dao{db: db, redis: redis.NewRedis(cfg.SpmodeRedis.Config)}
}
