package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management-job/api"
	"go-gateway/app/app-svr/app-gw/management-job/common"
	gwconfig "go-gateway/app/app-svr/app-gw/management-job/internal/model/gateway-config"
	mng "go-gateway/app/app-svr/app-gw/management/api"
	logutil "go-gateway/app/app-svr/app-gw/management/audit"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
)

func (s *Service) TaskDo(ctx context.Context, req *pb.TaskDoReq) (*pb.TaskDoReply, error) {
	var taskList = map[string]func(context.Context, *pb.TaskDoReq) error{
		"pushSingle":     s.PushSingle,
		"pushAll":        s.PushAll,
		"grpcPushSingle": s.GRPCPushSingle,
		"grpcPushAll":    s.GRPCPushAll,
	}
	f, ok := taskList[req.Name]
	if !ok {
		return nil, errors.New("no such task:" + req.Name)
	}
	out := &pb.TaskDoReply{
		TaskId: uuid.New().String(),
	}
	//nolint:errcheck
	s.worker.Do(ctx, func(ctx context.Context) {
		if err := f(ctx, req); err != nil {
			log.Error("Failed to execute task: %q with task id: %s, error from function: %+v", req.Name, out.TaskId, err)
			return
		}
	})
	return out, nil
}

func (s *Service) PushSingle(ctx context.Context, in *pb.TaskDoReq) error {
	gw, err := s.dao.Gateway(ctx, in.Params.Node, in.Params.Gateway)
	if err != nil {
		return err
	}
	pc := &gwconfig.PushConfigContext{}
	pc.FromTask(in)
	pc.Compare = false
	if err := s.PushProxyPassConfigs(ctx, gw, pc); err != nil {
		return err
	}
	return nil
}

func (s *Service) PushAll(ctx context.Context, in *pb.TaskDoReq) error {
	gateways, err := s.dao.ListGateway(ctx)
	if err != nil {
		return err
	}
	pc := &gwconfig.PushConfigContext{}
	pc.FromTask(in)
	pc.Compare = true
	for _, gw := range gateways {
		if err = s.PushProxyPassConfigs(ctx, gw, pc); err != nil {
			log.Error("Failed to push proxy configs:%+v", err)
			continue
		}
	}
	return nil
}

func (s *Service) GRPCPushSingle(ctx context.Context, in *pb.TaskDoReq) error {
	gw, err := s.dao.Gateway(ctx, in.Params.Node, in.Params.Gateway)
	if err != nil {
		return err
	}
	pc := &gwconfig.PushConfigContext{}
	pc.FromTask(in)
	pc.Compare = false
	if err := s.PushGRPCProxyPassConfigs(ctx, gw, pc); err != nil {
		return err
	}
	return nil
}

func (s *Service) GRPCPushAll(ctx context.Context, in *pb.TaskDoReq) error {
	gateways, err := s.dao.ListGateway(ctx)
	if err != nil {
		return err
	}
	pc := &gwconfig.PushConfigContext{}
	pc.FromTask(in)
	pc.Compare = true
	for _, gw := range gateways {
		if err = s.PushGRPCProxyPassConfigs(ctx, gw, pc); err != nil {
			log.Error("Failed to push grpc proxy configs:%+v", err)
			continue
		}
	}
	return nil
}

func (s *Service) PushProxyPassConfigs(ctx context.Context, gw *mng.Gateway, pc *gwconfig.PushConfigContext) error {
	proxyConf, err := s.ProxyConfig(ctx, gw.Node, gw.AppName)
	if err != nil {
		return err
	}
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(proxyConf); err != nil {
		return err
	}
	for _, cfg := range gw.Configs {
		if !cfg.Enable {
			continue
		}
		if pc.Compare {
			if s.configNotChanged(ctx, gw.Node, gw.AppName, gw.TreeId, cfg, tomlBuf.Bytes()) {
				continue
			}
		}
		req := &gwconfig.PushConfigReq{
			AppID:      formatAppID(gw.Node, gw.AppName),
			TreeID:     gw.TreeId,
			ConfigMeta: cfg,
			Buffer:     tomlBuf.Bytes(),
		}
		if err := s.dao.PushConfigs(ctx, req); err != nil {
			log.Error("Failed to push proxy config on: %+v: %+v", req, err)
			logutil.SendTaskLog(&logutil.ReportParam{
				GatewayGroup: gw.Node,
				GatewayName:  gw.AppName,
				Object:       logutil.LogTypeApplicationProxyTOML,
				Ctime:        pc.Ctime,
				Mtime:        pc.Mtime,
				Level:        logutil.LogLevelError,
				Sponsor:      pc.Sponsor,
				Action:       logutil.LogActionPush,
				Result:       logutil.LogResultFailure,
				Detail:       fmt.Sprintf("%+v", err),
				Identifier:   logutil.LogProxyConfigIdentifier,
			})
			continue
		}
		logutil.SendTaskLog(&logutil.ReportParam{
			GatewayGroup: gw.Node,
			GatewayName:  gw.AppName,
			Object:       logutil.LogTypeApplicationProxyTOML,
			Ctime:        pc.Ctime,
			Mtime:        pc.Mtime,
			Level:        logutil.LogLevelInfo,
			Sponsor:      pc.Sponsor,
			Action:       logutil.LogActionPush,
			Result:       logutil.LogResultSuccess,
			Detail:       jsonify(req),
			Identifier:   logutil.LogProxyConfigIdentifier,
		})
	}
	return nil
}

