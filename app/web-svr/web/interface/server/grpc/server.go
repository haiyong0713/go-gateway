package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/service"

	v1 "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
)

type server struct {
	srv *service.Service
}

// New Coin warden rpc server .
func New(c *conf.GRPCCfg, svr *service.Service) (ws *warden.Server) {
	var (
		err error
	)
	ws = warden.NewServer(c.GRPC)
	limiter := quota.New(c.QuotaCfg)
	ws.Use(limiter.Limit())
	v1.RegisterWebInterfaceServer(ws.Server(), &server{srv: svr})
	author := mauth.New(nil)
	ws.Add("/bilibili.web.interface.v1.WebInterface/ViewDetail", author.WebInterceptor(true))
	ws.Add("/bilibili.web.interface.v1.WebInterface/ActivitySeason", author.WebInterceptor(true))
	ws.Add("/bilibili.web.interface.v1.WebInterface/ActivityArchive", author.WebInterceptor(true))
	ws.Add("/bilibili.web.interface.v1.WebInterface/ActivityLiveTimeInfo", author.WebInterceptor(true))
	ws.Add("/bilibili.web.interface.v1.WebInterface/ClickActivitySeason", author.WebInterceptor(false))
	if ws, err = ws.Start(); err != nil {
		panic(err)
	}
	return ws
}

func (s *server) ViewDetail(c context.Context, req *v1.ViewDetailReq) (resp *v1.ViewDetailReply, err error) {
	resp = new(v1.ViewDetailReply)
	au, _ := auth.FromContext(c)
	v, err := s.srv.DetailGRPC(c, req, au.Mid, reqGRPCBuvid(c))
	if err != nil {
		return
	}
	resp = v
	return
}

func (s *server) ActivitySeason(c context.Context, req *v1.ActivitySeasonReq) (*v1.ActivitySeasonReply, error) {
	au, _ := auth.FromContext(c)
	return s.srv.ActivitySeason(c, au.Mid, req, reqGRPCBuvid(c))
}

func (s *server) ActivityArchive(c context.Context, req *v1.ActivityArchiveReq) (*v1.ActivityArchiveReply, error) {
	au, _ := auth.FromContext(c)
	return s.srv.ActivityArchive(c, au.Mid, req, reqGRPCBuvid(c))
}

func (s *server) ActivityLiveTimeInfo(c context.Context, req *v1.ActivityLiveTimeInfoReq) (*v1.ActivityLiveTimeInfoReply, error) {
	au, _ := auth.FromContext(c)
	return s.srv.ActivityLiveTimeInfo(c, au.Mid, req, reqGRPCBuvid(c))
}

func (s *server) ClickActivitySeason(c context.Context, req *v1.ClickActivitySeasonReq) (*v1.NoReply, error) {
	//au, _ := auth.FromContext(c)
	//err := s.srv.ClickActivitySeason(c, au.Mid, req, "")
	//if err != nil {
	//	return nil, err
	//}
	// only support http
	return nil, ecode.AccessDenied
}

func reqGRPCBuvid(ctx context.Context) string {
	buvid := ""
	if device, ok := device.FromContext(ctx); ok {
		buvid = device.Buvid3
	}
	return buvid
}
