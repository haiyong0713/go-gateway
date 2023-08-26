package dao

import (
	"context"

	"go-common/library/log"

	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
)

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
