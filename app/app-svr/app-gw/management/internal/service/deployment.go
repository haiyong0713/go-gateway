package service

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/management-job/common"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"

	"go-common/library/sync/errgroup.v2"

	"github.com/BurntSushi/toml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

/*
	deployment state:
		unconfirmed, deploying, persistence, finished, rollback
		When the deployment has not been confirmed, the state is unconfirmed. When the deployment has been confirmed and has not been persisted, the state is deploying.
		When all instances has been deployed and config is persistent, the state is persistence.
		When config was persistent, and user click on the finish button, the state is finished.
		When user click on the rollback button, the state is rollback
	instance state: wait, done, rollback
*/

const (
	_deploymentFinished    = "finished"
	_deploymentRollback    = "rollback"
	_deploymentClosed      = "closed"
	_deploymentPersistence = "persistence"
	_deploymentUnconfirmed = "unconfirmed"
	_deploymentDeploying   = "deploying"
	_deploymentCancel      = "cancel"
	_instanceDone          = "done"
	_instanceWait          = "wait"
	_instanceRollback      = "rollback"
)

var (
	isDeploymentFinishState = sets.NewString(_deploymentFinished, _deploymentRollback, _deploymentClosed, _deploymentCancel)
	isDeploymentDeployState = sets.NewString(_deploymentDeploying)
	canRollbackState        = sets.NewString(_deploymentPersistence, _deploymentDeploying)
)

func deploymentIDKey() string {
	nowTs := time.Now().Unix()
	return strconv.FormatInt(math.MaxInt64-nowTs, 10) + "/" + strconv.FormatInt(nowTs, 10)
}

func (s *DeploymentService) Deployment(ctx context.Context, req *pb.ListDeploymentReq) (*pb.ListDeploymentReply, error) {
	if req.Stime <= 0 {
		req.Stime = time.Now().AddDate(0, -1, 0).Unix()
	}
	if req.Etime <= 0 {
		req.Etime = time.Now().AddDate(0, 0, 1).Unix()
	}
	reply := &pb.ListDeploymentReply{
		Node:    req.Node,
		Gateway: req.Gateway,
		Pages: pb.Page{
			Num:   req.PageNum,
			Size_: req.Size_,
		},
	}
	meta, err := s.dao.ListDeployment(ctx, req)
	if err != nil {
		return nil, err
	}
	start := (req.PageNum - 1) * req.Size_
	end := start + req.Size_
	rawLen := int64(len(meta))
	if rawLen < end {
		end = rawLen
	}
	if start >= rawLen {
		reply.Pages.Total = 0
		return reply, nil
	}
	reply.Lists = meta[start:end]
	reply.Pages.Total = rawLen
	return reply, nil
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, req *pb.CreateDeploymentReq) (*pb.CreateDeploymentReply, error) {
	unfinished, err := s.hasUnfinishedDeployment(ctx, req.Node, req.Gateway, req.DeploymentType)
	if err != nil {
		return nil, err
	}
	if unfinished {
		return nil, ecode.Error(ecode.RequestErr, "存在未发布完成的发布单")
	}
	instanceStatus, err := s.instanceStatus(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	snapshot, err := s.AddSnapshot(ctx, &pb.AddSnapshotReq{Node: req.Node, Gateway: req.Gateway})
	if err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("构建失败，请重试: %+v", err))
	}
	deploymentID := deploymentIDKey()
	nowTs := time.Now().Unix()
	meta := &pb.DeploymentMeta{
		DeploymentId:   deploymentID,
		SnapshotUuid:   req.Uuid,
		Description:    req.Description,
		Sponsor:        req.Username,
		CreatedAt:      nowTs,
		UpdatedAt:      nowTs,
		State:          _deploymentUnconfirmed,
		RollbackUuid:   snapshot.Meta.Uuid,
		Node:           req.Node,
		Gateway:        req.Gateway,
		DeploymentType: req.DeploymentType,
	}
	if err = s.buildDigestAndState(ctx, instanceStatus, meta); err != nil {
		return nil, err
	}
	if err = checkDigestSimilarity(instanceStatus); err != nil {
		return nil, err
	}
	if err := s.dao.CreateDeploymentMeta(ctx, meta); err != nil {
		audit.SendDeploymentLog(meta, audit.LogActionCreate, audit.LogLevelError, audit.LogResultFailure, jsonify(req))
		return nil, err
	}
	audit.SendDeploymentLog(meta, audit.LogActionCreate, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req))
	return &pb.CreateDeploymentReply{DeploymentId: deploymentID}, nil
}

