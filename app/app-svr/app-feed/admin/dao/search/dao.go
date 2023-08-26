package search

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dataplat"

	filterGRPC "git.bilibili.co/bapis/bapis-go/filter/service"

	"github.com/jinzhu/gorm"
)

const (
	_searchOnlineUrl   = "http://s.search.bilibili.com/main/hotword"
	_statisticsUrl     = "/avenger/api/697/query"
	_statisticsUrlLive = "/avenger/api/858/query"

// _statisticsUrlLive_table = "bdp.ods_s_hot_search_stat_rt_view_dist"
)

// Dao struct user of color egg Dao.
type Dao struct {
	// db
	DB                      *gorm.DB
	MC                      *memcache.Memcache
	FilterGRPC              filterGRPC.FilterClient
	Client                  *bm.Client
	DataPlatClient          *dataplat.HttpClient
	DataPlatClient2         *dataplat.HttpClient
	SearchOnlineURL         string
	SearchStatisticsURL     string
	SearchStatisticsURLLive string
}

// New create a instance of color egg Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// db
		DB:                      orm.NewMySQL(c.ORMResource),
		MC:                      memcache.New(c.Memcache.Config),
		Client:                  bm.NewClient(c.HTTPClient.Read),
		DataPlatClient:          dataplat.New(c.HTTPClient.DataPlat),
		DataPlatClient2:         dataplat.New(c.HTTPClient.DataPlat2),
		SearchOnlineURL:         _searchOnlineUrl,
		SearchStatisticsURL:     c.Host.Berserker + _statisticsUrl,
		SearchStatisticsURLLive: c.Host.Berserker + _statisticsUrlLive,
	}
	var err error
	if d.FilterGRPC, err = filterGRPC.NewClient(c.FilGRPClient); err != nil {
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
