package dao

import (
	"context"
	"fmt"
	"strings"

	pb "go-gateway/app/app-svr/app-gw/management/api"
)

func grpcBreakerAPIKey(node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{grpc-breakerapi-%s}/")
	args := []interface{}{node}
	if gateway != "" {
		builder.WriteString("%s/")
		args = append(args, gateway)
	}
	if api != "" {
		builder.WriteString("%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

// ListBreakerAPI is
func (d *grpcResourceDao) ListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error) {
	return d.dao.scanBreakerAPI(ctx, gateway, grpcBreakerAPIKey(node, gateway, ""))
}

// SetBreakerAPI is
func (d *grpcResourceDao) SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq) error {
	return d.dao.setBreakerAPI(ctx, req, grpcBreakerAPIKey(req.Node, req.Gateway, req.Api))
}

// EnableBreakerAPI is
func (d *grpcResourceDao) EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq) error {
	return d.dao.enableBreakerAPI(ctx, req, grpcBreakerAPIKey(req.Node, req.Gateway, req.Api))
}

func (d *grpcResourceDao) DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq) error {
	return d.dao.deleteBreakerAPI(ctx, grpcBreakerAPIKey(req.Node, req.Gateway, req.Api))
}
