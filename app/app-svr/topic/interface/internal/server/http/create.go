package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func hasCreateJurisdiction(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	ctx.JSON(topicSvc.HasCreateJurisdiction(ctx, mid))
}

func createTopic(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.CreateTopicReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.CreateTopic(ctx, mid, params))
}

func webCreateTopic(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.CreateTopicReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebCreateTopic(ctx, mid, params))
}
