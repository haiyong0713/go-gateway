package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	mrestrict "go-common/component/restriction/middleware/grpc"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"go-gateway/app/app-svr/app-player/interface/model"
	"go-gateway/app/app-svr/app-player/interface/service"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	svr *service.Service
}

func New(c *warden.ServerConfig, svr *service.Service, _ *feature.Feature) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(c)
	s := &Server{
		svr: svr,
	}
	api.RegisterPlayURLServer(wsvr.Server(), s)
	// 用户鉴权
	author := mauth.New(nil)
	wsvr.Add("/bilibili.app.playurl.v1.PlayURL/PlayURL", author.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())                                  // guest: true/false
	wsvr.Add("/bilibili.app.playurl.v1.PlayURL/Project", author.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())                                  // guest: true/false
	wsvr.Add("/bilibili.app.playurl.v1.PlayURL/PlayView", author.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor(), SetDevContextInterceptor())     // guest: true/false
	wsvr.Add("/bilibili.app.playurl.v1.PlayURL/PlayConfEdit", author.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor(), SetDevContextInterceptor()) // guest: true/false
	wsvr.Add("/bilibili.app.playurl.v1.PlayURL/PlayConf", author.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor(), SetDevContextInterceptor())     // guest: true/false
	wsvr, err = wsvr.Start()
	return
}

func (s *Server) PlayURL(c context.Context, arg *api.PlayURLReq) (res *api.PlayURLReply, err error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	net, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	param := &model.Param{
		AID:       arg.Aid,
		CID:       arg.Cid,
		Qn:        arg.Qn,
		MobiApp:   dev.RawMobiApp,
		Fnval:     arg.Fnval,
		Fnver:     arg.Fnver,
		Build:     int32(dev.Build),
		Device:    dev.Device,
		ForceHost: arg.ForceHost,
		FourkBool: arg.Fourk,
		Buvid:     dev.Buvid,
		Platform:  dev.RawPlatform,
		Download:  arg.Download,
		NetType:   int32(net.Type),
		TfType:    int32(net.TF),
	}
	plat := model.Plat(param.MobiApp, param.Device)
	res, err = s.svr.PlayURLGRPC(c, au.Mid, param, plat)
	return
}

// Project is
func (s *Server) Project(c context.Context, arg *api.ProjectReq) (res *api.ProjectReply, err error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	res = new(api.ProjectReply)
	param := &model.Param{
		AID:     arg.Aid,
		CID:     arg.Cid,
		Qn:      arg.Qn,
		MobiApp: dev.RawMobiApp,
		// 兼容投屏老版本不支持dash，视频云以前不论fnval都不会返回dash格式
		Fnval:      0,
		Fnver:      arg.Fnver,
		Build:      int32(dev.Build),
		Device:     dev.Device,
		ForceHost:  arg.ForceHost,
		FourkBool:  arg.Fourk,
		Buvid:      dev.Buvid,
		Platform:   dev.RawPlatform,
		Download:   arg.Download,
		Protocol:   arg.Protocol,
		DeviceType: arg.DeviceType,
	}
	plat := model.Plat(param.MobiApp, param.Device)
	res.Project, err = s.svr.Project(c, au.Mid, param, plat)
	return
}

func (s *Server) PlayView(c context.Context, arg *api.PlayViewReq) (res *api.PlayViewReply, err error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	net, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	lessonsMode := int32(0)
	restrict, _ := restriction.FromContext(c)
	if restrict.IsLessons {
		lessonsMode = int32(1)
	}
	param := &model.Param{
		AID:           arg.Aid,
		CID:           arg.Cid,
		Qn:            arg.Qn,
		MobiApp:       dev.RawMobiApp,
		Fnval:         arg.Fnval,
		Fnver:         arg.Fnver,
		Build:         int32(dev.Build),
		Device:        dev.Device,
		ForceHost:     arg.ForceHost,
		FourkBool:     arg.Fourk,
		Buvid:         dev.Buvid,
		Platform:      dev.RawPlatform,
		Download:      arg.Download,
		TeenagersMode: arg.TeenagersMode,
		NetType:       int32(net.Type),
		TfType:        int32(net.TF),
		LessonsMode:   lessonsMode,
		IP:            net.RemoteIP,
		Business:      arg.Business,
		VoiceBalance:  arg.VoiceBalance,
	}
	//codecid=12则为h265
	switch arg.PreferCodecType {
	case api.CodeType_CODE264:
		param.PreferCodecID = model.CodeH264
	case api.CodeType_CODE265:
		param.PreferCodecID = model.CodeH265
	case api.CodeType_CODEAV1:
		param.PreferCodecID = model.CodeAV1
	default:
		// 默认返回264,和视频云确认，所有dash类视频都会有264
		param.PreferCodecID = model.CodeH264
	}
	plat := model.Plat(param.MobiApp, param.Device)
	res, err = s.svr.PlayView(c, au.Mid, param, plat)
	return
}

func (s *Server) PlayConfEdit(c context.Context, arg *api.PlayConfEditReq) (res *api.PlayConfEditReply, err error) {
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	if dev.Buvid == "" {
		err = ecode.RequestErr
		return
	}
	infoParams := &model.CloudEditParam{
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    dev.Build,
		Model:    dev.Model,
		Brand:    dev.Brand,
	}
	res, err = s.svr.PlayConfEdit(c, infoParams, arg)
	return
}

// PlayConf .
func (s *Server) PlayConf(c context.Context, arg *api.PlayConfReq) (res *api.PlayConfReply, err error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	if dev.Buvid == "" {
		err = ecode.RequestErr
		return
	}
	infoParams := &model.CloudEditParam{
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    dev.Build,
		Model:    dev.Model,
		Brand:    dev.Brand,
	}
	res, err = s.svr.PlayConf(c, infoParams, au.Mid)
	return
}

func SetDevContextInterceptor() grpc.UnaryServerInterceptor {
	const _grpcDeviceBin = "x-bili-device-bin"
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if gmd, ok := metadata.FromIncomingContext(ctx); ok {
			if clientDev := gmd.Get(_grpcDeviceBin); len(clientDev) > 0 {
				ctx = metadata.AppendToOutgoingContext(ctx, _grpcDeviceBin, clientDev[0])
			}
		}
		return handler(ctx, req)
	}
}
