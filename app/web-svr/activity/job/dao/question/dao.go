package question

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao .
type Dao struct {
	c     *conf.Config
	db    *sql.DB
	redis *redis.Pool
}

// New .
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:     c,
		db:    sql.NewMySQL(c.MySQL.Like),
		redis: redis.NewPool(c.Redis.Config),
	}
	return d
}
