package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming"
	"go-common/library/naming/discovery"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"
	"go-gateway/app/app-svr/app-gw/management/internal/model/tree"

	"go-common/library/sync/errgroup.v2"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func makeProjectSet(nodes []*tree.Node) sets.String {
	projectSet := sets.NewString()
	for _, n := range nodes {
		projectSet.Insert(nodeDir(n.Path))
	}
	return projectSet
}

func makeAppSet(nodes []*tree.Node) map[string]sets.String {
	appSet := map[string]sets.String{}
	for _, n := range nodes {
		node := nodeDir(n.Path)
		if _, ok := appSet[node]; !ok {
			appSet[node] = sets.NewString()
		}
		name, ok := nodeAppName(n.Path)
		if !ok {
			continue
		}
		appSet[node].Insert(name)
	}
	return appSet
}

func (s *CommonService) permittedNode(ctx context.Context, username, cookie, node string) error {
	nodes, err := s.dao.FetchRoleTree(ctx, username, cookie)
	if err != nil {
		return err
	}
	projectSet := makeProjectSet(nodes)
	if !projectSet.Has(node) {
		return ecode.Error(ecode.AccessDenied, fmt.Sprintf("denied on %s to %s", username, node))
	}
	return nil
}

// TOOD: simplify code with `permittedNode`
func (s *CommonService) permittedApp(ctx context.Context, username, cookie, node, appName string) error {
	nodes, err := s.dao.FetchRoleTree(ctx, username, cookie)
	if err != nil {
		return err
	}
	projectSet := makeProjectSet(nodes)
	if !projectSet.Has(node) {
		return ecode.Error(ecode.AccessDenied, fmt.Sprintf("denied on %s to %s", username, node))
	}

	appSet := makeAppSet(nodes)
	nodeAppSet, ok := appSet[node]
	if ok {
		if !nodeAppSet.Has(appName) {
			return ecode.Error(ecode.AccessDenied, fmt.Sprintf("denied on %s to %s.%s", username, node, appName))
		}
	}
	return nil
}

// Gateway is
func (s *CommonService) Gateway(ctx context.Context, req *pb.AuthZReq) (*pb.GatewayReply, error) {
	reply := &pb.GatewayReply{
		Gateways: []*pb.Gateway{},
	}
	if err := s.permittedNode(ctx, req.Username, req.Cookie, req.Node); err != nil {
		return nil, err
	}

	gateways, err := s.dao.ListGateway(ctx, req.Node)
	if err != nil {
		return nil, err
	}

	// filter gateway by regular tree node
	nodes, err := s.dao.FetchRoleTree(ctx, req.Username, req.Cookie)
	if err != nil {
		return nil, err
	}
	appSet := makeAppSet(nodes)
	nodeAppSet := appSet[req.Node]
	filtered := make([]*pb.Gateway, 0, len(nodes))
	for _, gw := range gateways {
		if !nodeAppSet.Has(gw.AppName) {
			continue
		}
		filtered = append(filtered, gw)
	}
	reply.Gateways = filtered
	return reply, nil
}

func (s *CommonService) getGateway(ctx context.Context, node, appName string) (*pb.Gateway, error) {
	gateways, err := s.dao.ListGateway(ctx, node)
	if err != nil {
		return nil, err
	}
	for _, gw := range gateways {
		if gw.Node == node && gw.AppName == appName {
			return gw, nil
		}
	}
	return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("no such gateway: %s.%s", node, appName))
}

func resolveFromDiscovery(ctx context.Context, appID, color string) (map[string][]*naming.Instance, error) {
	resolver := discovery.Build(appID)
	defer resolver.Close()

	// initial resolve
	event := resolver.Watch()
	select {
	case _, ok := <-event:
		if !ok {
			return nil, errors.Errorf("unexpected event on resolve %q", appID)
		}
		instances, ok := resolver.Fetch(ctx)
		if !ok {
			return nil, errors.Errorf("failed to fetch instance: %q", appID)
		}
		colorFiltered := func() {
			filtered := map[string][]*naming.Instance{}
			for zone, zoneInstances := range instances {
				if _, ok := filtered[zone]; !ok {
					filtered[zone] = []*naming.Instance{}
				}
				for _, instance := range zoneInstances {
					if instance.Metadata[naming.MetaCluster] != color {
						continue
					}
					filtered[zone] = append(filtered[zone], instance)
				}
			}
			instances = filtered
		}
		if color != "" {
			colorFiltered()
		}
		return instances, nil
	case <-ctx.Done():
		return nil, errors.Errorf("resolve %q deadline execeeded", appID)
	}
	//nolint:govet
	return nil, errors.Errorf("unexpected error on discovery resolve: %q", appID)
}

