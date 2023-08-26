package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/baas/api"
)

func mapperModelList(ctx *bm.Context) {
	req := &api.ModelListRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.ModelList(ctx, req))
}

func mapperItemList(ctx *bm.Context) {
	req := &api.ModelItemListRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.ModelItemList(ctx, req))
}

func mapperModelDetail(ctx *bm.Context) {
	req := &api.ModelDetailRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.ModelDetail(ctx, req))
}

func addMapperModel(ctx *bm.Context) {
	req := &api.AddModelRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddModel(ctx, req))
}

func mapperModelFieldList(ctx *bm.Context) {
	req := &api.ModelDetailRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.ModelFieldList(ctx, req))
}

func addMapperModelField(ctx *bm.Context) {
	req := &api.AddModelFieldRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddModelField(ctx, req))
}

func updateMapperModelField(ctx *bm.Context) {
	req := &api.UpdateModelFieldRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UpdateModelField(ctx, req))
}

func deleteMapperModelField(ctx *bm.Context) {
	req := &api.DeleteModelFieldRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.DeleteModelField(ctx, req))
}

func addMapperModelFieldRule(ctx *bm.Context) {
	req := &api.AddModelFieldRuleRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.AddModelFieldRule(ctx, req))
}

func updateMapperModelFieldRule(ctx *bm.Context) {
	req := &api.UpdateModelFieldRuleRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UpdateModelFieldRule(ctx, req))
}
