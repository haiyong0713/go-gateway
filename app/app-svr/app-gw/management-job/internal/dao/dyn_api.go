package dao

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/database/taishan"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/pkg/errors"
)

func dynPathKey(node, gateway, pattern string) string {
	builder := &strings.Builder{}
	builder.WriteString("{dynpath-%s}/")
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

func (d *dao) ListDynPath(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error) {
	key := dynPathKey(node, gateway, "")
	start, end := fullRange(key)
	out := []*pb.DynPath{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			dp := &pb.DynPath{}
			if err := dp.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal dyn path: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, dp)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}
	return out, nil
}

func (d *dao) GRPCListDynService(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error) {
	key := grpcDynPathKey(node, gateway, "")
	start, end := fullRange(key)
	out := []*pb.DynPath{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			dp := &pb.DynPath{}
			if err := dp.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal dyn service: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, dp)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}
	return out, nil
}
