package system

import (
	"context"
	http "go-common/library/net/http/blademaster"

	"go-common/library/database/orm"
	"go-gateway/app/web-svr/activity/admin/conf"

	"github.com/jinzhu/gorm"
)

type Dao struct {
	c      *conf.Config
	DB     *gorm.DB
	client *http.Client
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		DB:     orm.NewMySQL(c.ORM),
		client: http.NewClient(c.HTTPClient),
	}
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
