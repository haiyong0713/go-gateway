package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func hotWordVideos(ctx *bm.Context) {
	params := new(model.HotWordVideosReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.HotWordVideos(ctx, params))
}

func hotWordDynamics(ctx *bm.Context) {
	params := new(model.HotWordDynamicReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.HotWordDynamics(ctx, params))
}
