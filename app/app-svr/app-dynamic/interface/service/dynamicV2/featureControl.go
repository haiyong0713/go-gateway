package dynamicV2

import (
	"context"

	"go-common/component/metadata/device"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	dyncomn "git.bilibili.co/bapis/bapis-go/dynamic/common"
	"google.golang.org/grpc"
)

// 是否显示新话题信息
// 目前只在粉双端，粉Pad，iosHD上显示
func (s *Service) isDynNewTopicView(ctx context.Context, general *mdlv2.GeneralParam) (isNewTopic bool) {
	return feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynNewTopic, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynNewTopicAndroid, s.c.BuildLimit.DynNewTopicIOS) ||
			(general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DynNewTopicIOSHD) ||
			(general.IsPad() && general.GetBuild() >= s.c.BuildLimit.DynNewTopicIOS),
	})
}

// 是否在动态垂搜使用新话题搜索
// 目前只在粉双端使用
func (s *Service) isDynNewTopicVerticalSearch(ctx context.Context, general *mdlv2.GeneralParam) (isNewTopicSearch bool) {
	return feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynNewTopic, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynNewTopicAndroid, s.c.BuildLimit.DynNewTopicIOS),
	})
}

// 是否屏蔽 校园榜单自动开放功能
// 只在6.68以上的版本开启 低版本对于未开放的榜单tab直接屏蔽
func (s *Service) isDisableDynCampusBillboardAutoOpen(ctx context.Context, general *mdlv2.GeneralParam, tabType int64, v dyncomn.MajorTabStatus) bool {
	if tabType != _campusBillboard {
		return false
	}
	if feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynSchoolBillboardAutoOpen, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.Less, s.c.BuildLimit.DynSchoolBillboardAutoOpenAndroid, s.c.BuildLimit.DynSchoolBillboardAutoOpenIOS),
	}) {
		if v == dyncomn.MajorTabStatus_TAB_STATUS_UNOPEN {
			return true
		}
	}
	return false
}

// 是否启用新话题 话题集订阅更新卡
func (s *Service) isDynNewTopicSet(ctx context.Context, general *mdlv2.GeneralParam) bool {
	return feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynNewTopicSet, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynNewTopicSetAndroid, s.c.BuildLimit.DynNewTopicSetIOS),
	})
}

// 校园 是否在话题讨论页启用发布按钮
func (s *Service) isCampusTopicPublishBtn(ctx context.Context, general *mdlv2.GeneralParam) bool {
	return feature.GetBuildLimit(ctx, s.c.Feature.FeatureBuildLimit.DynSchoolTopicPublishBtn, &feature.OriginResutl{
		BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynSchoolTopicPublishBtnAndroid, s.c.BuildLimit.DynSchoolTopicPublishBtnIOS),
	})
}

func meet(ctx context.Context, fgt conf.FeatureGateItem) bool {
	if !fgt.Enable {
		return false
	}
	dev, _ := device.FromContext(ctx)
	fn := func(d device.Device, p conf.TargetPlatform) bool {
		ret := true
		if len(p.Device) > 0 {
			ret = ret && p.Device == d.Device
		}
		if p.Build != 0 {
			ret = ret && p.Build == d.Build
		}
		return ret
	}
	for _, p := range fgt.Platforms {
		// 只要命中一个规则直接跳出
		if p.MobiApp == dev.RawMobiApp && p.Channel == dev.Channel && fn(dev, p) {
			return true
		}
	}
	return false
}

// 特殊处理 去除游戏附加卡
func (s *Service) isNoGameAttach(ctx context.Context) bool {
	// 优先本地配置
	if s.c.FeatureGate.NoGameAttach.Enable {
		return meet(ctx, s.c.FeatureGate.NoGameAttach)
	}
	const (
		_dynamicOtype = 4
	)
	// 查询远程配置
	res, err := s.resourceDao.EntrancesIsHidden(ctx, []int64{}, _dynamicOtype)
	if err != nil {
		return false
	}
	return res
}

// 由于在单次请求中计算某些功能是否开启的 UnaryServerInterceptor
func (s *Service) FeatureGateUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = mdlv2.NewFeatureStatusCtx(ctx, &mdlv2.FeatureStatus{
			NoGameAttach: &mdlv2.FeatureLazySwitch{Fn: s.isNoGameAttach},
		})
		return handler(ctx, req)
	}
}

// 首映中稿件跳入时自动唤起首映浮层
func (s *Service) inArchivePremiereArg() func(uri string) string {
	return model.SuffixHandler("auto_float_layer=7")
}

// 是否启用 校园官号管理入口
func (s *Service) isCampusMngCapable(_ context.Context, general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit = 6790000
		_iosLimit     = 67900000
	)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit)
}

// IP地址展示支持的客户端
func (s *Service) isPubLocationCapable(_ context.Context, general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit   = 6800000
		_androidHDLimit = 1240000
		_iosLimit       = 68000000
		_iosHDLimit     = 34600100
	)
	return (general.IsAndroidPick() && general.GetBuild() >= _androidLimit) || (general.IsIPhonePick() && general.GetBuild() >= _iosLimit) ||
		(general.IsPad() && general.GetBuild() >= _iosLimit) || (general.IsPadHD() && general.GetBuild() >= _iosHDLimit) ||
		(general.IsAndroidHD() && general.GetBuild() >= _androidHDLimit)
}

// 校园 热议话题卡
func (s *Service) isCampusHotTopicCapable(_ context.Context, general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit = 6840000
		_iosLimit     = 68400000
	)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit)
}

// 校园 校园小黄条
func (s *Service) isCampusYellowBarCapable(_ context.Context, general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit = 6900000
		_iosLimit     = 69000000
	)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit)
}

// 是否支持c端动态编辑
func (s *Service) isDynEditCapable(_ context.Context, general *mdlv2.GeneralParam) bool {
	//const (
	//	_androidLimit = 6860000
	//	_iosLimit     = 68600000
	//)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynViewEditAndroid, s.c.BuildLimit.DynViewEditIOS)
}

// 是否显示动态附加卡上面的 HeadText，6.20 之后版本（true）可以不下发 HeadText，老版本（false）会展示一块空白
func (s *Service) isEmptyHeadTextCapable(general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit = 6200000
		_iosLimit     = 62000200
	)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit)
}

// 对于支持校园热议卡的版本，不出banner
func (s *Service) isCampusNoBanner(ctx context.Context, general *mdlv2.GeneralParam) bool {
	return s.isCampusHotTopicCapable(ctx, general)
}

// 是否支持UGC合集分享卡
func (s *Service) isUGCSeasonShareCapble(_ context.Context, general *mdlv2.GeneralParam) bool {
	const (
		_androidLimit = 6910000
		_iosLimit     = 69100000
	)
	return general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit)
}
