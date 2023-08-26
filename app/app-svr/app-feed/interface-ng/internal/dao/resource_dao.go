package dao

import (
	"context"
	"errors"

	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	rscrpc "go-gateway/app/app-svr/resource/service/rpc/client"
)

type resourceDao struct {
	resourceRPC  *rscrpc.Service
	resourceGRPC resourcegrpc.ResourceClient
}

func (d *resourceDao) Banner(ctx context.Context, req *model.ArgBanner) (map[int][]*model.Banner, string, error) {
	data, err := d.resourceRPC.Banners(ctx, req)
	if err != nil {
		return nil, "", err
	}
	if data == nil {
		return nil, "", errors.New("nil banner")
	}
	return data.Banner, data.Version, nil
}

func (d *resourceDao) CardPosRecs(ctx context.Context, ids []int64) (map[int64]*resourcegrpc.CardPosRec, error) {
	reply, err := d.resourceGRPC.CardPosRecs(ctx, &resourcegrpc.CardPosRecReplyRequest{CardIds: ids})
	if err != nil {
		return nil, err
	}
	return reply.Card, nil
}
