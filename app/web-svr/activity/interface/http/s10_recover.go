package http

import (
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func redeliveryGift(ctx *bm.Context) {
	var (
		err    error
		params = new(struct {
			ID  int64 `form:"id" validate:"gt=0"`
			Mid int64 `form:"mid" validate:"gt=0"`
			Gid int32 `form:"gid" validate:"gt=0"`
			Act int32 `form:"act" validate:"gte=0"`
		})
	)
	if err = ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(nil, service.S10Svc.RedeliveryGift(ctx, params.Gid, params.Act, params.ID, params.Mid))
}

func pointCache(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.PointFlush(ctx, mid))
}

func delGoodsStock(ctx *bm.Context) {
	params := ctx.Request.Form
	gidstr := params.Get("gid")
	gid, err := strconv.ParseInt(gidstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.DelGoodsStockByGid(ctx, int32(gid)))
}

func delRoundGoodsStock(ctx *bm.Context) {
	params := ctx.Request.Form
	gidstr := params.Get("gid")
	gid, err := strconv.ParseInt(gidstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.DelRoundGoodsStockByGid(ctx, int32(gid)))
}

func delUserStatic(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.DelUserStatic(ctx, mid))
}

func delRoundUserStatic(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.DelRoundUserStatic(ctx, mid))
}

func userCostStatic(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.UserCostStaticFlush(ctx, mid))
}

func userCostPointsDetail(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.UserCostPointsDetailFlush(ctx, mid))
}

func userLotteryInfo(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.UserLotteryInfoFlush(ctx, mid))
}
