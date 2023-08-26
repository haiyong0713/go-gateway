package dao

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/playlist/job/conf"
)

// Dao .
type Dao struct {
	c            *conf.Config
	db           *xsql.DB
	redis        *redis.Pool
	httpClient   *bm.Client
	viewCacheTTL int64
}

// New creates a dao instance.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		db:           xsql.NewMySQL(c.Mysql),
		redis:        redis.NewPool(c.Redis),
		httpClient:   bm.NewClient(c.HTTPClient),
		viewCacheTTL: int64(time.Duration(c.Job.ViewCacheTTL) / time.Second),
	}
	return
}

// PromInfo prometheus info count.
func PromInfo(name string) {
	prom.BusinessInfoCount.Incr(name)
}

// Ping reports the health of the db/cache etc.
func (d *Dao) Ping(c context.Context) (err error) {
	if err = d.db.Ping(c); err != nil {
		return
	}
	err = d.pingRedis(c)
	return
}

func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}
