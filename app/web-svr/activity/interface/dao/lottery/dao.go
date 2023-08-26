package lottery

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	_addTimes  = "add"
	_usedTimes = "used"
	_msgURL    = "/api/notify/send.user.notify.do"
)

// Dao dao struct.
type Dao struct {
	c                    *conf.Config
	db                   *xsql.DB
	mcLotteryExpire      int32
	lotteryIPExpire      int32
	lotteryExpire        int32
	lotteryTimesExpire   int32
	lotteryWinListExpire int32
	wxLotteryLogExpire   int32
	wxRedDotExpire       int32
	getAddressURL        string
	mc                   *memcache.Memcache
	redis                *redis.Pool
	cache                *fanout.Fanout
	client               *httpx.Client
	msgURL               string
	couponURL            string
	vipURL               string
	sourceItemURL        string
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                    c,
		db:                   component.GlobalDB,
		lotteryIPExpire:      int32(time.Duration(c.Redis.LotteryIPExpire) / time.Second),
		lotteryTimesExpire:   int32(time.Duration(c.Redis.LotteryTimesExpire) / time.Second),
		lotteryExpire:        int32(time.Duration(c.Redis.LotteryExpire) / time.Second),
		lotteryWinListExpire: int32(time.Duration(c.Redis.LotteryWinListExpire) / time.Second),
		wxLotteryLogExpire:   int32(time.Duration(c.Redis.WxLotteryLogExpire) / time.Second),
		wxRedDotExpire:       int32(time.Duration(c.Redis.WxRedDotExpire) / time.Second),
		mcLotteryExpire:      int32(time.Duration(c.Memcache.LotteryExpire) / time.Second),
		getAddressURL:        c.Host.ShowCo + _getAddressURL,
		mc:                   memcache.New(c.Memcache.Like),
		redis:                redis.NewPool(c.Redis.Store),
		cache:                fanout.New("cache"),
		client:               httpx.NewClient(c.HTTPClient),
		msgURL:               c.Host.Message + _msgURL,
		couponURL:            c.Host.APICo + _memberCouponURI,
		vipURL:               c.Host.APICo + _memberVipURI,
	}
	return d
}

// Close Dao
func (d *Dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
	if d.mc != nil {
		d.mc.Close()
	}
}
