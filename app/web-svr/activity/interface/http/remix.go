package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func remixMemberCount(ctx *bm.Context) {

	ctx.JSON(service.RemixSvc.MemberCount(ctx))
}

func remixPersonal(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.RemixSvc.Personal(ctx, midI))
}

func remixChildRank(ctx *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.RemixSvc.ChildRank(ctx, v.Sid))
}

func remixRank(ctx *bm.Context) {
	v := new(struct {
		RankType int `form:"rank_type" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.RemixSvc.Rank(ctx, v.RankType))
}
