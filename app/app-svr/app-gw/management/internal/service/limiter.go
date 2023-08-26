package service

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"
	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"go-common/library/sync/errgroup.v2"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

const (
	_pluginNameQuota = "quota"
	_httpProto       = "http"
)

func quotaID(appid, env, zone string) string {
	return fmt.Sprintf("%s.%s.%s", env, zone, appid)
}

func quotaResourceID(config model.QuotaConfig) string {
	return fmt.Sprintf("%s|%s|%s|%s", config.ServiceID, config.Protocol, config.Uri, config.Rule)
}

//nolint:deadcode,unused
func quotaHttpResourceID(env, zone string, method *pb.QuotaMethod) string {
	return fmt.Sprintf("%s.%s.%s.%s|http|%s|%s", env, zone, method.Node, method.Gateway, method.Api, method.Rule)
}

func (s *HttpService) ListLimiter(ctx context.Context, req *pb.ListLimiterReq) (*pb.ListLimiterReply, error) {
	prompt := &pb.ZonePromptAPIReq{
		Node:    req.Node,
		Gateway: req.Gateway,
	}
	zones, err := s.common.ZonePromptAPI(ctx, prompt)
	if err != nil {
		return nil, err
	}
	list := []*pb.Limiter{}
	var mutex sync.Mutex
	appID := fmt.Sprintf("%s.%s", req.Node, req.Gateway)
	eg := errgroup.WithContext(ctx)
	for _, zone := range zones.Zones {
		zone := zone
		eg.Go(func(ctx context.Context) error {
			field := quotaID(appID, env.DeployEnv, zone)
			plugin, err := s.dao.GetPlugin(ctx, _pluginNameQuota, field)
			if err != nil {
				return err
			}
			resources, err := s.dao.QuotaResources(ctx, quotaID(appID, env.DeployEnv, zone), plugin.Data)
			if err != nil {
				return err
			}
			mutex.Lock()
			list = append(list, resources...)
			mutex.Unlock()
			return nil
		})
	}
	quotaMethods := []*pb.QuotaMethod{}
	eg.Go(func(ctx context.Context) error {
		methods, err := s.resourceDao.GetQuotaMethods(ctx, req.Node, req.Gateway)
		if err != nil {
			return err
		}
		quotaMethods = methods
		return nil
	})
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	reply := &pb.ListLimiterReply{
		List: constructLimiterItems(quotaMethods, list),
	}
	return reply, nil
}

func convertLimiterMap(limiters []*pb.Limiter) map[string]map[string][]*pb.Limiter {
	out := make(map[string]map[string][]*pb.Limiter)
	for _, v := range limiters {
		cfg, ok := model.ParseQuotaConfig(v.Id)
		if !ok {
			log.Warn("No parsed config: %+v", v)
			continue
		}
		if _, ok := out[cfg.Uri]; !ok {
			out[cfg.Uri] = make(map[string][]*pb.Limiter)
		}
		out[cfg.Uri][cfg.Rule] = append(out[cfg.Uri][cfg.Rule], v)
	}
	return out
}

func constructLimiterItems(quotaMethods []*pb.QuotaMethod, limiters []*pb.Limiter) []*pb.LimiterListItem {
	quotaMethodsMap := make(map[string][]*pb.QuotaMethod)
	for _, quotaMethod := range quotaMethods {
		quotaMethodsMap[quotaMethod.Api] = append(quotaMethodsMap[quotaMethod.Api], quotaMethod)
	}
	limiterMap := convertLimiterMap(limiters)
	out := make([]*pb.LimiterListItem, 0, len(quotaMethodsMap))
	for api, methods := range quotaMethodsMap {
		item := &pb.LimiterListItem{Api: api}
		lms := mixedLimiterMetas(api, methods, limiterMap[api])
		sort.SliceStable(lms, func(i, j int) bool {
			return lms[i].Rule < lms[j].Rule
		})
		item.Limiters = lms
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Api < out[j].Api
	})
	return out
}

