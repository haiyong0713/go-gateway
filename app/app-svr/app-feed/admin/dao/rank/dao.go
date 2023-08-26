package rank

import (
	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/cache/redis"

	"github.com/jinzhu/gorm"

	"context"

	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dataplat"
)

// Dao struct user of color egg Dao.
type Dao struct {
	// db
	DB             *gorm.DB
	tagClient      tag.TagRPCClient
	DataPlatClient *dataplat.HttpClient
	Rds            *redis.Pool
}

// New create a instance of color egg Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// db
		DB:             orm.NewMySQL(c.ORM),
		DataPlatClient: dataplat.New(c.HTTPClient.DataPlat),
		Rds:            redis.NewPool(c.EntranceRedis.Config),
	}
	var err error
	if d.tagClient, err = tag.NewClient(c.TagGRPCClient); err != nil {
		panic(err)
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
