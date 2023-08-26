package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/esports/admin/client"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
)

func wallList(ctx *bm.Context) {
	p := new(pb.GetReplyWallModelReq)
	if err := ctx.Bind(p); err != nil {
		return
	}
	res, err := client.EsportsServiceClient.GetReplyWallModel(ctx, p)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	if len(res.ReplyList) == 0 {
		res.ReplyList = make([]*pb.ReplyWallModel, 0)
	}
	ctx.JSON(res, nil)
}

func wallSave(ctx *bm.Context) {
	p := new(pb.SaveReplyWallModel)
	if err := ctx.BindWith(p, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsServiceClient.SaveReplyWall(ctx, p))
}
