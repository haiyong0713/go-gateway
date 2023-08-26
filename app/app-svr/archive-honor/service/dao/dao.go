package dao

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"

	"go-gateway/app/app-svr/archive-honor/service/conf"
)

// Dao is archive-honor dao
type Dao struct {
	c *conf.Config
	// db
	db    *sql.DB
	redis *redis.Pool
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    sql.NewMySQL(c.HonorDB),
		redis: redis.NewPool(c.Redis),
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	d.db.Close()
	d.redis.Close()
}
