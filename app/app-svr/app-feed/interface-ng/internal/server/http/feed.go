package http

import (
	"bytes"
	"encoding/json"
	"html/template"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/server/http/asset"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"
)

func validateSession(ctx *bm.Context) {
	req := &session.IndexSession{}
	if err := ctx.BindWith(req, binding.JSON); err != nil {
		return
	}
	ctx.JSON(svc.ValidateSession(ctx, req))
}

func compareSession(ctx *bm.Context) {
	req := &session.IndexSession{}
	if err := ctx.BindWith(req, binding.JSON); err != nil {
		return
	}
	res, err := svc.CompareSession(ctx, req)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	compare, err := asset.Asset("static/compare.html")
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	tmpl, err := template.New("").Parse(string(compare))
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	b := bytes.NewBuffer(nil)
	if err := tmpl.Execute(b, res); err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.Bytes(200, "text/html", b.Bytes())
}

func comparePage(ctx *bm.Context) {
	page, err := asset.Asset("static/compare_index.html")
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.Bytes(200, "text/html", page)
}

func compareFormSession(ctx *bm.Context) {
	rawSession := ctx.Request.Form.Get("session")
	session := &session.IndexSession{}
	if err := json.Unmarshal([]byte(rawSession), &session); err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	res, err := svc.CompareSession(ctx, session)
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	compare, err := asset.Asset("static/compare.html")
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	tmpl, err := template.New("").Parse(string(compare))
	if err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	b := bytes.NewBuffer(nil)
	if err := tmpl.Execute(b, res); err != nil {
		ctx.String(500, "%+v", err)
		return
	}
	ctx.Bytes(200, "text/html", b.Bytes())
}

func index(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	header := ctx.Request.Header
	applist := header.Get("AppList")
	deviceInfo := header.Get("DeviceInfo")

	param := &feed.IndexParam{}
	if err := ctx.Bind(param); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}

	_, ok := cdm.Columnm[param.Column]
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	// 兼容老的style逻辑，3为新单列
	style := int(cdm.Columnm[param.Column])
	if style == 1 {
		style = 3
	}
	dev, _ := device.FromContext(ctx)
	ctxDevice := feedcard.NewCtxDevice(&dev)
	req := &model.IndexNgReq{
		Mid:        mid,
		FeedParam:  *param,
		Style:      style,
		AppList:    applist,
		DeviceInfo: deviceInfo,
		Device:     ctxDevice,
	}
	ctx.JSON(svc.IndexNg(ctx, req))
}
