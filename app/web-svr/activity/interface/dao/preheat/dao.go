package preheat

import (
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/conf"
)

type Dao struct {
	c     *conf.Config
	db    *xsql.DB
	redis *redis.Pool
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    xsql.NewMySQL(c.MySQL.Like),
		redis: redis.NewPool(c.Redis.Config),
	}
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
