package rank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	prefix  = "act_rank_new"
	ruleKey = "rule"
	baseKey = "base"

	separator = ":"
)

// Dao ...
type Dao struct {
	c              *conf.Config
	redisStore     *redis.Redis
	redisCache     *redis.Redis
	RuleExpire     int32
	SevenDayExpire int32
	OneDayExpire   int32
	singleClient   *httpx.Client
	qpsLimitExpire int32

	// newyear2021DataBusPub *databus.Databus
}

// New ...
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:              c,
		redisCache:     component.GlobalRedis,
		redisStore:     component.GlobalRedisStore,
		SevenDayExpire: int32(time.Duration(c.Redis.SevenDayExpire) / time.Second),
		OneDayExpire:   int32(time.Duration(c.Redis.OneDayExpire) / time.Second),
		singleClient:   httpx.NewClient(c.HTTPClientSingle),
		qpsLimitExpire: int32(time.Duration(c.Redis.QPSLimitExpire) / time.Second),

		// newyear2021DataBusPub: databus.New(c.DataBus.NewYear2021Pub),
	}
	return d
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return component.GlobalDB.Begin(c)
}
