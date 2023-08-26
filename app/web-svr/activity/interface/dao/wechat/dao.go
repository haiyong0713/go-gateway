package wechat

import (
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao.
type Dao struct {
	c      *conf.Config
	client *xhttp.Client
}

// New ...
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		client: xhttp.NewClient(c.HTTPClient),
	}
	return d
}
