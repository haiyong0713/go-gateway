package http

import (
	"time"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func searchPubTopics(ctx *bm.Context) {
	params := new(model.SearchPubTopicsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.SearchPubTopics(ctx, params))
}

func usrPubTopics(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.UsrPubTopicsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.UsrPubTopics(ctx, mid, params))
}

func isAlreadyExistedTopic(ctx *bm.Context) {
	params := new(model.IsAlreadyExistedTopicReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.IsAlreadyExistedTopic(ctx, params))
}

func topicPubEvents(ctx *bm.Context) {
	params := new(model.TopicPubEventsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.PubEvents(ctx, params, time.Now().Unix()))
}

func searchRcmdPubTopics(ctx *bm.Context) {
	params := new(model.SearchRcmdPubTopicsReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.SearchRcmdPubTopics(ctx, params))
}

func pubTopicEndpoint(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.PubTopicEndpointReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.PubTopicEndpoint(ctx, mid))
}

func pubUpload(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.PubTopicUploadReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	topicSvc.PubTopicUpload(ctx, mid, params)
	ctx.JSON(nil, nil)
}
