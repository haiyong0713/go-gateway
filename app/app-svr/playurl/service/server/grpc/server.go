package grpc

import (
	"context"
	abtest "go-common/component/tinker/middleware/grpc"
	"go-common/library/net/rpc/warden"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	pb "go-gateway/app/app-svr/playurl/service/api"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	svc *service.Service
}

type serverV2 struct {
	svc *service.Service
}

// PlayURL get playurl by grpc
func (s *serverV2) PlayURL(c context.Context, req *v2.PlayURLReq) (resp *v2.PlayURLReply, err error) {
	return s.svc.PlayURLV2(c, req)
}

// PlayURL get playurl by grpc
func (s *serverV2) PlayView(c context.Context, req *v2.PlayViewReq) (resp *v2.PlayViewReply, err error) {
	ctx := s.svc.FeatureSvc.BuildLimitManual(c, &feature.BuildLimitManual{
		Build:   int64(req.Build),
		MobiApp: req.MobiApp,
		Device:  req.Device,
	})
	return s.svc.PlayView(ctx, req)
}

func (s *serverV2) PlayConfEdit(c context.Context, req *v2.PlayConfEditReq) (resp *v2.PlayConfEditReply, err error) {
	return s.svc.PlayConfEdit(c, req, "0")
}

func (s *serverV2) PlayConf(c context.Context, req *v2.PlayConfReq) (resp *v2.PlayConfReply, err error) {
	return s.svc.PlayConf(c, req)
}

// New new warden rpc server
func New(c *warden.ServerConfig, svc *service.Service) *warden.Server {
	ws := warden.NewServer(c)
	ws.Use(abtest.UnaryServerInterceptor())
	ws.Use(SetDevContextInterceptor())
	pb.RegisterPlayURLServer(ws.Server(), &server{svc: svc})
	v2.RegisterPlayURLServer(ws.Server(), &serverV2{svc: svc})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

// PlayURL get playurl by http
func (s *server) PlayURL(c context.Context, req *pb.PlayURLReq) (resp *pb.PlayURLReply, err error) {
	return s.svc.PlayURL(c, req)
}

// SteinsPreview get SteinsPreview
func (s *server) SteinsPreview(c context.Context, req *pb.SteinsPreviewReq) (resp *pb.SteinsPreviewReply, err error) {
	resp = new(pb.SteinsPreviewReply)
	resp.Playurl, err = s.svc.SteinsPreview(c, req)
	return
}

// Project get tv project playurl
func (s *serverV2) Project(c context.Context, req *v2.ProjectReq) (resp *v2.ProjectReply, err error) {
	return s.svc.Project(c, req)
}

// Project get tv project playurl
func (s *serverV2) ChronosPkg(c context.Context, req *v2.ChronosPkgReq) (*v2.ChronosPkgReply, error) {
	return s.svc.ChronosPkg(c, req)
}

// HlsScheduler get tv project playurl
func (s *serverV2) HlsScheduler(c context.Context, req *v2.HlsCommonReq) (resp *v2.HlsSchedulerReply, err error) {
	return s.svc.HlsScheduler(c, req)
}

// MasterScheduler
func (s *serverV2) MasterScheduler(c context.Context, req *v2.HlsCommonReq) (resp *v2.MasterSchedulerReply, err error) {
	return s.svc.MasterScheduler(c, req)
}

// M3U8Scheduler .
func (s *serverV2) M3U8Scheduler(c context.Context, req *v2.HlsCommonReq) (resp *v2.M3U8SchedulerReply, err error) {
	return s.svc.M3U8Scheduler(c, req)
}

func (s *serverV2) PlayOnline(c context.Context, arg *v2.PlayOnlineReq) (*v2.PlayOnlineReply, error) {
	return s.svc.PlayOnlineGRPC(c, arg)
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
