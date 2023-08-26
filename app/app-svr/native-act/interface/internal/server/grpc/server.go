package grpc

import (
	"context"

	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcmidv1 "go-gateway/app/app-svr/archive/middleware/v1"
	pb "go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/service"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"

	"github.com/golang/protobuf/ptypes/empty"
)

type Server struct {
	svr *service.Service
}

// New new a grpc server.
func New(svc *service.Service) (ws *warden.Server, err error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	ws = warden.NewServer(&cfg)
	s := &Server{
		svr: svc,
	}
	pb.RegisterNativeActServer(ws.Server(), s)
	// 用户鉴权
	author := mauth.New(nil)
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Index", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Dynamic", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Editor", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Resource", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Video", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Vote", author.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Reserve", author.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/TimelineSupernatant", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/OgvSupernatant", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/FollowOgv", author.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/Progress", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/BottomTab", author.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.nativeact.v1.NativeAct/HandleClick", author.UnaryServerInterceptor(false))
	ws, err = ws.Start()
	return
}

func (s *Server) Index(c context.Context, arg *pb.IndexReq) (*pb.PageResp, error) {
	ctx, err := contextWithBatchPlayArg(c, arg.PlayerArgs)
	if err != nil {
		return nil, err
	}
	return s.svr.Index(ctx, arg)
}

func (s *Server) TopicIndex(c context.Context, arg *pb.TopicIndexReq) (*pb.PageResp, error) {
	return &pb.PageResp{}, nil
}

func (s *Server) InlineIndex(c context.Context, arg *pb.InlineIndexReq) (*pb.PageResp, error) {
	return &pb.PageResp{}, nil
}

func (s *Server) TabIndex(c context.Context, arg *pb.TabIndexReq) (*pb.PageResp, error) {
	return &pb.PageResp{}, nil
}

func (s *Server) Dynamic(c context.Context, req *pb.DynamicReq) (*pb.DynamicResp, error) {
	ctx, err := contextWithBatchPlayArg(c, req.PlayerArgs)
	if err != nil {
		return nil, err
	}
	return s.svr.Dynamic(ctx, req)
}

func (s *Server) Editor(c context.Context, req *pb.EditorReq) (*pb.EditorResp, error) {
	return s.svr.Editor(c, req)
}

func (s *Server) Resource(c context.Context, req *pb.ResourceReq) (*pb.ResourceResp, error) {
	return s.svr.Resource(c, req)
}

func (s *Server) Video(c context.Context, req *pb.VideoReq) (*pb.VideoResp, error) {
	return s.svr.Video(c, req)
}

func (s *Server) Vote(c context.Context, req *pb.VoteReq) (*pb.VoteResp, error) {
	return s.svr.Vote(c, req)
}

func (s *Server) Reserve(c context.Context, req *pb.ReserveReq) (*pb.ReserveRly, error) {
	return s.svr.Reserve(c, req)
}

func (s *Server) TimelineSupernatant(c context.Context, req *pb.TimelineSupernatantReq) (*pb.TimelineSupernatantResp, error) {
	return s.svr.TimelineSupernatant(c, req)
}

func (s *Server) OgvSupernatant(c context.Context, req *pb.OgvSupernatantReq) (*pb.OgvSupernatantResp, error) {
	return s.svr.OgvSupernatant(c, req)
}

func (s *Server) FollowOgv(c context.Context, req *pb.FollowOgvReq) (*pb.FollowOgvRly, error) {
	return s.svr.FollowOgv(c, req)
}

func (s *Server) Progress(c context.Context, req *pb.ProgressReq) (*pb.ProgressRly, error) {
	return s.svr.Progress(c, req)
}

func (s *Server) BottomTab(c context.Context, req *pb.BottomTabReq) (*pb.BottomTabRly, error) {
	return s.svr.BottomTab(c, req)
}

func (s *Server) HandleClick(c context.Context, req *pb.HandleClickReq) (*pb.HandleClickRly, error) {
	return s.svr.HandleClick(c, req)
}

func (s *Server) Ping(c context.Context, arg *empty.Empty) (*empty.Empty, error) {
	return nil, nil
}

func contextWithBatchPlayArg(c context.Context, args *arcmidv1.PlayerArgs) (context.Context, error) {
	au, _ := auth.FromContext(c)
	dev, ok := device.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	net, ok := network.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	batchArg := arcmid.MossBatchPlayArgs(args, dev, net, au.Mid)
	return arcmid.NewContext(c, batchArg), nil
}
