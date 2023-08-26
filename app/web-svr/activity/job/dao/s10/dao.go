package s10

import (
	"time"

	"go-gateway/app/web-svr/activity/job/conf"

	user "git.bilibili.co/bapis/bapis-go/passport/service/user"
)

type Dao struct {
	restPointExpire     int32
	lotteryExpire       int32
	exchangeExpire      int32
	roundExchangeExpire int32
	pointDetailExpire   int32
	userFlowExpire      int32
	userAccClient       user.PassportUserClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		restPointExpire:     int32(time.Duration(c.S10CacheExpire.RestPointExpire) / time.Second),
		lotteryExpire:       int32(time.Duration(c.S10CacheExpire.LotteryExpire) / time.Second),
		exchangeExpire:      int32(time.Duration(c.S10CacheExpire.ExchangeExpire) / time.Second),
		roundExchangeExpire: int32(time.Duration(c.S10CacheExpire.RoundExchangeExpire) / time.Second),
		pointDetailExpire:   int32(time.Duration(c.S10CacheExpire.PointDetailExpire) / time.Second),
		userFlowExpire:      int32(time.Duration(c.S10CacheExpire.UserFlowExpire) / time.Second),
	}
	var err error
	if d.userAccClient, err = user.NewClient(nil); err != nil {
		panic(err)
	}
	return
}
