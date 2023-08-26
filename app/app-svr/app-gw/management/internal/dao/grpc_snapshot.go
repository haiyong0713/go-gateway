package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/pkg/errors"
)

func snapshotGRPCBreakerAPIKey(uuid, node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{snapshot-grpc-%s-%s-%s}/breakerapi/")
	args := []interface{}{node, gateway, uuid}
	if api != "" {
		builder.WriteString("%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func snapshotGRPCDynPathKey(uuid, node, gateway, pattern string) string {
	builder := &strings.Builder{}
	builder.WriteString("{snapshot-grpc-%s-%s-%s}/dynpath/")
	args := []interface{}{node, gateway, uuid}
	if pattern != "" {
		builder.WriteString("%s")
		args = append(args, pattern)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *snapshotGRPCDao) ListBreakerAPI(ctx context.Context, node string, gateway string, uuid string) ([]*pb.BreakerAPI, error) {
	key := snapshotGRPCBreakerAPIKey(uuid, node, gateway, "")
	return d.dao.resource.scanBreakerAPI(ctx, gateway, key)
}

func (d *snapshotGRPCDao) SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) error {
	key := snapshotGRPCBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.setBreakerAPI(ctx, req, key)
}

func (d *snapshotGRPCDao) EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) error {
	key := snapshotGRPCBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.enableBreakerAPI(ctx, req, key)
}

func (d *snapshotGRPCDao) DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) error {
	key := snapshotGRPCBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.deleteBreakerAPI(ctx, key)
}

func (d *snapshotGRPCDao) ListDynPath(ctx context.Context, node string, gateway string, uuid string) ([]*pb.DynPath, error) {
	key := snapshotGRPCDynPathKey(uuid, node, gateway, "")
	return d.dao.resource.scanDynPath(ctx, key)
}

func (d *snapshotGRPCDao) SetDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) error {
	key := snapshotGRPCDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.setDynPath(ctx, req, key)
}

func (d *snapshotGRPCDao) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) error {
	key := snapshotGRPCDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.deleteDynPath(ctx, key)
}

func (d *snapshotGRPCDao) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) error {
	key := snapshotGRPCDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.enableDynPath(ctx, req, key)
}

func (d *snapshotGRPCDao) GetSnapshotMeta(ctx context.Context, node string, gateway string, uuid string) (*pb.SnapshotMeta, error) {
	key := snapshotKey(uuid, node, gateway)
	req := d.dao.taishan.NewGetReq([]byte(key))
	record, err := d.dao.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	meta := &pb.SnapshotMeta{}
	if err := meta.Unmarshal(record.Columns[0].Value); err != nil {
		return nil, err
	}
	return meta, nil
}

func (d *snapshotGRPCDao) rawBreakerAndDynPaths(ctx context.Context, node, gateway, uuid string) ([]*pb.BreakerAPI, []*pb.DynPath, []*pb.BreakerAPI, []*pb.DynPath, error) {
	snapshotBreakerAPIs, err := d.ListBreakerAPI(ctx, node, gateway, uuid)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	snapshotDynPaths, err := d.ListDynPath(ctx, node, gateway, uuid)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	grpcResourceDao := d.dao.CreateGRPCResourceDao()
	breakerAPIs, err := grpcResourceDao.ListBreakerAPI(ctx, node, gateway)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dynPaths, err := grpcResourceDao.ListDynPath(ctx, node, gateway)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return snapshotBreakerAPIs, snapshotDynPaths, breakerAPIs, dynPaths, nil
}

func (d *snapshotGRPCDao) BuildPlan(ctx context.Context, node, gateway, uuid string) (*pb.SnapshotRunPlan, error) {
	snapshotBreakerAPIs, snapshotDynPaths, breakerAPIs, dynPaths, err := d.rawBreakerAndDynPaths(ctx, node, gateway, uuid)
	if err != nil {
		return nil, err
	}
	return constructBatchReq(snapshotBreakerAPIs, breakerAPIs, snapshotDynPaths, dynPaths), nil
}

func (d *snapshotGRPCDao) RunPlan(ctx context.Context, req *pb.SnapshotRunPlan) error {
	if err := d.BatchSetBreakerAPIAndDynPath(ctx, req.SetBreakerReq, req.SetDynReq); err != nil {
		return err
	}
	if err := d.BatchDelBreakerAPIAndDynPath(ctx, req.DelBreakerReq, req.DelDynReq); err != nil {
		return err
	}
	return nil
}

func (d *snapshotGRPCDao) BatchSetBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.SetBreakerAPIReq, dpReq []*pb.SetDynPathReq) error {
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
		keys[grpcBreakerAPIKey(v.Node, v.Gateway, v.Api)] = value
	}
	for _, v := range dpReq {
		dp := &pb.DynPath{
			Node:       v.Node,
			Gateway:    v.Gateway,
			Pattern:    v.Pattern,
			ClientInfo: v.ClientInfo,
			UpdatedAt:  time.Now().Unix(),
			Enable:     v.Enable,
			Annotation: v.Annotation,
		}
		value, err := dp.Marshal()
		if err != nil {
			return err
		}
		keys[grpcDynPathKey(v.Node, v.Gateway, v.Pattern)] = value
	}
	batchPutReq := d.dao.taishan.NewBatchPutReq(ctx, keys, 0)
	resp, err := d.dao.taishan.BatchPut(ctx, batchPutReq)
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

func (d *snapshotGRPCDao) BatchDelBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.DeleteBreakerAPIReq, dpReq []*pb.DeleteDynPathReq) error {
	keys := make([]string, 0, len(bapiReq)+len(dpReq))
	for _, v := range bapiReq {
		keys = append(keys, grpcBreakerAPIKey(v.Node, v.Gateway, v.Api))
	}
	for _, v := range dpReq {
		keys = append(keys, grpcDynPathKey(v.Node, v.Gateway, v.Pattern))
	}
	batchDelReq := d.dao.taishan.NewBatchDelReq(ctx, keys)
	resp, err := d.dao.taishan.BatchDel(ctx, batchDelReq)
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

func (d *snapshotGRPCDao) SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.resource.setQuotaMethod(ctx, req)
}

func (d *snapshotGRPCDao) GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	return d.dao.resource.getQuotaMethods(ctx, node, gateway)
}

func (d *snapshotGRPCDao) EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error {
	return d.dao.resource.enableQuotaMethod(ctx, req)
}

func (d *snapshotGRPCDao) DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.resource.deleteQuotaMethod(ctx, req)
}
