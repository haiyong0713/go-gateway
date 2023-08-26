package monitor

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao is redis dao.
type Dao struct {
	c      *conf.Config
	client *bm.Client
	// url
	bapURL string
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
		// url
		bapURL: c.Host.Bap + _bapURL,
	}
	return d
}
