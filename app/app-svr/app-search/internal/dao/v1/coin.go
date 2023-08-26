package v1

import (
	"context"

	coinclient "git.bilibili.co/bapis/bapis-go/community/service/coin"
)

func (d *dao) ArchiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error) {
	const _coinBizAv = "archive"

	arg := &coinclient.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: _coinBizAv,
	}
	reply, err := d.coinClient.ItemsUserCoins(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Numbers, nil
}
