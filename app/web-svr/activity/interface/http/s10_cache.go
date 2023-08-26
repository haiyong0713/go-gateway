package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
)

func signed2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.SignIn(ctx, mid))
}

func tasksProgress2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.Tasks(ctx, mid))
}
func totalPoints2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.Points(ctx, mid))
}

func matchesStage2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.MatchesCategories(ctx, mid))
}

func userMatchesLotteryInfo2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.UserLotteryByRobin(ctx, mid, int32(robin)))
}

func updateUserLooteryState2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	number := params.Get("number")
	name := params.Get("name")
	addr := params.Get("addr")
	ctx.JSON(nil, service.S10Svc.UpdateUserLotteryState(ctx, int32(robin), mid, number, name, addr))
}

func stageLottery2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.StageLottery(ctx, int32(robin), mid))
}

func actGoods2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.ActGoods(ctx, mid))
}

func exchangeGoods2(ctx *bm.Context) {
	params := ctx.Request.Form
	midstr := params.Get("mid")
	mid, err := strconv.ParseInt(midstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	gidStr := params.Get("gid")
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http exchangeGoods error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.ExchangeGoods(ctx, mid, int32(gid)))
}
