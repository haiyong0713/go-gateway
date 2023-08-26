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

func breakerAPIKey(node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{breakerapi-%s}/")
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

func fullRange(prefix string) (string, string) {
	return fmt.Sprintf("%s\x00", prefix), fmt.Sprintf("%s\xFF", prefix)
}

// ListBreakerAPI is
func (d *httpResourceDao) ListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error) {
	return d.dao.scanBreakerAPI(ctx, gateway, breakerAPIKey(node, gateway, ""))
}

func (d *resourceDao) scanBreakerAPI(ctx context.Context, _ string, key string) ([]*pb.BreakerAPI, error) {
	start, end := fullRange(key)
	out := []*pb.BreakerAPI{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			bapi := &pb.BreakerAPI{
				Action:   &pb.BreakerAction{},
				FlowCopy: &pb.FlowCopy{},
			}
			if err := bapi.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal breaker api: %+v", errors.WithStack(err))
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

// SetBreakerAPI is
func (d *httpResourceDao) SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq) error {
	return d.dao.setBreakerAPI(ctx, req, breakerAPIKey(req.Node, req.Gateway, req.Api))
}

// SetBreakerAPI is
func (d *resourceDao) setBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, key string) error {
	bapi := &pb.BreakerAPI{
		Api:       req.Api,
		Ratio:     req.Ratio,
		Reason:    req.Reason,
		Condition: req.Condition,
		Action:    req.Action,
		Enable:    req.Enable,
		Node:      req.Node,
		Gateway:   req.Gateway,
		FlowCopy:  req.FlowCopy,
	}
	value, err := bapi.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	return d.taishan.Put(ctx, putReq)
}

func (d *resourceDao) getRawBreakerAPI(ctx context.Context, key string) ([]byte, error) {
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

// EnableBreakerAPI is
func (d *httpResourceDao) EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq) error {
	return d.dao.enableBreakerAPI(ctx, req, breakerAPIKey(req.Node, req.Gateway, req.Api))
}

func (d *httpResourceDao) DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq) error {
	return d.dao.deleteBreakerAPI(ctx, breakerAPIKey(req.Node, req.Gateway, req.Api))
}

// EnableBreakerAPI is
func (d *resourceDao) enableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, key string) error {
	raw, err := d.getRawBreakerAPI(ctx, key)
	if err != nil {
		return err
	}
	bapi := &pb.BreakerAPI{}
	if err := bapi.Unmarshal(raw); err != nil {
		return err
	}
	bapi.Enable = !req.Disable
	newRaw, err := bapi.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), raw, newRaw)
	return d.taishan.CAS(ctx, casReq)
}

func (d *resourceDao) deleteBreakerAPI(ctx context.Context, key string) error {
	delReq := d.taishan.NewDelReq([]byte(key))
	return d.taishan.Del(ctx, delReq)
}

func (d *dao) BatchSetBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.SetBreakerAPIReq, dpReq []*pb.SetDynPathReq) error {
	keys := make(map[string][]byte, len(bapiReq)+len(dpReq))
	for _, v := range bapiReq {
		bapi := &pb.BreakerAPI{
			Api:       v.Api,
			Ratio:     v.Ratio,
			Reason:    v.Reason,
			Condition: v.Condition,
			Action:    v.Action,
			Enable:    v.Enable,
			Node:      v.Node,
			Gateway:   v.Gateway,
			FlowCopy:  v.FlowCopy,
		}
		value, err := bapi.Marshal()
		if err != nil {
			return err
		}
		keys[breakerAPIKey(v.Node, v.Gateway, v.Api)] = value
	}
	for _, v := range dpReq {
		dp := &pb.DynPath{
			Node:       v.Node,
			Gateway:    v.Gateway,
			Pattern:    v.Pattern,
			ClientInfo: v.ClientInfo,
			UpdatedAt:  time.Now().Unix(),
			Enable:     v.Enable,
		}
		value, err := dp.Marshal()
		if err != nil {
			return err
		}
		keys[dynPathKey(v.Node, v.Gateway, v.Pattern)] = value
	}
	batchPutReq := d.taishan.NewBatchPutReq(ctx, keys, 0)
	resp, err := d.taishan.BatchPut(ctx, batchPutReq)
	if err != nil {
		return err
	}
	if resp.AllSucceed {
		return nil
	}
	errs := []string{}
	for _, v := range resp.Records {
		if v.Status.ErrNo != 0 {
			errs = append(errs, fmt.Sprintf("key: %+v, errno: %+v, errmsg: %+v", string(v.Key), v.Status.ErrNo, v.Status.Msg))
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%+v", errs)
	}
	return nil
}

func (d *dao) BatchDelBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.DeleteBreakerAPIReq, dpReq []*pb.DeleteDynPathReq) error {
	keys := make([]string, 0, len(bapiReq)+len(dpReq))
	for _, v := range bapiReq {
		keys = append(keys, breakerAPIKey(v.Node, v.Gateway, v.Api))
	}
	for _, v := range dpReq {
		keys = append(keys, dynPathKey(v.Node, v.Gateway, v.Pattern))
	}
	batchDelReq := d.taishan.NewBatchDelReq(ctx, keys)
	resp, err := d.taishan.BatchDel(ctx, batchDelReq)
	if err != nil {
		return err
	}
	if resp.AllSucceed {
		return nil
	}
	errs := []string{}
	for _, v := range resp.Records {
		if v.Status.ErrNo != 0 {
			errs = append(errs, fmt.Sprintf("key: %+v, errno: %+v, errmsg: %+v", string(v.Key), v.Status.ErrNo, v.Status.Msg))
		}
	}
	if len(errs) > 0 {
		return errors.Errorf("%+v", errs)
	}
	return nil
}