func mixedLimiterMetas(api string, methods []*pb.QuotaMethod, apiLimiterMap map[string][]*pb.Limiter) []*pb.LimiterMeta {
	var res []*pb.LimiterMeta
	for _, method := range methods {
		if _, ok := apiLimiterMap[method.Rule]; !ok {
			log.Error("Failed to find limiter: api(%s), rule(%s)", api, method.Rule)
			continue
		}
		for _, v := range apiLimiterMap[method.Rule] {
			cfg, _ := model.ParseQuotaConfig(v.Id)
			limiterMeta := &pb.LimiterMeta{
				Id:              v.Id,
				RefreshInterval: v.RefreshInterval,
				Algorithm:       v.Algorithm,
				Rule:            cfg.Rule,
				Enable:          method.Enable,
				Zone:            model.ParseZone(cfg.ServiceID),
			}
			switch model.ParseRuleType(cfg.Rule) {
			case model.TotalRule:
				limiterMeta.TotalRule = pb.TotalRule{Capacity: v.Capacity}
			case model.RefererRule:
				limiterMeta.RefererRule = pb.RefererRule{Capacity: v.Capacity}
			default:
				log.Error("UnExpected limiter rule type: %+v", limiterMeta)
				continue
			}
			res = append(res, limiterMeta)
		}
	}
	return res
}

func (s *HttpService) isRefererRuleLimited(ctx context.Context, req *pb.AddLimiterReq) bool {
	if model.ParseRuleType(req.Rule) == model.RefererRule && s.getLimiterCountByRuleType(ctx, req) > 20 {
		return true
	}
	return false
}

