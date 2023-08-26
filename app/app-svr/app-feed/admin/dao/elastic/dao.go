package elastic

import (
	"go-common/library/database/elastic"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao struct user of Dao.
type Dao struct {
	c        *conf.Config
	esClient *elastic.Elastic
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		esClient: elastic.NewElastic(&elastic.Config{
			Host:       c.Host.Manager,
			HTTPClient: c.HTTPClient.ES,
		}),
	}
	return
}
