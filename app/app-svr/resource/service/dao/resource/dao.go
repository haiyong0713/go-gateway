package resource

import (
	"context"
	"go-common/library/cache/credis"
	xsql "go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/resource/service/conf"
	"golang.org/x/sync/singleflight"
	"time"
)

// Dao is resource dao.
type Dao struct {
	db                   *xsql.DB
	c                    *conf.Config
	redis                credis.Redis
	cache                *fanout.Fanout
	redisFrontPageExpire int32
	showRedis            credis.Redis
	singleGetCC          singleflight.Group
}

// New init mysql db
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                    c,
		db:                   xsql.NewMySQL(c.DB.Res),
		redis:                credis.NewRedis(c.Redis.Res.Config),
		showRedis:            credis.NewRedis(c.Redis.Show),
		cache:                fanout.New("cache"),
		redisFrontPageExpire: int32(time.Duration(c.Redis.Res.FrontPageExpire) / time.Second),
		singleGetCC:          singleflight.Group{},
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Close close the resource.
func (d *Dao) Close() {
	d.db.Close()
}

// Ping check dao health.
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
