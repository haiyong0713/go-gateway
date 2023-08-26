package show

import (
	"context"
	"go-common/library/cache/credis"

	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/resource/service/conf"
)

// Dao is resource dao.
type Dao struct {
	db         *xsql.DB
	dbMgr      *xsql.DB
	c          *conf.Config
	httpClient *bm.Client
	redis      credis.Redis
}

// New init mysql db
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		db:         xsql.NewMySQL(c.DB.Show),
		dbMgr:      xsql.NewMySQL(c.DB.Manager),
		httpClient: bm.NewClient(c.HTTPClient),
		redis:      credis.NewRedis(c.Redis.Show),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.db.Close()
	d.dbMgr.Close()
}

// Ping check dao health.
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
