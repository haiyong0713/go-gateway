package dao

import (
	"context"

	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

type locationDao struct {
	location locationgrpc.LocationClient
}

func (d *locationDao) InfoGRPC(ctx context.Context, ipaddr string) (*locationgrpc.InfoReply, error) {
	reply, err := d.location.Info(ctx, &locationgrpc.InfoReq{Addr: ipaddr})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
