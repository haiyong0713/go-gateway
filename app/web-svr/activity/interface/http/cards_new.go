package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/model/like"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
)

// cardsIsJoin 是否加入
func cardsNewIsJoin(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	err := service.CardV2Svr.IsJoin(ctx, midI, arg.Activity)
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
}

// cardsJoin 加入
func cardsNewJoin(ctx *bm.Context) {
	arg := new(struct {
		like.HTTPReserveReport
		Activity string `form:"activity" validate:"required"`
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
	ctx.JSON(nil, service.CardV2Svr.Join(ctx, midI, report, arg.Activity))
}

// cardsNewBind 绑定
func cardsNewBind(ctx *bm.Context) {
	arg := new(struct {
		Token    string `form:"token" validate:"required"`
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)

	ctx.JSON(nil, service.CardV2Svr.Bind(ctx, midI, arg.Token, arg.Activity))
}

func cardsDrawNew(ctx *bm.Context) {
	arg := new(struct {
		Ts       int64  `form:"ts"`
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionLottery)
	ctx.JSON(service.CardV2Svr.Draw(ctx, midI, params, 1, arg.Ts, arg.Activity))
}

// cardsAddTimesNew 分享增加抽奖次数
func cardsAddTimesNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(nil, service.CardV2Svr.AddTimes(ctx, midI, arg.Activity))
}

// cardsTimes 次数
func cardsTimesNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardV2Svr.Times(ctx, midI, arg.Activity))
}

// cardsCardsNew 次数
func cardsCardsNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardV2Svr.Cards(ctx, midI, arg.Activity))
}

func cardsComposeNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionCompose)
	ctx.JSON(nil, service.CardV2Svr.Compose(ctx, midI, params, arg.Activity))
}

// cardsInviteShareNew 分享
func cardsInviteShareNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardV2Svr.InviteShare(ctx, midI, arg.Activity))
}

func cardsCardShareNew(ctx *bm.Context) {
	arg := new(struct {
		Activity string `form:"activity" validate:"required"`
		Card     int64  `form:"card_id" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.CardV2Svr.CardShare(ctx, midI, arg.Card, arg.Activity))
}

func cardsGetCardNew(ctx *bm.Context) {
	arg := new(struct {
		Token    string `form:"token" validate:"required"`
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSendCard)
	ctx.JSON(nil, service.CardV2Svr.GetCard(ctx, midI, arg.Token, params, arg.Activity))
}

func cardsCardTokenNew(ctx *bm.Context) {
	arg := new(struct {
		Token    string `form:"token" validate:"required"`
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(service.CardV2Svr.CardTokenToMid(ctx, arg.Token, arg.Activity))

}

func cardsShareTokenNew(ctx *bm.Context) {
	arg := new(struct {
		Token    string `form:"token" validate:"required"`
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	ctx.JSON(service.CardV2Svr.ShareTokenToMid(ctx, arg.Token, arg.Activity))

}

func cardsSendPoints(ctx *bm.Context) {
	arg := new(struct {
		Activity  string `form:"activity" validate:"required"`
		Business  string `form:"business" validate:"required"`
		Timestamp int64  `form:"timestamp"`
	})
	if err := ctx.Bind(arg); err != nil {
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, arg.Business)
	ctx.JSON(nil, service.TaskSvr.CardsActSend(ctx, midI, arg.Business, arg.Activity, arg.Timestamp, params, false))
}
