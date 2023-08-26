package dao

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

func (d *dao) Cards3GRPC(c context.Context, mids []int64) (res map[int64]*accountgrpc.Card, err error) {
	var cardsReply *accountgrpc.CardsReply
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.MidsReq{Mids: mids, RealIp: ip}
	if cardsReply, err = d.accountClient.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = cardsReply.Cards
	return
}

func (d *dao) CheckRegTime(ctx context.Context, req *accountgrpc.CheckRegTimeReq) bool {
	res, err := d.accountClient.CheckRegTime(ctx, req)
	if err != nil {
		log.Error("d.accGRPC.CheckRegTime req=%+v", req)
		return false
	}
	return res.GetHit()
}

func (d *dao) Card3(ctx context.Context, mid int64) (*accountgrpc.Card, error) {
	result, err := d.accountClient.Card3(ctx, &accountgrpc.MidReq{
		Mid:    mid,
		RealIp: metadata.String(ctx, metadata.RemoteIP),
	})
	if err != nil {
		return nil, err
	}
	return result.GetCard(), nil
}
