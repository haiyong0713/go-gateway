package steins

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/steins-gate/service/conf"
	aegis "go-main/app/archive/aegis/admin/server/databus"
)

const (
	_dimensionURI  = "/v2/dash/hd/query"
	_dimensionsURI = "/v2/dash/hd/batch-query"
)

// Dao dao.
type Dao struct {
	c                *conf.Config
	db               *sql.DB
	httpVideoClient  *bm.Client
	httpWechatClient *bm.Client
	bvcDimensionURL  string
	bvcDimensionsURL string
	cache            *fanout.Fanout
	graphPassPub     *databus.Databus
	steinsCidPub     *databus.Databus
	rds              *redis.Pool
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
		// mysql
		db: sql.NewMySQL(c.MySQL.Steinsgate),
		// redis
		rds: redis.NewPool(c.Redis.Graph),
		// http
		httpVideoClient:  bm.NewClient(c.VideoClient),
		httpWechatClient: bm.NewClient(c.WechatClient),
		cache:            fanout.New("graph_mc", fanout.Worker(1), fanout.Buffer(10240)),
		graphPassPub:     databus.New(c.ArcInteractivePub),
		steinsCidPub:     databus.New(c.SteinsGate),
		bvcDimensionURL:  c.Host.Bvc + _dimensionURI,
		bvcDimensionsURL: c.Host.Bvc + _dimensionsURI,
	}
	aegis.InitAegis(nil)
	return
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	return nil

}