func (s *Service) ProxyConfig(ctx context.Context, node, gateway string) (*gwconfig.ProxyConfigs, error) {
	dpReply, err := s.dao.ListDynPath(ctx, node, gateway)
	if err != nil {
		return nil, err
	}
	dynPaths := common.EnabledDynPath(dpReply)
	if len(dynPaths) == 0 {
		return nil, errors.New("all dynPath are disable")
	}
	pm := common.BuildPathMetaByDynPath(dynPaths)
	// kv获取配置
	baReply, err := s.dao.ListBreakerAPI(ctx, node, gateway)
	if err != nil {
		return nil, err
	}
	breakerAPIs := common.EnabledBreakerAPI(baReply)
	quotaReply, err := s.dao.GetQuotaMethods(ctx, node, gateway)
	if err != nil {
		return nil, err
	}
	quotaMethods := common.EnabledQuotaMethod(quotaReply)
	out, err := common.RunProcess(pm, common.PathMetaAppendBreakerAPIs(breakerAPIs), common.PathMetaAppendRateLimiter(quotaMethods))
	if err != nil {
		return nil, err
	}
	proxyConf := &gwconfig.ProxyConfigs{
		ProxyConfig: &gwconfig.ProxyConfig{
			DynPath: out,
		},
	}
	return proxyConf, nil
}

func (s *Service) PushGRPCProxyPassConfigs(ctx context.Context, gw *mng.Gateway, pc *gwconfig.PushConfigContext) error {
	proxyConf, err := s.GRPCProxyConfig(ctx, gw.Node, gw.AppName)
	if err != nil {
		return err
	}
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(proxyConf); err != nil {
		return err
	}
	for _, cfg := range gw.GrpcConfigs {
		if !cfg.Enable {
			continue
		}
		if pc.Compare {
			if s.configNotChanged(ctx, gw.Node, gw.AppName, gw.TreeId, cfg, tomlBuf.Bytes()) {
				continue
			}
		}
		req := &gwconfig.PushConfigReq{
			AppID:      formatAppID(gw.Node, gw.AppName),
			TreeID:     gw.TreeId,
			ConfigMeta: cfg,
			Buffer:     tomlBuf.Bytes(),
		}
		if err := s.dao.PushConfigs(ctx, req); err != nil {
			log.Error("Failed to push proxy config on: %+v: %+v", req, err)
			logutil.SendTaskLog(&logutil.ReportParam{
				GatewayGroup: gw.Node,
				GatewayName:  gw.AppName,
				Object:       logutil.LogTypeApplicationProxyTOML,
				Ctime:        pc.Ctime,
				Mtime:        pc.Mtime,
				Level:        logutil.LogLevelError,
				Sponsor:      pc.Sponsor,
				Action:       logutil.LogActionPush,
				Result:       logutil.LogResultFailure,
				Detail:       fmt.Sprintf("%+v", err),
				Identifier:   logutil.LogGRPCProxyConfigIdentifier,
			})
			continue
		}
		logutil.SendTaskLog(&logutil.ReportParam{
			GatewayGroup: gw.Node,
			GatewayName:  gw.AppName,
			Object:       logutil.LogTypeApplicationProxyTOML,
			Ctime:        pc.Ctime,
			Mtime:        pc.Mtime,
			Level:        logutil.LogLevelInfo,
			Sponsor:      pc.Sponsor,
			Action:       logutil.LogActionPush,
			Result:       logutil.LogResultSuccess,
			Detail:       jsonify(req),
			Identifier:   logutil.LogGRPCProxyConfigIdentifier,
		})
	}
	return nil
}

