package http

import (
	bm "go-common/library/net/http/blademaster"
	b "go-common/library/net/http/blademaster/binding"
	exammdl "go-gateway/app/web-svr/activity/interface/model/examination"
	"go-gateway/app/web-svr/activity/interface/service"
)

// 获取当前点赞数量
func examinationUp(ctx *bm.Context) {
	v := new(exammdl.UpReq)
	path := ctx.Request.URL.Path
	if err := ctx.BindWith(v, b.JSON); err != nil {
		return
	}
	var midI int64
	mid, ok := ctx.Get("mid")
	if ok {
		midI = mid.(int64)
	}
	ctx.JSON(service.ExaminationSvr.UpInfo(ctx, midI, path, v))
}
