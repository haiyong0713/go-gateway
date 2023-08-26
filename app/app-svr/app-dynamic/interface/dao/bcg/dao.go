package bcg

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	bcgGRPC "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/api"
)

type Dao struct {
	c *conf.Config
	// http client
	client  *bm.Client
	bcggrpc bcgGRPC.SunspotClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.bcggrpc, err = bcgGRPC.NewClient(c.BcgGRPC); err != nil {
		panic(err)
	}
	return d
}
