package push

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

const _pushURL = "/x/internal/push-strategy/task/add"

// Dao dao
type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// push service URL
	pushURL string
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:       c,
		client:  bm.NewClient(c.HTTPClient.Push),
		pushURL: c.Host.API + _pushURL,
	}
	return
}
