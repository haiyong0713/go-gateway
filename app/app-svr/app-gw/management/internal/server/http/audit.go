package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/prettyecode"
)

func listLog(ctx *bm.Context) {
	req := &pb.ListLogReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	res, err := rawSvc.HTTP.ListLog(ctx, req)
	if err != nil {
		ctx.JSON(nil, prettyecode.WithRawError(err))
		return
	}
	ctx.JSON(res, nil)
}
