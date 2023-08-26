package recommend

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/app-feed/interface/conf"
)

// Dao is show dao.
type Dao struct {
	c *conf.Config
	// http client
	client     *bm.Client
	clientAIAd *bm.Client
	clientAsyn *httpx.Client
	// hetongzi
	hot string
	// bigdata
	rcmd           string
	recommand      string
	group          string
	top            string
	followModeList string
	rcmdHot        string
	// databus
	databus *databus.Databus
	// mc
	mc       *memcache.Memcache
	expireMc int32
}

// New new a show dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// http client
		client:     bm.NewClient(c.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		clientAIAd: bm.NewClient(c.HTTPDataAd, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		// hetongzi
		hot: c.Host.Hetongzi + _hot,
		// bigdata
		rcmd:           c.HostDiscovery.Data + _rcmd,
		recommand:      c.HostDiscovery.Data + _recommand,
		group:          c.Host.BigData + _group,
		top:            c.Host.Data + _top,
		followModeList: c.Host.Data + _followModeList,
		rcmdHot:        c.Host.Data + _rcmdHot,
		// databus
		databus: databus.New(c.DislikeDatabus),
		// mc
		mc:       memcache.New(c.Memcache.Cache.Config),
		expireMc: int32(time.Duration(c.Memcache.Cache.ExpireCache) / time.Second),
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	if d.mc != nil {
		d.mc.Close()
	}
}
