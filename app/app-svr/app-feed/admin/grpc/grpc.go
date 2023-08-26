package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"

	pb "go-gateway/app/app-svr/app-feed/admin/api"
	"go-gateway/app/app-svr/app-feed/admin/service/pwd_appeal"
	"go-gateway/app/app-svr/app-feed/admin/service/search"
)

// RPC struct info.
type RpcService struct {
	s      *search.Service
	appeal *pwd_appeal.Service
}

// New new a grpc server.
func New(cfg *warden.ServerConfig, svc *search.Service, appeal *pwd_appeal.Service) *warden.Server {
	ws := warden.NewServer(cfg)
	pb.RegisterFeedAdminServer(ws.Server(), &RpcService{s: svc, appeal: appeal})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

// 给频道服务端用，返回管理后台所有配置过的频道id
func (r *RpcService) OpenChannelIds(c context.Context, req *pb.OpenChannelIdsReq) (res *pb.OpenChannelIdsReply, err error) {
	var (
		//nolint:ineffassign
		total = 0
	)
	res = &pb.OpenChannelIdsReply{
		Page: &pb.PageInfo{
			Num:   req.Pn,
			Size_: req.Ps,
		},
	}
	res.Ids, total, err = r.s.OpenChannelIdsCache(int(req.Ps), int(req.Pn))
	res.Page.Total = int32(total)
	return
}

func (r *RpcService) CreatePwdAppeal(c context.Context, req *pb.CreatePwdAppealReq) (*pb.CreatePwdAppealRly, error) {
	return r.appeal.CreatePwdAppeal(req)
}
