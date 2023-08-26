package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"go-common/library/log"

	errgroup2 "go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"
	"go-gateway/app/app-svr/app-gw/management/internal/model/tree"
)

const (
	_httpType = "http"
	_grpcType = "grpc"
)

func (s *CommonService) AppPromptAPI(ctx context.Context, req *api.AppPromptAPIReq) (*api.AppPromptAPIReply, error) {
	nodes, err := s.dao.FetchRoleTree(ctx, req.Username, req.Cookie)
	if err != nil {
		return nil, err
	}
	rsp := &api.AppPromptAPIReply{}
	nodeSet := map[string]*api.TreeNode{}
	for _, node := range nodes {
		if req.Node != "" && nodeDir(node.Path) != req.Node {
			continue
		}
		if req.OnlyGateway && !strings.HasSuffix(node.Name, "-gateway") {
			continue
		}
		nodeTmp := buildAPITreeNode(node)
		nodeSet[nodeTmp.Name] = nodeTmp
	}
	for _, node := range nodeSet {
		rsp.Nodes = append(rsp.Nodes, node)
	}
	return rsp, nil
}

func buildAPITreeNode(node *tree.Node) *api.TreeNode {
	t := &api.TreeNode{}
	t.Path = node.Path
	t.Name = node.Name
	t.TreeId = int64(node.TreeID)
	t.DiscoveryId = fmt.Sprintf("discovery://%s", strings.TrimPrefix(node.Path, "bilibili."))
	if node.DiscoveryID != "" {
		t.DiscoveryId = fmt.Sprintf("discovery://%s", node.DiscoveryID)
	}
	return t
}

func (s *CommonService) ConfigPromptAPI(ctx context.Context, req *api.ConfigPromptAPIReq) (*api.ConfigPromptAPIReply, error) {
	detail, err := s.dao.FetchConfig(ctx, req.TreeId, req.Cookie)
	if err != nil {
		return nil, err
	}
	rsp := &api.ConfigPromptAPIReply{}
	if req.Type == "" {
		req.Type = _httpType
	}
	for _, item := range detail {
		env := item.Env
		zone := item.Zone
		token := item.AppItem.Token
		buildName := item.Name
		for _, file := range item.Snapshot.Files {
			if file.IsDelete {
				continue
			}
			if req.Type == _httpType && file.Name != "proxy-config.toml" {
				continue
			}
			if req.Type == _grpcType && file.Name != "grpc-proxy-config.toml" {
				continue
			}
			config := &api.AppConfigItem{
				FileName:  file.Name,
				Env:       env,
				Zone:      zone,
				BuildName: buildName,
				Token:     token,
			}
			rsp.List = append(rsp.List, config)
		}
	}
	return rsp, nil
}

func (s *CommonService) AppPathPromptAPI(ctx context.Context, req *api.AppPathPromptAPIReq) (*api.AppPathPromptAPIReply, error) {
	dynPath, err := s.dao.CreateHTTPResourceDao().ListDynPath(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	set := sets.NewString()
	for _, item := range dynPath {
		if item.ClientInfo == nil {
			continue
		}
		if appID, ok := splitDiscovery(item.ClientInfo.Endpoint); ok {
			set.Insert(appID)
		}
	}
	eg := errgroup2.WithCancel(ctx)
	lock := sync.Mutex{}
	rsp := &api.AppPathPromptAPIReply{}
	pathsTmp := []string{}
	for _, appId := range set.List() {
		reqAppId := appId
		eg.Go(func(ctx context.Context) error {
			paths, err := s.dao.ServerMetadata(ctx, reqAppId)
			if err != nil {
				log.Errorc(ctx, "Failed to fetch server metadata: %s: %+v", reqAppId, err)
				return nil
			}
			paths = removeDebugPath(paths)
			lock.Lock()
			pathsTmp = append(pathsTmp, paths...)
			lock.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "Failed to errgroup.Wait(). error: %+v", err)
	}
	sort.Strings(pathsTmp)
	rsp.Paths = pathsTmp
	return rsp, nil
}

func (s *CommonService) GRPCAppMethodPromptAPI(ctx context.Context, req *api.AppPathPromptAPIReq) (*api.AppPathPromptAPIReply, error) {
	dynPath, err := s.dao.CreateGRPCResourceDao().ListDynPath(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	if len(dynPath) == 0 {
		return &api.AppPathPromptAPIReply{Paths: []string{}}, nil
	}
	set := sets.NewString()
	for _, item := range dynPath {
		if item.ClientInfo == nil {
			continue
		}
		if appID, ok := splitDiscovery(item.ClientInfo.Endpoint); ok {
			set.Insert(appID)
		}
	}
	svrMd, ok := s.dao.GRPCServerMethods(ctx, set.List())
	if !ok {
		return &api.AppPathPromptAPIReply{Paths: []string{}}, nil
	}
	rsp := &api.AppPathPromptAPIReply{}
	pathsTmp := []string{}
	for _, path := range svrMd {
		pathsTmp = append(pathsTmp, path...)
	}
	sort.Strings(pathsTmp)
	rsp.Paths = pathsTmp
	return rsp, nil
}

func (s *CommonService) GRPCAppPackagePromptAPI(ctx context.Context, req *api.GRPCAppPackagePromptAPIReq) (*api.GRPCAppPackagePromptAPIReply, error) {
	set := sets.NewString()
	if appID, ok := splitDiscovery(req.Endpoint); ok {
		set.Insert(appID)
	}
	svrMd, ok := s.dao.GRPCServerPackages(ctx, set.List())
	if !ok {
		return &api.GRPCAppPackagePromptAPIReply{}, nil
	}
	pkg := make(map[string]*api.AppPackageService)
	rsp := &api.GRPCAppPackagePromptAPIReply{Package: pkg}
	for _, pkgs := range svrMd {
		for pkg, services := range pkgs {
			svr, ok := rsp.Package[pkg]
			if !ok {
				svr = &api.AppPackageService{}
				rsp.Package[pkg] = svr
			}
			svr.Services = append(svr.Services, services...)
		}
	}
	return rsp, nil
}

func splitDiscovery(path string) (string, bool) {
	if !strings.HasPrefix(path, "discovery://") {
		return "", false
	}
	return strings.TrimPrefix(path, "discovery://"), true
}

func removeDebugPath(paths []string) []string {
	rsp := []string{}
	for _, path := range paths {
		if strings.HasPrefix(path, "/debug/pprof") || strings.HasPrefix(path, "/monitor/ping") ||
			path == "/metadata" || path == "/metrics" {
			continue
		}
		rsp = append(rsp, path)
	}
	return rsp
}

func (s *CommonService) ZonePromptAPI(ctx context.Context, req *api.ZonePromptAPIReq) (*api.ZonePromptAPIReply, error) {
	gateway, err := s.dao.ListGateway(ctx, req.Node)
	if err != nil {
		return nil, err
	}
	zoneSet := make(sets.String)
	for _, v := range gateway {
		if v.AppName != req.Gateway {
			continue
		}
		for _, config := range v.Configs {
			zoneSet.Insert(config.Zone)
		}
		break
	}
	reply := &api.ZonePromptAPIReply{
		Zones: zoneSet.List(),
	}
	return reply, nil
}
