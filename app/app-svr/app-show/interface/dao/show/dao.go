package show

import (
	"context"
	"go-gateway/app/app-svr/app-show/interface/component"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-show/interface/conf"
)

const (
	_aggKey         = "%d_hot_word"
	_aggregationURL = "/data/rank/hotword/list-%d.json"
)

// Dao is show dao.
type Dao struct {
	// mysql
	db *sql.DB
	// redis
	rcmmndRds *redis.Pool
	rcmmndExp int
	// http
	aggURL string
	client *bm.Client
	// memcache
	mc        *memcache.Memcache
	aggExpire int32
	redis     *redis.Pool
}

// New new a show dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// mysql
		db: component.GlobalShowDB,
		// redis
		rcmmndRds: redis.NewPool(c.Redis.Recommend.Config),
		rcmmndExp: int(time.Duration(c.Redis.Recommend.Expire) / time.Second),
		// memcache
		mc:        memcache.New(c.Memcache.Cards.Config),
		aggExpire: int32(time.Duration(c.Memcache.Cards.ExpireAggregation) / time.Second),
		// http
		client: bm.NewClient(c.HTTPClient),
		aggURL: c.Host.Data + _aggregationURL,
		redis:  redis.NewPool(c.Redis.Entrance),
	}
	return d
}

// Close close resource.
func (d *Dao) Close() (err error) {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
	if d.rcmmndRds != nil {
		return d.rcmmndRds.Close()
	}
	return nil
}

func (d *Dao) Ping(c context.Context) (err error) {
	conn := d.rcmmndRds.Get(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}
