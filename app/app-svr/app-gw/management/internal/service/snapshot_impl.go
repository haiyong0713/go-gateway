package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

type SnapshotResponseActionImplantation func(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error)
type SnapshotImpl struct {
	ReqParser  SnapshotRequestParser
	ActionImpl SnapshotResponseActionImplantation
}

type SnapshotService struct {
	snapshotParser
	dao      dao.SnapshotDao
	grpcDao  dao.SnapshotGRPCDao
	registry map[string]SnapshotImpl
}

func newSnapshot(d dao.SnapshotDao) *SnapshotService {
	s := &SnapshotService{
		dao:      d,
		grpcDao:  d.CreateSnapshotGRPCDao(),
		registry: map[string]SnapshotImpl{},
	}
	s.SetupRegistery()
	return s
}

func (s *SnapshotService) SetupRegistery() {
	s.addImpl("breakerAPI", "list", SnapshotImpl{
		ReqParser:  s.parseListBreakerAPI,
		ActionImpl: s.ListBreakerAPI,
	})
	s.addImpl("breakerAPI", "add", SnapshotImpl{
		ReqParser:  s.parseSetBreakerAPI,
		ActionImpl: s.AddBreakerAPI,
	})
	s.addImpl("breakerAPI", "update", SnapshotImpl{
		ReqParser:  s.parseSetBreakerAPI,
		ActionImpl: s.UpdateBreakerAPI,
	})
	s.addImpl("breakerAPI", "enable", SnapshotImpl{
		ReqParser:  s.parseEnableBreakerAPI,
		ActionImpl: s.EnableBreakerAPI,
	})
	s.addImpl("breakerAPI", "disable", SnapshotImpl{
		ReqParser:  s.parseEnableBreakerAPI,
		ActionImpl: s.DisableBreakerAPI,
	})
	s.addImpl("breakerAPI", "delete", SnapshotImpl{
		ReqParser:  s.parseDeleteBreakerAPI,
		ActionImpl: s.DeleteBreakerAPI,
	})

	s.addImpl("dynpath", "list", SnapshotImpl{
		ReqParser:  s.parseListDynPath,
		ActionImpl: s.ListDynPath,
	})
	s.addImpl("dynpath", "add", SnapshotImpl{
		ReqParser:  s.parseSetDynPath,
		ActionImpl: s.AddDynPath,
	})
	s.addImpl("dynpath", "update", SnapshotImpl{
		ReqParser:  s.parseSetDynPath,
		ActionImpl: s.UpdateDynPath,
	})
	s.addImpl("dynpath", "enable", SnapshotImpl{
		ReqParser:  s.parseEnableDynPath,
		ActionImpl: s.EnableDynPath,
	})
	s.addImpl("dynpath", "disable", SnapshotImpl{
		ReqParser:  s.parseEnableDynPath,
		ActionImpl: s.DisableDynPath,
	})
	s.addImpl("dynpath", "delete", SnapshotImpl{
		ReqParser:  s.parseDeleteDynPath,
		ActionImpl: s.DeleteDynPath,
	})

	s.addImpl("grpc_breakerAPI", "list", SnapshotImpl{
		ReqParser:  s.parseListBreakerAPI,
		ActionImpl: s.ListGRPCBreakerAPI,
	})
	s.addImpl("grpc_breakerAPI", "add", SnapshotImpl{
		ReqParser:  s.parseSetBreakerAPI,
		ActionImpl: s.AddGRPCBreakerAPI,
	})
	s.addImpl("grpc_breakerAPI", "update", SnapshotImpl{
		ReqParser:  s.parseSetBreakerAPI,
		ActionImpl: s.UpdateGRPCBreakerAPI,
	})
	s.addImpl("grpc_breakerAPI", "enable", SnapshotImpl{
		ReqParser:  s.parseEnableBreakerAPI,
		ActionImpl: s.EnableGRPCBreakerAPI,
	})
	s.addImpl("grpc_breakerAPI", "disable", SnapshotImpl{
		ReqParser:  s.parseEnableBreakerAPI,
		ActionImpl: s.DisableGRPCBreakerAPI,
	})
	s.addImpl("grpc_breakerAPI", "delete", SnapshotImpl{
		ReqParser:  s.parseDeleteBreakerAPI,
		ActionImpl: s.DeleteGRPCBreakerAPI,
	})

	s.addImpl("grpc_dynpath", "list", SnapshotImpl{
		ReqParser:  s.parseListDynPath,
		ActionImpl: s.ListGRPCDynPath,
	})
	s.addImpl("grpc_dynpath", "add", SnapshotImpl{
		ReqParser:  s.parseSetDynPath,
		ActionImpl: s.AddGRPCDynPath,
	})
	s.addImpl("grpc_dynpath", "update", SnapshotImpl{
		ReqParser:  s.parseSetDynPath,
		ActionImpl: s.UpdateGRPCDynPath,
	})
	s.addImpl("grpc_dynpath", "enable", SnapshotImpl{
		ReqParser:  s.parseEnableDynPath,
		ActionImpl: s.EnableGRPCDynPath,
	})
	s.addImpl("grpc_dynpath", "disable", SnapshotImpl{
		ReqParser:  s.parseEnableDynPath,
		ActionImpl: s.DisableGRPCDynPath,
	})
	s.addImpl("grpc_dynpath", "delete", SnapshotImpl{
		ReqParser:  s.parseDeleteDynPath,
		ActionImpl: s.DeleteGRPCDynPath,
	})
}