func (s *DeploymentService) hasUnfinishedDeployment(ctx context.Context, node, gateway, deploymentType string) (bool, error) {
	listDeploymentReq := &pb.ListDeploymentReq{
		Node:           node,
		Gateway:        gateway,
		PageNum:        1,
		Size_:          1,
		DeploymentType: deploymentType,
	}
	deployment, err := s.Deployment(ctx, listDeploymentReq)
	if err != nil {
		return true, err
	}
	for _, v := range deployment.Lists {
		if isUnfinishedState(v.State) {
			return true, nil
		}
	}
	return false, nil
}

func checkDigestSimilarity(is []*pb.InstanceStatus) error {
	instanceStatus := make(map[string]*pb.InstanceStatus, len(is))
	for _, v := range is {
		instanceStatus[v.Digest] = v
	}
	if len(instanceStatus) <= 1 {
		return nil
	}
	instances := make([]string, 0, len(instanceStatus))
	for _, v := range instanceStatus {
		instances = append(instances, v.Instance)
	}
	return ecode.Error(ecode.RequestErr, fmt.Sprintf("构建失败，请确保节点配置正常且一致: %+v", instances))
}

func isUnfinishedState(in string) bool {
	return !isDeploymentFinishState.Has(in)
}

func (s *DeploymentService) CompareDeployment(ctx context.Context, req *pb.DeploymentReq) (*pb.CompareDeploymentReply, error) {
	deploymentMeta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := checkDeploymentType(req.DeploymentType); err != nil {
		return nil, err
	}
	dynPaths := []*pb.DynPath{}
	breakerAPIs := []*pb.BreakerAPI{}
	eg := errgroup.WithContext(ctx)
	switch req.DeploymentType {
	case _httpType:
		func() {
			eg.Go(func(ctx context.Context) error {
				reply, err := s.http.resourceDao.ListDynPath(ctx, req.Node, req.Gateway)
				if err != nil {
					return err
				}
				dynPaths = reply
				return nil
			})
			eg.Go(func(ctx context.Context) error {
				reply, err := s.http.resourceDao.ListBreakerAPI(ctx, req.Node, req.Gateway)
				if err != nil {
					return err
				}
				breakerAPIs = reply
				return nil
			})
		}()
	case _grpcType:
		func() {
			eg.Go(func(ctx context.Context) error {
				reply, err := s.grpc.resourceDao.ListDynPath(ctx, req.Node, req.Gateway)
				if err != nil {
					return err
				}
				dynPaths = reply
				return nil
			})
			eg.Go(func(ctx context.Context) error {
				reply, err := s.grpc.resourceDao.ListBreakerAPI(ctx, req.Node, req.Gateway)
				if err != nil {
					return err
				}
				breakerAPIs = reply
				return nil
			})
		}()
	default:
		panic(fmt.Sprintf("No matched deployment type: %s", req.DeploymentType))
	}
	snapshotProfileReq := &pb.SnapshotProfileReq{
		Node:    req.Node,
		Gateway: req.Gateway,
		Uuid:    deploymentMeta.SnapshotUuid,
	}
	var snapshotProfile *pb.SnapshotProfileReply
	eg.Go(func(ctx context.Context) error {
		reply, err := s.SnapshotProfile(ctx, snapshotProfileReq)
		if err != nil {
			return err
		}
		snapshotProfile = reply
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	snapshotDynPaths, snapshotBreakerAPIs := constructSnapshotProfile(breakerAPIs, dynPaths, snapshotProfile, req.DeploymentType)
	out := &pb.CompareDeploymentReply{
		Id:      req.DeploymentId,
		Node:    req.Node,
		Gateway: req.Gateway,
		Type:    deploymentMeta.DeploymentType,
		OldConfigs: &pb.OldConfigs{
			BreakerApiList: breakerAPIs,
			DynPathList:    dynPaths,
		},
		NewConfigs: &pb.NewConfigs{
			SnapshotDynPaths:    snapshotDynPaths,
			SnapshotBreakerApis: snapshotBreakerAPIs,
		},
	}
	return out, nil
}

func checkDeploymentType(deploymentType string) error {
	if deploymentType == _httpType || deploymentType == _grpcType {
		return nil
	}
	return errors.Errorf("No matched deployment type: %s", deploymentType)
}

func (s *DeploymentService) ConfirmDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if meta.State != _deploymentUnconfirmed {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止确认，当前发布单状态为: %s", meta.State))
	}
	listDeploymentReq := &pb.ListDeploymentReq{
		Node:           req.Node,
		Gateway:        req.Gateway,
		PageNum:        1,
		Size_:          1,
		DeploymentType: req.DeploymentType,
	}
	deployment, err := s.Deployment(ctx, listDeploymentReq)
	if err != nil {
		return nil, err
	}
	// 非最新订单关闭处理
	if len(deployment.Lists) == 1 && req.DeploymentId != deployment.Lists[0].DeploymentId {
		dst := constructDstMeta(meta, _deploymentClosed)
		if err = s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
			return nil, err
		}
		return nil, ecode.Error(ecode.RequestErr, "当前不是最新的发布单，请回到详情列表查看最新发布单")
	}
	isConfirmed, err := s.dao.DeploymentIsConfirmed(ctx, req)
	if err != nil {
		return nil, err
	}
	// 确认成功但是写状态失败 补偿
	if isConfirmed && meta.State == _deploymentUnconfirmed {
		dst := constructDstMeta(meta, _deploymentDeploying)
		if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
			return nil, err
		}
		meta = dst
	}
	confirm := &pb.DeploymentConfirm{
		Sponsor:     req.Username,
		ConfirmedAt: time.Now().UnixNano(),
	}
	if err := s.dao.SetDeploymentConfirm(ctx, req, confirm); err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("发布单确认失败: %+v", err))
	}
	dst := constructDstMeta(meta, _deploymentDeploying)
	dst.Status.Initialized = true
	if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func constructDstMeta(meta *pb.DeploymentMeta, state string) *pb.DeploymentMeta {
	dst := *meta
	dst.State = state
	dst.UpdatedAt = time.Now().Unix()
	return &dst
}

