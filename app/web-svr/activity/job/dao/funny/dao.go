package funny

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/job/conf"
	"strings"
)

const (
	prefix    = "act_funny"
	separator = ":"
)

type Dao interface {
	Close()
	SetTask1Data(c context.Context, num int) (err error)
	SetTask2Data(c context.Context, num int) (err error)
	GetUserBatchData(c context.Context, sid, limit, page int64) ([]int64, error)
	IsNewUser(c context.Context, Mid int64) (bool, error)
}

// Dao .
type dao struct {
	c     *conf.Config
	db    *sql.DB
	redis *redis.Pool
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return &dao{
		c:     c,
		db:    sql.NewMySQL(c.MySQL.Like),
		redis: redis.NewPool(c.Redis.Config),
	}
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
