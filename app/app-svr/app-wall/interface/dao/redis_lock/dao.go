package redis_lock

import (
	"context"

	"go-common/library/cache/credis"
	"go-gateway/app/app-svr/app-wall/interface/conf"
)

// Dao def
type Dao struct {
	// cachel
	redis credis.Redis
}

// New create instance of Dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis: credis.NewRedis(c.Redis.Wall.Config),
	}
	return
}

// Ping dao.
func (d *Dao) Ping(c context.Context) (err error) {
	conn := d.redis.Conn(c)
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
