package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

func thresholdConfig(ctx *bm.Context) {
	req := new(model.ConfigRuleReq)
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(nil, svc.ThresholdConfig(ctx, req, cookie))
}

func customizedRules(ctx *bm.Context) {
	req := new(struct {
		Team string `form:"team"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(svc.CustomizedTeamRule(ctx, req.Team, cookie))
}

func editRule(ctx *bm.Context) {
	req := new(struct {
		ID        int64  `form:"id"`
		Team      string `form:"team"`
		Threshold int64  `form:"threshold"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(nil, svc.EditRule(ctx, req.Team, req.ID, req.Threshold, cookie))
}

func deleteRule(ctx *bm.Context) {
	req := new(struct {
		ID   int64  `form:"id"`
		Team string `form:"team"`
	})
	if err := ctx.Bind(req); err != nil {
		return
	}
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(nil, svc.DeleteRule(ctx, req.Team, req.ID, cookie))
}

func fetchRoleTree(ctx *bm.Context) {
	cookie := ctx.Request.Header.Get("Cookie")
	reply, err := svc.FetchRoleTree(ctx, cookie)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(reply, nil)
}

func rootReceiverGroup(ctx *bm.Context) {
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(nil, svc.RootReceiverGroups(ctx, cookie))
}

func myService(ctx *bm.Context) {
	cookie := ctx.Request.Header.Get("Cookie")
	ctx.JSON(svc.MyService(ctx, cookie))
}
