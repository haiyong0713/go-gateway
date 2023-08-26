package bnj

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao bnj dao.
type Dao struct {
	c              *conf.Config
	db             *sql.DB
	mc             *memcache.Memcache
	redis          *redis.Pool
	client         *xhttp.Client
	comicClient    *xhttp.Client
	resetExpire    int32
	rewardExpire   int32
	grantCouponURL string
	comicCouponURL string
}

// New init bnj dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		db:           component.GlobalDB,
		mc:           memcache.New(c.Memcache.Like),
		redis:        redis.NewPool(c.Redis.Config),
		client:       xhttp.NewClient(c.HTTPClientBnj),
		comicClient:  xhttp.NewClient(c.HTTPClientComic),
		resetExpire:  int32(time.Duration(c.Redis.ResetExpire) / time.Second),
		rewardExpire: int32(time.Duration(c.Redis.RewardExpire) / time.Second),
	}
	d.grantCouponURL = d.c.Host.Mall + _grantCouponURL
	d.comicCouponURL = d.c.Host.Comic + _comicCouponURI
	return d
}

// Close .
func (d *Dao) Close() {
	if d.mc != nil {
		d.mc.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}
