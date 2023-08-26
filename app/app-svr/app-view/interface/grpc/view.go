package grpc

import (
	"context"
	"strconv"
	"time"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/fawkes"
	mfawkes "go-common/component/fawkes/middleware/grpc"
	mlocale "go-common/component/locale/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	mrestrict "go-common/component/restriction/middleware/grpc"
	abtest "go-common/component/tinker/middleware/grpc"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	gwecode "go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	api "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/http"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/relates"
	viewmdl "go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/app/app-svr/app-view/interface/service/view"
	"go-gateway/app/app-svr/app-view/interface/tools"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	svr    *view.Service
	config *conf.Config
}

func New(c *warden.ServerConfig, hsvr *http.Server, svrConf *conf.Config) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(c)
	s := &Server{
		svr:    hsvr.ViewSvr,
		config: svrConf,
	}
	api.RegisterViewServer(wsvr.Server(), s)
	// 用户鉴权
	author := mauth.New(nil)
	fks := mfawkes.NewMiddleware(fawkes.Default())
	wsvr.Add("/bilibili.app.view.v1.View/View", author.UnaryServerInterceptor(true),
		fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(),
		anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC()) // guest: true/false
	wsvr.Add("/bilibili.app.view.v1.View/ViewTag", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/ViewMaterial", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/ViewProgress", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/ClickPlayerCard", author.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/ClickPlayerCardV2", author.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/ClickActivitySeason", author.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/Season", author.UnaryServerInterceptor(true), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/ExposePlayerCard", author.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/AddContract", author.UnaryServerInterceptor(false), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/FeedView", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/ChronosPkg", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/CacheView", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/RelatesFeed", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/ContinuousPlay", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/PremiereArchive", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/Reserve", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/GetArcsPlayer", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), anticrawler.ReportInterceptor(), hsvr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.view.v1.View/PlayerRelates", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/SeasonActivityRecord", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr.Add("/bilibili.app.view.v1.View/SeasonWidgetExpose", author.UnaryServerInterceptor(true), fks.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor(), mlocale.UnaryServerInterceptor(), anticrawler.ReportInterceptor())
	wsvr, err = wsvr.Start()
	return
}

func (s *Server) PlayerRelates(c context.Context, req *api.PlayerRelatesReq) (*api.PlayerRelatesReply, error) {
	if req.Aid == 0 {
		var err error
		if req.Aid, err = bvid.BvToAv(req.Bvid); err != nil || req.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}

	// 查看是否支持mid64
	c = context.WithValue(c, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(c, s.svr.VersionMapClient.AppkeyVersion))
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取网络信息
	nw, _ := network.FromContext(c)
	// 获取客户端模式 如课堂模式
	restrict, _ := restriction.FromContext(c)

	res := &api.PlayerRelatesReply{}
	//青少年模式&课堂模式&审核版本，不返回相关推荐
	if restrict.IsTeenagers || restrict.IsLessons || restrict.IsReview {
		log.Error("PlayerRelates restrict req(%+v), restrict.IsTeenagers(%t), restrict.IsLessons(%t), restrict.IsReview(%t)",
			req, restrict.IsTeenagers, restrict.IsLessons, restrict.IsReview)
		return res, nil
	}

	var disableRcmdMode int
	if restrict.DisableRcmd {
		disableRcmdMode = 1
	}

	//秒开参数设置
	batchArg := arcmid.MossBatchPlayArgs(req.PlayerArgs, dev, nw, au.Mid)
	c = arcmid.NewContext(c, batchArg)

	relatesFeedReq := &relates.RelatesFeedGRPCRequest{
		Aid:         req.Aid,
		Mid:         au.Mid,
		Build:       dev.Build,
		Buvid:       dev.Buvid,
		Spmid:       req.Spmid,
		FromSpmid:   req.FromSpmid,
		TrackId:     req.FromTrackId,
		Plat:        model.PlatNew(dev.RawMobiApp, dev.Device),
		MobileApp:   dev.RawMobiApp,
		Network:     dev.Network,
		Device:      dev.Device,
		DisableRcmd: disableRcmdMode,
		SessionId:   req.SessionId,
		Ip:          nw.RemoteIP,
		From:        req.From,
	}

	reply, err := s.svr.PlayerRelatesGRPC(c, relatesFeedReq)
	if err != nil {
		log.Error("PlayerRelates s.svr.RelatesFeedGRPC error req(%+v), err(%+v)", req, err)
		return nil, err
	}

	return reply, nil
}

