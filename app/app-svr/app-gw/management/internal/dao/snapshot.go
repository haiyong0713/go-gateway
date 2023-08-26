package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"

	"go-common/library/sync/errgroup.v2"

	"github.com/google/uuid"
)

func snapshotKey(uuid, node, gateway string) string {
	return fmt.Sprintf("{snapshot-%s-%s}/%s", node, gateway, uuid)
}

func snapshotBreakerAPIKey(uuid, node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{snapshot-%s-%s-%s}/breakerapi/")
	args := []interface{}{node, gateway, uuid}
	if api != "" {
		builder.WriteString("%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func snapshotDynPathKey(uuid, node, gateway, pattern string) string {
	builder := &strings.Builder{}
	builder.WriteString("{snapshot-%s-%s-%s}/dynpath/")
	args := []interface{}{node, gateway, uuid}
	if pattern != "" {
		builder.WriteString("%s")
		args = append(args, pattern)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *snapshotDao) ListBreakerAPI(ctx context.Context, node string, gateway string, uuid string) ([]*pb.BreakerAPI, error) {
	key := snapshotBreakerAPIKey(uuid, node, gateway, "")
	return d.dao.resource.scanBreakerAPI(ctx, gateway, key)
}

func (d *snapshotDao) SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) error {
	key := snapshotBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.setBreakerAPI(ctx, req, key)
}

func (d *snapshotDao) EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) error {
	key := snapshotBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.enableBreakerAPI(ctx, req, key)
}

func (d *snapshotDao) DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) error {
	key := snapshotBreakerAPIKey(uuid, req.Node, req.Gateway, req.Api)
	return d.dao.resource.deleteBreakerAPI(ctx, key)
}

func (d *snapshotDao) ListDynPath(ctx context.Context, node string, gateway string, uuid string) ([]*pb.DynPath, error) {
	key := snapshotDynPathKey(uuid, node, gateway, "")
	return d.dao.resource.scanDynPath(ctx, key)
}

func (d *snapshotDao) SetDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) error {
	key := snapshotDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.setDynPath(ctx, req, key)
}

func (d *snapshotDao) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) error {
	key := snapshotDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.deleteDynPath(ctx, key)
}

func (d *snapshotDao) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) error {
	key := snapshotDynPathKey(uuid, req.Node, req.Gateway, req.Pattern)
	return d.dao.resource.enableDynPath(ctx, req, key)
}

