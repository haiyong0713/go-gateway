package task

import (
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao dao struct.
type Dao struct {
	c               *conf.Config
	db              *sql.DB
	mc              *memcache.Memcache
	mcTaskExpire    int32
	redis           *redis.Pool
	userTaskExpire  int32
	taskStateExpire int32
	cache           *fanout.Fanout
}

const (
	prefix     = "act"
	separator  = ":"
	midRuleKey = "midRule"
	countKey   = "count"
)

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    component.GlobalDB,
		mc:    memcache.New(c.Memcache.Like),
		redis: redis.NewPool(c.Redis.Config),
		cache: fanout.New("cache"),
	}
	d.mcTaskExpire = int32(time.Duration(c.Memcache.TaskExpire) / time.Second)
	d.userTaskExpire = int32(time.Duration(c.Redis.UserTaskExpire) / time.Second)
	d.taskStateExpire = int32(time.Duration(c.Redis.TaskStateExpire) / time.Second)
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
