package recommend

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/app-svr/app-intl/interface/conf"
)

// Dao is show dao.
type Dao struct {
	// http client
	client     *httpx.Client
	clientAsyn *httpx.Client
	// hetongzi
	hot string
	// bigdata
	rcmd    string
	group   string
	top     string
	rcmdHot string
	// redis
	redis *redis.Pool
	// mc
	mc *memcache.Pool
	c  *conf.Config
}

// New new a show dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// http client
		client:     httpx.NewClient(c.HTTPData, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		// hetongzi
		hot: c.Host.Hetongzi + _hot,
		// bigdata
		rcmd:    c.HostDiscovery.Data + _rcmd,
		group:   c.Host.BigData + _group,
		top:     c.Host.Data + _top,
		rcmdHot: c.HostDiscovery.Data + _rcmdHot,
		redis:   redis.NewPool(c.Redis.Feed.Config),
		mc:      memcache.NewPool(c.Memcache.Cache.Config),
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}
