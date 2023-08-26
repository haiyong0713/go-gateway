package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/management/api"
)

func authz(ctx *bm.Context) {
	username, _ := ctx.Get("username")
	req := &pb.AuthZReq{
		Cookie:   ctx.Request.Header.Get("Cookie"),
		Username: username.(string),
	}
	reply, err := rawSvc.Common.AuthZ(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply.Projects, nil)
}

// sidebar auth
func authzSidebar(ctx *bm.Context) {
	reply, err := rawSvc.Common.AuthZSidebar(ctx, getUsername(ctx))
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply, nil)
}

func getUsername(ctx *bm.Context) string {
	username, _ := ctx.Get("username")
	return username.(string)
}
