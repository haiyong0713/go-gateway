package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go-common/library/ecode"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func (s *GrpcService) ListDynPath(ctx context.Context, req *pb.ListDynPathReq) (*pb.ListDynPathReply, error) {
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

func checkGRPCSetDynPathReq(ctx context.Context, req *pb.SetDynPathReq) error {
	if err := checkGRPCPattern(req.Pattern); err != nil {
		return err
	}
	if err := checkEndpoint(ctx, req.ClientInfo); err != nil {
		return err
	}
	return nil
}

// 应该支持service级别前缀匹配，精准方法匹配，还有正则匹配
// 前缀匹配  bilibili.app.test.v1.TestService
// 精准匹配  /bilibili.app.test.v1.TestService/Test
// 正则匹配  ~ /bilibili\.app\.test\.v.+
var (
	// 开头
	_rgxStart = regexp.MustCompile("^[~/a-zA-Z]")
	// 大概pattern
	_rgxPath = regexp.MustCompile("^[a-zA-Z0-9./_]+$")
)

func checkGRPCPattern(pattern string) error {
	if pattern == "" {
		return errors.New("empty pattern")
	}

	if !_rgxStart.MatchString(pattern) || !_rgxPath.MatchString(strings.Trim(pattern, "~ ")) {
		return errors.Errorf("malformed pattern: %q", pattern)
	}

	// 如果 带有的/不是两个而且不是正则匹配 那么一定是写错的pattern
	if count := strings.Count(pattern, "/"); count > 0 && count != 2 && !strings.HasPrefix(pattern, "~") {
		return errors.Errorf("malformed pattern: %q", pattern)
	}
	return nil
}

func (s *GrpcService) AddDynPath(ctx context.Context, req *pb.SetDynPathReq) (*empty.Empty, error) {
	if err := checkGRPCSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	if err := s.resourceDao.SetDynPath(ctx, req); err != nil {
		audit.SendSetDynPathLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendSetDynPathLog(req, audit.LogActionAdd, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _grpcType)
	return &empty.Empty{}, nil
}

func (s *GrpcService) hasGRPCDynPath(ctx context.Context, node, gateway, pattern string) (bool, error) {
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

func (s *GrpcService) UpdateDynPath(ctx context.Context, req *pb.SetDynPathReq) (*empty.Empty, error) {
	if err := checkGRPCSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	exist, err := s.hasGRPCDynPath(ctx, req.Node, req.Gateway, req.Pattern)
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
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _grpcType)
	return &empty.Empty{}, nil
}

func (s *GrpcService) DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq) (*empty.Empty, error) {
	if err := s.resourceDao.DeleteDynPath(ctx, req); err != nil {
		audit.SendDeleteDynPathLog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendDeleteDynPathLog(req, audit.LogActionDel, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}

func (s *GrpcService) EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq) (*empty.Empty, error) {
	exist, err := s.hasGRPCDynPath(ctx, req.Node, req.Gateway, req.Pattern)
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
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _grpcType)
	return &empty.Empty{}, nil
}

func (s *GrpcService) DisableDynPath(ctx context.Context, req *pb.EnableDynPathReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableDynPath(ctx, req)
}