func (s *Server) RelatesFeed(c context.Context, req *api.RelatesFeedReq) (*api.RelatesFeedReply, error) {
	if req.Aid == 0 {
		var err error
		if req.Aid, err = bvid.BvToAv(req.Bvid); err != nil || req.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}

	// 查看是否支持mid64
	c = context.WithValue(c, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(c, s.svr.VersionMapClient.AppkeyVersion))
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取网络信息
	nw, _ := network.FromContext(c)
	// 获取客户端模式 如课堂模式
	restrict, _ := restriction.FromContext(c)

	res := &api.RelatesFeedReply{}
	//青少年模式&课堂模式&审核版本，不返回相关推荐
	if restrict.IsTeenagers || restrict.IsLessons || restrict.IsReview {
		log.Error("RelatesFeed restrict req(%+v), restrict.IsTeenagers(%t), restrict.IsLessons(%t), restrict.IsReview(%t)",
			req, restrict.IsTeenagers, restrict.IsLessons, restrict.IsReview)
		return res, nil
	}

	var disableRcmdMode int
	if restrict.DisableRcmd {
		disableRcmdMode = 1
	}

	//秒开参数设置
	batchArg := arcmid.MossBatchPlayArgs(req.PlayerArgs, dev, nw, au.Mid)
	c = arcmid.NewContext(c, batchArg)

	relatesFeedReq := &relates.RelatesFeedGRPCRequest{
		Aid:         req.Aid,
		Mid:         au.Mid,
		Build:       dev.Build,
		Buvid:       dev.Buvid,
		Spmid:       req.Spmid,
		FromSpmid:   req.FromSpmid,
		TrackId:     req.FromTrackId,
		Plat:        model.PlatNew(dev.RawMobiApp, dev.Device),
		MobileApp:   dev.RawMobiApp,
		Network:     dev.Network,
		Device:      dev.Device,
		DisableRcmd: disableRcmdMode,
		PageIndex:   req.RelatesPage,
		SessionId:   req.SessionId,
		Ip:          nw.RemoteIP,
		From:        req.From,
		Pagination:  req.Pagination,
		RefreshNum:  req.RefreshNum,
	}

	reply, err := s.svr.RelatesFeedGRPC(c, relatesFeedReq)
	if err != nil {
		log.Error("RelatesFeed s.svr.RelatesFeedGRPC error req(%+v), err(%+v)", req, err)
		return nil, err
	}

	return reply, nil
}

func (s *Server) ViewMaterial(c context.Context, arg *api.ViewMaterialReq) (*api.ViewMaterialReply, error) {
	if arg.Aid == 0 {
		var err error
		if arg.Aid, err = bvid.BvToAv(arg.Bvid); err != nil || arg.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}
	// 查看是否支持mid64
	c = context.WithValue(c, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(c, s.svr.VersionMapClient.AppkeyVersion))
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	return s.svr.ViewMaterial(c, arg, au.Mid, int32(dev.Build), dev.Device, dev.RawMobiApp, dev.RawPlatform, dev.Buvid)
}

func (s *Server) ViewTag(c context.Context, arg *api.ViewTagReq) (*api.ViewTagReply, error) {
	if arg.Aid == 0 {
		var err error
		if arg.Aid, err = bvid.BvToAv(arg.Bvid); err != nil || arg.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}
	// 查看是否支持mid64
	c = context.WithValue(c, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(c, s.svr.VersionMapClient.AppkeyVersion))
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	plat := model.PlatNew(dev.RawMobiApp, dev.Device)
	if arg.Spmid == "" {
		arg.Spmid = "main.ugc-video-detail.0.0"
	}
	return s.svr.ViewTag(c, arg, au.Mid, plat, int(dev.Build), dev.Device, "v1", dev.Buvid, dev.RawMobiApp, dev.RawPlatform)
}

