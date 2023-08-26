package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/topic/interface/internal/model"
)

func subFavTopics(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.FavSubListReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(topicSvc.SubFavTopics(ctx, mid, params))
}

func addFav(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.AddFavReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(nil, topicSvc.AddFav(ctx, mid, params))
}

func cancelFav(ctx *bm.Context) {
	var mid int64
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	params := new(model.CancelFavReq)
	if err := ctx.Bind(params); err != nil {
		return
	}
	ctx.JSON(nil, topicSvc.CancelFav(ctx, mid, params))
}
