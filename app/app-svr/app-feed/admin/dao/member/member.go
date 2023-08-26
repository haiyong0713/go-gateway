package member

import (
	"context"

	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/net/metadata"
)

func (d *Dao) RealnameTeenAgeCheck(ctx context.Context, mid int64) (*membergrpc.RealnameTeenAgeCheckReply, error) {
	req := &membergrpc.MidReq{
		Mid:    mid,
		RealIP: metadata.String(ctx, metadata.RemoteIP),
	}
	rly, err := d.client.RealnameTeenAgeCheck(ctx, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
