package bml

import (
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"strings"
	"time"
)

const (
	BmlGuessActPrefix = "act_bml_guess_2021_pre"
	separator         = ":"
)

type Dao struct {
	c          *conf.Config
	db         *sql.DB
	redis      *redis.Redis
	dataExpire int32
	userExpire int32
	cache      *fanout.Fanout
}

// New new dao.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:          c,
		db:         sql.NewMySQL(c.MySQL.Like),
		redis:      component.GlobalStockRedis,
		dataExpire: int32(time.Duration(c.Redis.BwsOnlineExpire) / time.Second),
		userExpire: int32(time.Duration(c.Redis.BwsOnlineUserExpire) / time.Second),
		cache:      fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
	return d
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return BmlGuessActPrefix + separator + strings.Join(strArgs, separator)
}
