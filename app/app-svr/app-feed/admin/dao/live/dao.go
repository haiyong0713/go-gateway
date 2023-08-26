package live

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao .
type Dao struct {
	c              *conf.Config
	liveHTTPClient *bm.Client
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:              c,
		liveHTTPClient: bm.NewClient(c.HTTPClient.Read),
	}
}
