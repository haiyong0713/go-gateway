package brand

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package brand
import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	prefix    = "act_brand"
	separator = "_"
)

// Dao dao interface
type Dao interface {
	Close()
	CacheAddCouponTimes(c context.Context, mid int64) (couponNum int64, err error)
	CacheSetMinusCouponTimes(c context.Context, mid int64) (err error)
	CacheQPSLimit(c context.Context, typeName string) (num int64, err error)
}

// Dao dao.
type dao struct {
	c              *conf.Config
	redis          *redis.Pool
	qpsLimitExpire int32
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:              c,
		redis:          redis.NewPool(c.Redis.Config),
		qpsLimitExpire: int32(time.Duration(c.Brand.QPSLimitExpire) / time.Second),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (dao *dao) Close() {
	if dao.redis != nil {
		dao.redis.Close()
	}
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}
