package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"
)

func executeTask(ctx *bm.Context) {
	req := &pb.ExecuteTaskReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	req.Username = getUsername(ctx)
	res, err := rawSvc.Common.ExecuteTask(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
