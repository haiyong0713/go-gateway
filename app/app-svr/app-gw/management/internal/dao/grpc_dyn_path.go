package dao

import (
	"context"
	"fmt"
	"strings"

	pb "go-gateway/app/app-svr/app-gw/management/api"
)

func grpcDynPathKey(node, gateway, pattern string) string {
	builder := &strings.Builder{}
	builder.WriteString("{grpc-dynpath-%s}/")
	args := []interface{}{node}
	if gateway != "" {
		builder.WriteString("%s/")
		args = append(args, gateway)
	}
	if pattern != "" {
		builder.WriteString("%s")
		args = append(args, pattern)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *grpcResourceDao) ListDynPath(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error) {
	return d.dao.scanDynPath(ctx, grpcDynPathKey(node, gateway, ""))
}

func (d *grpcResourceDao) SetDynPath(ctx context.Context, req *pb.SetDynPathReq) error {
	return d.dao.setDynPath(ctx, req, grpcDynPathKey(req.Node, req.Gateway, req.Pattern))
}

func (d *grpcResourceDao) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq) error {
	return d.dao.deleteDynPath(ctx, grpcDynPathKey(req.Node, req.Gateway, req.Pattern))
}

func (d *grpcResourceDao) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq) error {
	return d.dao.enableDynPath(ctx, req, grpcDynPathKey(req.Node, req.Gateway, req.Pattern))
}
