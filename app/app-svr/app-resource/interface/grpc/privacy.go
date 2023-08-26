package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/library/net/rpc/warden"

	api "go-gateway/app/app-svr/app-resource/interface/api/privacy"
	"go-gateway/app/app-svr/app-resource/interface/http"
	privacysvr "go-gateway/app/app-svr/app-resource/interface/service/privacy"
)

type PrivacyServer struct {
	privacySvc *privacysvr.Service
}

func RegisterPrivacy(wsvr *warden.Server, svr *http.Server) {
	s := &PrivacyServer{
		privacySvc: svr.PrivacySvc,
	}
	api.RegisterPrivacyServer(wsvr.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	wsvr.Add("/bilibili.app.resource.privacy.v1.Privacy/PrivacyConfig", auther.UnaryServerInterceptor(true))
	wsvr.Add("/bilibili.app.resource.privacy.v1.Privacy/SetPrivacyConfig", auther.UnaryServerInterceptor(true))
	//nolint:gosimple
	return
}

func (s *PrivacyServer) PrivacyConfig(c context.Context, arg *api.NoArgRequest) (reply *api.PrivacyConfigReply, err error) {
	var mid int64
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	return s.privacySvc.PrivacyConfig(c, mid)
}

func (s *PrivacyServer) SetPrivacyConfig(c context.Context, arg *api.SetPrivacyConfigRequest) (reply *api.NoReply, err error) {
	var mid int64
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	return s.privacySvc.SetPrivacyConfig(c, mid, arg)
}
