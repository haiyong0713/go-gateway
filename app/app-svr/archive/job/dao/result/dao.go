package result

import (
	"context"

	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive/job/conf"
)

// Dao is redis dao.
type Dao struct {
	c      *conf.Config
	db     *sql.DB
	statDB *sql.DB
	client *bm.Client
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		db:     sql.NewMySQL(c.DB.Result),
		statDB: sql.NewMySQL(c.DB.Stat),
		client: bm.NewClient(c.HTTPClient),
	}
	return d
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}