func (d *snapshotDao) GetSnapshotMeta(ctx context.Context, node string, gateway string, uuid string) (*pb.SnapshotMeta, error) {
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

func (d *snapshotDao) createSnapshot(ctx context.Context, node string, gateway string, uuid string) (*pb.SnapshotMeta, error) {
	meta := &pb.SnapshotMeta{
		Uuid:    uuid,
		Node:    node,
		Gateway: gateway,
	}
	key := snapshotKey(uuid, node, gateway)
	newRaw, err := meta.Marshal()
	if err != nil {
		return nil, err
	}
	casReq := d.dao.taishan.NewCASReq([]byte(key), []byte{}, newRaw)
	if err := d.dao.taishan.CAS(ctx, casReq); err != nil {
		return nil, err
	}
	return meta, nil
}

func (d *snapshotDao) AddSnapshot(ctx context.Context, req *pb.AddSnapshotReq) (*pb.AddSnapshotReply, error) {
	uuid := uuid.New().String()
	httpDao := d.dao.CreateHTTPResourceDao()
	breakerAPI, err := httpDao.ListBreakerAPI(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	dynPath, err := httpDao.ListDynPath(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	grpcDao := d.dao.CreateGRPCResourceDao()
	grpcBreakerAPI, err := grpcDao.ListBreakerAPI(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	grpcDynPath, err := grpcDao.ListDynPath(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}

	eg := errgroup.WithCancel(ctx)
	for _, dp := range dynPath {
		dpReq := &pb.SetDynPathReq{
			Node:       dp.Node,
			Gateway:    dp.Gateway,
			Pattern:    dp.Pattern,
			ClientInfo: dp.ClientInfo,
			UpdatedAt:  time.Now().Unix(),
			Enable:     dp.Enable,
			Username:   req.Username,
		}
		eg.Go(func(ctx context.Context) error {
			return d.SetDynPath(ctx, dpReq, uuid)
		})
	}
	for _, ba := range breakerAPI {
		baReq := &pb.SetBreakerAPIReq{
			Api:       ba.Api,
			Ratio:     ba.Ratio,
			Reason:    ba.Reason,
			Condition: ba.Condition,
			Action:    ba.Action,
			Enable:    ba.Enable,
			Node:      ba.Node,
			Gateway:   ba.Gateway,
			Username:  req.Username,
		}
		eg.Go(func(ctx context.Context) error {
			return d.SetBreakerAPI(ctx, baReq, uuid)
		})
	}
	for _, dp := range grpcDynPath {
		dpReq := &pb.SetDynPathReq{
			Node:       dp.Node,
			Gateway:    dp.Gateway,
			Pattern:    dp.Pattern,
			ClientInfo: dp.ClientInfo,
			UpdatedAt:  time.Now().Unix(),
			Enable:     dp.Enable,
			Username:   req.Username,
		}
		eg.Go(func(ctx context.Context) error {
			return d.grpcDao.SetDynPath(ctx, dpReq, uuid)
		})
	}
	for _, ba := range grpcBreakerAPI {
		baReq := &pb.SetBreakerAPIReq{
			Api:       ba.Api,
			Ratio:     ba.Ratio,
			Reason:    ba.Reason,
			Condition: ba.Condition,
			Action:    ba.Action,
			Enable:    ba.Enable,
			Node:      ba.Node,
			Gateway:   ba.Gateway,
			Username:  req.Username,
		}
		eg.Go(func(ctx context.Context) error {
			return d.grpcDao.SetBreakerAPI(ctx, baReq, uuid)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	meta, err := d.createSnapshot(ctx, req.Node, req.Gateway, uuid)
	if err != nil {
		return nil, err
	}
	reply := &pb.AddSnapshotReply{
		Meta: meta,
	}
	return reply, nil
}

func (d *snapshotDao) rawBreakerAndDynPaths(ctx context.Context, node, gateway, uuid string) ([]*pb.BreakerAPI, []*pb.DynPath, []*pb.BreakerAPI, []*pb.DynPath, error) {
	snapshotBreakerAPIs, err := d.ListBreakerAPI(ctx, node, gateway, uuid)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	snapshotDynPaths, err := d.ListDynPath(ctx, node, gateway, uuid)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	httpResourceDao := d.dao.CreateHTTPResourceDao()
	breakerAPIs, err := httpResourceDao.ListBreakerAPI(ctx, node, gateway)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dynPaths, err := httpResourceDao.ListDynPath(ctx, node, gateway)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return snapshotBreakerAPIs, snapshotDynPaths, breakerAPIs, dynPaths, nil
}

func constructBatchReq(snapshotBreakerAPIs, breakerAPIs []*pb.BreakerAPI, snapshotDynPaths, dynPaths []*pb.DynPath) *pb.SnapshotRunPlan {
	baMap := make(map[string]*pb.BreakerAPI, len(breakerAPIs))
	for _, v := range breakerAPIs {
		baMap[v.Api] = v
	}
	dpsMap := make(map[string]*pb.DynPath, len(dynPaths))
	for _, v := range dynPaths {
		dpsMap[v.Pattern] = v
	}
	ssbaMap := make(map[string]*pb.BreakerAPI, len(snapshotBreakerAPIs))
	for _, v := range snapshotBreakerAPIs {
		ssbaMap[v.Api] = v
	}
	ssdpsMap := make(map[string]*pb.DynPath, len(snapshotDynPaths))
	for _, v := range snapshotDynPaths {
		ssdpsMap[v.Pattern] = v
	}
	dpsReq := make([]*pb.SetDynPathReq, 0, len(snapshotDynPaths))
	for _, dp := range snapshotDynPaths {
		val, ok := dpsMap[dp.Pattern]
		if ok && model.MatchDynPath(val, dp) {
			continue
		}
		dpReq := &pb.SetDynPathReq{
			Node:       dp.Node,
			Gateway:    dp.Gateway,
			Pattern:    dp.Pattern,
			ClientInfo: dp.ClientInfo,
			UpdatedAt:  time.Now().Unix(),
			Enable:     dp.Enable,
		}
		dpsReq = append(dpsReq, dpReq)
	}
	basReq := make([]*pb.SetBreakerAPIReq, 0, len(snapshotBreakerAPIs))
	for _, ba := range snapshotBreakerAPIs {
		bapi, ok := baMap[ba.Api]
		if ok && model.MatchBreakerAPI(bapi, ba) {
			continue
		}
		baReq := &pb.SetBreakerAPIReq{
			Api:       ba.Api,
			Ratio:     ba.Ratio,
			Reason:    ba.Reason,
			Condition: ba.Condition,
			Action:    ba.Action,
			Enable:    ba.Enable,
			Node:      ba.Node,
			Gateway:   ba.Gateway,
		}
		basReq = append(basReq, baReq)
	}
	baSets := sets.StringKeySet(baMap)
	dpSets := sets.StringKeySet(dpsMap)
	ssbaSets := sets.StringKeySet(ssbaMap)
	ssdpSets := sets.StringKeySet(ssdpsMap)
	delBaSets := baSets.Difference(ssbaSets)
	delDpSets := dpSets.Difference(ssdpSets)
	delDynReq := make([]*pb.DeleteDynPathReq, 0, len(delDpSets))
	for key := range delDpSets {
		delReq := &pb.DeleteDynPathReq{Node: dpsMap[key].Node, Gateway: dpsMap[key].Gateway, Pattern: key}
		delDynReq = append(delDynReq, delReq)
	}
	delBaReq := make([]*pb.DeleteBreakerAPIReq, 0, len(delBaSets))
	for key := range delBaSets {
		delReq := &pb.DeleteBreakerAPIReq{Node: baMap[key].Node, Gateway: baMap[key].Gateway, Api: key}
		delBaReq = append(delBaReq, delReq)
	}
	out := &pb.SnapshotRunPlan{
		SetDynReq:     dpsReq,
		SetBreakerReq: basReq,
		DelDynReq:     delDynReq,
		DelBreakerReq: delBaReq,
	}
	return out
}

func (d *snapshotDao) BuildPlan(ctx context.Context, node, gateway, uuid string) (*pb.SnapshotRunPlan, error) {
	snapshotBreakerAPIs, snapshotDynPaths, breakerAPIs, dynPaths, err := d.rawBreakerAndDynPaths(ctx, node, gateway, uuid)
	if err != nil {
		return nil, err
	}
	return constructBatchReq(snapshotBreakerAPIs, breakerAPIs, snapshotDynPaths, dynPaths), nil
}

func (d *snapshotDao) RunPlan(ctx context.Context, req *pb.SnapshotRunPlan) error {
	if err := d.dao.BatchSetBreakerAPIAndDynPath(ctx, req.SetBreakerReq, req.SetDynReq); err != nil {
		return err
	}
	if err := d.dao.BatchDelBreakerAPIAndDynPath(ctx, req.DelBreakerReq, req.DelDynReq); err != nil {
		return err
	}
	return nil
}

func (d *snapshotDao) CreateSnapshotGRPCDao() SnapshotGRPCDao {
	return &snapshotGRPCDao{dao: d.dao}
}

func (d *snapshotDao) SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.resource.setQuotaMethod(ctx, req)
}

func (d *snapshotDao) GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	return d.dao.resource.getQuotaMethods(ctx, node, gateway)
}

func (d *snapshotDao) EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error {
	return d.dao.resource.enableQuotaMethod(ctx, req)
}

func (d *snapshotDao) DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.resource.deleteQuotaMethod(ctx, req)
}
