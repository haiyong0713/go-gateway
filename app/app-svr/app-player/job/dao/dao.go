package dao

import (
	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-player/job/conf"
)

// Dao is dao.
type Dao struct {
	// redis
	c        *conf.Config
	mixRedis *redis.Pool
}

// New new a dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// reids
		c:        c,
		mixRedis: redis.NewPool(c.MixRedis),
	}
	return
}
