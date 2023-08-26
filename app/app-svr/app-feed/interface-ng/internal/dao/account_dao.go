package dao

import (
	"context"
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"
)

type accountDao struct {
	account accountgrpc.AccountClient
}

func (d *accountDao) IsAttentionGRPC(ctx context.Context, owners []int64, mid int64) map[int64]int8 {
	if len(owners) == 0 || mid == 0 {
		return nil
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	arg := &accountgrpc.RelationsReq{Mid: mid, Owners: owners, RealIp: ip}
	am, err := d.account.Relations3(ctx, arg)
	if err != nil {
		log.Error("Failed to raw relations3: %+v", err)
		return nil
	}
	isAtten := make(map[int64]int8, len(am.Relations))
	for mid, rel := range am.Relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return isAtten
}

func (d *accountDao) Cards3GRPC(c context.Context, mids []int64) (map[int64]*accountgrpc.Card, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.MidsReq{Mids: mids, RealIp: ip}
	cardsReply, err := d.account.Cards3(c, arg)
	if err != nil {
		return nil, err
	}
	return cardsReply.Cards, nil
}
