package s10

import (
	"time"

	"go-gateway/app/web-svr/activity/admin/conf"

	http "go-common/library/net/http/blademaster"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	blackList "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
)

type Dao struct {
	restPointExpire     int32
	lotteryExpire       int32
	exchangeExpire      int32
	roundExchangeExpire int32
	pointDetailExpire   int32
	accountClient       account.AccountClient
	backListClient      blackList.AccountControlPlaneClient
	httpClient          *http.Client
	redeliveryURL       string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		restPointExpire:     int32(time.Duration(c.S10CacheExpire.RestPointExpire) / time.Second),
		lotteryExpire:       int32(time.Duration(c.S10CacheExpire.LotteryExpire) / time.Second),
		exchangeExpire:      int32(time.Duration(c.S10CacheExpire.ExchangeExpire) / time.Second),
		roundExchangeExpire: int32(time.Duration(c.S10CacheExpire.RoundExchangeExpire) / time.Second),
		pointDetailExpire:   int32(time.Duration(c.S10CacheExpire.PointDetailExpire) / time.Second),
		httpClient:          http.NewClient(c.HTTPClient),
		redeliveryURL:       c.S10General.RedeliveryHost + _redeliveryPath,
	}
	var err error
	if d.accountClient, err = account.NewClient(nil); err != nil {
		panic(err)
	}
	if d.backListClient, err = blackList.NewClient(nil); err != nil {
		panic(err)
	}
	return
}
