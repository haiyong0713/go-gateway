package actplat

import (
	"fmt"
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
	"time"
)

const (
	taskPrefix = "act_task"
	separator  = ":"
	taskKey    = "task"
)

// Dao dao.
type Dao struct {
	c                 *conf.Config
	redisStore        *redis.Redis
	redisCache        *redis.Redis
	FiveMinutesExpire int32
}

// New dao new.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                 c,
		redisCache:        component.GlobalRedis,
		redisStore:        component.GlobalRedisStore,
		FiveMinutesExpire: int32(time.Duration(c.Redis.FiveMinutesExpire) / time.Second),
	}
	return
}

// Close Dao
func (d *Dao) Close() {

}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return taskPrefix + separator + strings.Join(strArgs, separator)
}
