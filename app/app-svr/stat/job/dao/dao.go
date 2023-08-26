package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-gateway/app/app-svr/stat/job/conf"
)

// Dao is stat job dao.
type Dao struct {
	c  *conf.Config
	db *xsql.DB
}

// New add a feed job dao.
func New(c *conf.Config) *Dao {
	return &Dao{
		c:  c,
		db: xsql.NewMySQL(c.DB),
	}
}

// Ping ping health of db.
func (d *Dao) Ping(c context.Context) (err error) {
	return d.db.Ping(c)
}
