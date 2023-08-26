package dubbing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/model/dubbing"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package dubbing

const (
	prefix     = "act_dubbing"
	separator  = ":"
	archiveKey = "archive"
	// awardCountKey 获奖总人数
	awardCountKey = "awardCount"
	// activityMid 参与活动的所有mid
	activityMid = "activityMid"
	midScoreKey = "midScore"
)

// Dao dao interface
type Dao interface {
	Close()
	AddArchiveScore(c context.Context, sid int64, batch int64, index int, list map[int64]int64) (err error)
	GetArchiveScore(c context.Context, sid int64, batch int64, index int) (res map[int64]int64, err error)
	SetDubbingMidScore(c context.Context, mid int64, midScore *dubbing.MapMidDubbingScore) (err error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c                  *conf.Config
	redis              *redis.Pool
	db                 *xsql.DB
	archiveScoreExpire int32
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:                  c,
		db:                 sql.NewMySQL(c.MySQL.Like),
		redis:              redis.NewPool(c.Redis.Config),
		archiveScoreExpire: int32(time.Duration(c.Dubbing.ArchiveScoreExpire) / time.Second),
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
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
