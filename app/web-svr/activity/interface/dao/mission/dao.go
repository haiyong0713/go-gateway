package mission

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao.
type Dao struct {
	db    *sql.DB
	redis *redis.Redis
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:    component.GlobalDB,
		redis: component.GlobalRedis,
	}
	return
}
