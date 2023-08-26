package frontpage

import (
	"github.com/jinzhu/gorm"

	locationGRPC "git.bilibili.co/bapis/bapis-go/platform/admin/location"
	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	// db
	ORMResource        *gorm.DB
	ORMManager         *gorm.DB
	HttpClient         *bm.Client
	locationGRPCClient locationGRPC.PolicyClient
}

// New create a instance of color tianma Dao and return.
func New(c *conf.Config) (d *Dao) {
	var err error
	d = &Dao{
		// db
		ORMResource: orm.NewMySQL(c.ORMResource),
		ORMManager:  orm.NewMySQL(c.ORMManager),
		HttpClient:  bm.NewClient(c.HTTPClient.Read),
	}
	if d.locationGRPCClient, err = locationGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	return
}
