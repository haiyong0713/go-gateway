package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"

	"github.com/pkg/errors"
)

func snapshotAction(ctx *bm.Context) {
	snapshotUUID := ctx.Params.ByName("snapshot_id")
	resource := ctx.Params.ByName("resource")
	action := ctx.Params.ByName("action")

	impl, ok := rawSvc.Snapshot.ResolveImpl(resource, action)
	if !ok {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, errors.Errorf("no such resource: %q and action: %q", resource, action)))
		return
	}
	req, err := impl.ReqParser(ctx, snapshotUUID, resource, action)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	reply, err := rawSvc.Snapshot.SnapshotAction(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	innerReply := func() interface{} {
		switch inner := reply.SnapshotReply.(type) {
		case *pb.SnapshotActionReply_Empty:
			return inner.Empty
		case *pb.SnapshotActionReply_ListBreakerAPI:
			return inner.ListBreakerAPI
		case *pb.SnapshotActionReply_ListDynPath:
			return inner.ListDynPath
		default:
			log.Warn("Unrecognized snapshot reply: %+v", reply)
			return inner
		}
	}()
	ctx.JSON(innerReply, nil)
}

func snapshotGRPCAction(ctx *bm.Context) {
	snapshotUUID := ctx.Params.ByName("snapshot_id")
	resource := ctx.Params.ByName("resource")
	action := ctx.Params.ByName("action")

	impl, ok := rawSvc.Snapshot.ResolveGRPCImpl(resource, action)
	if !ok {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, errors.Errorf("no such resource: %q and action: %q", resource, action)))
		return
	}
	req, err := impl.ReqParser(ctx, snapshotUUID, resource, action)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithError(ecode.RequestErr, err))
		return
	}
	reply, err := rawSvc.Snapshot.SnapshotGRPCAction(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	innerReply := func() interface{} {
		switch inner := reply.SnapshotReply.(type) {
		case *pb.SnapshotActionReply_Empty:
			return inner.Empty
		case *pb.SnapshotActionReply_ListBreakerAPI:
			return inner.ListBreakerAPI
		case *pb.SnapshotActionReply_ListDynPath:
			return inner.ListDynPath
		default:
			log.Warn("Unrecognized snapshot reply: %+v", reply)
			return inner
		}
	}()
	ctx.JSON(innerReply, nil)
}

func addSnapshot(ctx *bm.Context) {
	req := &pb.AddSnapshotReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Snapshot.AddSnapshot(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}

func snapshotProfile(ctx *bm.Context) {
	req := &pb.SnapshotProfileReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	res, err := rawSvc.Snapshot.SnapshotProfile(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