func (s *Server) View(c context.Context, arg *api.ViewReq) (*api.ViewReply, error) {
	if arg.Aid == 0 {
		var err error
		if arg.Aid, err = bvid.BvToAv(arg.Bvid); err != nil || arg.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}
	if arg.PageVersion == "" {
		arg.PageVersion = model.PageVersionV1
	}
	// 查看是否支持mid64
	c = context.WithValue(c, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(c, s.svr.VersionMapClient.AppkeyVersion))

	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取客户端模式 如课堂模式
	restrict, _ := restriction.FromContext(c)
	var (
		teenagersMode, lessonsMode, disableRcmdMode int
		net, filterd                                string
	)
	if restrict.IsTeenagers {
		teenagersMode = 1
	}
	if restrict.IsLessons {
		lessonsMode = 1
	}
	if restrict.DisableRcmd {
		disableRcmdMode = 1
	}
	if restrict.IsReview {
		filterd = "1"
	}
	nw, _ := network.FromContext(c)
	if nw.Type == network.TypeCellular {
		net = "mobile"
	} else if nw.Type == network.TypeWIFI {
		net = "wifi"
	}
	var isMelloi string
	if gmd, ok := metadata.FromIncomingContext(c); ok {
		if values := gmd.Get("x-melloi"); len(values) > 0 {
			isMelloi = values[0]
		}
	}
	now := time.Now()
	plat := model.PlatNew(dev.RawMobiApp, dev.Device)
	slocale := ""
	clocale := ""
	if locale, ok := locale.FromContext(c); ok {
		slocale, clocale = i18n.ConstructLocaleID(locale)
	}
	//高版本秒开参数
	var batchArg *arcgrpc.BatchPlayArg
	if arg.PlayerArgs != nil && s.config.Custom.PlayerArgs {
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, nw, au.Mid)
	} else {
		batchArg = &arcgrpc.BatchPlayArg{
			Ip:        nw.RemoteIP,
			Build:     dev.Build,
			Device:    dev.Device,
			NetType:   arcgrpc.NetworkType(nw.Type),
			Qn:        int64(arg.Qn),
			MobiApp:   dev.RawMobiApp,
			Fnver:     int64(arg.Fnver),
			Fnval:     int64(arg.Fnval),
			ForceHost: int64(arg.ForceHost),
			Buvid:     dev.Buvid,
			Mid:       au.Mid,
			Fourk:     int64(arg.Fourk),
			TfType:    arcgrpc.TFType(nw.TF),
		}
	}
	c = arcmid.NewContext(c, batchArg)

	res, err := s.svr.ViewGRPC(c, arg, au.Mid, plat, teenagersMode, lessonsMode, int(dev.Build), dev.RawMobiApp,
		dev.Buvid, dev.Device, net, dev.RawPlatform, nw.RemoteIP, nw.WebcdnIP, filterd, isMelloi, dev.Brand,
		slocale, clocale, now, disableRcmdMode)

	// view曝光上报 首次调用view接口才上报
	if arg.PlayMode != "background" && arg.Refresh == 0 {
		s.svr.ViewInfoc(au.Mid, int(plat), arg.Trackid, strconv.FormatInt(arg.Aid, 10), nw.RemoteIP, model.PathView, strconv.FormatInt(dev.Build, 10), dev.Buvid, arg.From, now, err, int(arg.Autoplay), arg.Spmid, arg.FromSpmid, net, isMelloi)
	}
	// 404错误页面有运营配置的特殊处理
	if err != nil {
		if ecode.EqualError(gwecode.AppViewForRetry, err) {
			return nil, status.Error(codes.DataLoss, "DegradeForView")
		}
		if ecode.EqualError(ecode.NothingFound, err) && s.svr.HasCustomConfig(c, arg.Aid) {
			return &api.ViewReply{Ecode: api.ECode_CODE404, CustomConfig: &api.CustomConfig{RedirectUrl: s.svr.NothingFoundUrl(arg.Aid)}}, nil
		}
		return nil, err
	}
	return res, nil
}

func (s *Server) ViewProgress(c context.Context, arg *api.ViewProgressReq) (*api.ViewProgressReply, error) {
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	rly, err := s.svr.ViewProgress(c, arg, au.Mid, dev)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &api.ViewProgressReply{}, nil
	}
	return rly, nil
}

func (s *Server) ClickPlayerCard(c context.Context, arg *api.ClickPlayerCardReq) (*api.NoReply, error) {
	dev, _ := device.FromContext(c)
	au, _ := auth.FromContext(c)
	if err := s.svr.ClickPlayerCard(c, arg, au.Mid, dev); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}

