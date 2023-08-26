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

func gatewayKey(node, appName string) string {
	builder := &strings.Builder{}
	builder.WriteString("{gateway}/%s")
	args := []interface{}{node}
	if appName != "" {
		builder.WriteString("/%s")
		args = append(args, appName)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *dao) ListGateway(ctx context.Context) ([]*pb.Gateway, error) {
	key := "{gateway}"
	start, end := fullRange(key)
	out := []*pb.Gateway{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			gw := &pb.Gateway{
				Configs: []*pb.ConfigMeta{},
			}
			if err := gw.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal gateway: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, gw)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: reply.NextKey,
		}
	}
	return out, nil
}

func (d *dao) Gateway(ctx context.Context, node, gateway string) (*pb.Gateway, error) {
	key := gatewayKey(node, gateway)
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	gw := &pb.Gateway{}
	if err := gw.Unmarshal(record.Columns[0].Value); err != nil {
		return nil, err
	}
	return gw, nil
}
