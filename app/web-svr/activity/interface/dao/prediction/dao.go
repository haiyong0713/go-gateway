package prediction

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao struct
type Dao struct {
	db                *xsql.DB
	mc                *memcache.Memcache
	mcPerpetualExpire int32
	redis             *redis.Pool
}

// New init
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		db:                component.GlobalDB,
		mc:                memcache.New(c.Memcache.Like),
		mcPerpetualExpire: int32(time.Duration(c.Memcache.PerpetualExpire) / time.Second),
		redis:             redis.NewPool(c.Redis.Config),
	}
	return
}

// Close Dao
func (dao *Dao) Close() {
	if dao.redis != nil {
		dao.redis.Close()
	}
	if dao.mc != nil {
		dao.mc.Close()
	}
}

// Ping Dao
func (dao *Dao) Ping(c context.Context) error {
	return dao.db.Ping(c)
}

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	log.Error(format, args...)
}
