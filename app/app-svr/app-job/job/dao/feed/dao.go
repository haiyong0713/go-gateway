package feed

import (
	"time"

	"go-common/library/cache/memcache"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/archive/service/api"
)

type Dao struct {
	c *conf.Config
	// grpc
	arcGRPC api.ArchiveClient
	// http client
	client     *httpx.Client
	clientAsyn *httpx.Client
	// hetongzi
	hot string
	// tag
	tags string
	// rcmdUp
	rcmdUp   string
	mcRcmd   *memcache.Pool
	expireMC int32
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// http client
		client:     httpx.NewClient(c.HTTPClient),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		// hetongzi
		hot: c.Host.Hetongzi + _hot,
		// tag
		tags: c.Host.APICo + _tags,
		// rcmdUp
		rcmdUp: c.Host.APP + _rcmdUp,
		// mc
		mcRcmd:   memcache.NewPool(c.Memcache.Cache.Config),
		expireMC: int32(time.Duration(c.Memcache.Feed.ExpireCache) / time.Second),
	}
	var err error
	if d.arcGRPC, err = api.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	return d
}
