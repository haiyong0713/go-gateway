package http

import (
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func topicReport(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.TopicReportReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.TopicReport(ctx, mid, params))
}

func topicResReport(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.TopicResReportReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.TopicResReport(ctx, mid, params))
}

func topicLike(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	params := new(model.TopicLikeReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.TopicLike(ctx, mid, dev, params))
}

func topicDislike(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.TopicDislikeReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.TopicDisLike(ctx, mid, params))
}
