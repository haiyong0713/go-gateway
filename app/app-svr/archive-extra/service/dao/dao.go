package dao

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"

	"go-gateway/app/app-svr/archive-extra/service/conf"
)

// Dao is archive-extra dao
type Dao struct {
	c *conf.Config
	// db
	db    *sql.DB
	redis *redis.Redis
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    sql.NewMySQL(c.ExtraDB),
		redis: redis.NewRedis(c.Redis),
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	d.db.Close()
}
