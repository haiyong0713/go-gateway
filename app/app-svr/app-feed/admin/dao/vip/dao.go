package vip

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao .
type Dao struct {
	c             *conf.Config
	vipHTTPClient *bm.Client
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:             c,
		vipHTTPClient: bm.NewClient(c.HTTPClient.Read),
	}
}
