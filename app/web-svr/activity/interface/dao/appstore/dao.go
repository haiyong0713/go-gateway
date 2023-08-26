package appstore

import (
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	log.Error(format, args...)
}

// Dao dao.
type Dao struct {
	redis          *redis.Pool
	db             *xsql.DB
	appstoreExpire int32
	cache          *fanout.Fanout
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis:          redis.NewPool(c.Redis.Config),
		db:             xsql.NewMySQL(c.MySQL.Like),
		appstoreExpire: int32(time.Duration(c.Redis.AppstoreExpire) / time.Second),
		cache:          fanout.New("cache", fanout.Worker(1), fanout.Buffer(10240)),
	}
	return
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}
