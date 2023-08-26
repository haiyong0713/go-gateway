package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *SnapshotService) listGRPCBreakerAPI(ctx context.Context, req *pb.ListBreakerAPIReq, uuid string) (*pb.ListBreakerAPIReply, error) {
	bapis, err := s.grpcDao.ListBreakerAPI(ctx, req.Node, req.Gateway, uuid)
	if err != nil {
		return nil, err
	}
	reply := &pb.ListBreakerAPIReply{
		BreakerApiList: bapis,
	}
	return reply, nil
}

func (s *SnapshotService) ListGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_ListBreakerAPIReq).ListBreakerAPIReq
	out, err := s.listGRPCBreakerAPI(ctx, innerReq, req.Uuid)
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

func (s *SnapshotService) HasGRPCBreakerAPI(ctx context.Context, node, gateway, api, uuid string) (bool, error) {
	bapis, err := s.grpcDao.ListBreakerAPI(ctx, node, gateway, uuid)
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

func (s *SnapshotService) AddGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetBreakerAPIReq).SetBreakerAPIReq
	exist, err := s.HasGRPCBreakerAPI(ctx, innerReq.Node, innerReq.Gateway, innerReq.Api, req.Uuid)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("api %q is conflicated ", innerReq.Api))
	}
	if err := s.grpcDao.SetBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) updateGRPCBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasGRPCBreakerAPI(ctx, req.Node, req.Gateway, req.Api, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.grpcDao.SetBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) UpdateGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetBreakerAPIReq).SetBreakerAPIReq
	if _, err := s.updateGRPCBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) enableGRPCBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasGRPCBreakerAPI(ctx, req.Node, req.Gateway, req.Api, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("api %q is not exist", req.Api))
	}
	if err := s.grpcDao.EnableBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) EnableGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableBreakerAPIReq).EnableBreakerAPIReq
	if _, err := s.enableGRPCBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) disableGRPCBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) (*empty.Empty, error) {
	req.Disable = true
	return s.enableGRPCBreakerAPI(ctx, req, uuid)
}

func (s *SnapshotService) DisableGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableBreakerAPIReq).EnableBreakerAPIReq
	if _, err := s.disableGRPCBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) deleteGRPCBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) (*empty.Empty, error) {
	if err := s.grpcDao.DeleteBreakerAPI(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) DeleteGRPCBreakerAPI(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_DeleteBreakerAPIReq).DeleteBreakerAPIReq
	if _, err := s.deleteGRPCBreakerAPI(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) listGRPCDynPath(ctx context.Context, req *pb.ListDynPathReq, uuid string) (*pb.ListDynPathReply, error) {
	dps, err := s.grpcDao.ListDynPath(ctx, req.Node, req.Gateway, uuid)
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

func (s *SnapshotService) ListGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_ListDynPathReq).ListDynPathReq
	out, err := s.listGRPCDynPath(ctx, innerReq, req.Uuid)
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

func (s *SnapshotService) addGRPCDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) (*empty.Empty, error) {
	if err := checkGRPCSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	if err := s.grpcDao.SetDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) AddGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetDynPathReq).SetDynPathReq
	if _, err := s.addGRPCDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) HasGRPCDynPath(ctx context.Context, node, gateway, pattern, uuid string) (bool, error) {
	dps, err := s.grpcDao.ListDynPath(ctx, node, gateway, uuid)
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

func (s *SnapshotService) updateGRPCDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) (*empty.Empty, error) {
	if err := checkGRPCSetDynPathReq(ctx, req); err != nil {
		return nil, ecode.Error(ecode.RequestErr, err.Error())
	}
	exist, err := s.HasGRPCDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("no such dyn path: %+v", req))
	}
	if err := s.grpcDao.SetDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) UpdateGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_SetDynPathReq).SetDynPathReq
	if _, err := s.updateGRPCDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) deleteGRPCDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasGRPCDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.NothingFound
	}
	if err := s.grpcDao.DeleteDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) DeleteGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_DeleteDynPathReq).DeleteDynPathReq
	if _, err := s.deleteGRPCDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) enableGRPCDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) (*empty.Empty, error) {
	exist, err := s.HasGRPCDynPath(ctx, req.Node, req.Gateway, req.Pattern, uuid)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, ecode.NothingFound
	}
	if err := s.grpcDao.EnableDynPath(ctx, req, uuid); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *SnapshotService) EnableGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableDynPathReq).EnableDynPathReq
	if _, err := s.enableGRPCDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}

func (s *SnapshotService) disableGRPCDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) (*empty.Empty, error) {
	req.Disable = true
	return s.enableGRPCDynPath(ctx, req, uuid)
}

func (s *SnapshotService) DisableGRPCDynPath(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	innerReq := req.SnapshotReq.(*pb.SnapshotActionReq_EnableDynPathReq).EnableDynPathReq
	if _, err := s.disableGRPCDynPath(ctx, innerReq, req.Uuid); err != nil {
		return nil, err
	}
	reply := &pb.SnapshotActionReply{
		SnapshotReply: &pb.SnapshotActionReply_Empty{
			Empty: &pb.Empty{},
		},
	}
	return reply, nil
}
