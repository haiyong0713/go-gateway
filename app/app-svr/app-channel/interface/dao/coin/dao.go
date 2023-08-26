package coin

import (
	"context"

	"go-common/library/log"

	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-gateway/app/app-svr/app-channel/interface/conf"

	"github.com/pkg/errors"
)

// Dao is coin dao
type Dao struct {
	c          *conf.Config
	coinClient coingrpc.CoinClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.coinClient, err = coingrpc.NewClient(c.CoinGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) IsCoins(c context.Context, aids []int64, mid int64) (res map[int64]int64, err error) {
	var (
		args  = &coingrpc.ItemsUserCoinsReq{Mid: mid, Aids: aids, Business: "archive"}
		coins *coingrpc.ItemsUserCoinsReply
	)
	if coins, err = d.coinClient.ItemsUserCoins(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = coins.Numbers
	return
}

func (d *Dao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error) {
	const (
		_coinBizAv = "archive"
	)
	arg := &coingrpc.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: _coinBizAv,
	}
	reply, err := d.coinClient.ItemsUserCoins(ctx, arg)
	if err != nil {
		return nil, errors.WithMessagef(err, "ArchiveUserCoins arg=%+v", arg)
	}
	return reply.Numbers, nil
}
