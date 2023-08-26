package dao

import (
	"context"

	"go-common/library/log"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

type tunnelDao struct {
	tunnel tunnelgrpc.TunnelClient
}

func (d *tunnelDao) FeedCards(ctx context.Context, req *tunnelgrpc.FeedCardsReq) (map[int64]*tunnelgrpc.FeedCard, error) {
	resp, err := d.tunnel.FeedCards(ctx, req)
	if err != nil {
		log.Error("tunnel grpc FeedCards error(%v) or is resp null", err)
		return nil, err
	}
	return resp.FeedCards, nil
}
