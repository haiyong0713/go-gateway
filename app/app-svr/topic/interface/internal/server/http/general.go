package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func generalFeedList(ctx *bm.Context) {
	params := new(model.GeneralFeedListReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.GeneralFeedList(ctx, params))
}

func topicTimeLine(ctx *bm.Context) {
	params := new(model.TopicTimeLineReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.TopicTimeLine(ctx, params))
}
