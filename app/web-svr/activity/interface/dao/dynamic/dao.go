package dynamic

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
)

type Dao struct {
	c              *conf.Config
	client         *httpx.Client
	dynamicInfoURL string
	feedDynamicURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         httpx.NewClient(c.HTTPDynamic),
		dynamicInfoURL: c.Host.Dynamic + _dynamicInfoURI,
		feedDynamicURL: c.Host.Dynamic + _feedDynamicURI,
	}
	return
}
