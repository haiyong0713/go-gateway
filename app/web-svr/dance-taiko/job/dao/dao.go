package dao

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"time"

	"go-gateway/app/web-svr/dance-taiko/job/conf"
)

// Dao dao
type Dao struct {
	c *conf.Config
	// db
	db *sql.DB
	// redis
	redis *redis.Pool
	// expire
	gameExpire int64
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:          c,
		db:         sql.NewMySQL(c.Mysql),
		redis:      redis.NewPool(c.Redis.Config),
		gameExpire: int64(time.Duration(c.Redis.GameExp) / time.Second),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
}