func (s *Service) GRPCProxyConfig(ctx context.Context, node, gateway string) (*gwconfig.GrpcProxyConfig, error) {
	dsReply, err := s.dao.GRPCListDynService(ctx, node, gateway)
	if err != nil {
		return nil, err
	}
	dynService := common.EnabledDynPath(dsReply)
	if len(dynService) == 0 {
		return nil, errors.New("all dynService are disable")
	}
	sm := common.BuildServiceMetaByDynPath(dynService)
	baReply, err := s.dao.GRPCListBreakerAPI(ctx, node, gateway)
	if err != nil {
		return nil, err
	}
	breakerAPIs := common.EnabledBreakerAPI(baReply)
	out, err := common.RunGRPCProcess(sm, common.ServiceMetaAppendBreakerAPIs(breakerAPIs))
	if err != nil {
		return nil, err
	}
	grpcProxyConf := &gwconfig.GrpcProxyConfig{
		ProxyConfig: &sdkwarden.Config{
			DynService: out,
		},
	}
	return grpcProxyConf, nil
}

func (s *Service) configNotChanged(ctx context.Context, node, appName string, treeId int64, cfg *mng.ConfigMeta, content []byte) bool {
	req := &gwconfig.RawConfigReq{
		AppID:      formatAppID(node, appName),
		TreeID:     treeId,
		ConfigMeta: cfg,
	}
	conf, err := s.dao.RawConfigs(ctx, req)
	if err != nil {
		log.Error("Failed to raw proxy config on: %+v: %+v", req, err)
		logutil.SendTaskLog(&logutil.ReportParam{
			GatewayGroup: node,
			GatewayName:  appName,
			Object:       logutil.LogTypeApplicationProxyTOML,
			Ctime:        time.Now().Unix(),
			Mtime:        time.Now().Unix(),
			Level:        logutil.LogLevelError,
			Sponsor:      logutil.LogMngSponsor,
			Action:       logutil.LogActionPush,
			Result:       logutil.LogResultFailure,
			Detail:       fmt.Sprintf("%+v", err),
			Identifier:   logutil.LogProxyConfigIdentifier})
		return true
	}
	if bytes.Equal(conf, content) {
		log.Info("Same configs(%s: %+v, %+v)", appName, string(conf), string(content))
		result := ""
		in := &gwconfig.RawLogReq{
			Node:       node,
			Gateway:    appName,
			Order:      "ctime",
			ObjectType: logutil.LogTypeApplicationProxyTOML,
			Pn:         1,
			Ps:         1,
		}
		if result, err = s.firstLog(ctx, in); err != nil {
			log.Error("Failed to raw first log(%s,%+v)", in.Gateway, err)
			return true
		}
		if result != logutil.LogResultNone {
			logutil.SendTaskLog(&logutil.ReportParam{
				GatewayGroup: node,
				GatewayName:  appName,
				Object:       logutil.LogTypeApplicationProxyTOML,
				Ctime:        time.Now().Unix(),
				Mtime:        time.Now().Unix(),
				Level:        logutil.LogLevelInfo,
				Sponsor:      logutil.LogMngSponsor,
				Action:       logutil.LogActionPush,
				Result:       logutil.LogResultNone,
				Detail:       jsonify(req),
				Identifier:   logutil.LogProxyConfigIdentifier})
		}
		return true
	}
	return false
}

func (s *Service) firstLog(ctx context.Context, req *gwconfig.RawLogReq) (string, error) {
	res, err := s.dao.RawTaskLog(ctx, req)
	if err != nil {
		return "", err
	}
	if len(res.Result) <= 0 {
		log.Warn("Audit Log is nil(%s)", req.Gateway)
		return "", nil
	}
	extraData := new(gwconfig.Extra)
	if err = json.Unmarshal([]byte(res.Result[0].ExtraData), &extraData); err != nil {
		log.Error("failed to json unmarshal:%+v", err)
		return "", err
	}
	return extraData.Result, nil
}

func formatAppID(node, gateway string) string {
	return fmt.Sprintf("%s.%s", node, gateway)
}

func jsonify(in interface{}) string {
	out, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(out)
}

func (s *Service) RawConfig(ctx context.Context, req *pb.RawConfigReq) (*pb.RawConfigReply, error) {
	gw, err := s.dao.Gateway(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	proxyConf, err := s.ProxyConfig(ctx, gw.Node, gw.AppName)
	if err != nil {
		return nil, err
	}
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(proxyConf); err != nil {
		return nil, err
	}
	out := &pb.RawConfigReply{
		Config: tomlBuf.String(),
	}
	return out, nil
}
