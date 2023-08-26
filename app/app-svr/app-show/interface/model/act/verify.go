package act

import (
	"context"

	"go-gateway/app/app-svr/app-show/interface/conf"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	IphoneMobi  = "iphone"
	IpadMobi    = "ipad"
	AndroidMobi = "android"
	AndroidBlue = "android_b"
	IphoneBlue  = "iphone_b"
	AndroidI    = "android_i"
	IphoneI     = "iphone_i"
)

// IsInlineLow inline tab组件低版本62以前版本
// ios 蓝版是 10030
// 安卓所有蓝版
func IsInlineLow(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.InlineLow, &feature.OriginResutl{
		BuildLimit: (mobiApp == IphoneMobi && build <= 10020) || (mobiApp == AndroidMobi && build < 6020000) || (mobiApp == IpadMobi && build <= 12390) || mobiApp == AndroidBlue || (mobiApp == IphoneBlue && build <= 10030),
	})
}

// IsNewFeed .
func IsNewFeed(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.NewFeed, &feature.OriginResutl{
		BuildLimit: (mobiApp == IphoneMobi && build >= 10085) || (mobiApp == AndroidMobi && build >= 6040000) || (mobiApp == IpadMobi && build > 12390) || mobiApp == AndroidI || mobiApp == IphoneI,
	})
}

// IsSelectLow select组件低版本68以前版本
func IsSelectLow(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.SelectLow, &feature.OriginResutl{
		BuildLimit: (mobiApp == IphoneMobi && build <= 10200) || (mobiApp == AndroidMobi && build < 6080000) || (mobiApp == IpadMobi && build <= 12430) || mobiApp == AndroidBlue || (mobiApp == IphoneBlue && build <= 10200),
	})
}

// IsVersion615Low .
func IsVersion615Low(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.Version615, &feature.OriginResutl{
		BuildLimit: (mobiApp == IphoneMobi && build < 61500000) || (mobiApp == AndroidMobi && build < 6150000) || mobiApp == IpadMobi || mobiApp == AndroidI || mobiApp == IphoneI,
	})
}

// IsVersion618Low .
func IsVersion618Low(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(c, featureCfg.FeatureBuildLimit.Version618, &feature.OriginResutl{
		BuildLimit: (mobiApp == IphoneMobi && build < 61800000) || (mobiApp == AndroidMobi && build < 6180000) || mobiApp == IpadMobi || mobiApp == AndroidI || mobiApp == IphoneI,
	})
}
