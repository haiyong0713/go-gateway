package dao

import (
	"context"

	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
)

const (
	_coinBizAv = "archive"
)

type coinDao struct {
	coin coingrpc.CoinClient
}

func (d *coinDao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error) {
	arg := &coingrpc.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: _coinBizAv,
	}
	reply, err := d.coin.ItemsUserCoins(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Numbers, nil
}
