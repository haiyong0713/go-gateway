package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/interface/model"
)

func favNav(c *bm.Context) {
	var mid int64
	v := new(struct {
		VMid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.FavNav(c, mid, v.VMid))
}

func favArc(c *bm.Context) {
	var mid int64
	v := new(model.FavArcArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.FavArchive(c, mid, v))
}

func favSeasonList(ctx *bm.Context) {
	v := &struct {
		SeasonID int64 `form:"season_id" validate:"min=1"`
	}{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(spcSvc.FavSeasonList(ctx, v.SeasonID))
}
