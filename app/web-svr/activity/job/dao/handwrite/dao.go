package handwrite

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/model/handwrite"
	mdl "go-gateway/app/web-svr/activity/job/model/handwrite"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package handwrite

const (
	prefix    = "act_handwrite"
	separator = "_"
	midKey    = "mid"
	awardKey  = "award"
	taskKey   = "task"
	// initFansKey 初始化粉丝key
	initFansKey = "init_fans"
	// awardCountKey 获奖总人数
	awardCountKey = "awardCount"
	// activityMid 参与活动的所有mid
	activityMid = "activityMid"
	// taskCountKey 任务完成总人数
	taskCountKey = "taskCountAll"
)

// Dao dao interface
type Dao interface {
	Close()
	AddMidAward(c context.Context, midMap map[int64]*mdl.MidAward) (err error)
	GetMidAward(c context.Context, mid int64) (res *mdl.MidAward, err error)
	SetAwardCount(c context.Context, awardCount *mdl.AwardCount) (err error)
	GetAwardCount(c context.Context) (res *mdl.AwardCount, err error)
	SetMidInitFans(c context.Context, midMap map[int64]int64) (err error)
	MidListDistinct(ctx context.Context, mids []int64) (rs []*handwrite.Mid, err error)
	GetActivityMember(c context.Context) (res []int64, err error)
	CacheActivityMember(c context.Context, mids []int64) (err error)
	GetMidsAward(c context.Context, mids []int64) (res map[int64]*mdl.MidAward, err error)
	GetMidInitFans(c context.Context) (values map[string]int64, err error)
	MidTask(ctx context.Context, mids []int64, taskType int) (rs []*handwrite.MidTaskDB, err error)
	AddMidTask(c context.Context, midMap map[int64]*mdl.MidTaskAll) (err error)
	SetTaskCount(c context.Context, awardCount *mdl.AwardCountNew) (err error)
	BatchAddTask(c context.Context, tasks []*handwrite.MidTaskDB) (err error)
	GetAllMidTask(ctx context.Context, offset, limit int64) (rs []*handwrite.MidTaskDB, err error)
	GetTaskCount(c context.Context) (res *mdl.AwardCountNew, err error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c     *conf.Config
	redis *redis.Pool
	// qpsLimitExpire int32
	db     *xsql.DB
	prefix string
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:      c,
		db:     sql.NewMySQL(c.MySQL.Like),
		redis:  redis.NewPool(c.Redis.Store),
		prefix: c.Handwrite2021.Prefix,
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

// Ping ping
func (d *dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

// buildKey build key
func (d *dao) buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return d.prefix + separator + strings.Join(strArgs, separator)
}
