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

func (s *Server) DynLight(c context.Context, req *api.DynLightReq) (*api.DynLightReply, error) {
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
		LocalTime:   req.LocalTime,
	}
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	// 秒开信息处理
	ctx := s.buildPlayerArgs(c, nil, req.PlayerArgs)
	return s.dynSvr.DynLight(ctx, general, req)
}