func (s *DeploymentService) DeployDeploymentProfile(ctx context.Context, req *pb.DeploymentReq) (*pb.DeployDeploymentProfileReply, error) {
	out := &pb.DeployDeploymentProfileReply{
		Id:      req.DeploymentId,
		Node:    req.Node,
		Gateway: req.Gateway,
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.GetDeploymentConfirm(ctx, req)
		if err != nil {
			return err
		}
		out.Confirm = reply.Sponsor
		return nil
	})
	var meta *pb.DeploymentMeta
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.GetDeploymentMeta(ctx, req)
		if err != nil {
			return err
		}
		meta = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.GetDeploymentActionLog(ctx, req)
		if err != nil {
			return err
		}
		out.ActionLog = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.instanceStatus(ctx, req.Node, req.Gateway)
		if err != nil {
			return err
		}
		out.InstanceStatus = reply
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to fetch deployment list data: %+v", err)
		return nil, err
	}
	if err := s.buildDigestAndState(ctx, out.InstanceStatus, meta); err != nil {
		return nil, err
	}
	out.Sponsor = meta.Sponsor
	out.State = meta.State
	out.Description = meta.Description
	out.Status = meta.Status
	out.Type = meta.DeploymentType
	return out, nil
}

func (s *DeploymentService) DeployDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if !isDeploymentDeployState.Has(meta.State) {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止发布，当前发布单状态为：%s", meta.State))
	}
	// 获取所有节点host
	instances, err := s.instanceStatus(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	if err := s.buildDigestAndState(ctx, instances, meta); err != nil {
		return nil, err
	}
	inWaitInstance := FilterInstance(instances, func(in *pb.InstanceStatus) bool {
		return in.State == _instanceWait
	})
	// 节点发布完毕，发布单状态变更失败
	if len(inWaitInstance) == 0 {
		if err = s.deploymentConfigPersist(ctx, req, meta); err != nil {
			return nil, err
		}
		return &empty.Empty{}, nil
	}
	digest, content, err := s.buildConfigDigestAndToml(ctx, meta)
	if err != nil {
		return nil, err
	}
	reloadConfigReq := &model.ReloadConfigReq{
		Host:           inWaitInstance[0].Addr,
		Content:        content,
		Digest:         digest,
		OriginalDigest: inWaitInstance[0].Digest,
		IsGRPC:         isGRPC(req.DeploymentType),
	}
	addActionLogReq := &pb.AddActionLogReq{
		Node:         req.Node,
		Gateway:      req.Gateway,
		DeploymentId: req.DeploymentId,
		ActionLog: pb.ActionLog{
			Instance:  inWaitInstance[0].Instance,
			Action:    "deploy",
			Level:     "INFO",
			CreatedAt: time.Now().UnixNano(),
			Sponsor:   req.Username,
		},
	}
	reloadReply, err := s.dao.ReloadConfig(ctx, reloadConfigReq)
	if err != nil {
		addActionLogReq.ActionLog.Level = "ERROR"
		s.dao.AddActionLog(ctx, addActionLogReq)
		return nil, err
	}
	if reloadReply.Digest != digest {
		addActionLogReq.ActionLog.Level = "WARN"
	}
	s.dao.AddActionLog(ctx, addActionLogReq)
	if !meta.Status.SingleDeployed {
		dst := constructDstMeta(meta, _deploymentDeploying)
		dst.Status.SingleDeployed = true
		if err = s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
			return nil, err
		}
		meta = dst
	}
	// 发布到最后一个节点时，进行持久化操作
	if len(inWaitInstance) == 1 {
		if err = s.deploymentConfigPersist(ctx, req, meta); err != nil {
			return nil, err
		}
	}
	return &empty.Empty{}, nil
}

