package teenagers

import (
	"go-common/library/cache/redis"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

type Dao struct {
	attRedis *redis.Pool
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		attRedis: redis.NewPool(c.Redis.Attention.Config),
	}
	return
}
