package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model"
)

func activityMovieList(ctx *bm.Context) {
	v := new(model.ActivityMovieListReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(webSvc.ActivityMovieList(ctx, mid, v))
}