func (s *DeploymentService) deploymentConfigPersist(ctx context.Context, req *pb.DeploymentReq, meta *pb.DeploymentMeta) error {
	if err := s.deployAndExecuteTask(ctx, req, meta); err != nil {
		return err
	}
	dst := constructDstMeta(meta, _deploymentPersistence)
	dst.Status.SingleDeployed = true
	dst.Status.Deployed = true
	dst.Status.Persisted = true
	if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
		return err
	}
	return nil
}

func (s *DeploymentService) buildConfigDigestAndToml(ctx context.Context, meta *pb.DeploymentMeta) (string, string, error) {
	switch meta.DeploymentType {
	case _httpType:
		return s.buildHTTPConfigDigestAndToml(ctx, meta)
	case _grpcType:
		return s.buildGRPCConfigDigestAndToml(ctx, meta)
	default:
		return "", "", ecode.Error(ecode.RequestErr, fmt.Sprintf("No matched deployment type: %s", meta.DeploymentType))
	}
}

func (s *DeploymentService) buildHTTPConfigDigestAndToml(ctx context.Context, meta *pb.DeploymentMeta) (string, string, error) {
	cfg, err := s.proxyConfig(ctx, meta.Node, meta.Gateway, meta.SnapshotUuid)
	if err != nil {
		return "", "", err
	}
	raw := struct {
		ProxyConfig *sdk.Config
	}{
		ProxyConfig: cfg,
	}
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(raw); err != nil {
		return "", "", err
	}
	return cfg.Digest(), tomlBuf.String(), nil
}

func (s *DeploymentService) buildGRPCConfigDigestAndToml(ctx context.Context, meta *pb.DeploymentMeta) (string, string, error) {
	cfg, err := s.grpcProxyConfig(ctx, meta.Node, meta.Gateway, meta.SnapshotUuid)
	if err != nil {
		return "", "", err
	}
	raw := struct {
		ProxyConfig *sdkwarden.Config
	}{
		ProxyConfig: cfg,
	}
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(raw); err != nil {
		return "", "", err
	}
	return cfg.Digest(), tomlBuf.String(), nil
}

