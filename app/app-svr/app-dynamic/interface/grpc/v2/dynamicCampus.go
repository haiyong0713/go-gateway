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

// 动态综合
func (s *Server) CampusRcmd(c context.Context, req *api.CampusRcmdReq) (*api.CampusRcmdReply, error) {
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
	return s.dynSvr.CampusRcmd(s.buildPlayerArgs(c, nil, req.PlayerArgs), general, req)
}
