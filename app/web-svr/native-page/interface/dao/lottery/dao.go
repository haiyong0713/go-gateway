package lottery

import (
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

// Dao dao struct.
type Dao struct {
	c      *conf.Config
	client *httpx.Client
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: httpx.NewClient(c.HTTPClient),
	}
	return d
}

// Close Dao
func (d *Dao) Close() {
}
