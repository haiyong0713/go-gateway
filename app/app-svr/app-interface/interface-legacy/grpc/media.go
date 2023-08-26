package grpc

import (
	"context"
	"go-common/component/metadata/network"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcApi "go-gateway/app/app-svr/archive/service/api"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/media"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/media"
)

type MediaServer struct {
	srcSvr *media.Service
}

func newMedia(ws *warden.Server, svr *http.Server) error {
	s := &MediaServer{
		srcSvr: svr.MediaSvr,
	}
	api.RegisterMediaServer(ws.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.Media/MediaTab", auther.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.interface.v1.Media/MediaDetail", auther.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.interface.v1.Media/MediaVideo", auther.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.interface.v1.Media/MediaRelation", auther.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.interface.v1.Media/MediaFollow", auther.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.interface.v1.Media/MediaComment", auther.UnaryServerInterceptor(false))
	return nil
}

func (s *MediaServer) MediaComment(c context.Context, arg *api.MediaCommentReq) (*api.MediaCommentReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	return s.srcSvr.MediaComment(c, arg, mid)
}

func (s *MediaServer) MediaTab(c context.Context, arg *api.MediaTabReq) (*api.MediaTabReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	if arg.Args == nil {
		arg.Args = make(map[string]string)
	}
	arg.Args["source"] = arg.Source
	return s.srcSvr.MediaTab(c, arg, mid, dev)
}

func (s *MediaServer) MediaDetail(c context.Context, arg *api.MediaDetailReq) (*api.MediaDetailReply, error) {
	return s.srcSvr.MediaDetail(c, arg)
}

func (s *MediaServer) MediaVideo(c context.Context, arg *api.MediaVideoReq) (*api.MediaVideoReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	network, ok := network.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	var batchArg *arcApi.BatchPlayArg
	if arg.PlayerArgs != nil {
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, network, mid)
	} else {
		batchArg = buildPlayerArg(dev, network, mid)
	}
	ctx := arcmid.NewContext(c, batchArg)
	return s.srcSvr.MediaVideo(ctx, arg, mid, dev)
}

func buildPlayerArg(dev device.Device, network network.Network, mid int64) *arcApi.BatchPlayArg {
	batchArg := &arcApi.BatchPlayArg{}
	batchArg.Ip = network.RemoteIP
	batchArg.Buvid = dev.Buvid
	batchArg.MobiApp = dev.RawMobiApp
	batchArg.NetType = arcApi.NetworkType(network.Type)
	batchArg.TfType = arcApi.TFType(network.TF)
	batchArg.Build = dev.Build
	batchArg.Mid = mid
	batchArg.Device = dev.Device
	return batchArg
}

func (s *MediaServer) MediaRelation(c context.Context, arg *api.MediaRelationReq) (*api.MediaRelationReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	return s.srcSvr.MediaRelation(c, arg, mid, dev)
}

func (s *MediaServer) MediaFollow(c context.Context, arg *api.MediaFollowReq) (*api.MediaFollowReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	return s.srcSvr.MediaFollow(c, arg, mid)
}