func (s *Server) ClickPlayerCardV2(ctx context.Context, arg *api.ClickPlayerCardReq) (*api.ClickPlayerCardReply, error) {
	dev, _ := device.FromContext(ctx)
	au, _ := auth.FromContext(ctx)
	res, err := s.svr.ClickPlayerCardV2(ctx, arg, au.Mid, dev)
	if err != nil {
		return nil, err
	}
	return &api.ClickPlayerCardReply{
		Message: res,
	}, nil
}

func (s *Server) ShortFormVideoDownload(ctx context.Context, arg *api.ShortFormVideoDownloadReq) (*api.ShortFormVideoDownloadReply, error) {
	_, tfType := model.TrafficFree(arg.TfIsp)
	videoDownloadReq := &viewmdl.VideoDownloadReq{
		ShortFormVideoDownloadReq: arg,
		TfType:                    tfType,
	}
	reply, err := s.svr.VideoDownload(ctx, videoDownloadReq)
	if err != nil {
		return nil, err
	}
	return reply.ShortFormVideoDownloadReply, nil
}

func (s *Server) ClickActivitySeason(c context.Context, arg *api.ClickActivitySeasonReq) (*api.NoReply, error) {
	dev, _ := device.FromContext(c)
	au, _ := auth.FromContext(c)
	if err := s.svr.ClickActivitySeason(c, arg, au.Mid, dev); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}

func (s *Server) Season(c context.Context, arg *api.SeasonReq) (*api.SeasonReply, error) {
	return s.svr.Season(c, arg)
}

func (s *Server) ExposePlayerCard(c context.Context, arg *api.ExposePlayerCardReq) (*api.NoReply, error) {
	dev, _ := device.FromContext(c)
	au, _ := auth.FromContext(c)
	if err := s.svr.ExposePlayerCard(c, arg, au.Mid, dev); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}

func (s *Server) AddContract(c context.Context, arg *api.AddContractReq) (*api.NoReply, error) {
	dev, _ := device.FromContext(c)
	au, _ := auth.FromContext(c)
	if err := s.svr.AddContract(c, arg, au.Mid, dev); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}

func (s *Server) FeedView(ctx context.Context, arg *api.FeedViewReq) (*api.FeedViewReply, error) {
	if arg.Aid == 0 {
		_aid, err := bvid.BvToAv(arg.Bvid)
		if err != nil || _aid == 0 {
			return nil, ecode.Errorf(ecode.RequestErr, "%s", err)
		}
		arg.Aid = _aid
	}
	authN, _ := auth.FromContext(ctx)
	device, _ := device.FromContext(ctx)
	network, _ := network.FromContext(ctx)
	batchArg := arcmid.MossBatchPlayArgs(arg.PlayerArgs, device, network, authN.Mid)
	ctx = arcmid.NewContext(ctx, batchArg)
	return s.svr.FeedView(ctx, arg)
}

