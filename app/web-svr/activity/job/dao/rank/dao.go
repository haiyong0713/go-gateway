package rank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/model/rank"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

const (
	prefix       = "act_rank"
	separator    = "_"
	handWriteKey = "handWrite"
	rankKey      = "rank"
	midKey       = "mid"
)

// Dao dao interface
type Dao interface {
	Close()
	BatchAddRank(c context.Context, rank []*rank.DB) (err error)
	SetRank(c context.Context, rankNameKey string, rankBatch []*rank.Redis) (err error)
	GetRank(c context.Context, rankNameKey string) (res []*rank.Redis, err error)
	SetMidRank(c context.Context, rankNameKey string, midRank []*rank.Redis) (err error)
	GetRankListByBatch(c context.Context, sid, batch int64) (rs []*rank.DB, err error)
	GetMemberRankTimes(c context.Context, sid, startBatch, endBatch int64, mids []int64) (rs []*rank.MemberRankTimes, err error)
	GetMemberHighest(c context.Context, sid, startBatch, endBatch int64, mids []int64) (rs []*rank.MemberRankHighest, err error)
	GetRankListByBatchPatch(c context.Context, sid, batch int64, offset, limit int) (rs []*rank.DB, err error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c          *conf.Config
	redis      *redis.Pool
	db         *xsql.DB
	dataExpire int32
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:          c,
		db:         sql.NewMySQL(c.MySQL.Like),
		redis:      redis.NewPool(c.Redis.Config),
		dataExpire: int32(time.Duration(c.Redis.RankDataExpire) / time.Second),
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
