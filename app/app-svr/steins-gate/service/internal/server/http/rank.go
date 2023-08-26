package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

func rankList(ctx *bm.Context) {
	req := &api.RankListReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	mid, ok := ctx.Get("mid")
	if ok {
		req.CurrentMid, _ = mid.(int64)
	}
	ctx.JSON(svc.RankList(ctx, req))
}

func rankScoreSubmit(ctx *bm.Context) {
	req := &api.RankScoreSubmitReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	mid, ok := ctx.Get("mid")
	if ok {
		req.CurrentMid, _ = mid.(int64)
	}
	ctx.JSON(nil, svc.RankScoreSubmit(ctx, req))

}
