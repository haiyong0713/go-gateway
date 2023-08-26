package knowledge

import (
	"fmt"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
	"time"
)

const (
	prefix    = "knowledge"
	separator = ":"
)

type Dao struct {
	c                 *conf.Config
	db                *sql.DB
	mc                *memcache.Memcache
	cache             *fanout.Fanout
	redis             *redis.Redis
	FiveMinutesExpire int32
	singleClient      *httpx.Client
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:            c,
		db:           component.GlobalDB,
		redis:        component.GlobalRedis,
		mc:           component.GlobalMC,
		singleClient: httpx.NewClient(c.HTTPClientSingle),

		cache:             fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		FiveMinutesExpire: int32(time.Duration(c.Redis.FiveMinutesExpire) / time.Second),
	}
	return d
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
