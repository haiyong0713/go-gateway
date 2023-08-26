package rank

import (
	"go-common/library/cache/redis"
	httpx "go-common/library/net/http/blademaster"
	"time"

	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao is rank dao.
type Dao struct {
	client *httpx.Client
	// rank
	rankBangumiAppURL string
	rankRegionAppURL  string
	rankAllAppURL     string
	rankOriginAppURL  string
	expireRedis       int32
	// redis
	redis *redis.Pool
}

// New rank dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient),
		// rank
		rankBangumiAppURL: c.Host.Data + _rankBangumiAppURL,
		rankRegionAppURL:  c.Host.Data + _rankRegionAppURL,
		rankAllAppURL:     c.Host.Data + _rankAllAppURL,
		rankOriginAppURL:  c.Host.Data + _rankOriginAppURL,
		// redis
		redis:       redis.NewPool(c.Redis.Recommend.Config),
		expireRedis: int32(time.Duration(c.Redis.Recommend.ExpireRank) / time.Second),
	}
	return
}
