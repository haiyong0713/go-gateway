package dao

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/dynamic/service/conf"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	_regionURL    = "/dynamic/region"
	_regionTagURL = "/dynamic/tag"
	_liveURL      = "/room/v1/Area/dynamic"
	_hotURL       = "/x/internal/tag/hotmap"
)

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}

// Dao dao.
type Dao struct {
	conf *conf.Config
	// http
	httpR *bm.Client
	// bigData api
	regionURI    string
	lpRegionURI  string
	regionTagURI string
	// live api
	liveURI string
	// tag api
	hotURI string
	// memcache
	mc       *memcache.Pool
	mcExpire int32
	// cache Prom
	cacheProm *prom.Prom
	// region archive redis
	rgRds *redis.Pool
	dbArc *sql.DB
	// content.flow.control.service gRPC
	cfcGRPC cfcgrpc.FlowControlClient
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf:         c,
		httpR:        bm.NewClient(c.HTTPClient.Read),
		regionURI:    c.Host.BigDataURI + _regionURL,
		lpRegionURI:  c.Host.LpBigDataURI + _regionURL,
		regionTagURI: c.Host.BigDataURI + _regionTagURL,
		liveURI:      c.Host.LiveURI + _liveURL,
		hotURI:       c.Host.APIURI + _hotURL,
		// memcache
		mc:       memcache.NewPool(c.Memcache.Config),
		mcExpire: int32(time.Duration(c.Memcache.Expire) / time.Second),
		// region redis
		rgRds: redis.NewPool(c.Redis.Archive.Config),
		dbArc: sql.NewMySQL(c.DB.ArcResult),
	}
	var err error
	if d.cfcGRPC, err = cfcgrpc.NewClient(c.CfcGRPC); err != nil {
		panic(err)
	}
	d.cacheProm = prom.CacheHit
	return
}

// Ping check connection success.
func (dao *Dao) Ping(c context.Context) (err error) {
	err = dao.pingMC(c)
	return
}

// Close close memcache resource.
func (dao *Dao) Close() {
	if dao.mc != nil {
		dao.mc.Close()
	}
	if dao.rgRds != nil {
		dao.rgRds.Close()
	}
}
