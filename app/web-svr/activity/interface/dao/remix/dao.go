package remix

import (
	"fmt"
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package handwrite

const (
	prefix    = "act_handwrite"
	separator = "_"
	midKey    = "mid"
	scoreKey  = "score"
)

// Dao dao interface
type Dao interface {
	Close()
}

// Dao dao.
type dao struct {
	c     *conf.Config
	redis *redis.Pool
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:     c,
		redis: redis.NewPool(c.Redis.Config),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
	if d.redis != nil {
		d.redis.Close()
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
