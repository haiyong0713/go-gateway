package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/client"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"
)

func addAccountBindRouter(group *bm.RouterGroup) {
	group.SetMethodConfig(&bm.MethodConfig{Timeout: xtime.Duration(time.Second * 5)})
	bindGroup := group.Group("/bind")
	{
		bindGroup.GET("/games", getGames)
		bindGroup.GET("/config", getConfig)
		bindGroup.POST("/config/save", saveConfig)
		bindGroup.GET("/config/list", getConfigList)
		bindGroup.GET("/config/external", getExternal)
	}
}
func getExternal(ctx *bm.Context) {
	ctx.JSON(client.ActivityClient.GetBindExternals(ctx, &api.NoReply{}))
}

func getGames(ctx *bm.Context) {
	ctx.JSON(client.ActivityClient.GetBindGames(ctx, &api.NoReply{}))
}

func getConfig(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetBindConfig(ctx, &api.GetBindConfigReq{
		ID:        v.Id,
		SkipCache: false,
	}))
}

func saveConfig(ctx *bm.Context) {
	v := new(struct {
		Id           int64  `json:"id" form:"id"`
		BindPhone    int64  `json:"bind_phone" form:"bind_phone"`
		BindAccount  int64  `json:"bind_account" form:"bind_account"`
		BindType     int64  `json:"bind_type" form:"bind_type"`
		GameType     int64  `json:"game_type" form:"game_type" validate:"required,min=1"`
		ActId        string `json:"act_id" form:"act_id"`
		BindExternal int64  `json:"bind_external" form:"bind_external"`
		Status       int64  `json:"status" form:"status"`
	})
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.SaveBindConfig(ctx, &api.BindConfigInfo{
		ID:           v.Id,
		BindPhone:    v.BindPhone,
		BindAccount:  v.BindAccount,
		BindType:     v.BindType,
		GameType:     v.GameType,
		ActId:        v.ActId,
		BindExternal: v.BindExternal,
		Status:       v.Status,
	}))
}

func getConfigList(ctx *bm.Context) {
	v := new(struct {
		Id int64 `form:"id"`
		Pn int64 `form:"pn" validate:"required,min=1"`
		Ps int64 `form:"ps" validate:"required,max=100"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.GetBindConfigList(ctx, &api.GetBindConfigListReq{
		ID: v.Id,
		Pn: v.Pn,
		Ps: v.Ps,
	}))
}
