package lottery

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package lottery
import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
)

const (
	lotteryPrefix  = "lottery_new"
	separator      = ":"
	infoKey        = "info"
	timesConfKey   = "times_conf"
	giftKey        = "gift"
	memberGroupKey = "member_group"
	addressKey     = "address"
	ipKey          = "ip"
	winKey         = "win"
	actionKey      = "action"
	realyWinKey    = "realy_win"
	timesKey       = "times"
	giftNumKey     = "giftNum"
	giftDayNumKey  = "giftDayNum"
	msgURL         = "/api/notify/send.user.notify.do"
	vipBuyURL      = "/mall-marketing/coupon_code/createV2"
	qpsLimitKey    = "qps"
)

// Dao dao interface
type Dao interface {
	Close()
	RawLottery(c context.Context, sid string) (res *lottery.Lottery, err error)
	RawLotteryInfo(c context.Context, sid string) (res *lottery.Info, err error)
	RawLotteryTimesConfig(c context.Context, sid string) (res []*lottery.TimesConfig, err error)
	RawLotteryGift(c context.Context, sid string) (res []*lottery.GiftDB, err error)
	RawLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error)
	InsertLotteryAddr(c context.Context, id, mid, addressID int64) (ef int64, err error)
	InsertLotteryAddTimes(c context.Context, id int64, mid int64, addType, num int, cid int64, ip, orderNo string) (ef int64, err error)
	InsertLotteryRecard(c context.Context, id int64, record []*lottery.InsertRecord, gid []int64, ip string) (count int64, err error)
	InsertLotteryRecardOrderNo(c context.Context, id int64, record []*lottery.InsertRecord, gid []int64, ip string) (count int64, err error)
	RawLotteryUsedTimes(c context.Context, id int64, mid int64) (res []*lottery.RecordDetail, err error)
	RawLotteryAddTimes(c context.Context, id int64, mid int64) (res []*lottery.AddTimes, err error)
	RawMemberGroup(c context.Context, sid string) (res []*lottery.MemberGroupDB, err error)
	UpdatelotteryGiftNumSQL(c context.Context, id int64, num int) (ef int64, err error)
	InsertLotteryWin(c context.Context, id, giftID, mid int64, ip string) (ef int64, err error)
	UpdateLotteryWin(c context.Context, id int64, mid int64, giftID int64, ip string) (ef int64, err error)
	RawLotteryWinOne(c context.Context, id, mid, giftID int64) (res string, err error)
	RawLotteryWinList(c context.Context, id int64, giftID []int64, num int64) (res []*lottery.GiftMid, err error)
	RawLotteryMidWinList(c context.Context, sid, mid, offset, limit int64) (res []*lottery.MidWinList, err error)
	RawLotteryActionByOrderNo(c context.Context, sid int64, orderNo string) (res *lottery.InsertRecord, err error)

	CacheLottery(c context.Context, sid string) (res *lottery.Lottery, err error)
	AddCacheLottery(c context.Context, sid string, val *lottery.Lottery) (err error)
	DeleteLottery(c context.Context, sid string) (err error)
	CacheLotteryInfo(c context.Context, sid string) (res *lottery.Info, err error)
	AddCacheLotteryInfo(c context.Context, sid string, val *lottery.Info) (err error)
	DeleteLotteryInfo(c context.Context, sid string) (err error)
	CacheLotteryTimesConfig(c context.Context, sid string) (res []*lottery.TimesConfig, err error)
	AddCacheLotteryTimesConfig(c context.Context, sid string, list []*lottery.TimesConfig) (err error)
	DeleteLotteryTimesConfig(c context.Context, sid string) (err error)
	CacheLotteryGift(c context.Context, sid string) (res []*lottery.Gift, err error)
	AddCacheLotteryGift(c context.Context, sid string, list []*lottery.Gift) (err error)
	DeleteLotteryGift(c context.Context, sid string) (err error)
	CacheLotteryAddrCheck(c context.Context, id, mid int64) (res int64, err error)
	AddCacheLotteryAddrCheck(c context.Context, id, mid int64, val int64) (err error)
	CacheIPRequestCheck(c context.Context, ip string) (res int, err error)
	AddCacheIPRequestCheck(c context.Context, ip string, val int) (err error)
	AddCacheLotteryTimes(c context.Context, sid int64, mid int64, remark string, list map[string]int) (err error)
	CacheLotteryTimes(c context.Context, sid int64, mid int64, remark string) (list map[string]int, err error)
	IncrTimes(c context.Context, sid int64, mid int64, list map[string]int, status string) (err error)
	CacheLotteryWinList(c context.Context, sid int64) (res []*lottery.GiftMid, err error)
	CacheLotteryActionLog(c context.Context, sid int64, mid int64, start, end int64) (res []*lottery.RecordDetail, err error)
	AddCacheLotteryActionLog(c context.Context, sid int64, mid int64, list []*lottery.RecordDetail) (err error)
	AddLotteryActionLog(c context.Context, sid int64, mid int64, list []*lottery.RecordDetail) (err error)
	CacheSendGiftNum(c context.Context, sid int64, giftIds []int64) (num map[int64]int64, err error)
	IncrGiftSendNum(c context.Context, sid int64, giftIDNum map[int64]int) (resMap map[int64]int64, err error)
	CacheSendDayGiftNum(c context.Context, sid int64, day string, giftKeys []string) (num map[string]int64, err error)
	AddCacheMemberGroup(c context.Context, sid string, list []*lottery.MemberGroup) (err error)
	DeleteMemberGroup(c context.Context, sid string) (err error)
	CacheMemberGroup(c context.Context, sid string) (res []*lottery.MemberGroup, err error)
	IncrGiftSendDayNum(c context.Context, sid int64, day string, giftKeysNum map[string]int, expireTime int64) (resGiftKeysNum map[string]int64, err error)
	CacheLotteryMcNum(c context.Context, sid int64, high, mc int) (res int64, err error)
	AddCacheLotteryMcNum(c context.Context, sid int64, high, mc int, val int64) (err error)
	AddCacheLotteryWinList(c context.Context, sid int64, list []*lottery.GiftMid) (err error)
	CacheQPSLimit(c context.Context, mid int64) (num int64, err error)
	DeleteLotteryActionLog(c context.Context, sid int64, mid int64) (err error)
	AddCacheLotteryWinLog(c context.Context, sid, mid int64, list []*lottery.MidWinList) (err error)
	CacheLotteryWinLog(c context.Context, sid, mid int64, start, end int64) (res []*lottery.MidWinList, err error)
	DeleteLotteryWinLog(c context.Context, sid, mid int64) (err error)
	GiftOtherSendNumIncr(c context.Context, key string, num int) (res bool, err error)

	SendSysMsg(c context.Context, uids []int64, mc, title string, context string, ip string) (err error)
	GetMemberAddress(c context.Context, id, mid int64) (val *lottery.AddressInfo, err error)
	SendVipBuyCoupon(c context.Context, clientIP, couponID, sourceActivityID, sourceBizID, uname string, sourceID, mid int64) (err error)
	SendLetter(c context.Context, l *lottery.LetterParam) (code int, err error)
	ComicsIsRookie(ctx context.Context, mid int64) (isRookie int, err error)
	GetCluesSrc(c context.Context, uri string, timeStamp int64) ([]*lottery.Item, error)
}

