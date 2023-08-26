package dao

import (
	"context"

	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

func (d *dao) ArcPassed(ctx context.Context, mid int64) (*uparcapi.ArcPassedReply, error) {
	arg := &uparcapi.ArcPassedReq{
		Mid:          mid,
		Pn:           1,
		Ps:           10,
		WithoutStaff: false,
	}
	reply, err := d.uparc.ArcPassed(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
