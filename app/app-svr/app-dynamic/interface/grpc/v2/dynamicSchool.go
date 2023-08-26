package v2

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	xmetadata "go-common/library/net/metadata"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

func (s *Server) CampusEntryTab(ctx context.Context, req *api.CampusEntryTabReq) (*api.CampusEntryTabResp, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.CampusEntryTab(ctx, general, req)
}

func (s *Server) AlumniDynamics(ctx context.Context, req *api.AlumniDynamicsReq) (*api.AlumniDynamicsReply, error) {
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
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.DynAlumniDynamics(c, general, req)
}

func (s *Server) SchoolRecommend(ctx context.Context, req *api.SchoolRecommendReq) (*api.SchoolRecommendReply, error) {
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
	return s.dynSvr.SchoolRecommend(ctx, general, req)
}

func (s *Server) SchoolSearch(ctx context.Context, req *api.SchoolSearchReq) (*api.SchoolSearchReply, error) {
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
	return s.dynSvr.SchoolSearch(ctx, general, req)
}

func (s *Server) OfficialAccounts(ctx context.Context, req *api.OfficialAccountsReq) (*api.OfficialAccountsReply, error) {
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
	return s.dynSvr.OfficialAccounts(ctx, general, req)
}

func (s *Server) OfficialDynamics(ctx context.Context, req *api.OfficialDynamicsReq) (*api.OfficialDynamicsReply, error) {
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
	return s.dynSvr.OfficialDynamics(ctx, general, req)
}

func (s *Server) CampusRedDot(ctx context.Context, req *api.CampusRedDotReq) (*api.CampusRedDotReply, error) {
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
	return s.dynSvr.CampusRedDot(ctx, general, req)
}

func (s *Server) CampusRcmdFeed(ctx context.Context, req *api.CampusRcmdFeedReq) (*api.CampusRcmdFeedReply, error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	nw, _ := network.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Network:     &nw,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		LocalTime:   req.LocalTime,
	}
	if general.LocalTime < -12 || general.LocalTime > 14 {
		general.LocalTime = 8
	}
	// 秒开信息处理
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.CampusRcmdFeed(c, general, req)
}

func (s *Server) TopicSquare(ctx context.Context, req *api.TopicSquareReq) (*api.TopicSquareReply, error) {
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
	return s.dynSvr.TopicSquare(ctx, general, req)
}

func (s *Server) TopicList(ctx context.Context, req *api.TopicListReq) (*api.TopicListReply, error) {
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
	return s.dynSvr.TopicList(ctx, general, req)
}

func (s *Server) CampusMateLikeList(ctx context.Context, req *api.CampusMateLikeListReq) (*api.CampusMateLikeListReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.CampusMateLikeList(ctx, general, req)
}

func (s *Server) CampusFeedback(ctx context.Context, req *api.CampusFeedbackReq) (*api.CampusFeedbackReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.CampusFeedback(ctx, general, req)
}

func (s *Server) CampusBillboard(ctx context.Context, req *api.CampusBillBoardReq) (*api.CampusBillBoardReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.CampusBillboard(c, general, req)
}

func (s *Server) CampusBillboardInternal(ctx context.Context, req *api.CampusBillboardInternalReq) (*api.CampusBillBoardReply, error) {
	general := &mdlv2.GeneralParam{Mid: req.Mid, Device: &device.Device{}}
	return s.dynSvr.CampusBillboard(ctx, general, &api.CampusBillBoardReq{CampusId: req.CampusId, VersionCode: req.VersionCode})
}

func (s *Server) CampusTopicRcmdFeed(ctx context.Context, req *api.CampusTopicRcmdFeedReq) (*api.CampusTopicRcmdFeedReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	general.SetLocalTime(req.LocalTime)
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.CampusTopicRcmdFeed(c, general, req)
}

func (s *Server) FetchTabSetting(ctx context.Context, _ *api.NoReq) (*api.FetchTabSettingReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.FetchTabSetting(ctx, general)
}

func (s *Server) UpdateTabSetting(ctx context.Context, req *api.UpdateTabSettingReq) (*api.NoReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.UpdateTabSetting(ctx, general, req)
}

func (s *Server) CampusSquare(ctx context.Context, req *api.CampusSquareReq) (*api.CampusSquareReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.CampusSquare(ctx, general, req)
}

func (s *Server) CampusRecommend(ctx context.Context, req *api.CampusRecommendReq) (*api.CampusRecommendReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.CampusRecommend(c, general, req)
}

func (s *Server) CampusHomePages(ctx context.Context, req *api.CampusHomePagesReq) (*api.CampusHomePagesReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.HomePages(c, general, req)
}

func (s *Server) HomeSubscribe(ctx context.Context, req *api.HomeSubscribeReq) (*api.HomeSubscribeReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	return s.dynSvr.HomeSubscribe(ctx, general, req)
}

func (s *Server) CampusMngDetail(ctx context.Context, req *api.CampusMngDetailReq) (*api.CampusMngDetailReply, error) {
	return s.dynSvr.CampusMngDetail(ctx, mdlv2.NewGeneralParamFromCtx(ctx), req)
}

func (s *Server) CampusMngSubmit(ctx context.Context, req *api.CampusMngSubmitReq) (*api.CampusMngSubmitReply, error) {
	return s.dynSvr.CampusMngSubmit(ctx, mdlv2.NewGeneralParamFromCtx(ctx), req)
}

func (s *Server) CampusMngQuizOperate(ctx context.Context, req *api.CampusMngQuizOperateReq) (*api.CampusMngQuizOperateReply, error) {
	return s.dynSvr.CampusMngQuizOperate(ctx, mdlv2.NewGeneralParamFromCtx(ctx), req)
}
