package music

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	return
}
