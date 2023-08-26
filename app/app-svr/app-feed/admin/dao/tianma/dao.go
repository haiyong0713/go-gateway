package tianma

import (
	"context"

	"go-common/library/database/orm"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	bgroupGRPC "git.bilibili.co/bapis/bapis-go/platform/service/bgroup"
	"github.com/jinzhu/gorm"
	bm "go-common/library/net/http/blademaster"
)

// Dao struct user of color tianma Dao.
type Dao struct {
	// db
	DB            *gorm.DB
	bgroupClient  bgroupGRPC.BGroupServiceClient
	CmmngHost     string
	BerserkerHost string
	HttpClient    *bm.Client
}

// New create a instance of color tianma Dao and return.
func New(c *conf.Config) (d *Dao) {
	var err error
	d = &Dao{
		// db
		DB:            orm.NewMySQL(c.ORMManager),
		CmmngHost:     c.Host.Cmmng,
		BerserkerHost: c.Host.Berserker,
		HttpClient:    bm.NewClient(c.HTTPClient.Read),
	}
	d.initORM()
	if d.bgroupClient, err = bgroupGRPC.NewClient(c.BGroupClient); err != nil {
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
