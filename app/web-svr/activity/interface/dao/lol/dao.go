package lol

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao.
type Dao struct {
	// 存放用户竞猜列表
	redis *redis.Pool
	// 存放赛事ID
	mc *memcache.Memcache
	// key
	eSportsKey, contestDetailKey, GuessMainDetailsKey string
	// expire
	userExpire, listExpire int32
	cache                  *fanout.Fanout
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis:               redis.NewPool(c.S10Cache.Redis),
		mc:                  memcache.New(c.S10Cache.Mc),
		cache:               fanout.New("cache", fanout.Worker(1), fanout.Buffer(512)),
		eSportsKey:          c.S10Cache.S10Key.ESportsKey,
		contestDetailKey:    c.S10Cache.S10Key.ContestDetailKey,
		GuessMainDetailsKey: c.S10Cache.S10Key.GuessMainDetailsKey,
		userExpire:          int32(time.Duration(c.S10Cache.S10Key.UserExpire) / time.Second),
		listExpire:          int32(time.Duration(c.S10Cache.S10Key.ListExpire) / time.Second),
	}
	return
}

// Close close  resource.
func (d *Dao) Close() {
	d.mc.Close()
	d.redis.Close()
}
