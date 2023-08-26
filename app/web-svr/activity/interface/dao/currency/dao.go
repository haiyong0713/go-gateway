package currency

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao struct.
type Dao struct {
	c                    *conf.Config
	db                   *sql.DB
	mc                   *memcache.Memcache
	currencyExpire       int32
	liveItemPub          *databus.Databus
	redis                *redis.Pool
	userCurrencyExpire   int32
	cache                *fanout.Fanout
	singleCurrencyExpire int32
	client               *httpx.Client
	couponURL            string
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:           c,
		db:          component.GlobalDB,
		mc:          memcache.New(c.Memcache.Like),
		liveItemPub: databus.New(c.DataBus.LiveItemPub),
		redis:       redis.NewPool(c.Redis.Config),
		cache:       fanout.New("cache"),
		client:      httpx.NewClient(c.HTTPClient),
	}
	d.currencyExpire = int32(time.Duration(c.Memcache.CurrencyExpire) / time.Second)
	d.userCurrencyExpire = int32(time.Duration(c.Redis.UserCurrencyExpire) / time.Second)
	d.singleCurrencyExpire = int32(time.Duration(c.Redis.SingleCurrencyExpire) / time.Second)
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}
