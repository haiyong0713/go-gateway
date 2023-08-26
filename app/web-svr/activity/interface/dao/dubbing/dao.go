package dubbing

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/dubbing"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

const (
	prefix      = "act_dubbing"
	separator   = ":"
	rankKey     = "rank"
	midKey      = "mid"
	midScoreKey = "midScore"
)

// Dao dao interface
type Dao interface {
	Close()
	GetMidDubbingScore(c context.Context, mid int64) (res *dubbing.MapMidDubbingScore, err error)
	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c     *conf.Config
	redis *redis.Pool
	// qpsLimitExpire int32
	db *xsql.DB
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:  c,
		db: sql.NewMySQL(c.MySQL.Like),

		redis: redis.NewPool(c.Redis.Config),
		// qpsLimitExpire: int32(time.Duration(c.Brand.QPSLimitExpire) / time.Second),
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
