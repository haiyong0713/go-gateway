package http

import bm "go-common/library/net/http/blademaster"

func authorBindState(ctx *bm.Context) {
	mid, _ := ctx.Get("mid")
	author, err := svrThirdsdk.AuthorBindState(ctx, mid.(int64))
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["author"] = author
	ctx.JSON(res, nil)
}
