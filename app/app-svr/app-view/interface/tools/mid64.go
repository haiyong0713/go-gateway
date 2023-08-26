package tools

import (
	"context"
	"go-common/library/stat/prom"
	"math"
	"strconv"

	"go-common/component/metadata/device"
	meta "go-common/component/metadata/fawkes"
	"go-common/library/account/int64mid"
	bm "go-common/library/net/http/blademaster"
)

const (
	_appKey = "appkey"
	_build  = "build"
)

type SupportMid64App struct{}

func CheckMid64SupportGRPC(ctx context.Context, appkeyVersion *int64mid.AppkeyVersion) bool {
	dev, ok := device.FromContext(ctx)
	if !ok {
		prom.BusinessInfoCount.Incr("mid64:dev为空")
		return false
	}

	if appkeyVersion == nil || len(appkeyVersion.AppkeyToVersion) == 0 {
		prom.BusinessInfoCount.Incr("mid64:appkeyVersion为空")
		return defaultCheckMid64Support(dev)
	}

	fwk, ok := meta.FromContext(ctx)
	if !ok {
		prom.BusinessInfoCount.Incr("mid64:fwk为空")
		return defaultCheckMid64Support(dev)
	}

	supportMid64App := appkeyVersion.SupportInt64(strconv.FormatInt(dev.Build, 10), fwk.AppKey)
	prom.BusinessInfoCount.Incr("mid64:SupportMid64App:" + strconv.FormatBool(supportMid64App) + ":" + fwk.AppKey)
	return supportMid64App
}

func CheckMid64Support(appkeyVersion *int64mid.AppkeyVersion) bm.HandlerFunc {
	return func(ctx *bm.Context) {
		req := ctx.Request
		appKey := req.Form.Get(_appKey)
		build := req.Form.Get(_build)

		if appkeyVersion == nil || len(appkeyVersion.AppkeyToVersion) == 0 {
			ctx.Context = context.WithValue(ctx.Context, SupportMid64App{}, false)
			prom.BusinessInfoCount.Incr("mid64:appkeyVersion为空-http")
			return
		}

		supportMid64App := false
		if appkeyVersion != nil && len(appkeyVersion.AppkeyToVersion) > 0 {
			supportMid64App = appkeyVersion.SupportInt64(build, appKey)
		}

		ctx.Context = context.WithValue(ctx.Context, SupportMid64App{}, supportMid64App)
	}
}

func CheckNeedFilterMid64(c context.Context) bool {
	v, _ := c.Value(SupportMid64App{}).(bool)
	return !v
}

func IsInt32Mid(v int64) bool {
	return v <= math.MaxInt32
}

func defaultCheckMid64Support(dev device.Device) bool {
	return (dev.RawMobiApp == "android" && dev.Build >= 6500000) ||
		(dev.RawMobiApp == "iphone" && dev.Build >= 65000000) ||
		(dev.RawMobiApp == "ipad" && dev.Build >= 33000000) ||
		(dev.RawMobiApp == "iphone_i" && dev.Build >= 65000000) ||
		(dev.RawMobiApp == "android_i" && dev.Build >= 6500000) ||
		(dev.RawMobiApp == "android_b" && dev.Build >= 6500000) ||
		(dev.RawMobiApp == "iphone_b" && dev.Build >= 65000000) ||
		(dev.RawMobiApp == "android_hd" && dev.Build >= 1070000)
}
