package http

import (
	bm "go-common/library/net/http/blademaster"
)

func getOnelink(ctx *bm.Context) {
	var (
		mid int64
		gid int64 = 1667
		err error
	)
	arg := new(struct {
		BSource string `form:"bsource"`
	})
	if err = ctx.Bind(arg); err != nil {
		return
	}
	if midInt, ok := ctx.Get("mid"); ok {
		mid = midInt.(int64)
	}
	ctx.JSON(webSvc.GetOnelink(ctx, mid, gid, arg.BSource))
}
