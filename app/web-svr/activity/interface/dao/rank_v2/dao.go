package rank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/rank"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

const (
	prefix    = "act_rank"
	separator = ":"
	rankKey   = "rank"
	midKey    = "mid"
	configKey = "config"
)

// Dao dao interface
type Dao interface {
	Close()
	GetRank(c context.Context, rankKey string) (res []*rank.MidRank, err error)
	GetMidRank(c context.Context, rankActivityKey string, mid int64) (res *rank.MidRank, err error)
	GetRankConfig(c context.Context, sid int64, sidSource int) (rankConfig *rank.Rank, err error)
	SetRankConfig(c context.Context, sid int64, sidSource int, rankConfig *rank.Rank) (err error)

	GetRankConfigBySid(c context.Context, sid int64, sidSource int) (rank *rankmdl.Rank, err error)

	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c                *conf.Config
	redis            *redis.Pool
	db               *xsql.DB
	rankConfigExpire int32
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:                c,
		db:               sql.NewMySQL(c.MySQL.Like),
		redis:            redis.NewPool(c.Redis.Store),
		rankConfigExpire: int32(time.Duration(c.Redis.RankConfigExpire) / time.Second),
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
