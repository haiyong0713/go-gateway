package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func vertSearchTopics(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.VertSearchTopicsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.VertSearchTopics(ctx, mid, params))
}

func vertTopicCenter(ctx *bm.Context) {
	params := new(model.VertTopicCenterReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.VertTopicCenter(ctx, params))
}

func vertTopicOnline(ctx *bm.Context) {
	params := new(model.VertTopicOnlineReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.VertTopicOnline(ctx, params))
}
