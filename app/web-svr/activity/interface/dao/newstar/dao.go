package newstar

import (
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/conf"
)

type Dao struct {
	c             *conf.Config
	db            *xsql.DB
	redis         *redis.Pool
	newstarExpire int32
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		db:            xsql.NewMySQL(c.MySQL.Like),
		redis:         redis.NewPool(c.Redis.Config),
		newstarExpire: int32(time.Duration(c.Redis.NewstarExpire) / time.Second),
	}
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
