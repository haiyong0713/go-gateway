package bubble

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	c  *conf.Config
	db *sql.DB
	// memcache
	bubbleMc *memcache.Pool
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		db:       sql.NewMySQL(c.MySQL.Show),
		bubbleMc: memcache.NewPool(c.BubbleMemcache.Config),
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}
