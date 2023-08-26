package http

import (
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-gw/baas/api"
)

func authz(ctx *bm.Context) {
	username, _ := ctx.Get("username")
	req := &pb.AuthZReq{
		Cookie:   ctx.Request.Header.Get("Cookie"),
		Username: username.(string),
	}
	reply, err := svc.AuthZ(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply.Projects, nil)
}
