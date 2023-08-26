package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
	"strconv"
)

func signed(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.SignIn(ctx, mid))
}

func tasksProgress(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.Tasks(ctx, mid))
}

func totalPoints(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.Points(ctx, mid))
}

func matchesStage(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid, _ := midStr.(int64)
	ctx.JSON(service.S10Svc.MatchesCategories(ctx, mid))
}

func actGoods(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid, _ := midStr.(int64)
	ctx.JSON(service.S10Svc.ActGoods(ctx, mid))
}

func userMatchesLotteryInfo(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	params := ctx.Request.Form
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(service.S10Svc.UserLotteryByRobin(ctx, mid, int32(robin)))
}

func updateUserLooteryState(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	params := ctx.Request.Form
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

func stageLottery(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	params := ctx.Request.Form
	robinstr := params.Get("robin")
	robin, err := strconv.ParseInt(robinstr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http updateUserLooteryState error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.StageLottery(ctx, int32(robin), mid))
}

func exchangeGoods(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	if mid <= 0 {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	params := ctx.Request.Form
	gidStr := params.Get("gid")
	gid, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		log.Errorc(ctx, "http exchangeGoods error(%v)", err)
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, service.S10Svc.ExchangeGoods(ctx, mid, int32(gid)))
}

func checkUserProfile(ctx *bm.Context) {
	params := ctx.Request.Form
	tel := params.Get("tel")
	if len(tel) != 11 {
		ctx.JSON(nil, ecode.RequestErr)
	}
	signstr := params.Get("sign")
	timestamp := params.Get("timestamp")
	sourcestr := params.Get("source")
	var (
		source int64
		err    error
	)
	if sourcestr != "" {
		source, err = strconv.ParseInt(sourcestr, 10, 64)
		if err != nil {
			log.Errorc(ctx, "http checkUserProfile error(%v)", err)
			ctx.JSON(nil, ecode.RequestErr)
			return
		}
	}
	ctx.JSON(service.S10Svc.CheckUserProfile(ctx, int32(source), tel, timestamp, signstr))
}

func ackSendFreeFlow(ctx *bm.Context) {
	params := ctx.Request.Form
	tel := params.Get("tel")
	if len(tel) != 11 {
		ctx.JSON(nil, ecode.RequestErr)
	}
	signstr := params.Get("sign")
	timestamp := params.Get("timestamp")
	sourcestr := params.Get("source")
	var (
		source int64
		err    error
	)
	if sourcestr != "" {
		source, err = strconv.ParseInt(sourcestr, 10, 64)
		if err != nil {
			log.Errorc(ctx, "http ackSendFreeFlow error(%v)", err)
			ctx.JSON(nil, ecode.RequestErr)
			return
		}
	}
	ctx.JSON(nil, service.S10Svc.AckSendFreeFlow(ctx, int32(source), tel, timestamp, signstr))
}

func userFlow(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	mid, _ := midStr.(int64)
	ctx.JSON(service.S10Svc.UserFlow(ctx, mid))
}

func otherActivity(ctx *bm.Context) {
	ctx.JSON(service.S10Svc.OtherActivity(ctx))
}
