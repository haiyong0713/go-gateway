package http

import (
	bm "go-common/library/net/http/blademaster"

	model "go-gateway/app/app-svr/app-feed/admin/model/family"
)

func searchFamily(ctx *bm.Context) {
	req := new(model.SearchFamilyReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(spmodeSvc.SearchFamily(ctx, req))
}

func familyBindList(ctx *bm.Context) {
	req := new(model.BindListReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(spmodeSvc.BindList(ctx, req))
}

func unbindFamily(ctx *bm.Context) {
	req := new(model.UnbindReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	var userid int64
	if uid, ok := ctx.Get("uid"); ok {
		userid = uid.(int64)
	}
	var username string
	if uname, ok := ctx.Get("username"); ok {
		username = uname.(string)
	}
	ctx.JSON(nil, spmodeSvc.UnbindFamily(ctx, req, userid, username))
}
