package http

import bm "go-common/library/net/http/blademaster"

func renewPosterFromAllTeam(ctx *bm.Context) {
	ctx.JSON(esSvc.RenewPosterFromTeam(ctx))
}

func renewPosterByTeamIDList(ctx *bm.Context) {
	v := new(struct {
		List []int64 `form:"list,split" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(esSvc.RenewPosterByTeamID(ctx, v.List))
}
