package dao

import (
	"go-common/library/database/sql"

	"go-gateway/app/app-svr/archive-shjd/job/conf"
)

// Dao is redis dao.
type Dao struct {
	c      *conf.Config
	db     *sql.DB
	statDB *sql.DB
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		db:     sql.NewMySQL(c.DB.Result),
		statDB: sql.NewMySQL(c.DB.Stat),
	}
	return d
}