func (s *HttpService) AddLimiter(ctx context.Context, req *pb.AddLimiterReq) (*empty.Empty, error) {
	appID := fmt.Sprintf("%s.%s", req.Node, req.Gateway)
	field := quotaID(appID, env.DeployEnv, req.Zone)

	if s.isRefererRuleLimited(ctx, req) {
		err := errors.Wrapf(ecode.RequestErr, "新增referer限流达到上限")
		audit.SendAddLimiterLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	plugin, err := s.dao.GetPlugin(ctx, _pluginNameQuota, field)
	if err != nil {
		return nil, err
	}
	quotaConfig := model.QuotaConfig{
		ServiceID: field,
		Protocol:  _httpProto,
		Uri:       req.Api,
		Rule:      req.Rule,
	}
	limiter := &pb.Limiter{
		Id:              quotaResourceID(quotaConfig),
		Capacity:        req.Capacity,
		RefreshInterval: req.RefreshInterval,
		Algorithm:       req.Algorithm,
	}
	if err = s.dao.AddQuotaResources(ctx, limiter, plugin.Data); err != nil {
		audit.SendAddLimiterLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	quotaMethod := &pb.QuotaMethod{
		Api:     req.Api,
		Rule:    req.Rule,
		Node:    req.Node,
		Gateway: req.Gateway,
		Enable:  req.Enable,
	}
	if err := s.resourceDao.SetQuotaMethod(ctx, quotaMethod); err != nil {
		audit.SendAddLimiterLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendAddLimiterLog(req, audit.LogActionAdd, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) UpdateLimiter(ctx context.Context, req *pb.SetLimiterReq) (*empty.Empty, error) {
	quotaConfig, ok := model.ParseQuotaConfig(req.Id)
	if !ok {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("id %s is invalid ", req.Id))
	}
	plugin, err := s.dao.GetPlugin(ctx, _pluginNameQuota, quotaConfig.ServiceID)
	if err != nil {
		return nil, err
	}
	limiter := &pb.Limiter{
		Id:              req.Id,
		Capacity:        req.Capacity,
		RefreshInterval: req.RefreshInterval,
		Algorithm:       req.Algorithm,
	}
	if err = s.dao.UpdateQuotaResources(ctx, limiter, plugin.Data); err != nil {
		audit.SendSetLimiterLog(req, audit.LogActionUpdate, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetLimiterLog(req, audit.LogActionUpdate, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) getLimiterListItem(ctx context.Context, node, gateway, api string) (*pb.LimiterListItem, error) {
	listReq := &pb.ListLimiterReq{
		Node:    node,
		Gateway: gateway,
	}
	limiterList, err := s.ListLimiter(ctx, listReq)
	if err != nil {
		return nil, err
	}
	for _, v := range limiterList.List {
		if v.Api == api {
			return v, nil
		}
	}
	return nil, errors.Errorf("Failed to get limiterList: %s, %s, %s", node, gateway, api)
}

func (s *HttpService) getLimiterListByRule(ctx context.Context, req *pb.DeleteLimiterReq, quotaConfig model.QuotaConfig) ([]*pb.LimiterMeta, error) {
	listItem, err := s.getLimiterListItem(ctx, req.Node, req.Gateway, quotaConfig.Uri)
	if err != nil {
		return nil, err
	}
	var res []*pb.LimiterMeta
	for _, v := range listItem.Limiters {
		if v.Rule != quotaConfig.Rule {
			continue
		}
		res = append(res, v)
	}
	return res, nil
}

func (s *HttpService) getLimiterCountByRuleType(ctx context.Context, req *pb.AddLimiterReq) int {
	listItem, err := s.getLimiterListItem(ctx, req.Node, req.Gateway, req.Api)
	if err != nil {
		log.Warn("s.getLimiterListItem err(%+v)", err)
		return 0
	}
	res := 0
	for _, v := range listItem.Limiters {
		if model.ParseRuleType(v.Rule) != model.ParseRuleType(req.Rule) {
			continue
		}
		res++
	}
	return res
}

func (s *HttpService) DeleteLimiter(ctx context.Context, req *pb.DeleteLimiterReq) (*empty.Empty, error) {
	quotaConfig, ok := model.ParseQuotaConfig(req.Id)
	if !ok {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("id %s is invalid ", req.Id))
	}
	plugin, err := s.dao.GetPlugin(ctx, _pluginNameQuota, quotaConfig.ServiceID)
	if err != nil {
		return nil, err
	}
	if err = s.dao.DeleteQuotaResources(ctx, req.Id, plugin.Data); err != nil {
		audit.SendDeleteLimiterLog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	items, err := s.getLimiterListByRule(ctx, req, quotaConfig)
	if err != nil {
		return nil, err
	}
	if len(items) <= 1 {
		quotaMethod := &pb.QuotaMethod{
			Api:     quotaConfig.Uri,
			Rule:    quotaConfig.Rule,
			Node:    req.Node,
			Gateway: req.Gateway,
		}
		if err := s.resourceDao.DeleteQuotaMethod(ctx, quotaMethod); err != nil {
			audit.SendDeleteLimiterLog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
			return nil, err
		}
	}
	audit.SendDeleteLimiterLog(req, audit.LogActionDel, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *HttpService) EnableLimiter(ctx context.Context, req *pb.EnableLimiterReq) (*empty.Empty, error) {
	if err := s.resourceDao.EnableQuotaMethod(ctx, req); err != nil {
		audit.SendEnableLimiterLog(req, action(req.Disable), audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendEnableLimiterLog(req, action(req.Disable), audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) DisableLimiter(ctx context.Context, req *pb.EnableLimiterReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableLimiter(ctx, req)
}

func (s *HttpService) SetupPlugin(ctx context.Context, req *pb.PluginReq) (*empty.Empty, error) {
	if err := s.dao.SetupPlugin(ctx, req.PluginName, req.Field, req.Plugin); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *HttpService) PluginList(ctx context.Context, req *pb.PluginListReq) (*pb.PluginListReply, error) {
	list, err := s.dao.PluginList(ctx, req)
	if err != nil {
		return nil, err
	}
	return &pb.PluginListReply{List: list}, nil
}
