package fit

import (
	"context"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"
)

type Dao struct {
	c  *conf.Config
	db *sql.DB
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.MySQL.Lottery),
	}
	return
}

// Ping Dao
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

// Close .
func (d *Dao) Close() error {
	return d.db.Close()
}
