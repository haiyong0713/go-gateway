package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func rankv3Result(ctx *bm.Context) {
	v := new(struct {
		RankID int64 `form:"rank_id" validate:"required"`
		Pn     int   `form:"pn" validate:"min=1" default:"1"`
		Ps     int   `form:"ps" validate:"min=1" default:"10"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.Rankv3Svc.GetRankByID(ctx, v.RankID, v.Pn, v.Ps))
}
