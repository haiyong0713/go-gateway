package grpc

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	v1 "go-gateway/app/app-svr/app-wall/interface/api"
	"go-gateway/app/app-svr/app-wall/interface/http"
	operatorsvr "go-gateway/app/app-svr/app-wall/interface/service/operator"
	unicomsvr "go-gateway/app/app-svr/app-wall/interface/service/unicom"

	client "git.bilibili.co/bapis/bapis-go/bilibili/app/wall/v1"
)

// Server struct
type Server struct {
	unicomSvc   *unicomsvr.Service
	operatorSvc *operatorsvr.Service
}

// New Coin warden rpc server
func New(c *warden.ServerConfig, svr *http.Server) (wsvr *warden.Server, err error) {
	s := &Server{
		unicomSvc:   svr.UnicomSvc,
		operatorSvc: svr.OperatorSvc,
	}
	wsvr = warden.NewServer(c)
	v1.RegisterAppWallServer(wsvr.Server(), s)
	client.RegisterWallServer(wsvr.Server(), s)
	wsvr, err = wsvr.Start()
	return
}

// UnicomBindInfo
func (s *Server) UnicomBindInfo(c context.Context, arg *v1.UsersRequest) (reply *v1.UsersReply, err error) {
	if arg == nil {
		err = ecode.RequestErr
		return
	}
	if reply, err = s.unicomSvc.UnicomBindInfosGRPC(c, arg.Mids); err != nil {
		log.Error("s.unicomSvc.UnicomBindInfosGRPC error(%v)", err)
		return
	}
	return
}

func (s *Server) RuleInfo(c context.Context, arg *client.RuleRequest) (reply *client.RulesReply, err error) {
	if reply, err = s.operatorSvc.RuleInfo(c); err != nil {
		log.Error("%+v", err)
	}
	return
}
