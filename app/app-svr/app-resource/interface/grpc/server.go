package grpc

import (
	"context"

	"go-common/library/net/rpc/warden"
	v1 "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/app-svr/app-resource/interface/http"
	entrancesvr "go-gateway/app/app-svr/app-resource/interface/service/entrance"
	privacsvr "go-gateway/app/app-svr/app-resource/interface/service/privacy"
)

// Server struct
type Server struct {
	privacSvc   *privacsvr.Service
	entranceSvc *entrancesvr.Service
}

// New Coin warden rpc server
func New(c *warden.ServerConfig, svr *http.Server) (wsvr *warden.Server, err error) {
	wsvr = warden.NewServer(c)
	v1.RegisterAppResourceServer(wsvr.Server(), &Server{
		privacSvc:   svr.PrivacySvc,
		entranceSvc: svr.EntranceSvc,
	})
	RegisterModule(wsvr, svr)
	RegisterPrivacy(wsvr, svr)
	wsvr, err = wsvr.Start()
	return
}

// ModuleUpdateCache update module cache
func (s *Server) ModuleUpdateCache(c context.Context, noArg *v1.NoArgRequest) (noReply *v1.NoReply, err error) {
	noReply = &v1.NoReply{}
	return
}

// PrivacyConfig 获取app端隐私设置
func (s *Server) PrivacyConfig(c context.Context, arg *v1.NoArgRequest) (reply *v1.PrivacyConfigReply, err error) {
	reply = new(v1.PrivacyConfigReply)

	return
}

// SetPrivacyConfig 设置app端隐私设置
func (s *Server) SetPrivacyConfig(c context.Context, arg *v1.SetPrivacyConfigRequest) (reply *v1.NoReply, err error) {
	reply = new(v1.NoReply)

	return
}

// CheckEntranceInfoc 检查入口上报信息是否存在
func (s *Server) CheckEntranceInfoc(c context.Context, arg *v1.CheckEntranceInfocRequest) (*v1.CheckEntranceInfocReply, error) {
	reply, err := s.entranceSvc.CheckEntranceInfoc(c, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
