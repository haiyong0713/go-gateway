package duertv

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/job/conf"
)

type Dao struct {
	c *conf.Config
	// api
	client        *httpx.Client
	duertvPush    string
	duertvPushUGC string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:             c,
		client:        httpx.NewClient(c.HTTPDuertv),
		duertvPush:    c.Host.Duertv + _duertvPush,
		duertvPushUGC: c.Host.Duertv + _duertvPushUGC,
	}
	return d
}
