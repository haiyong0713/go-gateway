package bubble

import (
	"go-common/library/cache/memcache"
	xsql "go-common/library/database/sql"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

type Dao struct {
	c  *conf.Config
	db *xsql.DB
	// memcache
	bubbleMc *memcache.Memcache
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		db:       component.GlobalDB,
		bubbleMc: memcache.New(c.Memcache.Bubble.Config),
	}
	return
}
