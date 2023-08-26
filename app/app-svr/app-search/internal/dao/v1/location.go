package v1

import (
	"context"
	"go-common/library/log"

	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

func (d *dao) LocationInfo(ctx context.Context, ipaddr string) (info *locationgrpc.InfoReply, err error) {
	if info, err = d.locationClient.Info(ctx, &locationgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("%+v", err)
	}
	return
}
