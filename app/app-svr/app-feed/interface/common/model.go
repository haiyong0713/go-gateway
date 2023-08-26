package common

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/stat"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/pkg/adresource"
)

func CanEnableInlineBanner(mobiApp string, build int64) bool {
	return (mobiApp == "android" && build >= 6150400) ||
		(mobiApp == "iphone" && build >= 61500000) ||
		(mobiApp == "ipad" && build >= 30800100) ||
		(mobiApp == "android_b" && build >= 6190120) ||
		(mobiApp == "iphone_b" && build >= 21503100) ||
		(mobiApp == "android_i" && build >= 3000100) ||
		(mobiApp == "iphone_i" && build >= 64400200) ||
		(mobiApp == "win")
}

// 给闪屏使用，闪屏plat不区分蓝粉版及HD版本
func BannerResourceID(ctx context.Context, mobiApp string, build int64, lessonMode int, plat int8) int64 {
	if lessonMode == 1 {
		return BannerLessonResource(ctx, plat)
	}
	if (mobiApp == "iphone" && build > 8510) || (mobiApp == "android" && build > 5415000) ||
		(mobiApp == "ipad" && build > 12110) || (mobiApp == "iphone_b" && build > 8110) ||
		(mobiApp == "android_b" && build > 591240) || mobiApp == "android_i" ||
		(mobiApp == "iphone_i" && build >= 64400200) {
		if IsInlineBanner(mobiApp, build) {
			return InlineBannerResource(ctx, plat)
		}
	}
	return OldBannerResource(ctx, plat)
}

func IsInlineBanner(mobiApp string, build int64) bool {
	return CanEnableInlineBanner(mobiApp, build)
}

func InlineBannerResource(ctx context.Context, plat int8) int64 {
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusInlineBannerIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusInlineBannerIPad
	case model.PlatAndroid, model.PlatAndroidB:
		scene = adresource.PegasusInlineBannerAndroid
	case model.PlatAndroidI:
		scene = adresource.PegasusInlineBannerAndroidI
	case model.PlatIPhoneI:
		scene = adresource.PegasusInlineBannerIPhoneI
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func OldBannerResource(ctx context.Context, plat int8) int64 {
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusOldBannerIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusOldBannerIPad
	case model.PlatAndroid, model.PlatAndroidB:
		scene = adresource.PegasusOldBannerAndroid
	case model.PlatAndroidI:
		scene = adresource.PegasusOldBannerAndroidI
	case model.PlatAndroidG:
		scene = adresource.PegasusOldBannerAndroidG
	case model.PlatIPhoneI:
		scene = adresource.PegasusOldBannerIOSI
	case model.PlatIPadI:
		scene = adresource.PegasusOldBannerIPadI
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func BannerLessonResource(ctx context.Context, plat int8) int64 {
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusBannerLessonIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusBannerLessonIPad
	case model.PlatAndroid, model.PlatAndroidB:
		scene = adresource.PegasusBannerLessonAndroid
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func BannerTeenagerResource(ctx context.Context, plat int8) int64 {
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusBannerTeenagerIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusBannerTeenagerIPad
	case model.PlatAndroid, model.PlatAndroidB:
		scene = adresource.PegasusBannerTeenagerAndroid
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func Ffcover(cover, from string) string {
	if cover == "" {
		stat.MetricFfCoverTotal.Inc(from)
	}
	return cover
}
