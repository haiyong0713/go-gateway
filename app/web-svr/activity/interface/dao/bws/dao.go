package bws

import (
	"context"
	"runtime"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao.
type Dao struct {
	// config
	c *conf.Config
	// db
	db                   *xsql.DB
	mc                   *memcache.Memcache
	mcExpire             int32
	redis                *redis.Pool
	redisExpire          int32
	userAchExpire        int32
	userPointExpire      int32
	achCntExpire         int32
	mcItemExpire         int32
	bluetoothExpire      int32
	bwsOfflineUserExpire int32
	bwsRankUserExpire    int32
	bwsUserExpire        int32
	bws2019              []int64
	cacheCh              chan func()
	cache                *fanout.Fanout
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:                    c,
		db:                   component.GlobalDB,
		mc:                   memcache.New(c.Memcache.Like),
		mcExpire:             int32(time.Duration(c.Memcache.BwsExpire) / time.Second),
		redis:                redis.NewPool(c.Redis.Config),
		cacheCh:              make(chan func(), 1024),
		redisExpire:          int32(time.Duration(c.Redis.Expire) / time.Second),
		userAchExpire:        int32(time.Duration(c.Redis.UserAchExpire) / time.Second),
		userPointExpire:      int32(time.Duration(c.Redis.UserPointExpire) / time.Second),
		achCntExpire:         int32(time.Duration(c.Redis.AchCntExpire) / time.Second),
		mcItemExpire:         int32(time.Duration(c.Memcache.ItemExpire) / time.Second),
		bluetoothExpire:      int32(time.Duration(c.Redis.BwsbluetoothExpire) / time.Second),
		bwsUserExpire:        int32(time.Duration(c.Redis.BwsOnlineUserExpire) / time.Second),
		bwsOfflineUserExpire: int32(time.Duration(c.Redis.BwsOfflineUserExpire) / time.Second),
		bwsRankUserExpire:    int32(time.Duration(c.Redis.BwsRankUserExpire) / time.Second),
		bws2019:              c.Bws.Bws2019,
		cache:                fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go d.cacheproc()
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go d.cacheproc()
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// PromError stat and log.
func PromError(name string, format string, args ...interface{}) {
	log.Error(format, args...)
}

func (d *Dao) addCache(f func()) {
	select {
	case d.cacheCh <- f:
	default:
		log.Warn("d.cacheCh is full")
	}
}

func (d *Dao) cacheproc() {
	for {
		f, ok := <-d.cacheCh
		if !ok {
			return
		}
		f()
	}
}