type FilterFunc func(in *pb.InstanceStatus) bool

func FilterInstance(slice []*pb.InstanceStatus, filter FilterFunc) []*pb.InstanceStatus {
	out := []*pb.InstanceStatus{}
	for _, v := range slice {
		if filter(v) {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Instance < out[j].Instance
	})
	return out
}

func (s *DeploymentService) instanceStatus(ctx context.Context, node, appName string) ([]*pb.InstanceStatus, error) {
	gw, err := s.common.getGateway(ctx, node, appName)
	if err != nil {
		return nil, err
	}
	color := ""
	if gw.DiscoveryAppid != "" {
		node = nodeDir(gw.DiscoveryAppid)
		_appName, ok := nodeAppName(gw.DiscoveryAppid)
		if !ok {
			return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("Invalid discovery appid: %q", gw.DiscoveryAppid))
		}
		appName = _appName
		color = gw.DiscoveryColor
	}
	appID := fmt.Sprintf("%s.%s", node, appName)
	instances, err := resolveFromDiscovery(ctx, appID, color)
	if err != nil {
		return nil, err
	}
	instanceStatus := make([]*pb.InstanceStatus, 0)
	for _, ists := range instances {
		for _, ist := range ists {
			instance := &pb.InstanceStatus{
				Instance:  ist.Hostname,
				Addr:      matchHttpUrl(ist.Addrs),
				CreatedAt: ist.LastTs,
			}
			instanceStatus = append(instanceStatus, instance)
		}
	}
	sort.SliceStable(instanceStatus, func(i, j int) bool {
		return instanceStatus[i].Instance < instanceStatus[j].Instance
	})
	return instanceStatus, nil
}

func (s *DeploymentService) buildDigestAndState(ctx context.Context, instanceStatus []*pb.InstanceStatus, meta *pb.DeploymentMeta) error {
	switch meta.DeploymentType {
	case _httpType:
		return s.buildHTTPDigestAndState(ctx, instanceStatus, meta)
	case _grpcType:
		return s.buildGRPCDigestAndState(ctx, instanceStatus, meta)
	default:
		return ecode.Error(ecode.RequestErr, fmt.Sprintf("No matched deployment type: %s", meta.DeploymentType))
	}
}

