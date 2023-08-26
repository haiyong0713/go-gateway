package v2

import (
	"context"
	"strconv"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	mrestrict "go-common/component/restriction/middleware/grpc"
	abtest "go-common/component/tinker/middleware/grpc"
	"go-common/library/ecode"
	xmetadata "go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/http"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	dynsvrv2 "go-gateway/app/app-svr/app-dynamic/interface/service/dynamicV2"
	"go-gateway/app/app-svr/archive/middleware"
	middlewarev1 "go-gateway/app/app-svr/archive/middleware/v1"
	arcApi "go-gateway/app/app-svr/archive/service/api"
)

type Server struct {
	dynSvr *dynsvrv2.Service
	mauth  *mauth.Auth
	config *conf.Config
}

func New(wsvr *warden.Server, mauth *mauth.Auth, svr *http.Server) (*Server, error) {
	s := &Server{
		dynSvr: svr.DynamicSvcV2,
		config: svr.Config,
		mauth:  mauth,
	}
	api.RegisterDynamicServer(wsvr.Server(), s)
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynVideo", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynDetails", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynVideoPersonal", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynVideoUpdOffset", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynAll", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynFakeCard", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynRcmdUpExchange", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynAllPersonal", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynAllUpdOffset", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynVote", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynMixUpListViewMore", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynDetail", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/LikeList", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/RepostList", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynSpace", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynLight", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynUnLoginRcmd", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/AlumniDynamics", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/SchoolRecommend", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/SchoolSearch", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusRcmd", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/SetDecision", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/SubscribeCampus", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/SetRecentCampus", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynTab", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/DynSearch", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/OfficialAccounts", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/OfficialDynamics", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusRedDot", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusRcmdFeed", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/TopicSquare", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/TopicList", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusMateLikeList", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusFeedback", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusBillboard", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusBillboardInternal", svr.Config.AppAuth.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusTopicRcmdFeed", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/UnfollowMatch", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/FetchTabSetting", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/UpdateTabSetting", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusSquare", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusRecommend", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusHomePages", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/HomeSubscribe", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusEntryTab", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/LbsPoi", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), svr.FeatureSvc.BuildLimitGRPC(), mrestrict.UnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/FeedFilter", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusMngDetail", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusMngSubmit", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/CampusMngQuizOperate", s.mauth.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.dynamic.v2.Dynamic/LegacyTopicFeed", s.mauth.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), mrestrict.UnaryServerInterceptor(), s.dynSvr.FeatureGateUnaryServerInterceptor())

	return s, nil
}

func (s *Server) DynDetails(ctx context.Context, req *api.DynDetailsReq) (*api.DynDetailsReply, error) {
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
	return s.dynSvr.DynDetails(c, general, req)
}

// 暂时不做，客户端调用老接口
func (s *Server) DynAdditionCommonFollow(ctx context.Context, req *api.DynAdditionCommonFollowReq) (*api.DynAdditionCommonFollowReply, error) {
	dynamicID, _ := strconv.ParseInt(req.DynId, 10, 64)
	if dynamicID == 0 {
		return nil, ecode.RequestErr
	}
	if err := s.dynSvr.AdditionFollow(ctx, req.CardType, dynamicID, req.Status.String()); err != nil {
		return nil, err
	}
	return &api.DynAdditionCommonFollowReply{}, nil
}

// 暂时不做，客户端调用老接口
func (s *Server) DynThumb(_ context.Context, _ *api.DynThumbReq) (*api.NoReply, error) {
	return &api.NoReply{}, nil
}

func (s *Server) DynVote(ctx context.Context, req *api.DynVoteReq) (*api.DynVoteReply, error) {
	if req.VoteId == 0 {
		return nil, ecode.RequestErr
	}
	au, _ := auth.FromContext(ctx)
	gen := &mdlv2.GeneralParam{
		Mid: au.Mid,
	}
	return s.dynSvr.DynVote(ctx, gen, req)
}

func (s *Server) DynMixUpListViewMore(c context.Context, req *api.DynMixUpListViewMoreReq) (res *api.DynMixUpListViewMoreReply, err error) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	limit, _ := restriction.FromContext(c)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(c, xmetadata.RemoteIP),
	}
	return s.dynSvr.DynMixUpListViewMore(c, req, general)
}

func (s *Server) buildPlayerArgs(ctx context.Context, params *api.PlayurlParam, playArg *middlewarev1.PlayerArgs) context.Context {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取网络信息
	net, _ := network.FromContext(ctx)
	//高版本秒开参数
	var batchArg *arcApi.BatchPlayArg
	if playArg != nil {
		batchArg = middleware.MossBatchPlayArgs(playArg, dev, net, au.Mid)
	} else if params != nil {
		batchArg = &arcApi.BatchPlayArg{
			Ip:        net.RemoteIP,
			Build:     dev.Build,
			Device:    dev.Device,
			NetType:   arcApi.NetworkType(net.Type),
			Qn:        int64(params.Qn),
			MobiApp:   dev.RawMobiApp,
			Fnver:     int64(params.Fnver),
			Fnval:     int64(params.Fnval),
			ForceHost: int64(params.ForceHost),
			Buvid:     dev.Buvid,
			Mid:       au.Mid,
			Fourk:     int64(params.Fourk),
			TfType:    arcApi.TFType(net.TF),
		}
	}
	return middleware.NewContext(ctx, batchArg)
}

func (s *Server) DynUnLoginRcmd(c context.Context, req *api.DynRcmdReq) (*api.DynRcmdReply, error) {
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
	return s.dynSvr.DynUnLoginRcmd(ctx, general, req)
}

func (s *Server) DynTab(ctx context.Context, req *api.DynTabReq) (*api.DynTabReply, error) {
	if req.TeenagersMode != model.TeenagersClose && req.TeenagersMode != model.TeenagersOpen {
		return nil, ecode.RequestErr
	}
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	limit, _ := restriction.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
	}
	return s.dynSvr.DynTab(ctx, general, req)
}

func (s *Server) SetDecision(ctx context.Context, req *api.SetDecisionReq) (*api.NoReply, error) {
	au, _ := auth.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Mid: au.Mid,
	}
	return s.dynSvr.SetDecision(ctx, general, req)
}

func (s *Server) SubscribeCampus(ctx context.Context, req *api.SubscribeCampusReq) (*api.NoReply, error) {
	au, _ := auth.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Mid: au.Mid,
	}
	return s.dynSvr.SubscribeCampus(ctx, general, req)
}

func (s *Server) SetRecentCampus(ctx context.Context, req *api.SetRecentCampusReq) (*api.NoReply, error) {
	au, _ := auth.FromContext(ctx)
	general := &mdlv2.GeneralParam{
		Mid: au.Mid,
	}
	return s.dynSvr.SetRecentCampus(ctx, general, req)
}

func (s *Server) LbsPoi(ctx context.Context, req *api.LbsPoiReq) (*api.LbsPoiReply, error) {
	general := mdlv2.NewGeneralParamFromCtx(ctx)
	c := s.buildPlayerArgs(ctx, nil, req.PlayerArgs)
	return s.dynSvr.LbsPoiList(c, general, req)
}
