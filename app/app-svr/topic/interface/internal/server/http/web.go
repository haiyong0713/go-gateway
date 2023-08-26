package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func webTopicInfo(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.WebTopicInfoReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebTopicInfo(ctx, mid, params))
}

func webTopicCards(ctx *bm.Context) {
	params := new(model.WebTopicCardsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebTopicCards(ctx, params))
}

func webTopicFoldCards(ctx *bm.Context) {
	params := new(model.WebTopicFoldCardsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebTopicFoldCards(ctx, params))
}

func webSubFavTopics(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.WebFavSubListReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebSubFavTopics(ctx, mid, params))
}

func webDynamicRcmdTopics(ctx *bm.Context) {
	params := new(model.WebDynamicRcmdReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.WebDynamicRcmdTopics(ctx, params))
}
