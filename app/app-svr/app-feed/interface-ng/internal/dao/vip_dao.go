package dao

import (
	"context"
	"strconv"

	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
)

type vipDao struct {
	vip vipgrpc.VipClient
}

func (d *vipDao) TipsRenew(ctx context.Context, build, platform, mid int64) (*vipgrpc.TipsRenewReply, error) {
	arg := &vipgrpc.TipsRenewReq{
		Position: 7,
		Mid:      mid,
		Platform: platform,
		Build:    strconv.FormatInt(build, 10),
	}
	reply, err := d.vip.TipsRenew(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
