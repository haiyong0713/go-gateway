package lottery

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/conf"

	"github.com/jinzhu/gorm"
)

const (
	_addrDetail    = "/api/basecenter/addr/view"
	_vipInfo       = "/x/admin/vip/act/info"
	_couponInfo    = "/x/admin/coupon/allowance/batch/info"
	lotteryPrefix  = "lottery_new"
	separator      = "_"
	memberGroupKey = "member_group"
)

// Dao struct user of Dao
type Dao struct {
	// config
	c *conf.Config
	// db
	db            *sql.DB
	orm           *gorm.DB
	httpClient    *bm.Client
	addrDetailURL string
	vipInfoURL    string
	couponInfoURL string
	redis         *redis.Pool
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		db:            sql.NewMySQL(c.MySQL.Lottery),
		orm:           orm.NewMySQL(c.ORM),
		httpClient:    bm.NewClient(c.HTTPClient),
		addrDetailURL: c.Host.SHOW + _addrDetail,
		vipInfoURL:    c.Host.MNG + _vipInfo,
		couponInfoURL: c.Host.MNG + _couponInfo,
		redis:         redis.NewPool(c.Redis.Config),
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}

// Ping Dao
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return lotteryPrefix + separator + strings.Join(strArgs, separator)
}
