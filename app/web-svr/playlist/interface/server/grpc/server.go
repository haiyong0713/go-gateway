package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"
	api "go-gateway/app/web-svr/playlist/interface/api/v1"
	"go-gateway/app/web-svr/playlist/interface/model"
	"go-gateway/app/web-svr/playlist/interface/service"

	"github.com/golang/protobuf/ptypes/empty"
)

type server struct {
	srv *service.Service
}

// New Coin warden rpc server .
func New(c *warden.ServerConfig, svr *service.Service) (ws *warden.Server) {
	var (
		err error
	)
	ws = warden.NewServer(c)
	api.RegisterPlaylistServer(ws.Server(), &server{srv: svr})
	if ws, err = ws.Start(); err != nil {
		panic(err)
	}
	return ws
}

// SetStat set playlist stat cache.
func (s *server) SetStat(ctx context.Context, req *api.PlStatReq) (res *empty.Empty, err error) {
	err = s.srv.SetStat(ctx, &model.PlStat{
		ID:    req.Id,
		Mid:   req.Mid,
		Fid:   req.Fid,
		View:  req.View,
		Reply: req.Reply,
		Fav:   req.Fav,
		Share: req.Share,
		MTime: req.Mtime})
	return &empty.Empty{}, err
}
