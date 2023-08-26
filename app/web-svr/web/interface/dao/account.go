package dao

import (
	"context"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"
)

func (d *Dao) Card3(ctx context.Context, mid int64) (*accgrpc.Card, error) {
	req := &accgrpc.MidReq{Mid: mid}
	reply, err := d.accClient.Card3(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", req)
	}
	return reply.GetCard(), nil
}

func (d *Dao) Cards3(ctx context.Context, mids []int64) (map[int64]*accgrpc.Card, error) {
	req := &accgrpc.MidsReq{Mids: mids}
	reply, err := d.accClient.Cards3(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.accClient.Cards3 req=%+v", req)
	}
	return reply.Cards, nil
}
