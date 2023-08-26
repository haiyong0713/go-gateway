package service

import (
	"context"

	pb "go-gateway/app/app-svr/app-gw/management/api"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

func (s *SnapshotService) SnapshotAction(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	impl, ok := s.ResolveImpl(req.Resource, req.Action)
	if !ok {
		return nil, errors.Errorf("undefined resource: %q and action: %q", req.Resource, req.Action)
	}
	reply, err := impl.ActionImpl(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *SnapshotService) SnapshotGRPCAction(ctx context.Context, req *pb.SnapshotActionReq) (*pb.SnapshotActionReply, error) {
	impl, ok := s.ResolveGRPCImpl(req.Resource, req.Action)
	if !ok {
		return nil, errors.Errorf("undefined resource: %q and action: %q", req.Resource, req.Action)
	}
	reply, err := impl.ActionImpl(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *SnapshotService) AddSnapshot(ctx context.Context, req *pb.AddSnapshotReq) (*pb.AddSnapshotReply, error) {
	return s.dao.AddSnapshot(ctx, req)
}

func (s *SnapshotService) SnapshotProfile(ctx context.Context, req *pb.SnapshotProfileReq) (*pb.SnapshotProfileReply, error) {
	meta, err := s.dao.GetSnapshotMeta(ctx, req.Node, req.Gateway, req.Uuid)
	if err != nil {
		return nil, err
	}

	reply := &pb.SnapshotProfileReply{
		Meta: meta,
	}

	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		lbReq := &pb.ListBreakerAPIReq{
			Node:    req.Node,
			Gateway: req.Gateway,
		}
		bapis, err := s.listBreakerAPI(ctx, lbReq, req.Uuid)
		if err != nil {
			return err
		}
		reply.BreakerApi = bapis
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		ldpReq := &pb.ListDynPathReq{
			Node:    req.Node,
			Gateway: req.Gateway,
		}
		dynPath, err := s.listDynPath(ctx, ldpReq, req.Uuid)
		if err != nil {
			return err
		}
		reply.DynPath = dynPath
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		lbReq := &pb.ListBreakerAPIReq{
			Node:    req.Node,
			Gateway: req.Gateway,
		}
		bapis, err := s.listGRPCBreakerAPI(ctx, lbReq, req.Uuid)
		if err != nil {
			return err
		}
		reply.GrpcBreakerApi = bapis
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		ldpReq := &pb.ListDynPathReq{
			Node:    req.Node,
			Gateway: req.Gateway,
		}
		dynPath, err := s.listGRPCDynPath(ctx, ldpReq, req.Uuid)
		if err != nil {
			return err
		}
		reply.GrpcDynPath = dynPath
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		quotaMethod, err := s.dao.GetQuotaMethods(ctx, req.Node, req.Gateway)
		if err != nil {
			return err
		}
		reply.QuotaMethod = quotaMethod
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return reply, nil
}

func (s *SnapshotService) snapshotDeploy(ctx context.Context, node, gateway, uuid, deploymentType string) error {
	switch deploymentType {
	case _httpType:
		reply, err := s.dao.BuildPlan(ctx, node, gateway, uuid)
		if err != nil {
			return err
		}
		if err := s.dao.RunPlan(ctx, reply); err != nil {
			return err
		}
	case _grpcType:
		grpcReply, err := s.grpcDao.BuildPlan(ctx, node, gateway, uuid)
		if err != nil {
			return err
		}
		if err := s.grpcDao.RunPlan(ctx, grpcReply); err != nil {
			return err
		}
	default:
		return errors.Errorf("Unrecognized deployment type: %q", deploymentType)
	}
	return nil
}
