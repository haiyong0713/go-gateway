package recommend

import (
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
)

type Dao struct {
	c *conf.Config
	// http client
	client    *bm.Client
	recommand string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// http client
		client:    bm.NewClient(c.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		recommand: c.Hosts.Data + _recommand,
	}
	return d
}
