package dao

import (
	"github.com/jinzhu/gorm"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"

	"go-gateway/app/app-svr/resource/job/conf"
)

type Dao struct {
	c          *conf.Config
	showDB     *gorm.DB
	managerDB  *sql.DB
	resourceDB *sql.DB
	redisShow  *redis.Pool
}

func New(c *conf.Config) *Dao {
	db := orm.NewMySQL(c.MySQL.Show)
	db.SingularTable(true)

	return &Dao{
		c:          c,
		showDB:     db,
		resourceDB: sql.NewMySQL(c.MySQL.Resource),
		managerDB:  sql.NewMySQL(c.MySQL.Manager),
		redisShow:  redis.NewPool(c.Redis.Show.Config),
	}
}
