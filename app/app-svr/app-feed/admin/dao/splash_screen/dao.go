package splash_screen

import (
	"context"

	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/jinzhu/gorm"
)

// Dao struct user of color splash_screen Dao.
type Dao struct {
	// db
	DB *gorm.DB
}

// New create a instance of color splash_screen Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// db
		DB: orm.NewMySQL(c.ORM),
	}
	d.initORM()
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
