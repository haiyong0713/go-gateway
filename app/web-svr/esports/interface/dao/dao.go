package dao

import (
	"context"
	"net/http"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
)

const (
	_searchURL = "/esports/search"
	_esSQLUrl  = "/x/admin/search"
)

// Dao dao struct.
type Dao struct {
	// config
	c *conf.Config
	// db
	db *sql.DB
	// redis
	redis                                             *redis.Pool
	filterExpire, listExpire, guessExpire, treeExpire int32
	// http client
	http      *bm.Client
	ldClient  *http.Client
	searchURL string
	ela       *elastic.Elastic
	cache     *fanout.Fanout
	esSQLURL  string
	tunnelPub *databus.Databus
	memcache  *memcache.Memcache
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:            c,
		db:           sql.NewMySQL(c.Mysql),
		redis:        redis.NewPool(c.Redis.Config),
		filterExpire: int32(time.Duration(c.Redis.FilterExpire) / time.Second),
		listExpire:   int32(time.Duration(c.Redis.ListExpire) / time.Second),
		guessExpire:  int32(time.Duration(c.Redis.GuessExpire) / time.Second),
		treeExpire:   int32(time.Duration(c.Redis.TreeExpire) / time.Second),
		http:         bm.NewClient(c.HTTPClient),
		ldClient:     http.DefaultClient,
		searchURL:    c.Host.Search + _searchURL,
		esSQLURL:     c.Host.Es + _esSQLUrl,
		ela:          elastic.NewElastic(c.Elastic),
		cache:        fanout.New("fanout"),
		tunnelPub:    databus.New(c.TunnelDatabusPub),
		memcache:     component.GlobalMemcached4UserGuess,
	}
	return
}

// Ping ping dao
func (d *Dao) Ping(c context.Context) (err error) {
	if err = component.GlobalDBOfMaster.Ping(c); err != nil {
		return
	}

	if err = d.db.Ping(c); err != nil {
		return
	}
	return
}
