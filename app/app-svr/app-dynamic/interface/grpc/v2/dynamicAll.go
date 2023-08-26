package v2

import (
	"context"

	"go-common/component/metadata/network"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	"go-common/library/ecode"
	xmetadata "go-common/library/net/metadata"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/app/app-svr/archive/middleware"
)

// 动态综合筛选器
func (s *Server) FeedFilter(ctx context.Context, req *api.FeedFilterReq) (*api.FeedFilterReply, error) {
	if len(req.Tab) <= 0 {
		return nil, ecode.RequestErr
	}
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	general.LocalTime = req.LocalTime
	if req.LocalTime < -12 || req.LocalTime > 14 {
		general.LocalTime = 8
	}
	return s.dynSvr.FeedFilter(s.buildPlayerArgs(ctx, nil, req.PlayerArgs), general, req)
}

// 动态综合
func (s *Server) DynAll(ctx context.Context, req *api.DynAllReq) (*api.DynAllReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		LocalTime:   req.LocalTime,
	}
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	if req.RefreshType == api.Refresh_refresh_history && req.Offset == "" {
		return nil, ecode.RequestErr
	}
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, req.PlayurlParam, req.PlayerArgs)
	return s.dynSvr.DynAll(c, general, req)
}

// 动态假卡片
func (s *Server) DynFakeCard(c context.Context, req *api.DynFakeCardReq) (*api.DynFakeCardReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取限制条件
	limit, _ := restriction.FromContext(c)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(c, xmetadata.RemoteIP),
	}
	if req.Content == "" {
		return nil, ecode.RequestErr
	}
	return s.dynSvr.DynFakeCard(c, general, req)
}

// 0关注推荐用户换一换
func (s *Server) DynRcmdUpExchange(c context.Context, req *api.DynRcmdUpExchangeReq) (*api.DynRcmdUpExchangeReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取限制条件
	limit, _ := restriction.FromContext(c)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(c, xmetadata.RemoteIP),
	}
	return s.dynSvr.DynRcmdUpExchange(c, general, req)
}

func (s *Server) DynAllPersonal(ctx context.Context, req *api.DynAllPersonalReq) (*api.DynAllPersonalReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		LocalTime:   req.LocalTime,
	}
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, req.PlayurlParam, req.PlayerArgs)
	return s.dynSvr.DynAllPersonal(c, general, req)
}

func (s *Server) DynAllUpdOffset(ctx context.Context, req *api.DynAllUpdOffsetReq) (*api.NoReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	limit, _ := restriction.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Mid:         au.Mid,
		Device:      &dev,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
	}
	return s.dynSvr.DynAllUpdOffset(ctx, general, req)
}

func (s *Server) DynServerDetails(ctx context.Context, req *api.DynServerDetailsReq) (*api.DynServerDetailsReply, error) {
	general := &mdlv2.GeneralParam{
		Mid: req.Mid,
		Device: &device.Device{
			Build:       req.Build,
			RawPlatform: req.Platform,
			RawMobiApp:  req.MobiApp,
			Buvid:       req.Buvid,
			Device:      req.Device,
		},
		IP: xmetadata.String(ctx, xmetadata.RemoteIP),
	}
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	return s.dynSvr.DynServerDetails(c, general, req)
}

func (s *Server) DynSpaceSearchDetails(ctx context.Context, req *api.DynSpaceSearchDetailsReq) (*api.DynSpaceSearchDetailsReply, error) {
	general := &mdlv2.GeneralParam{
		Mid: req.Mid,
		Device: &device.Device{
			Build:       req.Build,
			RawPlatform: req.Platform,
			RawMobiApp:  req.MobiApp,
			Buvid:       req.Buvid,
			Device:      req.Device,
		},
		IP:        req.Ip,
		LocalTime: req.LocalTime,
	}
	net := network.Network{
		Type:     network.NetworkType(req.NetType),
		TF:       network.TFType(req.TfType),
		RemoteIP: req.Ip,
	}
	// 秒开信息处理
	batchArg := middleware.MossBatchPlayArgs(req.PlayerArgs, *general.Device, net, req.Mid)
	c := middleware.NewContext(ctx, batchArg)
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	return s.dynSvr.DynSpaceSearchDetails(c, general, req)
}

func (s *Server) UnfollowMatch(ctx context.Context, req *api.UnfollowMatchReq) (*api.NoReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
	}
	if err := s.dynSvr.UnfollowMatch(ctx, general, req); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}
