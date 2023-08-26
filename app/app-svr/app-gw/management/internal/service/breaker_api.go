package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/audit"

	"github.com/golang/protobuf/ptypes/empty"
)

// ListBreakerAPI is
func (s *HttpService) ListBreakerAPI(ctx context.Context, req *pb.ListBreakerAPIReq) (*pb.ListBreakerAPIReply, error) {
	bapis, err := s.resourceDao.ListBreakerAPI(ctx, req.Node, req.Gateway)
	if err != nil {
		return nil, err
	}
	return &pb.ListBreakerAPIReply{BreakerApiList: bapis}, nil
}

func jsonify(in interface{}) string {
	out, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(out)
}

// SetBreakerAPI is
func (s *HttpService) SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq) (*empty.Empty, error) {
	exist, err := s.hasBreakerAPI(ctx, req.Node, req.Gateway, req.Api)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("api %q is conflicated ", req.Api))
	}

	if err := s.resourceDao.SetBreakerAPI(ctx, req); err != nil {
		audit.SendBreakApiLog(req, audit.LogActionAdd, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendBreakApiLog(req, audit.LogActionAdd, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

func (s *HttpService) hasBreakerAPI(ctx context.Context, node, gateway, api string) (bool, error) {
	bapis, err := s.resourceDao.ListBreakerAPI(ctx, node, gateway)
	if err != nil {
		return false, err
	}
	for _, ba := range bapis {
		if ba.Api == api {
			return true, nil
		}
	}
	return false, nil
}

// UpdateBreakerAPI is
func (s *HttpService) UpdateBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq) (*empty.Empty, error) {
	exist, err := s.hasBreakerAPI(ctx, req.Node, req.Gateway, req.Api)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.resourceDao.SetBreakerAPI(ctx, req); err != nil {
		audit.SendBreakApiLog(req, audit.LogActionUpdate, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendBreakApiLog(req, audit.LogActionUpdate, audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

// EnableBreakerAPI is
func (s *HttpService) EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq) (*empty.Empty, error) {
	exist, err := s.hasBreakerAPI(ctx, req.Node, req.Gateway, req.Api)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.resourceDao.EnableBreakerAPI(ctx, req); err != nil {
		audit.SendEnableBreakApiLog(req, action(req.Disable), audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendEnableBreakApiLog(req, action(req.Disable), audit.LogLevelInfo, audit.LogResultSuccess, jsonify(req), 0, 0)
	s.common.asyncTriggerConfigPush(ctx, req.Node, req.Gateway, req.Username, _httpType)
	return &empty.Empty{}, nil
}

// DisableBreakerAPI is
func (s *HttpService) DisableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq) (*empty.Empty, error) {
	req.Disable = true
	return s.EnableBreakerAPI(ctx, req)
}

// DeleteBreakerAPI is
func (s *HttpService) DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq) (*empty.Empty, error) {
	if err := s.resourceDao.DeleteBreakerAPI(ctx, req); err != nil {
		audit.SendDeleteBreakerAPILog(req, audit.LogActionDel, audit.LogLevelError, audit.LogResultFailure, fmt.Sprintf("%+v", err), 0, 0)
		return nil, err
	}
	audit.SendDeleteBreakerAPILog(req, audit.LogActionDel, audit.LogLevelWarn, audit.LogResultSuccess, jsonify(req), 0, 0)
	return &empty.Empty{}, nil
}