func (s *CommonService) fetchWorkerNodes(ctx context.Context, node, appName, color, currentHost string) ([]*pb.WorkerNode, error) {
	appID := fmt.Sprintf("%s.%s", node, appName)
	instances, err := resolveFromDiscovery(ctx, appID, color)
	if err != nil {
		return nil, err
	}
	nodes := []*pb.WorkerNode{}
	eg := errgroup.WithContext(ctx)
	for _, ists := range instances {
		for _, ist := range ists {
			advertiseAddr := matchHttpUrl(ist.Addrs)
			tokenPayload := &model.JWTTokenPayload{
				Addr:    advertiseAddr,
				Node:    node,
				Gateway: appName,
			}
			node := &pb.WorkerNode{
				Hostname:      ist.Hostname,
				Zone:          ist.Zone,
				AdvertiseAddr: advertiseAddr,
				Addrs:         ist.Addrs,
				Status:        ist.Status,
				// RegTimestamp:    ist.RegTimestamp,
				// UpTimestamp:     ist.UpTimestamp,
				// RenewTimestamp:  ist.RenewTimestamp,
				// DirtyTimestamp:  ist.DirtyTimestamp,
				LatestTimestamp: ist.LastTs,
				Version:         ist.Version,
				Metadata:        ist.Metadata,
				MonitorUrl:      s.formatURL(ctx, tokenPayload, currentHost, "/metrics"),
				ConfigApi:       s.formatURL(ctx, tokenPayload, currentHost, "/configs.toml"),
				GrpcConfigApi:   s.formatURL(ctx, tokenPayload, currentHost, "/grpc-configs.toml"),
			}
			nodes = append(nodes, node)
			eg.Go(func(ctx context.Context) error {
				gatewayProfile, err := s.dao.GatewayProfile(ctx, advertiseAddr, false)
				if err != nil {
					log.Error("Failed to fetch gateway profile: %+v", err)
					return nil
				}
				node.GatewayVersion = gatewayProfile.GatewayVersion
				node.SdkVersion = gatewayProfile.SDKVersion
				node.ConfigDigest = gatewayProfile.ConfigDigest
				return nil
			})
			eg.Go(func(ctx context.Context) error {
				grpcProfile, err := s.dao.GatewayProfile(ctx, advertiseAddr, true)
				if err != nil {
					log.Error("Failed to fetch gateway grpc profile: %+v", err)
					return nil
				}
				node.GrpcConfigDigest = grpcProfile.ConfigDigest
				return nil
			})
		}
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	return nodes, nil
}

func matchHttpUrl(addrs []string) string {
	for _, v := range addrs {
		if !strings.Contains(v, _httpType) {
			continue
		}
		return v
	}
	return ""
}

func (s *CommonService) formatURL(ctx context.Context, req *model.JWTTokenPayload, host, suffix string) string {
	secret, err := s.dao.TokenSecret(ctx)
	if err != nil {
		log.Error("Failed to get token:%s %+v", suffix, err)
		return ""
	}
	token, err := s.newToken(req, secret)
	if err != nil {
		log.Error("Failed to create token:%s %+v", suffix, err)
		return ""
	}
	return buildProxyURL(host, token, suffix)
}

func buildProxyURL(host, token, suffix string) string {
	baseURL := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path.Join("/x/admin/app-gw/gateway/proxy", token, suffix),
	}
	return baseURL.String()
}

// GatewayProfile is
func (s *CommonService) GatewayProfile(ctx context.Context, req *pb.GatewayProfileReq) (*pb.GatewayProfileReply, error) {
	if err := s.permittedApp(ctx, req.Username, req.Cookie, req.Node, req.AppName); err != nil {
		return nil, err
	}
	gw, err := s.getGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return nil, err
	}

	node := req.Node
	appName := gw.AppName
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

	nodes, err := s.fetchWorkerNodes(ctx, node, appName, color, req.Host)
	if err != nil {
		return nil, err
	}
	workerNodes := &pb.WorkerNodes{
		Nodes: nodes,
	}
	workerNodes.GatewayVersion, workerNodes.SdkVersion = profileVersion(nodes)

	out := &pb.GatewayProfileReply{
		ProjectName: gw.ProjectName,
		AppName:     gw.AppName,
		Node:        gw.Node,
		TreeId:      gw.TreeId,
		Zones:       matchZones(gw.Configs),
		Configs:     gw.Configs,
		GrpcConfigs: gw.GrpcConfigs,
		UpdatedAt:   gw.UpdatedAt,
		Envs:        matchEnvs(gw.Configs),
		WorkerNodes: workerNodes,
	}
	return out, nil
}

func matchEnvs(configs []*pb.ConfigMeta) []string {
	envSet := make(sets.String)
	for _, config := range configs {
		envSet.Insert(config.Env)
	}
	return envSet.List()
}

