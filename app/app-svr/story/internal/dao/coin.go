package dao

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"github.com/pkg/errors"
)

const (
	_coinBizAv = "archive"
)

func (d *dao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (res map[int64]int64, err error) {
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
