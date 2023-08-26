package dao

import (
	"context"

	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	xsql "go-common/library/database/sql"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/conf"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/jinzhu/gorm"
)

const (
	_actURLAddTags                 = "/x/internal/tag/activity/add"
	_songsURL                      = "/x/internal/v1/audio/songs/activity/filter/info"
	_actReserveIncrURI             = "/x/internal/activity/reserve/incr"
	_actBwsReserveGiftURI          = "/x/internal/activity/bws/online/reserve/award"
	_contentFeatureSingleImportURL = "/x/internal/feature-admin/content/single/import"
	_nativePageURL                 = "/x/admin/native_page/native/topic/upgrade"
)

// Dao struct user of Dao.
type Dao struct {
	c                             *conf.Config
	DB                            *gorm.DB
	db                            *xsql.DB
	client                        *xhttp.Client
	es                            *elastic.Elastic
	tagGRPC                       tagrpc.TagRPCClient
	actURLAddTags                 string
	songsURL                      string
	actReserveURL                 string
	actBwsReserveGiftURL          string
	ContentFeatureSingleImportURL string
	actNativeURL                  string
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                             c,
		DB:                            orm.NewMySQL(c.ORM),
		db:                            xsql.NewMySQL(c.MySQL.Lottery),
		client:                        xhttp.NewClient(c.HTTPClient),
		es:                            elastic.NewElastic(c.EsClient),
		actURLAddTags:                 c.Host.API + _actURLAddTags,
		songsURL:                      c.Host.API + _songsURL,
		actReserveURL:                 c.Host.API + _actReserveIncrURI,
		actBwsReserveGiftURL:          c.Host.API + _actBwsReserveGiftURI,
		ContentFeatureSingleImportURL: c.Host.MNG + _contentFeatureSingleImportURL,
		actNativeURL:                  c.Host.MNG + _nativePageURL,
	}
	var err error
	if d.tagGRPC, err = tagrpc.NewClient(c.TagGRPC); err != nil {
		panic(err)
	}
	d.initORM()
	return
}

func (d *Dao) initORM() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if defaultTableName == "act_matchs" {
			return defaultTableName
		}
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
