package knowledgetask

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao  dao
type Dao struct {
	c     *conf.Config
	db    *sql.DB
	redis *redis.Redis
}

// New init
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    component.GlobalDB,
		redis: component.GlobalRedis,
	}
	return d
}
