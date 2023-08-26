package cost

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	"go-gateway/app/web-svr/activity/interface/dao/reward_conf"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
	"strings"
	"time"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package cost

const (
	prefix    = "cost"
	separator = ":"
	// cost:user_cost_SummerCamp_1234
	userCostRedisKey = "user_cost_%s_%d"
	// cost:user_exchange_orderid_2058174640_20210709_69
	userExchangeFlagKey = "user_exchange_orderid_%s"
)

// Dao dao interface
type Dao interface {
	GetUserTotalPoint(ctx context.Context, mid int64, activityId string) (totalPoint int64, costPoint int64, obtainPoint int64, err error)
	IsUserExchOrder(ctx context.Context, mid int64, orderId string) (bool, error)
	GetUserAllCost(ctx context.Context, activityId string, mid int64, isSplitTable bool) (total int, list []*cmdl.UserCostInfoDB, dbErr error)
	TaskFormulaTotal(ctx context.Context, mid int64, activityId string) (int64, error)
	fetchCostFromDB(ctx context.Context, activityId string, mid int64, lastID int64) ([]*cmdl.UserCostInfoDB, error)
	CacheGetUserCostPoint(ctx context.Context, mid int64, activityId string) (points int64, err error)
	CacheSetUserCostPoint(ctx context.Context, mid int64, activityId string, data int64) (err error)
	CacheDelUserCostPoint(ctx context.Context, mid int64, activityId string) (err error)
	getUserCostByOrderId(ctx context.Context, orderId string) (res *cmdl.UserCostInfoDB, err error)
	CacheGetUserExchangeFlag(ctx context.Context, orderId string) (flag int, err error)
	CacheSetUserExchangeFlag(ctx context.Context, orderId string, flag int) (err error)
	InsertOneUserCost(ctx context.Context, record *cmdl.UserCostInfoDB, isSplitTable bool) (int64, error)
	UserCostForExchange(ctx context.Context, activityId string, awardId string, mid int64, orderId string) error
	UserCostForLottery(ctx context.Context, mid int64, orderId string, awardPoolId string, activityId string) error
	GetUserCostListByDate(ctx context.Context, mid int64, activityId string, costTyp int, timeS xtime.Time) (list []*cmdl.UserCostInfoDB, err error)
	TodayUserHasExchangedPrizes(ctx context.Context, mid int64, activityId string) (list []*cmdl.UserCostInfoDB, err error)
}

// Dao dao.
type dao struct {
	c                *conf.Config
	redis            *redis.Redis
	db               *xsql.DB
	rewardConfDao    reward_conf.Dao
	stockDao         *bwsonline.Dao
	userCostExpire   int32
	userExangeExpire int32
}

// New init
func newDao(c *conf.Config) (nd Dao) {
	nd = &dao{
		c:                c,
		redis:            component.GlobalRedis,
		db:               component.GlobalDB,
		rewardConfDao:    reward_conf.New(c),
		stockDao:         bwsonline.New(c),
		userCostExpire:   int32(time.Duration(c.Redis.SummerCampUserCostExpire) / time.Second),
		userExangeExpire: int32(time.Duration(c.Redis.SummerCampUserExchangeFlagExpire) / time.Second),
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
