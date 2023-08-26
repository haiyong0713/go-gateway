package dao

import (
	"context"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

func (d *dao) InfoGRPC(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	return d.locationClient.Info(c, &locgrpc.InfoReq{Addr: ipaddr})
}
