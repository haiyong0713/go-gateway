package olympic

import (
	"context"
	"github.com/bluele/gcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
)

// Dao dao.
type Dao struct {
	conf         *conf.Config
	db           *sql.DB
	redis        *redis.Redis
	contestCache gcache.Cache
}

var localD *Dao

// New dao new.
func New(c *conf.Config) (d *Dao) {
	if localD != nil {
		return localD
	}
	d = &Dao{
		conf:         c,
		db:           component.GlobalDB,
		redis:        component.GlobalRedis,
		contestCache: gcache.New(c.OlympicConf.ValidContestSize).LFU().Build(),
	}
	d.refreshValidContestCache(context.Background())
	go initialize.CallC(d.refreshValidContestCacheTicker)
	localD = d
	return
}
