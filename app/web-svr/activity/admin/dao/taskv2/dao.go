package taskv2

import (
	"context"

	"go-common/library/database/orm"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"

	"github.com/jinzhu/gorm"
)

// Dao ...
type Dao struct {
	c  *conf.Config
	DB *gorm.DB
	db *xsql.DB
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: xsql.NewMySQL(c.MySQL.Lottery),
		DB: orm.NewMySQL(c.ORM),
	}
	d.DB.LogMode(true)
	return
}

// Ping .
func (d *Dao) Ping(c context.Context) error {
	return d.DB.DB().PingContext(c)
}

// Close .
func (d *Dao) Close() error {
	return d.DB.Close()
}
