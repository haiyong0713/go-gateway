package http

import bm "go-common/library/net/http/blademaster"

func hisSearch(ctx *bm.Context) {
	v := new(struct {
		Keyword  string `form:"keyword" validate:"required"`
		Pn       int64  `form:"pn" validate:"min=1"`
		Business string `form:"business" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	midInter, _ := ctx.Get("mid")
	mid := midInter.(int64)
	var buvid string
	if ck, err := ctx.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	ctx.JSON(srvWeb.HisSearch(ctx, mid, buvid, v.Keyword, v.Pn, v.Business))
}