func (s *SnapshotService) ResolveImpl(resource, action string) (SnapshotImpl, bool) {
	key := fmt.Sprintf("%s_%s", resource, action)
	impl, ok := s.registry[key]
	return impl, ok
}

func (s *SnapshotService) ResolveGRPCImpl(resource, action string) (SnapshotImpl, bool) {
	key := fmt.Sprintf("grpc_%s_%s", resource, action)
	impl, ok := s.registry[key]
	return impl, ok
}

func (s *SnapshotService) addImpl(resource, action string, impl SnapshotImpl) {
	key := fmt.Sprintf("%s_%s", resource, action)
	if _, ok := s.registry[key]; ok {
		panic(errors.Errorf("already has snapshot resource implantation: %q", key))
	}
	s.registry[key] = impl
}

func (s *SnapshotService) listBreakerAPI(ctx context.Context, req *pb.ListBreakerAPIReq, uuid string) (*pb.ListBreakerAPIReply, error) {
	bapis, err := s.dao.ListBreakerAPI(ctx, req.Node, req.Gateway, uuid)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListBreakerAPIReply{
		BreakerApiList: bapis,
	}
	return reply, nil
}

func (s *SnapshotService) ListBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_ListBreakerAPIReq).ListBreakerAPIReq
	out, err := s.listBreakerAPI(ctx, innerReq, req.Uuid)
	if err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_ListBreakerAPI{
			ListBreakerAPI: out,
		},
	}
	return reply, nil
}

func (s *SnapshotService) HasBreakerAPI(ctx context.Context, node, gateway, api, uuid string) (bool, error) {
	bapis, err := s.dao.ListBreakerAPI(ctx, node, gateway, uuid)
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

func (s *SnapshotService) AddBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetBreakerAPIReq).SetBreakerAPIReq
	exist, err := s.HasBreakerAPI(ctx, innerReq.Node, innerReq.Gateway, innerReq.Api, req.Uuid)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("api %q is conflicated ", innerReq.Api))
	}
	if err := s.dao.SetBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) updateBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasBreakerAPI(ctx, req.Node, req.Gateway, req.Api, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.dao.SetBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) UpdateBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetBreakerAPIReq).SetBreakerAPIReq
	if _, err := s.updateBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) enableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasBreakerAPI(ctx, req.Node, req.Gateway, req.Api, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.dao.EnableBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) EnableBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableBreakerAPIReq).EnableBreakerAPIReq
	if _, err := s.enableBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) disableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) (*empty.Empty, error) {
	req.Disable = true
	return s.enableBreakerAPI(ctx, req, uuid)
}

func (s *SnapshotService) DisableBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableBreakerAPIReq).EnableBreakerAPIReq
	if _, err := s.disableBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) deleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) (*empty.Empty, error) {
	if err := s.dao.DeleteBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) DeleteBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_DeleteBreakerAPIReq).DeleteBreakerAPIReq
	if _, err := s.deleteBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) listDynPath(ctx context.Context, req *pb.ListDynPathReq, uuid string) (*pb.ListDynPathReply, error) {
	dps, err := s.dao.ListDynPath(ctx, req.Node, req.Gateway, uuid)
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

func (s *SnapshotService) ListDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_ListDynPathReq).ListDynPathReq
	out, err := s.listDynPath(ctx, innerReq, req.Uuid)
	if err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_ListDynPath{
			ListDynPath: out,
		},
	}
	return reply, nil
}

func (s *SnapshotService) addDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) (*empty.Empty, error) {
	if err := checkSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	if err := s.dao.SetDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) AddDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetDynPathReq).SetDynPathReq
	if _, err := s.addDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) HasDynPath(ctx context.Context, node, gateway, pattern, uuid string) (bool, error) {
	dps, err := s.dao.ListDynPath(ctx, node, gateway, uuid)
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

func (s *SnapshotService) updateDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) (*empty.Empty, error) {
	if err := checkSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	exist, err := s.HasDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("no such dyn path: %+v", req))
	}
	if err := s.dao.SetDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) UpdateDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetDynPathReq).SetDynPathReq
	if _, err := s.updateDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) deleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.NothingFound
	}
	if err := s.dao.DeleteDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) DeleteDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_DeleteDynPathReq).DeleteDynPathReq
	if _, err := s.deleteDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) enableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.NothingFound
	}
	if err := s.dao.EnableDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) EnableDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableDynPathReq).EnableDynPathReq
	if _, err := s.enableDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) disableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) (*empty.Empty, error) {
	req.Disable = true
	return s.enableDynPath(ctx, req, uuid)
}

func (s *SnapshotService) DisableDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableDynPathReq).EnableDynPathReq
	if _, err := s.disableDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}
