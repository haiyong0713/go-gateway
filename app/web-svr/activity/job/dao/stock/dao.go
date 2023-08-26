package stock

import (
	"go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao  dao
type Dao struct {
	c          *conf.Config
	httpClient *blademaster.Client
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:          c,
		httpClient: blademaster.NewClient(c.HTTPClient),
	}
}
