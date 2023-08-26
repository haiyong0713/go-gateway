package v2

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	xmetadata "go-common/library/net/metadata"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

func (s *Server) DynDetail(ctx context.Context, req *api.DynDetailReq) (*api.DynDetailReply, error) {
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
	// 广告
	general.Pattern = req.Pattern
	// 分享ID
	general.ShareID = req.ShareId
	general.ShareMode = req.ShareMode
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.DynDetail(c, general, req)
}

func (s *Server) LikeList(ctx context.Context, req *api.LikeListReq) (*api.LikeListReply, error) {
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
	return s.dynSvr.LikeList(ctx, general, req)
}

func (s *Server) RepostList(ctx context.Context, req *api.RepostListReq) (*api.RepostListRsp, error) {
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
	return s.dynSvr.RepostList(ctx, general, req)
}
