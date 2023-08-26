package dao

import (
	"context"
	"go-common/library/cache/memcache"
	"net/http"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/esports/job/conf"
)

const _pushURL = "/x/internal/push-strategy/task/add"

// Dao dao
type Dao struct {
	c *conf.Config
	// http client
	http              *bm.Client
	messageHTTPClient *bm.Client
	// push service URL
	pushURL string
	// db
	db *sql.DB
	// leidata client
	ldClient        *http.Client
	bfsClient       *bfs.BFS
	ldHttp          *bm.Client
	redis           *redis.Pool
	scoreLiveExpire int32
	tunnelPub       *databus.Databus
	memcache        *memcache.Memcache
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:                 c,
		http:              bm.NewClient(c.HTTPClient),
		messageHTTPClient: bm.NewClient(c.MessageHTTPClient),
		db:                sql.NewMySQL(c.Mysql),
		pushURL:           c.Host.API + _pushURL,
		ldClient:          http.DefaultClient,
		bfsClient:         bfs.New(nil),
		ldHttp:            bm.NewClient(c.LeidaHTTPClient),
		redis:             redis.NewPool(c.Redis.Config),
		tunnelPub:         databus.New(c.TunnelDatabusPub),
		memcache:          memcache.New(c.Memcache),
	}
	dao.scoreLiveExpire = int32(time.Duration(c.Redis.ScoreLiveExpire) / time.Second)
	return
}

// New init mysql db
func V2New(c *conf.Config, db *sql.DB) (dao *Dao) {
	dao = &Dao{
		c:                 c,
		http:              bm.NewClient(c.HTTPClient),
		messageHTTPClient: bm.NewClient(c.MessageHTTPClient),
		db:                db,
		pushURL:           c.Host.API + _pushURL,
		ldClient:          http.DefaultClient,
		bfsClient:         bfs.New(nil),
		ldHttp:            bm.NewClient(c.LeidaHTTPClient),
	}
	return
}

func (d *Dao) GetMc() *memcache.Memcache {
	return d.memcache
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping ping dao
func (d *Dao) Ping(c context.Context) (err error) {
	if err = d.db.Ping(c); err != nil {
		return
	}
	return
}
