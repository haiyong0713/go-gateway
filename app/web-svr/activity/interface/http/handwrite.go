package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func memberCount(ctx *bm.Context) {

	ctx.JSON(service.HandWriteSvc.AwardMemberCount(ctx))
}

func rank(ctx *bm.Context) {

	ctx.JSON(service.HandWriteSvc.Rank(ctx))
}

func personal(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.HandWriteSvc.Personal(ctx, midI))

}

func addHwLotteryTimes(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.HandWriteSvc.AddLotteryTimes(ctx, midI))
}

func coin(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.HandWriteSvc.Coin(ctx, midI))
}

func handwriteMemberCount(ctx *bm.Context) {
	ctx.JSON(service.HandWriteSvc.AwardMemberCount2021(ctx))
}

func handwritePersonal(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	midI := mid.(int64)
	ctx.JSON(service.HandWriteSvc.Personal2021(ctx, midI))
}

func handwritePersonalInternal(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.HandWriteSvc.Personal2021(ctx, v.Mid))
}
