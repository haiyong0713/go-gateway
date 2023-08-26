package aggregation_v2

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	c         *conf.Config
	db        *sql.DB
	mc        *memcache.Pool
	client    *bm.Client
	tagClient tag.TagRPCClient
	// host
	aggURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		db:     sql.NewMySQL(c.MySQL.Show),
		mc:     memcache.NewPool(c.AggregationMemcache.Config),
		client: bm.NewClient(c.HTTPClient.Read),
		aggURL: c.Host.BigData + _aiAggregationURL,
	}
	var err error
	if d.tagClient, err = tag.NewClient(c.TagGRPCClient); err != nil {
		panic(err)
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}