func matchZones(configMeta []*pb.ConfigMeta) []string {
	zoneSet := make(sets.String)
	for _, config := range configMeta {
		zoneSet.Insert(config.Zone)
	}
	return zoneSet.List()
}

func profileVersion(nodes []*pb.WorkerNode) ([]string, []string) {
	gatewayVersionSet := make(sets.String)
	sdkVersionSet := make(sets.String)
	for _, node := range nodes {
		if node.GatewayVersion != "" {
			gatewayVersionSet.Insert(node.GatewayVersion)
		}
		if node.SdkVersion != "" {
			sdkVersionSet.Insert(node.SdkVersion)
		}
	}
	return gatewayVersionSet.List(), sdkVersionSet.List()
}

func (s *CommonService) hasGateway(ctx context.Context, node, gateway string) (bool, error) {
	gateways, err := s.dao.ListGateway(ctx, node)
	if err != nil {
		return false, err
	}
	for _, gw := range gateways {
		if gw.AppName == gateway {
			return true, nil
		}
	}
	return false, nil
}

// AddGateway is
func (s *CommonService) AddGateway(ctx context.Context, req *pb.SetGatewayReq) (*empty.Empty, error) {
	exist, err := s.hasGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("gateway %s is exist", req.AppName))
	}
	if err := s.dao.SetGateway(ctx, req); err != nil {
		audit.SendSetGatewayLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetGatewayLog(req, audit.LogActionAdd, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

// UpdateGateway is
func (s *CommonService) UpdateGateway(ctx context.Context, req *pb.SetGatewayReq) (*empty.Empty, error) {
	exist, err := s.hasGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("gateway %q is not exist", req.AppName))
	}
	if err := s.dao.SetGateway(ctx, req); err != nil {
		audit.SendSetGatewayLog(req, audit.LogActionUpdate, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetGatewayLog(req, audit.LogActionUpdate, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *CommonService) DeleteGateway(ctx context.Context, req *pb.DeleteGatewayReq) (*empty.Empty, error) {
	if err := s.dao.DeleteGateway(ctx, req); err != nil {
		audit.SendDeleteGatewayLog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendDeleteGatewayLog(req, audit.LogActionDel, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *CommonService) EnableALLGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) (*empty.Empty, error) {
	exist, err := s.hasGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("gateway %q is not exist", req.AppName))
	}
	if err := s.dao.EnableALLGatewayConfig(ctx, req); err != nil {
		audit.SendUpdateALLGatewayConfigLog(req, action(req.Disable), audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendUpdateALLGatewayConfigLog(req, action(req.Disable), audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *CommonService) DisableALLGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableALLGatewayConfig(ctx, req)
}

func action(disable bool) string {
	if disable {
		return audit.LogActionDisable
	}
	return audit.LogActionEnable
}

func (s *CommonService) GatewayProxy(ctx context.Context, req *pb.GatewayProxyReq) (*pb.GatewayProxyReply, error) {
	secret, err := s.dao.TokenSecret(ctx)
	if err != nil {
		return nil, err
	}
	tokenPayload, err := s.parseToken(req.Token, secret)
	if err != nil {
		return nil, err
	}
	proxyPage, err := s.dao.ProxyPage(ctx, tokenPayload.Addr, req.Suffix)
	if err != nil {
		return nil, errors.Errorf("Failed to proxy page: %s %s %s %+v", tokenPayload.Node, tokenPayload.Gateway, tokenPayload.Addr, err)
	}
	return proxyPage, nil
}

func (s *CommonService) InitGatewayConfigs(ctx context.Context, req *pb.InitGatewayConfigsReq) (*empty.Empty, error) {
	//创建版本
	createConfigBuildReq := &model.CreateConfigBuildReq{
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		TreeId:    req.TreeId,
		Cookie:    req.Cookie,
	}
	if err := s.dao.CreateGatewayConfigBuild(ctx, createConfigBuildReq); err != nil {
		return nil, err
	}
	//获取版本信息
	fetchConfigBuildMetaReq := &model.FetchConfigBuildMetaReq{
		TreeId: req.TreeId,
		Cookie: req.Cookie,
		Env:    createConfigBuildReq.Env,
	}
	configBuildMeta, err := s.fetchConfigBuildMeta(ctx, fetchConfigBuildMetaReq)
	if err != nil {
		return nil, err
	}
	//创建配置文件并发版
	if err := s.createProxyConfigFile(ctx, req, configBuildMeta.Token); err != nil {
		return nil, err
	}
	if err := s.createHTTPConfigFile(ctx, req, configBuildMeta.Token); err != nil {
		return nil, err
	}
	if err := s.createGRPCConfigFile(ctx, req, configBuildMeta.Token); err != nil {
		return nil, err
	}
	if err := s.createApplicationConfigFile(ctx, req, configBuildMeta.Token); err != nil {
		return nil, err
	}
	if err := s.createGRPCProxyConfigFile(ctx, req, configBuildMeta.Token); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *CommonService) fetchConfigBuildMeta(ctx context.Context, req *model.FetchConfigBuildMetaReq) (*model.ConfigBuildMeta, error) {
	configBuildDetail, err := s.dao.FetchConfig(ctx, req.TreeId, req.Cookie)
	if err != nil {
		return nil, err
	}
	metas := make(map[string]*model.ConfigBuildMeta, len(configBuildDetail))
	for _, item := range configBuildDetail {
		key := fmt.Sprintf("%s,%s", item.Name, item.Env)
		metas[key] = &model.ConfigBuildMeta{
			Token: item.AppItem.Token,
		}
	}
	key := fmt.Sprintf("%s,%s", "docker-1", env.DeployEnv)
	meta, ok := metas[key]
	if !ok {
		return nil, errors.Errorf("Failed to fetch configBuildMeta %q.", key)
	}
	return meta, nil
}

func formatAppID(node, gateway string) string {
	return fmt.Sprintf("%s.%s", node, gateway)
}

func (s *CommonService) createProxyConfigFile(ctx context.Context, req *pb.InitGatewayConfigsReq, token string) error {
	configMeta := &pb.ConfigMeta{
		Token:     token,
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		Filename:  "proxy-config.toml",
	}
	pcReq := &model.AddConfigFileReq{
		AppID:      formatAppID(req.Node, req.AppName),
		TreeID:     req.TreeId,
		ConfigMeta: configMeta,
		Buffer:     []byte("[ProxyConfig]"),
	}
	if err := s.dao.AddGatewayConfigFile(ctx, pcReq); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) createApplicationConfigFile(ctx context.Context, req *pb.InitGatewayConfigsReq, token string) error {
	configMeta := &pb.ConfigMeta{
		Token:     token,
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		Filename:  "application.toml",
	}
	pcReq := &model.AddConfigFileReq{
		AppID:      formatAppID(req.Node, req.AppName),
		TreeID:     req.TreeId,
		ConfigMeta: configMeta,
		Buffer:     []byte("#"),
	}
	if err := s.dao.AddGatewayConfigFile(ctx, pcReq); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) createHTTPConfigFile(ctx context.Context, req *pb.InitGatewayConfigsReq, token string) error {
	configMeta := &pb.ConfigMeta{
		Token:     token,
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		Filename:  "http.toml",
	}
	pcReq := &model.AddConfigFileReq{
		AppID:      formatAppID(req.Node, req.AppName),
		TreeID:     req.TreeId,
		ConfigMeta: configMeta,
		Buffer: []byte(`[Server]
addr = "0.0.0.0:8000"
timeout = "1s"`),
	}
	if err := s.dao.AddGatewayConfigFile(ctx, pcReq); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) createGRPCConfigFile(ctx context.Context, req *pb.InitGatewayConfigsReq, token string) error {
	configMeta := &pb.ConfigMeta{
		Token:     token,
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		Filename:  "grpc.toml",
	}
	pcReq := &model.AddConfigFileReq{
		AppID:      formatAppID(req.Node, req.AppName),
		TreeID:     req.TreeId,
		ConfigMeta: configMeta,
		Buffer:     []byte("[Server]"),
	}
	if err := s.dao.AddGatewayConfigFile(ctx, pcReq); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) createGRPCProxyConfigFile(ctx context.Context, req *pb.InitGatewayConfigsReq, token string) error {
	configMeta := &pb.ConfigMeta{
		Token:     token,
		Env:       env.DeployEnv,
		Zone:      env.Zone,
		BuildName: "docker-1",
		Filename:  "grpc-proxy-config.toml",
	}
	pcReq := &model.AddConfigFileReq{
		AppID:      formatAppID(req.Node, req.AppName),
		TreeID:     req.TreeId,
		ConfigMeta: configMeta,
		Buffer:     []byte("[ProxyConfig]"),
	}
	if err := s.dao.AddGatewayConfigFile(ctx, pcReq); err != nil {
		return err
	}
	return nil
}

func (s *CommonService) EnableAllGRPCGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) (*empty.Empty, error) {
	exist, err := s.hasGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("gateway %q is not exist", req.AppName))
	}
	if err := s.dao.EnableAllGRPCGatewayConfig(ctx, req); err != nil {
		audit.SendUpdateALLGatewayConfigLog(req, action(req.Disable), audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendUpdateALLGatewayConfigLog(req, action(req.Disable), audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *CommonService) DisableAllGRPCGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableAllGRPCGatewayConfig(ctx, req)
}
