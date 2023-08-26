package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (d *httpResourceDao) ListDynPath(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error) {
	return d.dao.scanDynPath(ctx, dynPathKey(node, gateway, ""))
}

func (d *resourceDao) scanDynPath(ctx context.Context, key string) ([]*pb.DynPath, error) {
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
			Key: reply.NextKey,
		}
	}
	return out, nil
}

func (d *httpResourceDao) SetDynPath(ctx context.Context, req *pb.SetDynPathReq) error {
	return d.dao.setDynPath(ctx, req, dynPathKey(req.Node, req.Gateway, req.Pattern))
}

func (d *resourceDao) setDynPath(ctx context.Context, req *pb.SetDynPathReq, key string) error {
	dp := &pb.DynPath{
		Node:       req.Node,
		Gateway:    req.Gateway,
		Pattern:    req.Pattern,
		ClientInfo: req.ClientInfo,
		UpdatedAt:  time.Now().Unix(),
		Enable:     req.Enable,
		Annotation: req.Annotation,
	}
	value, err := dp.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	return d.taishan.Put(ctx, putReq)
}

func (d *httpResourceDao) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq) error {
	return d.dao.deleteDynPath(ctx, dynPathKey(req.Node, req.Gateway, req.Pattern))
}

func (d *resourceDao) deleteDynPath(ctx context.Context, key string) error {
	delReq := d.taishan.NewDelReq([]byte(key))
	return d.taishan.Del(ctx, delReq)
}

func (d *resourceDao) getRawDynPath(ctx context.Context, key string) ([]byte, error) {
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

func (d *httpResourceDao) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq) error {
	return d.dao.enableDynPath(ctx, req, dynPathKey(req.Node, req.Gateway, req.Pattern))
}

func (d *resourceDao) enableDynPath(ctx context.Context, req *pb.EnableDynPathReq, key string) error {
	raw, err := d.getRawDynPath(ctx, key)
	if err != nil {
		return err
	}
	dp := &pb.DynPath{}
	if err := dp.Unmarshal(raw); err != nil {
		return err
	}
	dp.Enable = !req.Disable
	newRaw, err := dp.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), raw, newRaw)
	return d.taishan.CAS(ctx, casReq)
}
