package pay

import (
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao  dao
type Dao struct {
	c   *conf.Config
	db  *sql.DB
	pay *Pay
}

// New init
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:   c,
		db:  sql.NewMySQL(c.MySQL.Like),
		pay: newPayClient(bm.NewClient(c.HTTPClient)),
	}
	return d
}
