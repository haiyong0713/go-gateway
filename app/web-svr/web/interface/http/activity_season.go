package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	webgrpc "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
)

func activitySeason(ctx *bm.Context) {
	v := new(webgrpc.ActivitySeasonReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	// get mid
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(webSvc.ActivitySeason(ctx, mid, v, reqBuvid(ctx)))
}

func activityArchive(ctx *bm.Context) {
	v := new(webgrpc.ActivityArchiveReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	// get mid
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(webSvc.ActivityArchive(ctx, mid, v, reqBuvid(ctx)))
}

func activityLiveTimeInfo(ctx *bm.Context) {
	v := new(webgrpc.ActivityLiveTimeInfoReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	// get mid
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(webSvc.ActivityLiveTimeInfo(ctx, mid, v, reqBuvid(ctx)))
}

func activitySeasonClick(ctx *bm.Context) {
	v := new(struct {
		OrderType int32  `form:"order_type" validate:"required"`
		ReserveID int64  `form:"reserve_id"`
		From      string `form:"from"`
		Type      string `form:"type"`
		Oid       int64  `form:"oid"`
		SeasonID  int64  `form:"season_id"`
		Spmid     string `form:"spmid"`
		Action    int64  `form:"action" validate:"min=0,max=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	req := &webgrpc.ClickActivitySeasonReq{Spmid: v.Spmid, Action: v.Action}
	switch v.OrderType {
	case int32(webgrpc.OrderType_TypeOrderActivity):
		req.OrderType = webgrpc.OrderType_TypeOrderActivity
		if v.ReserveID <= 0 || v.From == "" || v.Type == "" || v.Oid <= 0 {
			ctx.JSON(nil, ecode.RequestErr)
			return
		}
		req.Param = &webgrpc.ClickActivitySeasonReq_ReserveParam{ReserveParam: &webgrpc.ReserveActivityParam{ReserveId: v.ReserveID, From: v.From, Type: v.Type, Oid: v.Oid}}
	case int32(webgrpc.OrderType_TypeFavSeason):
		req.OrderType = webgrpc.OrderType_TypeFavSeason
		if v.SeasonID <= 0 {
			ctx.JSON(nil, ecode.RequestErr)
			return
		}
		req.Param = &webgrpc.ClickActivitySeasonReq_FavParam{FavParam: &webgrpc.FavSeasonParam{SeasonId: v.SeasonID}}
	default:
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	buvid := reqBuvid(ctx)
	ctx.JSON(nil, webSvc.ClickActivitySeason(ctx, mid, req, buvid))
}
