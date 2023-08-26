package task

import (
	"context"

	"go-common/library/database/orm"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/conf"

	"github.com/jinzhu/gorm"
)

type Dao struct {
	c           *conf.Config
	DB          *gorm.DB
	client      *xhttp.Client
	addAwardURL string
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:           c,
		DB:          orm.NewMySQL(c.ORM),
		client:      xhttp.NewClient(c.HTTPClient),
		addAwardURL: c.Host.API + _addAwardURI,
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
