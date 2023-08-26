package rank

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v3"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

// Dao dao interface
type Dao interface {
	Close()
	BeginTran(c context.Context) (tx *xsql.Tx, err error)
	Ping(c context.Context) error
	GetBaseOnline(c context.Context, now time.Time) (list []*rankmdl.Base, err error)
	GetRuleOnline(c context.Context, now time.Time) (list []*rankmdl.Rule, err error)
	GetRankLog(c context.Context, rankID []int64, thisDate string) (list []*rankmdl.Log, err error)
	InsertRankLog(c context.Context, rank []*rankmdl.Log) (err error)
}

// BeginTran begin transcation.
func (d *dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Dao dao.
type dao struct {
	c      *conf.Config
	redis  *redis.Pool
	db     *xsql.DB
	client *xhttp.Client
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:      c,
		db:     xsql.NewMySQL(c.MySQL.Like),
		redis:  redis.NewPool(c.Redis.Store),
		client: xhttp.NewClient(c.HTTPClient),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}

// Ping ping
func (d *dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
