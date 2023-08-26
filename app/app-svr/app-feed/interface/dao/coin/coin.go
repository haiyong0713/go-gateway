package coin

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/pkg/errors"
)

const (
	_coinBizAv = "archive"
)

// Dao is coin dao
type Dao struct {
	coinClient api.CoinClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.coinClient, err = api.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	return
}

// ArchiveUserCoins .
func (d *Dao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (res map[int64]int64, err error) {
	var reply *api.ItemsUserCoinsReply
	arg := &api.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: _coinBizAv,
	}
	if reply, err = d.coinClient.ItemsUserCoins(ctx, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if reply == nil {
		return
	}
	res = reply.Numbers
	return
}
