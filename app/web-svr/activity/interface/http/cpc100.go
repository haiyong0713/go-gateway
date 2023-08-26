package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/service"
	"time"
)

func cpc100Info(ctx *bm.Context) {
	midStr, _ := ctx.Get("mid")
	var mid int64
	if midStr != nil {
		mid = midStr.(int64)
	}
	ctx.JSON(service.Cpc100Svr.Info(ctx, mid))
}

func cpc100Reset(ctx *bm.Context) {
	ctx.JSON(nil, service.Cpc100Svr.Reset(ctx))
}

func cpc100Unlock(ctx *bm.Context) {
	v := new(struct {
		Key string `form:"key" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	err := service.Cpc100Svr.Unlock(ctx, mid, v.Key)
	if err == nil {
		err = service.TaskSvr.ActSend(ctx, mid, conf.Conf.Cpc100.UnlockBusiness, conf.Conf.Cpc100.Activity, time.Now().Unix())
	}
	ctx.JSON(nil, err)
}

func cpc100Pv(ctx *bm.Context) {
	ctx.JSON(service.Cpc100Svr.PageView(ctx))
}

func cpc100Total(ctx *bm.Context) {
	ctx.JSON(service.Cpc100Svr.TotalView(ctx))
}
