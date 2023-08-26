package gameholiday

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package gameholiday

const (
	prefix    = "act:gameholiday"
	separator = ":"
	midKey    = "mid"
	awardKey  = "award"
	// awardCountKey 获奖总人数
	awardCountKey = "awardCount"
	// addTimesLock 增加次数锁
	addTimesLock = "addTimesLock"
	// alreadyAddTimes 已经增加次数
	alreadyAddTimes = "alreadAt"
)

// Dao dao interface
type Dao interface {
	Close()
	AddTimeLock(c context.Context, mid int64) (err error)
	AddTimesRecord(c context.Context, mid int64, day string) (err error)
	GetAddTimesRecord(c context.Context, mid int64, day string) (res string, err error)
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
		redis: redis.NewPool(c.Redis.Store),
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
