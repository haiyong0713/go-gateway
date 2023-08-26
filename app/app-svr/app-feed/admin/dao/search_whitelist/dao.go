package search_whitelist

import (
	api "git.bilibili.co/bapis/bapis-go/archive/service"
	"github.com/jinzhu/gorm"

	"context"

	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	DB        *gorm.DB
	arcClient api.ArchiveClient
}

func New(c *conf.Config) (d *Dao) {
	var err error
	d = &Dao{
		// db
		DB: orm.NewMySQL(c.ORMResource),
	}
	d.initORM()
	if d.arcClient, err = api.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) initORM() {
	d.DB.LogMode(true)
}

// Ping check connection of db , mc.
func (d *Dao) Ping(c context.Context) (err error) {
	if d.DB != nil {
		err = d.DB.DB().PingContext(c)
		return
	}
	return
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
