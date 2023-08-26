package stat

import (
	"go-gateway/app/app-svr/ugc-season/job/conf"

	"go-common/library/database/sql"
)

// Dao is redis dao.
type Dao struct {
	c  *conf.Config
	db *sql.DB
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.DB.Stat),
	}
	return d
}