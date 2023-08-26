package reward_conf

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
	"strings"
	"time"
)

const (
	prefix    = "reward_conf"
	separator = ":"
	// reward_conf:user_cost_SummerCamp_1234
	userCostRedisKey    = "user_cost_%s_%d"
	userExchangeFlagKey = "user_exchange_orderid_%s"
)

// Dao dao interface
type Dao interface {
	GetTodayAwardList(ctx context.Context, activityId string, costType int) ([]*cmdl.AwardConfigDataDB, error)
	GetAwardConfByIdAndDate(ctx context.Context, sid string, awardId string, costType int, timeS xtime.Time) (res *cmdl.AwardConfigDataDB, err error)
	FetchAwardFromDB(ctx context.Context, activityId string, costType int, timeS xtime.Time) ([]*cmdl.AwardConfigDataDB, error)
	IsAwardCanExchange(ctx context.Context, activityId string, awardId string, mid int64) (hasStock bool, res *cmdl.AwardConfigDataDB, err error)
	IsAwardCanLottery(ctx context.Context, activityId string, awardId string) (canLottery bool, res *cmdl.AwardConfigDataDB, err error)
}

// Dao dao.
type dao struct {
	c               *conf.Config
	redis           *redis.Redis
	db              *xsql.DB
	stockDao        *bwsonline.Dao
	awardListExpire int32
}

// New init
func newDao(c *conf.Config) (nd Dao) {
	nd = &dao{
		c:               c,
		redis:           component.GlobalRedis,
		db:              component.GlobalDB,
		stockDao:        bwsonline.New(c),
		awardListExpire: int32(time.Duration(c.Redis.RewardConfListExpire) / time.Second),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
