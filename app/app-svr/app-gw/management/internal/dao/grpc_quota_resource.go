package dao

import (
	"context"

	pb "go-gateway/app/app-svr/app-gw/management/api"
)

func (d *grpcResourceDao) GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	return d.dao.getQuotaMethods(ctx, node, gateway)
}

func (d *grpcResourceDao) SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.setQuotaMethod(ctx, req)
}

func (d *grpcResourceDao) DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.deleteQuotaMethod(ctx, req)
}

func (d *grpcResourceDao) EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error {
	return d.dao.enableQuotaMethod(ctx, req)
}
