package aggregation

import (
	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/jinzhu/gorm"
)

const (
	_aggregationURL = "/data/rank/hotword/list-%d.json"
)

// Dao struct user of color egg Dao.
type Dao struct {
	c *conf.Config
	// db
	DB        *gorm.DB
	tagClient tag.TagRPCClient
	client    *bm.Client
	AggURL    string
}

// New create a instance of color egg Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// db
		DB:     orm.NewMySQL(c.ORM),
		client: bm.NewClient(c.HTTPClient.Read),
		AggURL: c.Host.BigData + _aggregationURL,
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
