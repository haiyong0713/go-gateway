package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func rankPersonal(ctx *bm.Context) {
	v := new(struct {
		SID           int64 `form:"sid" validate:"required"`
		AttributeType int   `form:"rank_attribute" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.RankSvc.Personal(ctx, v.SID, v.AttributeType, midI))
}

func rankResult(ctx *bm.Context) {
	v := new(struct {
		SID           int64 `form:"sid" validate:"required"`
		AttributeType int   `form:"rank_attribute" validate:"required"`
		Pn            int   `form:"pn" validate:"min=1" default:"1"`
		Ps            int   `form:"ps" validate:"min=1" default:"10"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.RankSvc.GetRankDetail(ctx, v.SID, v.AttributeType, v.Ps, v.Pn))
}
