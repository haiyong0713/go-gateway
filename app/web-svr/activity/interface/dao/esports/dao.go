package esports

import (
	"fmt"
	"go-common/library/cache/redis"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
	"time"
)

const (
	esportsActivityPrefix = "esports:arena:activity"
	separator             = ":"
)

// Dao ...
type Dao struct {
	c                 *conf.Config
	redisStore        *redis.Redis
	redisCache        *redis.Redis
	MidCardExpire     int32
	SevenDayExpire    int32
	ActivityEndExpire int32
	singleClient      *httpx.Client
	qpsLimitExpire    int32
}

// New ...
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:                 c,
		redisCache:        component.GlobalRedis,
		redisStore:        component.GlobalRedisStore,
		MidCardExpire:     int32(time.Duration(c.Redis.MidCardExpire) / time.Second),
		SevenDayExpire:    int32(time.Duration(c.Redis.SevenDayExpire) / time.Second),
		ActivityEndExpire: int32(time.Duration(c.Redis.SpringFestivialActivityEndExpire) / time.Second),
		singleClient:      httpx.NewClient(c.HTTPClientSingle),
		qpsLimitExpire:    int32(time.Duration(c.Redis.QPSLimitExpire) / time.Second),
	}
	return d
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return esportsActivityPrefix + separator + strings.Join(strArgs, separator)
}
