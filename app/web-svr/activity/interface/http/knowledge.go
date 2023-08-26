package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/service"
)

func badgeProgress(ctx *bm.Context) {
	var mid int64
	p := new(struct {
		ConfigId int64 `form:"config_id" validate:"min=1"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(service.KnowledgeSvr.BadgeProgress(ctx, mid, p.ConfigId))
}

func badgeShare(ctx *bm.Context) {
	p := new(struct {
		ConfigId  int64  `form:"config_id" validate:"min=1"`
		ShareName string `form:"share_name" validate:"required"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	midStr, _ := ctx.Get("mid")
	mid := midStr.(int64)
	ctx.JSON(nil, service.KnowledgeSvr.BadgeShare(ctx, mid, p.ConfigId, p.ShareName))
}

func badgeConfig(ctx *bm.Context) {
	p := new(struct {
		ConfigId   int64  `form:"config_id" validate:"min=1"`
		JsonConfig string `form:"json_config" validate:"required"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	ctx.JSON(nil, service.KnowledgeSvr.BadgeConfig(ctx, p.JsonConfig, p.ConfigId))
}

func knowledgeUser(ctx *bm.Context) {
	p := new(struct {
		Sid int64 `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	midStr, ok := ctx.Get("mid")
	var mid int64
	if ok {
		mid = midStr.(int64)
	}
	ctx.JSON(service.KnowledgeSvr.UserInfo(ctx, mid, p.Sid))

}
