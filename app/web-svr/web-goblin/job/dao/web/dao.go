package web

import (
	"context"

	"go-common/library/database/bfs"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/web-goblin/job/conf"
)

const (
	_broadURL = "/x/internal/broadcast/push/all"
)

// Dao dao
type Dao struct {
	c *conf.Config
	// db
	db *sql.DB
	// http client
	http *bm.Client
	// bfs client
	bfsClient *bfs.BFS
	// http client
	xiaomiClient *bm.Client
	// broadcast URL
	broadcastURL string
	xiaomiURL    string
	ela          *elastic.Elastic
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:            c,
		db:           sql.NewMySQL(c.Mysql),
		http:         bm.NewClient(c.HTTPClient),
		bfsClient:    bfs.New(nil),
		xiaomiClient: bm.NewClient(c.XiaomiClient),
		broadcastURL: c.Host.API + _broadURL,
		xiaomiURL:    c.Host.Xiaomi + _xiaomiURI,
		ela:          elastic.NewElastic(nil),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping dao ping
func (d *Dao) Ping(c context.Context) error {
	return nil
}

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}
