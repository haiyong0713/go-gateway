package grpc

import (
	"context"
	"time"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-common/library/xstr"

	mrestrict "go-common/component/restriction/middleware/grpc"
	api "go-gateway/app/app-svr/app-show/interface/api/popular"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/http"
	"go-gateway/app/app-svr/app-show/interface/model"
	showSvr "go-gateway/app/app-svr/app-show/interface/service/show"
	"go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

type PopularServer struct {
	showSvc *showSvr.Service
	config  *conf.Config
}

func PopularGRPC(wsvr *warden.Server, svr *http.Server) {
	s := &PopularServer{
		showSvc: svr.ShowSvc,
		config:  svr.Config,
	}
	api.RegisterPopularServer(wsvr.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	wsvr.Add("/bilibili.app.show.v1.Popular/Index", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())
}

func (s *PopularServer) Index(c context.Context, arg *api.PopularResultReq) (res *api.PopularReply, err error) {
	var (
		mid     int64
		ver     string
		config  *api.Config
		topAids []int64
	)
	if arg.Idx < 0 {
		arg.Idx = 0
	}
	if arg.LocationIds != "" {
		topAids, _ = xstr.SplitInts(arg.LocationIds)
	}
	// 获取鉴权mid
	au, ok := auth.FromContext(c)
	if ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	// 获取网络信息
	net, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	// 获取语言信息
	loc, _ := locale.FromContext(c)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	data, ok := s.showSvc.AuditFeed3(c, dev.RawMobiApp, plat, int(dev.Build), mid, dev.Device)
	if !ok {
		log.Warn("popular_index api(%s) mid(%d) buvid(%s) mobi_app(%s) device(%s) idx(%d) flush(%d) lastparam(%s) source_id(%d) topAids(%v)", "grpc_popular_index", mid, dev.Buvid, dev.RawMobiApp, dev.Device, arg.Idx, arg.Flush, arg.LastParam, arg.SourceId, topAids)
		//高版本秒开参数
		var batchArg *arcgrpc.BatchPlayArg
		if arg.PlayerArgs != nil && s.config.Custom.PlayerArgs {
			batchArg = middleware.MossBatchPlayArgs(arg.PlayerArgs, dev, net, mid)
		} else {
			batchArg = &arcgrpc.BatchPlayArg{
				Ip:        metadata.String(c, metadata.RemoteIP),
				Build:     dev.Build,
				Device:    dev.Device,
				NetType:   arcgrpc.NetworkType(net.Type),
				Qn:        int64(arg.Qn),
				MobiApp:   dev.RawMobiApp,
				Fnver:     int64(arg.Fnver),
				Fnval:     int64(arg.Fnval),
				ForceHost: int64(arg.ForceHost),
				Buvid:     dev.Buvid,
				Mid:       au.Mid,
				Fourk:     int64(arg.Fourk),
				TfType:    arcgrpc.TFType(net.TF),
			}
		}
		c = middleware.NewContext(c, batchArg)
		data, ver, config, err = s.showSvc.FeedIndex3(c, mid, arg.EntranceId, arg.Idx, plat, int(dev.Build), arg.LoginEvent, dev.RawMobiApp, dev.Device, dev.Buvid, arg.Spmid, topAids, arg.SourceId, time.Now(), loc, arg.Flush, arg.PopularAd)
	}
	res = &api.PopularReply{
		Items:  data,
		Ver:    ver,
		Config: config,
	}
	return
}
