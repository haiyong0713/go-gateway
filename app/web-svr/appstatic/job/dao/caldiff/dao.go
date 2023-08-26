package caldiff

import (
	"go-common/library/database/boss"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/appstatic/job/conf"
)

// Dao .
type Dao struct {
	c      *conf.Config
	db     *xsql.DB
	client *bm.Client
	boss   *boss.Boss
	host   *conf.Host
}

// New creates a dao instance.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		db:     xsql.NewMySQL(c.MySQL),
		client: bm.NewClient(c.HTTPClient),
		boss:   boss.New(c.Boss),
		host:   c.Host,
	}
	return
}
