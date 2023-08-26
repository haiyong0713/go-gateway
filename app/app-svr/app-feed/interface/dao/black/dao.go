package black

import (
	"context"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/interface/conf"
)

// Dao is black dao.
type Dao struct {
	// http clientAsyn
	clientAsyn *httpx.Client
	// redis
	redis     credis.Redis
	expireRds int32
	aCh       chan func()
	// url
	blackURL string
}

// New new a black dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		// redis init
		redis:     credis.NewRedis(c.Redis.Feed.Config),
		expireRds: int32(time.Duration(c.Redis.Feed.ExpireBlack) / time.Second),
		aCh:       make(chan func(), 1024),
		// url
		blackURL: c.Host.Black + _blackURL,
	}
	// nolint: biligowordcheck
	go d.cacheproc()
	return
}

// Ping  Ping check redis connection
func (d *Dao) Ping(c context.Context) (err error) {
	connRedis := d.redis.Conn(c)
	_, err = connRedis.Do("SET", "PING", "PONG")
	connRedis.Close()
	return
}

func (d *Dao) AddBlacklist(mid, aid int64) {
	d.addCache(func() {
		if err := d.addBlackCache(context.Background(), mid, aid); err != nil {
			log.Error("Failed to addBlackCache: %+v", err)
		}
	})
}

func (d *Dao) DelBlacklist(mid, aid int64) {
	d.addCache(func() {
		if err := d.delBlackCache(context.Background(), mid, aid); err != nil {
			log.Error("Failed to delBlackCache: %+v", err)
		}
	})
}

// addCache add cache to mc by goroutine
func (d *Dao) addCache(i func()) {
	select {
	case d.aCh <- i:
	default:
		log.Warn("cacheproc chan full")
	}
}

// cacheproc cache proc
func (d *Dao) cacheproc() {
	for {
		f, ok := <-d.aCh
		if !ok {
			log.Warn("cache proc exit")
			return
		}
		f()
	}
}
