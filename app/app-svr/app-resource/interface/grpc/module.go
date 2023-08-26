package grpc

import (
	"context"
	"time"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/fawkes"
	mfawkes "go-common/component/fawkes/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	cfawkes "go-common/component/metadata/fawkes"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-resource/interface/http"
	"go-gateway/app/app-svr/app-resource/interface/service/mod"

	v1 "git.bilibili.co/bapis/bapis-go/bilibili/app/resource/v1"
)

type ModuleServer struct {
	modSvc *mod.Service
}

func RegisterModule(wsvr *warden.Server, svr *http.Server) {
	v1.RegisterModuleServer(wsvr.Server(), &ModuleServer{
		modSvc: svr.ModSvc,
	})
	// 用户鉴权
	auther := mauth.New(nil)
	faw := mfawkes.NewMiddleware(fawkes.Default())
	wsvr.Add("/bilibili.app.resource.v1.Module/List", anticrawler.ReportInterceptor(), auther.UnaryServerInterceptor(true), faw.UnaryServerInterceptor())
	//nolint:gosimple
	return
}

func (s *ModuleServer) List(ctx context.Context, in *v1.ListReq) (*v1.ListReply, error) {
	var mid int64
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	dev, _ := device.FromContext(ctx)
	if dev.RawMobiApp == "" {
		return nil, ecode.RequestErr
	}
	if dev.Build == 0 {
		return nil, ecode.RequestErr
	}
	faw, _ := cfawkes.FromContext(ctx)
	// 针对ipad HD2需要做特殊兼容，fawkes-key为ipad的，映射为ipad2
	if faw.AppKey == "ipad" {
		faw.AppKey = "ipad2"
	}
	now := time.Now()
	return s.modSvc.GRPCListWrap(ctx, faw.AppKey, dev.Buvid, mid, dev.Build, dev.Device, in, now)
}
