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

func quotaMethodKey(node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{rate-limiter-%s}/%s")
	args := []interface{}{node, gateway}
	if api != "" {
		builder.WriteString("/%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *dao) GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	key := quotaMethodKey(node, gateway, "")
	start, end := fullRange(key)
	out := []*pb.QuotaMethod{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			bapi := &pb.QuotaMethod{}
			if err := bapi.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal quota method: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, bapi)
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
