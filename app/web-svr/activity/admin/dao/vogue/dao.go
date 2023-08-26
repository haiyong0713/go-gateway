package vogue

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/admin/conf"

	silver "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	"github.com/jinzhu/gorm"
)

// Dao struct
type Dao struct {
	c     *conf.Config
	lotDB *sql.DB
	DB    *gorm.DB
	acc   accApi.AccountClient
	sbc   silver.SilverbulletProxyClient
	redis *redis.Pool
}

// New create a instance of Dao and return
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// conf
		c: c,
		// db
		DB:    orm.NewMySQL(c.ORM),
		lotDB: sql.NewMySQL(c.MySQL.Lottery),
		redis: redis.NewPool(c.Redis.Config),
	}
	d.initORM()
	var err error
	if d.acc, err = accApi.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if d.sbc, err = silver.NewClient(c.SilverBulletClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) initORM() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName
	}
	d.DB.LogMode(true)
}

// Ping check connection of db , mc.
func (d *Dao) Ping(c context.Context) (err error) {
	if d.DB != nil {
		err = d.DB.DB().PingContext(c)
	}
	return
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
