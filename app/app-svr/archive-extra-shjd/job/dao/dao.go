package dao

import (
	"go-common/library/cache/redis"

	"go-gateway/app/app-svr/archive-extra-shjd/job/conf"
)

// Dao is archive-extra-job dao
type Dao struct {
	c     *conf.Config
	redis *redis.Redis
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		redis: redis.NewRedis(c.Redis),
	}
	return
}
