package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/model/like"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
)

const (
	ActivityYouth = "YouthWithYou"
)

// cardsIsJoin 是否加入
func cardsIsJoin(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	err := service.CardSvc.IsJoin(ctx, midI)
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
}

// cardsDraw 抽卡
func cardsDraw(ctx *bm.Context) {
	arg := new(struct {
		Ts int64 `form:"ts"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)

	params := risk(ctx, midI, riskmdl.ActionLottery)
	ctx.JSON(service.CardSvc.Draw(ctx, midI, params, 1, arg.Ts, ActivityYouth))
}

// cardsAddTimes 分享增加抽奖次数
func cardsAddTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(nil, service.CardSvc.AddTimes(ctx, midI))
}

// cardsJoin 加入
func cardsJoin(ctx *bm.Context) {
	arg := new(struct {
		like.HTTPReserveReport
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	report := &like.ReserveReport{
		From:     arg.From,
		Typ:      arg.Typ,
		Oid:      arg.Oid,
		Ip:       metadata.String(ctx, metadata.RemoteIP),
		Platform: arg.Platform,
		Mobiapp:  arg.Mobiapp,
		Buvid:    arg.Buvid,
		Spmid:    arg.Spmid,
	}

	ctx.JSON(nil, service.CardSvc.Join(ctx, midI, report, ActivityYouth))
}

// cardsTask 任务列表
func cardsTask(ctx *bm.Context) {

	mid, _ := ctx.Get("mid")
	midI := mid.(int64)

	ctx.JSON(service.CardSvc.Task(ctx, midI))
}

// cardsFollow 一键关注
func cardsFollow(ctx *bm.Context) {
	arg := new(struct {
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Follow)

	ctx.JSON(nil, service.CardSvc.Follow(ctx, midI, params, arg.MobiApp))
}

// cardsFollower 关注列表
func cardsFollower(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardSvc.Follower(ctx, midI))

}

// cardsTimes 次数
func cardsTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardSvc.Times(ctx, midI))
}

// cardsCards 次数
func cardsCards(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardSvc.Cards(ctx, midI, ActivityYouth))
}

// cardsInviteShare 分享
func cardsInviteShare(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardSvc.InviteShare(ctx, midI, ActivityYouth))
}

// cardsBind 绑定
func cardsBind(ctx *bm.Context) {
	arg := new(struct {
		Token   string `form:"token" validate:"required"`
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Join)

	ctx.JSON(nil, service.CardSvc.Bind(ctx, midI, arg.Token, params, arg.MobiApp, ActivityYouth))
}

// cardsCompose 合成
func cardsCompose(ctx *bm.Context) {
	arg := new(struct {
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Compose)
	ctx.JSON(nil, service.CardSvc.Compose(ctx, midI, params, arg.MobiApp, ActivityYouth))
}

func cardsClick(ctx *bm.Context) {
	arg := new(struct {
		Business string `form:"business" validate:"required"`
		MobiApp  string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Sign)
	ctx.JSON(nil, service.CardSvc.ClickTask(ctx, midI, arg.Business, params, arg.MobiApp))
}

func cardsCardShare(ctx *bm.Context) {
	arg := new(struct {
		Card int64 `form:"card_id" validate:"min=1,max=9"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardSvc.CardShare(ctx, midI, arg.Card, ActivityYouth))
}

func cardsGetCard(ctx *bm.Context) {
	arg := new(struct {
		Token   string `form:"token" validate:"required"`
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21SendCard)
	ctx.JSON(nil, service.CardSvc.GetCard(ctx, midI, arg.Token, params, arg.MobiApp, ActivityYouth))
}

func cardsCardToken(ctx *bm.Context) {
	arg := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(service.CardSvc.CardTokenToMid(ctx, arg.Token, ActivityYouth))

}

func cardsShareToken(ctx *bm.Context) {
	arg := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(service.CardSvc.ShareTokenToMid(ctx, arg.Token, ActivityYouth))

}
