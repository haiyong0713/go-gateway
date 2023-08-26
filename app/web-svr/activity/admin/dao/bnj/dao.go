package bnj

import (
	"go-common/library/database/elastic"
	"go-gateway/app/web-svr/activity/admin/conf"
)

// Dao struct user of Dao.
type Dao struct {
	c        *conf.Config
	esClient *elastic.Elastic
}

// New init dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		esClient: elastic.NewElastic(c.Elastic),
	}
	return d
}