func (s *Server) CacheView(c context.Context, arg *api.CacheViewReq) (*api.CacheViewReply, error) {
	if arg.Aid == 0 {
		var err error
		if arg.Aid, err = bvid.BvToAv(arg.Bvid); err != nil || arg.Aid == 0 {
			return nil, ecode.RequestErr
		}
	}
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取客户端模式
	restrict, _ := restriction.FromContext(c)
	var (
		teenagersMode, lessonsMode, disableRcmdMode int
		net, filterd                                string
	)
	if restrict.IsTeenagers {
		teenagersMode = 1
	}
	if restrict.IsLessons {
		lessonsMode = 1
	}
	if restrict.DisableRcmd {
		disableRcmdMode = 1
	}
	if restrict.IsReview {
		filterd = "1"
	}
	nw, _ := network.FromContext(c)
	if nw.Type == network.TypeCellular {
		net = "mobile"
	} else if nw.Type == network.TypeWIFI {
		net = "wifi"
	}
	//获取压测标识
	var isMelloi string
	if gmd, ok := metadata.FromIncomingContext(c); ok {
		if values := gmd.Get("x-melloi"); len(values) > 0 {
			isMelloi = values[0]
		}
	}
	now := time.Now()
	plat := model.PlatNew(dev.RawMobiApp, dev.Device)
	slocale := ""
	clocale := ""
	if locale, ok := locale.FromContext(c); ok {
		slocale, clocale = i18n.ConstructLocaleID(locale)
	}

	res, err := s.svr.CacheViewGRPC(c, arg, au.Mid, plat, teenagersMode, lessonsMode, int(dev.Build), dev.RawMobiApp,
		dev.Buvid, dev.Device, net, dev.RawPlatform, nw.WebcdnIP, filterd, dev.Brand, slocale,
		clocale, disableRcmdMode)

	//cacheView曝光上报
	s.svr.ViewInfoc(au.Mid, int(plat), "", strconv.FormatInt(arg.Aid, 10), nw.RemoteIP, model.PathCacheView,
		strconv.FormatInt(dev.Build, 10), dev.Buvid, arg.From, now, err, 0, arg.Spmid, arg.FromSpmid,
		net, isMelloi)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) GetArcsPlayer(c context.Context, arg *api.GetArcsPlayerReq) (*api.GetArcsPlayerReply, error) {
	au, _ := auth.FromContext(c)
	dev, _ := device.FromContext(c)
	nw, _ := network.FromContext(c)
	//高版本秒开参数
	var batchArg *arcgrpc.BatchPlayArg
	if arg.PlayerArgs != nil {
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, nw, au.Mid)
	}
	c = arcmid.NewContext(c, batchArg)
	res, err := s.svr.GetArcsPlayerGRPC(c, arg, dev)
	if err != nil {
		log.Warn("s.svr.GetArcsPlayerGRPC() err : %+v ", err)
		return nil, err
	}
	return res, nil
}

func (s *Server) ContinuousPlay(c context.Context, arg *api.ContinuousPlayReq) (*api.ContinuousPlayReply, error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	nw, _ := network.FromContext(c)
	net := ""
	if nw.Type == network.TypeCellular {
		net = "mobile"
	} else if nw.Type == network.TypeWIFI {
		net = "wifi"
	}
	var isMelloi string
	if gmd, ok := metadata.FromIncomingContext(c); ok {
		if values := gmd.Get("x-melloi"); len(values) > 0 {
			isMelloi = values[0]
		}
	}
	plat := model.PlatNew(dev.RawMobiApp, dev.Device)
	//高版本秒开参数
	var batchArg *arcgrpc.BatchPlayArg
	if arg.PlayerArgs != nil && s.config.Custom.PlayerArgs {
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, nw, au.Mid)
	}
	c = arcmid.NewContext(c, batchArg)
	res, err := s.svr.ContinuousPlayGRPC(c, arg, au.Mid, plat, dev, net, isMelloi)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) PremiereArchive(ctx context.Context, arg *api.PremiereArchiveReq) (*api.PremiereArchiveReply, error) {
	// 查看是否支持mid64
	ctx = context.WithValue(ctx, tools.SupportMid64App{}, tools.CheckMid64SupportGRPC(ctx, s.svr.VersionMapClient.AppkeyVersion))
	res, err := s.svr.GetPremiereGRPC(ctx, arg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) Reserve(ctx context.Context, arg *api.ReserveReq) (*api.ReserveReply, error) {
	au, _ := auth.FromContext(ctx)
	res, err := s.svr.ReserveRPC(ctx, arg, au.Mid)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "网络错误")
	}
	return res, nil
}

func (s *Server) SeasonActivityRecord(ctx context.Context, arg *api.SeasonActivityRecordReq) (*api.SeasonActivityRecordReply, error) {
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	// 获取网络信息
	nw, _ := network.FromContext(ctx)
	res, err := s.svr.SeasonActivityRecordRPC(ctx, arg, dev, au.Mid, nw)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) SeasonWidgetExpose(ctx context.Context, arg *api.SeasonWidgetExposeReq) (*api.SeasonWidgetExposeReply, error) {
	res, err := s.svr.SeasonWidgetExpose(ctx, arg)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) ChronosPkg(ctx context.Context, arg *api.ChronosPkgReq) (*api.Chronos, error) {
	return s.svr.ChronosPkg(ctx, &viewmdl.ChronosPkgReq{
		ServiceKey:    arg.ServiceKey,
		EngineVersion: arg.EngineVersion,
	})
}
