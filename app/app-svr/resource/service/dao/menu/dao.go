package menu

import (
	"context"
	"go-common/library/cache/credis"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/resource/service/conf"
)

type Dao struct {
	db    *sql.DB
	c     *conf.Config
	redis credis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    sql.NewMySQL(c.DB.Show),
		redis: credis.NewRedis(c.Redis.Show),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.db.Close()
}

// Ping check dao health.
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
