package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

func addExternalBindRouter(group *bm.RouterGroup) {
	rewardsGroup := group.Group("/bind")
	{
		rewardsGroup.GET("/config", bindConfig)
		rewardsGroup.GET("/info", authSvc.User, bindInfo)
		rewardsGroup.GET("/handler", authSvc.User, bindHandler)
		// rewardsGroup.GET("/bindOpenId", bindOpenId)
	}
}

func bindConfig(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.ExternalBindSvr.GetBindConfig(ctx, &api.GetBindConfigReq{
		ID: v.Id,
	}))
}

func bindInfo(ctx *bm.Context) {
	v := new(struct {
		Id      int64 `form:"id"`
		Refresh int32 `form:"refresh"`
	})
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.ExternalBindSvr.GetBindInfo(ctx, v.Id, mid, v.Refresh))
}

func bindHandler(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id"`
	})
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.ExternalBindSvr.GetBindHandler(ctx, v.Id, mid))
}

func bindOpenId(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id"`
	})
	var mid int64
	if midStr, ok := ctx.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(service.ExternalBindSvr.GetBindOpenId(ctx, v.Id, mid))
}
