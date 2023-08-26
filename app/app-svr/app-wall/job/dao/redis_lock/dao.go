package redis_lock

import (
	"context"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-wall/job/conf"
)

// Dao def
type Dao struct {
	// cachel
	redis *redis.Pool
}

// New create instance of Dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis: redis.NewPool(c.Redis.Wall.Config),
	}
	return
}

// Ping dao.
func (d *Dao) Ping(c context.Context) (err error) {
	conn := d.redis.Get(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

// Close dao.
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}
