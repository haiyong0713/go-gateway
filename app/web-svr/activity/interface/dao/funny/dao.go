package funny

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
)

const (
	prefix    = "act_funny"
	separator = ":"
)

// Dao dao interface
type Dao interface {
	Close()
	GetUserTodayIsAdded(c context.Context, mid int64) (IsAdd int, err error)
	GetTask1Num(c context.Context) (count int, err error)
	GetTask2Num(c context.Context) (count int, err error)
	SetUserAddedTimes(c context.Context, mid int64) (err error)
}

// Dao dao.
type dao struct {
	c           *conf.Config
	redis       *redis.Pool
	limitExpire int32
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:           c,
		redis:       redis.NewPool(c.Redis.Config),
		limitExpire: 3 * 24 * 60 * 60,
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
