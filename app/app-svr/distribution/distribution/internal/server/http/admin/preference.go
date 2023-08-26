package http

import (
	bm "go-common/library/net/http/blademaster"

	adminservice "go-gateway/app/app-svr/distribution/distribution/internal/service/admin"
)

func userDevice(ctx *bm.Context) {
	req := &adminservice.UserDeviceRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.UserDevice(ctx, req))
}

func devicePreference(ctx *bm.Context) {
	req := &adminservice.DevicePreferenceRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.DevicePreference(ctx, req))
}
