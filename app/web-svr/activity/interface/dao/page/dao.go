package page

import (
	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"time"
)

type Dao struct {
	c          *conf.Config
	db         *sql.DB
	mc         *memcache.Memcache
	pageExpire int32
}

// New init page dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		db:         component.GlobalDB,
		mc:         memcache.New(c.Memcache.Like),
		pageExpire: int32(time.Duration(c.Redis.PageExpire) / time.Second),
	}
	return d
}

// Close .
func (d *Dao) Close() {
	if d.mc != nil {
		d.mc.Close()
	}
}
