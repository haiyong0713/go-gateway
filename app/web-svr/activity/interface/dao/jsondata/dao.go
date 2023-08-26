package jsondata

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao ...
type Dao struct {
	c            *conf.Config
	singleClient *httpx.Client

	// newyear2021DataBusPub *databus.Databus
}

// New ...
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:            c,
		singleClient: httpx.NewClient(c.HTTPClientSingle),
	}
	return d
}
