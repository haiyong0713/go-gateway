package handwrite

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
	mdl "go-gateway/app/web-svr/activity/interface/model/handwrite"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package handwrite

const (
	prefix    = "act_handwrite"
	separator = "_"
	midKey    = "mid"
	awardKey  = "award"
	// awardCountKey 获奖总人数
	awardCountKey = "awardCount"
	// addTimesLock 增加次数锁
	addTimesLock = "addTimesLock"
	// alreadyAddTimes 已经增加次数
	alreadyAddTimes = "alreadAt"
	taskKey         = "task"
	// taskCountKey 任务完成总人数
	taskCountKey = "taskCountAll"
)

// Dao dao interface
type Dao interface {
	Close()
	GetMidAward(c context.Context, mid int64) (res *mdl.MidAward, err error)
	GetAwardCount(c context.Context) (res *mdl.AwardCount, err error)
	AddTimeLock(c context.Context, mid int64) (err error)
	AddTimesRecord(c context.Context, mid int64, day string) (err error)
	GetAddTimesRecord(c context.Context, mid int64, day string) (res string, err error)
	GetMidTask(c context.Context, mid int64) (res *mdl.MidTaskAll, err error)
	GetTaskCount(c context.Context) (res *mdl.AwardCountNew, err error)
}

// Dao dao.
type dao struct {
	c      *conf.Config
	redis  *redis.Pool
	prefix string
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:      c,
		redis:  redis.NewPool(c.Redis.Store),
		prefix: c.HandWrite2021.Prefix,
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
func (d *dao) buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return d.prefix + separator + strings.Join(strArgs, separator)
}
