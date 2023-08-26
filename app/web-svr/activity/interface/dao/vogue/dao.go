package dao

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// dao dao.
type Dao struct {
	mc          *memcache.Memcache
	db          *sql.DB
	redis       *redis.Pool
	client      *httpx.Client
	cache       *fanout.Fanout
	confExpire  int32
	goodsExpire int32
	taskExpire  int32

	winListURL string
	mytimesURL string
	favURL     string
}

// New new a dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		mc:          memcache.New(c.Memcache.Like),
		db:          sql.NewMySQL(c.MySQL.Like),
		redis:       redis.NewPool(c.Redis.Config),
		client:      httpx.NewClient(c.HTTPClientKfc),
		cache:       fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		confExpire:  int32(time.Duration(c.Memcache.VogueConfExpire) / time.Second),
		goodsExpire: int32(time.Duration(c.Memcache.VogueGoodsExpire) / time.Second),
		taskExpire:  int32(time.Duration(c.Memcache.VogueTaskExpire) / time.Second),

		winListURL: c.Host.APICo + "/x/activity/lottery/win/list",
		mytimesURL: c.Host.APICo + "/x/activity/lottery/mytimes",
		favURL:     c.Host.APICo + "/x/v3/fav/resource/ids",
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {

}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	return nil
}
