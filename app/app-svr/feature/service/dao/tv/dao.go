package tv

import (
	xsql "go-common/library/database/sql"

	"go-gateway/app/app-svr/feature/service/conf"
)

type Dao struct {
	c  *conf.Config
	db *xsql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: xsql.NewMySQL(c.DB.TV),
	}
	return
}
