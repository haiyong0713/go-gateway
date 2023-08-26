package dm

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao dao struct.
type Dao struct {
	// http
	broadcastURL string
	httpCli      *bm.Client
}

// New return dm dao instance.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		broadcastURL: c.Host.APICo + _broadcastURI,
		httpCli:      bm.NewClient(c.HTTPClient),
	}
	return
}
