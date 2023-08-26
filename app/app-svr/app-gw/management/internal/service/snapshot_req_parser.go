package service

import (
	"fmt"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/server/http/util"
)

type SnapshotRequestParser func(ctx *bm.Context, uuid, resource, action string) (*pb.SnapshotActionReq, error)

type snapshotParser struct{}

func (sp snapshotParser) parseListBreakerAPI(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.ListBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_ListBreakerAPIReq{ListBreakerAPIReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseSetBreakerAPI(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.SetBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	if err := util.ParseAction(req, ctx.Request); err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("%+v", err))
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_SetBreakerAPIReq{SetBreakerAPIReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseDeleteBreakerAPI(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.DeleteBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_DeleteBreakerAPIReq{DeleteBreakerAPIReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseEnableBreakerAPI(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.EnableBreakerAPIReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_EnableBreakerAPIReq{EnableBreakerAPIReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseListDynPath(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.ListDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_ListDynPathReq{ListDynPathReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseSetDynPath(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.SetDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	if err := util.ParseClientInfo(req, ctx.Request); err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("%+v", err))
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_SetDynPathReq{SetDynPathReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseDeleteDynPath(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.DeleteDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_DeleteDynPathReq{DeleteDynPathReq: req},
	}
	return out, nil
}

func (sp snapshotParser) parseEnableDynPath(ctx *bm.Context, snapshotID, resource, action string) (*pb.SnapshotActionReq, error) {
	req := &pb.EnableDynPathReq{}
	if err := ctx.Bind(req); err != nil {
		return nil, err
	}
	out := &pb.SnapshotActionReq{
		Uuid:        snapshotID,
		Resource:    resource,
		Action:      action,
		SnapshotReq: &pb.SnapshotActionReq_EnableDynPathReq{EnableDynPathReq: req},
	}
	return out, nil
}
