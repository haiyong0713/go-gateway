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
)

func (s *Server) DynVideo(ctx context.Context, req *api.DynVideoReq) (*api.DynVideoReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	// 获取客户端网络信息
	nw, _ := network.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		LocalTime:   req.LocalTime,
		Network:     &nw,
	}
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	if req.RefreshType == api.Refresh_refresh_history && req.Offset == "" {
		return nil, ecode.RequestErr
	}
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, req.PlayurlParam, req.PlayerArgs)
	return s.dynSvr.DynVideo(c, general, req)
}

func (s *Server) DynVideoPersonal(ctx context.Context, req *api.DynVideoPersonalReq) (*api.DynVideoPersonalReply, error) {
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
	return s.dynSvr.DynVideoPersonal(c, general, req)
}

func (s *Server) DynVideoUpdOffset(ctx context.Context, req *api.DynVideoUpdOffsetReq) (*api.NoReply, error) {
	au, _ := auth.FromContext(ctx)
	gen := &mdlv2.GeneralParam{
		Mid: au.Mid,
	}
	return s.dynSvr.DynVideoUpdateOffset(ctx, gen, req)
}
