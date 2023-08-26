package http

import (
	"strconv"

	bm "go-common/library/net/http/blademaster"
	model "go-gateway/app/app-svr/app-show/interface/model/rank-list"
)

func rankListIndex(ctx *bm.Context) {
	req := &model.IndexReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	mid, ok := ctx.Get("mid")
	if ok {
		req.Mid = mid.(int64)
	}
	req.FromViewAid, _ = strconv.ParseInt(req.RawFromViewAid, 10, 64)
	ctx.JSON(rankListSvc.RankListIndex(ctx, req))
}