func (s *DeploymentService) buildHTTPDigestAndState(ctx context.Context, instanceStatus []*pb.InstanceStatus, meta *pb.DeploymentMeta) error {
	cfg, err := s.proxyConfig(ctx, meta.Node, meta.Gateway, meta.SnapshotUuid)
	if err != nil {
		return err
	}
	digest := cfg.Digest()
	eg := errgroup.WithContext(ctx)
	for _, v := range instanceStatus {
		v := v
		eg.Go(func(ctx context.Context) error {
			gatewayProfile, err := s.dao.GatewayProfile(ctx, v.Addr, false)
			if err != nil {
				return err
			}
			v.Digest = gatewayProfile.ConfigDigest
			v.State = constructInstanceState(v.Digest, digest, meta)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *DeploymentService) buildGRPCDigestAndState(ctx context.Context, instanceStatus []*pb.InstanceStatus, meta *pb.DeploymentMeta) error {
	grpcCfg, err := s.grpcProxyConfig(ctx, meta.Node, meta.Gateway, meta.SnapshotUuid)
	if err != nil {
		return err
	}
	digest := grpcCfg.Digest()
	eg := errgroup.WithContext(ctx)
	for _, v := range instanceStatus {
		v := v
		eg.Go(func(ctx context.Context) error {
			gatewayProfile, err := s.dao.GatewayProfile(ctx, v.Addr, true)
			if err != nil {
				return err
			}
			v.Digest = gatewayProfile.ConfigDigest
			v.State = constructInstanceState(v.Digest, digest, meta)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func constructInstanceState(gatewayDigest, snapshotDigest string, meta *pb.DeploymentMeta) string {
	if meta.State == _deploymentFinished {
		return _instanceDone
	}
	if meta.State == _deploymentRollback {
		return _instanceRollback
	}
	if gatewayDigest == snapshotDigest {
		return _instanceDone
	}
	return _instanceWait
}

func (s *DeploymentService) DeployDeploymentAll(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	deploymentMeta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if !isDeploymentDeployState.Has(deploymentMeta.State) {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止发布，当前发布单状态为：%s", deploymentMeta.State))
	}
	// 获取所有节点host
	instances, err := s.instanceStatus(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	if err := s.buildDigestAndState(ctx, instances, deploymentMeta); err != nil {
		return nil, err
	}
	inWaitInstance := FilterInstance(instances, func(in *pb.InstanceStatus) bool {
		return in.State == _instanceWait
	})
	digest, content, err := s.buildConfigDigestAndToml(ctx, deploymentMeta)
	if err != nil {
		return nil, err
	}
	eg := errgroup.WithCancel(ctx)
	for _, v := range inWaitInstance {
		v := v
		eg.Go(func(ctx context.Context) error {
			reloadConfigReq := &model.ReloadConfigReq{
				Host:           v.Addr,
				Content:        content,
				Digest:         digest,
				OriginalDigest: v.Digest,
				IsGRPC:         isGRPC(req.DeploymentType),
			}
			addActionLogReq := &pb.AddActionLogReq{
				Node:         req.Node,
				Gateway:      req.Gateway,
				DeploymentId: req.DeploymentId,
				ActionLog: pb.ActionLog{
					Instance:  v.Instance,
					Action:    "deploy",
					Level:     "INFO",
					CreatedAt: time.Now().UnixNano(),
					Sponsor:   req.Username,
				},
			}
			reloadReply, err := s.dao.ReloadConfig(ctx, reloadConfigReq)
			if err != nil {
				addActionLogReq.ActionLog.Level = "ERROR"
				s.dao.AddActionLog(ctx, addActionLogReq)
				return err
			}
			if reloadReply.Digest != digest {
				addActionLogReq.ActionLog.Level = "WARN"
				s.dao.AddActionLog(ctx, addActionLogReq)
				return nil
			}
			s.dao.AddActionLog(ctx, addActionLogReq)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if err = s.deploymentConfigPersist(ctx, req, deploymentMeta); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func isGRPC(deploymentType string) bool {
	return deploymentType == "grpc"
}

func (s *DeploymentService) FinishDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if meta.State != _deploymentPersistence {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止结单，当前发布单状态为：%s", meta.State))
	}
	dst := constructDstMeta(meta, _deploymentFinished)
	dst.Status.Finished = true
	if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
		return nil, err
	}
	addActionLogReq := &pb.AddActionLogReq{
		Node:         req.Node,
		Gateway:      req.Gateway,
		DeploymentId: req.DeploymentId,
		ActionLog: pb.ActionLog{
			Instance:  "",
			Action:    "finish",
			Level:     "INFO",
			CreatedAt: time.Now().UnixNano(),
			Sponsor:   req.Username,
		},
	}
	s.dao.AddActionLog(ctx, addActionLogReq)
	audit.SendDeploymentLog(meta, audit.LogActionFinish, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req))
	return &empty.Empty{}, nil
}

func (s *DeploymentService) CloseDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if meta.State != _deploymentUnconfirmed {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止关闭，当前发布单状态为：%s", meta.State))
	}
	dst := constructDstMeta(meta, _deploymentClosed)
	if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
		return nil, err
	}
	audit.SendDeploymentLog(meta, audit.LogActionClose, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req))
	return &empty.Empty{}, nil
}

func (s *DeploymentService) CancelDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if meta.State != _deploymentDeploying {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止撤销，当前发布单状态为：%s", meta.State))
	}
	dst := constructDstMeta(meta, _deploymentCancel)
	if err := s.dao.UpdateDeploymentState(ctx, meta, dst); err != nil {
		return nil, err
	}
	addActionLogReq := &pb.AddActionLogReq{
		Node:         req.Node,
		Gateway:      req.Gateway,
		DeploymentId: req.DeploymentId,
		ActionLog: pb.ActionLog{
			Instance:  "",
			Action:    "cancel",
			Level:     "INFO",
			CreatedAt: time.Now().UnixNano(),
			Sponsor:   req.Username,
		},
	}
	s.dao.AddActionLog(ctx, addActionLogReq)
	return &empty.Empty{}, nil
}

func (s *DeploymentService) RollbackDeployment(ctx context.Context, req *pb.DeploymentReq) (*empty.Empty, error) {
	meta, err := s.dao.GetDeploymentMeta(ctx, req)
	if err != nil {
		return nil, err
	}
	if !canRollbackState.Has(meta.State) {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("禁止回滚，当前发布单状态为：%s", meta.State))
	}
	if meta.Status.Persisted {
		if err := s.snapshotDeploy(ctx, req.Node, req.Gateway, meta.RollbackUuid, meta.DeploymentType); err != nil {
			return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("回滚失败 %+v", err))
		}
	}
	taskReq := &pb.ExecuteTaskReq{
		Node:     req.Node,
		Gateway:  req.Gateway,
		Task:     matchTask(req.DeploymentType),
		Username: req.Username,
	}
	if _, err := s.common.ExecuteTask(ctx, taskReq); err != nil {
		return nil, err
	}
	meta.State = _deploymentRollback
	meta.Status.Rollbacked = true
	meta.UpdatedAt = time.Now().Unix()
	if err := s.dao.SetDeploymentMeta(ctx, meta); err != nil {
		return nil, err
	}
	addActionLogReq := &pb.AddActionLogReq{
		Node:         req.Node,
		Gateway:      req.Gateway,
		DeploymentId: req.DeploymentId,
		ActionLog: pb.ActionLog{
			Action:    "rollback",
			Level:     "WARN",
			CreatedAt: time.Now().UnixNano(),
			Sponsor:   req.Username,
		},
	}
	s.dao.AddActionLog(ctx, addActionLogReq)
	return &empty.Empty{}, nil
}

func (s *DeploymentService) proxyConfig(ctx context.Context, node, gateway, uuid string) (*sdk.Config, error) {
	req := &pb.SnapshotProfileReq{
		Node:    node,
		Gateway: gateway,
		Uuid:    uuid,
	}
	profile, err := s.SnapshotProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	dynPaths := common.EnabledDynPath(profile.DynPath.DynPaths)
	if len(dynPaths) == 0 {
		return nil, errors.New("all dynPath are disable")
	}
	pm := common.BuildPathMetaByDynPath(dynPaths)
	breakerAPIs := common.EnabledBreakerAPI(profile.BreakerApi.BreakerApiList)
	quotaMethods := common.EnabledQuotaMethod(profile.QuotaMethod)
	out, err := common.RunProcess(pm, common.PathMetaAppendBreakerAPIs(breakerAPIs),
		common.PathMetaAppendRateLimiter(quotaMethods))
	if err != nil {
		return nil, err
	}
	reply := &sdk.Config{DynPath: out}
	return reply, nil
}

func (s *DeploymentService) grpcProxyConfig(ctx context.Context, node, gateway, uuid string) (*sdkwarden.Config, error) {
	req := &pb.SnapshotProfileReq{
		Node:    node,
		Gateway: gateway,
		Uuid:    uuid,
	}
	profile, err := s.SnapshotProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	dynService := common.EnabledDynPath(profile.GrpcDynPath.DynPaths)
	if len(dynService) == 0 {
		return nil, errors.New("all dynService are disable")
	}
	sm := common.BuildServiceMetaByDynPath(dynService)
	breakerAPIs := common.EnabledBreakerAPI(profile.GrpcBreakerApi.BreakerApiList)
	out, err := common.RunGRPCProcess(sm, common.ServiceMetaAppendBreakerAPIs(breakerAPIs))
	if err != nil {
		return nil, err
	}
	reply := &sdkwarden.Config{
		DynService: out,
	}
	return reply, nil
}

func constructSnapshotProfile(bas []*pb.BreakerAPI, dps []*pb.DynPath, snapshotProfile *pb.SnapshotProfileReply, deploymentType string) ([]*pb.SnapshotDynPath, []*pb.SnapshotBreakerApi) {
	baMap := make(map[string]*pb.BreakerAPI, len(bas))
	for _, v := range bas {
		baMap[v.Api] = v
	}
	dpsMap := make(map[string]*pb.DynPath, len(dps))
	for _, v := range dps {
		dpsMap[v.Pattern] = v
	}
	breakerAPIs := constructBreakerAPIs(baMap, snapshotProfile, deploymentType)
	dynPaths := constructDynPath(dpsMap, snapshotProfile, deploymentType)
	return dynPaths, breakerAPIs
}

func constructBreakerAPIs(baMap map[string]*pb.BreakerAPI, snapshotProfile *pb.SnapshotProfileReply, deploymentType string) []*pb.SnapshotBreakerApi {
	out := []*pb.SnapshotBreakerApi{}
	//nolint:ineffassign,staticcheck
	list := []*pb.BreakerAPI{}
	switch deploymentType {
	case _httpType:
		list = snapshotProfile.BreakerApi.BreakerApiList
	case _grpcType:
		list = snapshotProfile.GrpcBreakerApi.BreakerApiList
	default:
		log.Error("No matched deployment type: %s", deploymentType)
		return out
	}
	for _, ba := range list {
		bapi, ok := baMap[ba.Api]
		if !ok {
			breakerAPI := &pb.SnapshotBreakerApi{
				BreakerApi: ba,
				HasChanged: true,
			}
			out = append(out, breakerAPI)
			continue
		}
		breakerAPI := &pb.SnapshotBreakerApi{
			BreakerApi: ba,
			HasChanged: !model.MatchBreakerAPI(bapi, ba),
		}
		out = append(out, breakerAPI)
	}
	return out
}

func constructDynPath(dpsMap map[string]*pb.DynPath, snapshotProfile *pb.SnapshotProfileReply, deploymentType string) []*pb.SnapshotDynPath {
	out := []*pb.SnapshotDynPath{}
	//nolint:ineffassign,staticcheck
	list := []*pb.DynPath{}
	switch deploymentType {
	case _httpType:
		list = snapshotProfile.DynPath.DynPaths
	case _grpcType:
		list = snapshotProfile.GrpcDynPath.DynPaths
	default:
		log.Error("No matched deployment type: %s", deploymentType)
		return out
	}
	for _, dp := range list {
		val, ok := dpsMap[dp.Pattern]
		if !ok {
			dynPath := &pb.SnapshotDynPath{
				DynPath:    dp,
				HasChanged: true,
			}
			out = append(out, dynPath)
			continue
		}
		dynPath := &pb.SnapshotDynPath{
			DynPath:    dp,
			HasChanged: !model.MatchDynPath(dp, val),
		}
		out = append(out, dynPath)
	}
	return out
}

func (s *DeploymentService) deployAndExecuteTask(ctx context.Context, req *pb.DeploymentReq, meta *pb.DeploymentMeta) error {
	addActionLogReq := &pb.AddActionLogReq{
		Node:         req.Node,
		Gateway:      req.Gateway,
		DeploymentId: req.DeploymentId,
		ActionLog: pb.ActionLog{
			Instance:  "",
			Action:    "push",
			Level:     "INFO",
			CreatedAt: time.Now().UnixNano(),
			Sponsor:   req.Username,
		},
	}
	if err := s.snapshotDeploy(ctx, req.Node, req.Gateway, meta.SnapshotUuid, meta.DeploymentType); err != nil {
		log.Error("Failed to set snapshot to sys: %+v", err)
		addActionLogReq.ActionLog.Level = "ERROR"
		s.dao.AddActionLog(ctx, addActionLogReq)
		return err
	}
	taskReq := &pb.ExecuteTaskReq{
		Node:     req.Node,
		Gateway:  req.Gateway,
		Task:     matchTask(req.DeploymentType),
		Username: req.Username,
	}
	if _, err := s.common.ExecuteTask(ctx, taskReq); err != nil {
		return err
	}
	s.dao.AddActionLog(ctx, addActionLogReq)
	return nil
}

func matchTask(req string) string {
	if req == _httpType {
		return "pushSingle"
	}
	return "grpcPushSingle"
}
