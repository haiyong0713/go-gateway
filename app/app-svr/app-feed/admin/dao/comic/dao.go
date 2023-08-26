package comic

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao .
type Dao struct {
	c               *conf.Config
	comicHTTPClient *bm.Client
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:               c,
		comicHTTPClient: bm.NewClient(c.HTTPClient.Read),
	}
}
