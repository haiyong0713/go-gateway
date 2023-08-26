package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/model/like"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/service"
)

// springFestivalIsJoin 是否加入
func springFestivalIsJoin(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	err := service.SpringFestival2021Svc.IsJoin(ctx, midI)
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
}

// springFestivalDraw 抽卡
func springFestivalDraw(ctx *bm.Context) {
	arg := new(struct {
		Ts int64 `form:"ts"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)

	params := risk(ctx, midI, riskmdl.ActionLottery)
	ctx.JSON(service.SpringFestival2021Svc.Draw(ctx, midI, params, 1, arg.Ts))
}

// springFestivalAddTimes 分享增加抽奖次数
func springFestivalAddTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(nil, service.SpringFestival2021Svc.AddTimes(ctx, midI))
}

// springFestivalJoin 加入
func springFestivalJoin(ctx *bm.Context) {
	arg := new(struct {
		like.HTTPReserveReport
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
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

	ctx.JSON(nil, service.SpringFestival2021Svc.Join(ctx, midI, report))
}

// springFestivalTask 任务列表
func springFestivalTask(ctx *bm.Context) {

	mid, _ := ctx.Get("mid")
	midI := mid.(int64)

	ctx.JSON(service.SpringFestival2021Svc.Task(ctx, midI))
}

// springFestivalFollow 一键关注
func springFestivalFollow(ctx *bm.Context) {
	arg := new(struct {
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Follow)

	ctx.JSON(nil, service.SpringFestival2021Svc.Follow(ctx, midI, params, arg.MobiApp))
}

// springFestivalFollower 关注列表
func springFestivalFollower(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.SpringFestival2021Svc.Follower(ctx, midI))

}

// springFestivalTimes 次数
func springFestivalTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.SpringFestival2021Svc.Times(ctx, midI))
}

// springFestivalCards 次数
func springFestivalCards(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.SpringFestival2021Svc.Cards(ctx, midI))
}

// springFestivalInviteShare 分享
func springFestivalInviteShare(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.SpringFestival2021Svc.InviteShare(ctx, midI))
}

// springFestivalBind 绑定
func springFestivalBind(ctx *bm.Context) {
	arg := new(struct {
		Token   string `form:"token" validate:"required"`
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Join)

	ctx.JSON(nil, service.SpringFestival2021Svc.Bind(ctx, midI, arg.Token, params, arg.MobiApp))
}

// springFestivalCompose 合成
func springFestivalCompose(ctx *bm.Context) {
	arg := new(struct {
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Compose)
	ctx.JSON(nil, service.SpringFestival2021Svc.Compose(ctx, midI, params, arg.MobiApp))
}

func springFestivalClick(ctx *bm.Context) {
	arg := new(struct {
		Business string `form:"business" validate:"required"`
		MobiApp  string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21Sign)
	ctx.JSON(nil, service.SpringFestival2021Svc.ClickTask(ctx, midI, arg.Business, params, arg.MobiApp))
}

func springFestivalCardShare(ctx *bm.Context) {
	arg := new(struct {
		Card int64 `form:"card_id" validate:"min=1,max=5"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.SpringFestival2021Svc.CardShare(ctx, midI, arg.Card))
}

func springFestivalGetCard(ctx *bm.Context) {
	arg := new(struct {
		Token   string `form:"token" validate:"required"`
		MobiApp string `form:"mobiapp"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	params := risk(ctx, midI, riskmdl.ActionSf21SendCard)
	ctx.JSON(nil, service.SpringFestival2021Svc.GetCard(ctx, midI, arg.Token, params, arg.MobiApp))
}

func springFestivalCardToken(ctx *bm.Context) {
	arg := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(service.SpringFestival2021Svc.CardTokenToMid(ctx, arg.Token))

}

func springFestivalShareToken(ctx *bm.Context) {
	arg := new(struct {
		Token string `form:"token" validate:"required"`
	})
	if err := ctx.Bind(arg); err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(service.SpringFestival2021Svc.ShareTokenToMid(ctx, arg.Token))

}
