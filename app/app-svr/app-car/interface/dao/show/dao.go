package show

import (
	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-car/interface/conf"
)

type Dao struct {
	// redis
	redis *redis.Pool
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		// redis
		redis: redis.NewPool(c.Redis.Entrance),
	}
	return d
}
