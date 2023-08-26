package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func dmVote(ctx *bm.Context) {
	req := &model.DmVoteReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	midInter, _ := ctx.Get("mid")
	req.Mid = midInter.(int64)
	if ck, err := ctx.Request.Cookie("buvid3"); err == nil {
		req.Buvid = ck.Value
	}
	ctx.JSON(webSvc.DmVote(ctx, req))
}
