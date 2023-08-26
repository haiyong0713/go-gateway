package s10

import (
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	blackList "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	user "git.bilibili.co/bapis/bapis-go/passport/service/user"

	http "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
)

type Dao struct {
	signedDataBusPub          *databus.Databus
	liveDataBusPub            *databus.Databus
	freeFlowDataBusPub        *databus.Databus
	signedExpire              int64
	taskProgressExpire        int64
	restPointExpire           int64
	lotteryExpire             int64
	exchangeExpire            int64
	roundExchangeExipre       int64
	restCountGoodsExpire      int64
	roundRestCountGoodsExpire int64
	userFlowExpire            int64
	couponURL                 string
	memberCouponURL           string
	httpClient                *http.Client
	accountClient             account.AccountClient
	backListClient            blackList.AccountControlPlaneClient
	userAccClient             user.PassportUserClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		signedDataBusPub:          databus.New(c.DataBus.SignedInPub),
		liveDataBusPub:            databus.New(c.DataBus.LiveItemPub),
		freeFlowDataBusPub:        databus.New(c.DataBus.FreeFlowPub),
		signedExpire:              int64(time.Duration(c.S10CacheExpire.SignedExpire) / time.Second),
		taskProgressExpire:        int64(time.Duration(c.S10CacheExpire.TaskProgressExpire) / time.Second),
		restPointExpire:           int64(time.Duration(c.S10CacheExpire.RestPointExpire) / time.Second),
		lotteryExpire:             int64(time.Duration(c.S10CacheExpire.LotteryExpire) / time.Second),
		exchangeExpire:            int64(time.Duration(c.S10CacheExpire.ExchangeExpire) / time.Second),
		roundExchangeExipre:       int64(time.Duration(c.S10CacheExpire.RoundExchangeExpire) / time.Second),
		restCountGoodsExpire:      int64(time.Duration(c.S10CacheExpire.RestCountGoodsExpire) / time.Second),
		roundRestCountGoodsExpire: int64(time.Duration(c.S10CacheExpire.RoundRestCountGoodsExpire) / time.Second),
		userFlowExpire:            int64(time.Duration(c.S10CacheExpire.UserFlowExpire) / time.Second),
		couponURL:                 c.Host.Mall + _mallCouponURI,
		memberCouponURL:           c.Host.APICo + _memberCouponURI,
		httpClient:                http.NewClient(c.HTTPClient),
	}
	var err error
	if d.accountClient, err = account.NewClient(nil); err != nil {
		panic(err)
	}
	if d.backListClient, err = blackList.NewClient(nil); err != nil {
		panic(err)
	}
	if d.userAccClient, err = user.NewClient(nil); err != nil {
		panic(err)
	}
	return
}
