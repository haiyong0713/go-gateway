package dao

import (
	"context"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

func (d *dao) AccountInfo3(ctx context.Context, mid int64) (*accgrpc.Info, error) {
	reply, err := d.account.Info3(ctx, &accgrpc.MidReq{Mid: mid})
	if err != nil {
		return nil, err
	}
	return reply.Info, nil
}
