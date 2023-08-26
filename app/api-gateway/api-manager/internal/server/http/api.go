package http

import (
	"encoding/json"
	"go-common/library/ecode"
	"go-gateway/app/api-gateway/api-manager/internal/model"
	"io/ioutil"

	bm "go-common/library/net/http/blademaster"
)

func discoveryList(ctx *bm.Context) {
	ctx.JSON(svc.DiscoveryList(ctx), nil)
}

func appList(ctx *bm.Context) {
	form := new(struct {
		Key         string `form:"key"`
		Tp          int8   `form:"tp"`
		DiscoveryID string `form:"discovery_id"`
		Pn          int64  `form:"pn" validate:"required"`
		Ps          int64  `form:"ps" validate:"required"`
	})
	if err := ctx.Bind(form); err != nil {
		return
	}
	ctx.JSON(svc.AppList(ctx, form.Key, form.DiscoveryID, form.Tp, form.Pn, form.Ps))
}

func addApi(ctx *bm.Context) {
	req := &model.AddApiReq{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	if ok := req.Check(); !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, svc.AddApi(ctx, req))
}

func editApi(ctx *bm.Context) {
	req := &model.AddApiReq{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	if ok := req.Check(); !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, svc.EditApi(ctx, req))

}

func delApi(ctx *bm.Context) {
	req := &struct {
		ID int64 `json:"id" validate:"required"`
	}{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, svc.DelApi(ctx, req.ID))
}
