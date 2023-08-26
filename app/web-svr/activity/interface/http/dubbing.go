package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func dubbingPersonal(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.DubbingSvc.Personal(ctx, midI))
}

func dubbingRank(ctx *bm.Context) {
	v := new(struct {
		RankType int   `form:"rank_type" validate:"required"`
		Sid      int64 `form:"sid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.DubbingSvc.Rank(ctx, v.RankType, v.Sid))
}