const (
	getAddressURL   = "/api/basecenter/addr/view"
	memberCouponURI = "/x/internal/coupon/allowance/receive"
	memberVipURI    = "/x/internal/vip/resources/grant"
)

// Dao dao.
type dao struct {
	c                     *conf.Config
	redis                 *redis.Pool
	redisNew              *redis.Redis
	db                    *xsql.DB
	mcLotteryExpire       int32
	lotteryIPExpire       int32
	lotteryExpire         int32
	lotteryTimesExpire    int32
	lotteryWinListExpire  int32
	wxLotteryLogExpire    int32
	wxRedDotExpire        int32
	qpsLimitExpire        int32
	getAddressURL         string
	mc                    *memcache.Memcache
	client                *httpx.Client
	couponURL             string
	vipURL                string
	sourceItemURL         string
	msgURL                string
	vipBuyURL             string
	comicBnj2021CouponURL string
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:                     c,
		db:                    component.GlobalDB,
		redisNew:              component.GlobalRedis,
		redis:                 redis.NewPool(c.Redis.Store),
		lotteryIPExpire:       int32(time.Duration(c.Redis.LotteryIPExpire) / time.Second),
		lotteryTimesExpire:    int32(time.Duration(c.Redis.LotteryTimesExpire) / time.Second),
		lotteryExpire:         int32(time.Duration(c.Redis.LotteryExpire) / time.Second),
		lotteryWinListExpire:  int32(time.Duration(c.Redis.LotteryWinListExpire) / time.Second),
		mcLotteryExpire:       int32(time.Duration(c.Memcache.LotteryExpire) / time.Second),
		qpsLimitExpire:        int32(time.Duration(c.Redis.QPSLimitExpire) / time.Second),
		getAddressURL:         c.Host.ShowCo + getAddressURL,
		mc:                    memcache.New(c.Memcache.Like),
		client:                httpx.NewClient(c.HTTPClient),
		couponURL:             c.Host.APICo + memberCouponURI,
		vipURL:                c.Host.APICo + memberVipURI,
		msgURL:                c.Host.Message + msgURL,
		vipBuyURL:             c.Host.VipBuy + vipBuyURL,
		comicBnj2021CouponURL: c.Host.Comic + _cartoonBnj2021URI,
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
	return lotteryPrefix + separator + strings.Join(strArgs, separator)
}

// lotteryTimesMapKey ...
func lotteryTimesMapKey(ltc *lottery.TimesConfig) string {
	return strconv.Itoa(ltc.Type) + strconv.FormatInt(ltc.ID, 10)
}
