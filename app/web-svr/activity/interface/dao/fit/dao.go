package fit

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/fit"
	"strings"
	"time"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package fit

const (
	prefix      = "fit"
	separator   = ":"
	planListKey = "plan_list"
	planIdKey   = "plan_id"
)

// Dao dao interface
type Dao interface {
	GetPlanList(ctx context.Context, offset, limit int) ([]*fit.PlanRecordRes, error)
	GetPlanById(ctx context.Context, planId int64) (res *fit.DBActFitPlanConfig, err error)
	CacheGetPlanList(ctx context.Context) (res []*fit.PlanRecordRes, err error)
	CacheSetPlanList(ctx context.Context, data []*fit.PlanRecordRes) (err error)
	CacheGetPlanDeatailById(ctx context.Context, planId int64) (res *fit.PlanWeekBodanList, err error)
	CacheSetPlanDeatailById(ctx context.Context, planId int64, data *fit.PlanWeekBodanList) (err error)
}

// Dao dao.
type dao struct {
	c                       *conf.Config
	redis                   *redis.Redis
	db                      *xsql.DB
	fitPlanListExpire       int32
	fitPlanDetailByIdExpire int32
}

// New init
func newDao(c *conf.Config) (nd Dao) {
	nd = &dao{
		c:                       c,
		redis:                   component.GlobalRedis,
		db:                      component.GlobalDB,
		fitPlanListExpire:       int32(time.Duration(c.Redis.FitPlanListExpire) / time.Second),
		fitPlanDetailByIdExpire: int32(time.Duration(c.Redis.FitPlanDetailByIdExpire) / time.Second),
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
