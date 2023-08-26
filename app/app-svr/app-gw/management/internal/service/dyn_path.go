package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go-common/library/ecode"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func (s *HttpService) ListDynPath(ctx context.Context, req *pb.ListDynPathReq) (*pb.ListDynPathReply, error) {
	dps, err := s.resourceDao.ListDynPath(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	out := &pb.ListDynPathReply{
		Node:     req.Node,
		Gateway:  req.Gateway,
		DynPaths: dps,
	}
	return out, nil
}

func checkPattern(pattern string) error {
	if pattern == "" {
		return errors.New("empty pattern")
	}
	if strings.HasPrefix(pattern, "/") {
		return nil
	}
	if strings.HasPrefix(pattern, "= ") {
		return nil
	}
	if strings.HasPrefix(pattern, "~ ") {
		rawExp := strings.TrimPrefix(pattern, "~ ")
		if _, err := regexp.Compile(rawExp); err != nil {
			return err
		}
		return nil
	}
	if strings.HasPrefix(pattern, "~") && !strings.HasPrefix(pattern, "~ ") {
		return errors.Errorf("invalid regexp pattern: %s", pattern)
	}
	return errors.Errorf("unrecognized pattern: %s", pattern)
}

func checkEndpoint(ctx context.Context, arg *pb.ClientInfo) error {
	if arg.SkipEndpointCheck {
		return nil
	}
	u, err := url.Parse(arg.Endpoint)
	if err != nil {
		return errors.WithStack(err)
	}
	switch u.Scheme {
	case "discovery":
		hostname := u.Hostname()
		if _, err := resolveFromDiscovery(ctx, hostname, ""); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func checkSetDynPathReq(ctx context.Context, req *pb.SetDynPathReq) error {
	if err := checkPattern(req.Pattern); err != nil {
		return err
	}
	if err := checkEndpoint(ctx, req.ClientInfo); err != nil {
		return err
	}
	return nil
}

func (s *HttpService) AddDynPath(ctx context.Context, req *pb.SetDynPathReq) (*empty.Empty, error) {
	if err := checkSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	if err := s.resourceDao.SetDynPath(ctx, req); err != nil {
		audit.SendSetDynPathLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetDynPathLog(req, audit.LogActionAdd, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) hasDynPath(ctx context.Context, node, gateway, pattern string) (bool, error) {
	dps, err := s.resourceDao.ListDynPath(ctx, node, gateway)
	if err != nil {
		return false, err
	}
	for _, dp := range dps {
		if dp.Pattern == pattern {
			return true, nil
		}
	}
	return false, nil
}

func (s *HttpService) UpdateDynPath(ctx context.Context, req *pb.SetDynPathReq) (*empty.Empty, error) {
	if err := checkSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	exist, err := s.hasDynPath(ctx, req.Node, req.Gateway, req.Pattern)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("dynPath %q is not exist", req.Pattern))
	}
	if err := s.resourceDao.SetDynPath(ctx, req); err != nil {
		audit.SendSetDynPathLog(req, audit.LogActionUpdate, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetDynPathLog(req, audit.LogActionUpdate, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq) (*empty.Empty, error) {
	if err := s.resourceDao.DeleteDynPath(ctx, req); err != nil {
		audit.SendDeleteDynPathLog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendDeleteDynPathLog(req, audit.LogActionDel, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *HttpService) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq) (*empty.Empty, error) {
	exist, err := s.hasDynPath(ctx, req.Node, req.Gateway, req.Pattern)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("dynPath %q is not exist", req.Pattern))
	}
	if err := s.resourceDao.EnableDynPath(ctx, req); err != nil {
		audit.SendEnableDynPathLog(req, action(req.Disable), audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendEnableDynPathLog(req, action(req.Disable), audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) DisableDynPath(ctx context.Context, req *pb.EnableDynPathReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableDynPath(ctx, req)
}
