package view

import (
	"go-common/library/cache/redis"

	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao is dao.
type Dao struct {
	redis *redis.Pool
}

// New new a dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis: redis.NewPool(c.Redis.Feed.Config),
	}
	return
}
