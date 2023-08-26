package grpc

import (
	"context"
	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	mrestrict "go-common/component/restriction/middleware/grpc"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	hisSvr "go-gateway/app/app-svr/app-interface/interface-legacy/service/history"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcApi "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

type HistoryServer struct {
	hisSvr *hisSvr.Service
	config *conf.Config
}

//nolint:unparam
func newHistory(ws *warden.Server, svr *http.Server) (err error) {
	s := &HistoryServer{
		hisSvr: svr.HistorySvr,
		config: svr.Config,
	}
	api.RegisterHistoryServer(ws.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.History/HistoryTab", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor()) // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/Cursor", auther.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())       // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/CursorV2", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())      // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/Delete", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())        // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/Search", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())        // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/Clear", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())         // guest: true/false
	ws.Add("/bilibili.app.interface.v1.History/LatestHistory", auther.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.interface.v1.History/HistoryTabV2", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor())
	return
}

func (s *HistoryServer) HistoryTabV2(c context.Context, arg *api.HistoryTabReq) (*api.HistoryTabReply, error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err := ecode.RequestErr
		return nil, err
	}
	rest, ok := restriction.FromContext(c)
	if !ok {
		err := ecode.RequestErr
		return nil, err
	}
	return s.hisSvr.HistoryTabGRPCV2(c, au.Mid, dev.Buvid, arg, rest.IsLessons)
}

// HistoryTab is
func (s *HistoryServer) HistoryTab(c context.Context, arg *api.HistoryTabReq) (res *api.HistoryTabReply, err error) {
	res = new(api.HistoryTabReply)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	rest, ok := restriction.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	return s.hisSvr.HistoryTabGRPC(c, au.Mid, dev, arg, rest.IsLessons)
}

func (s *HistoryServer) Cursor(c context.Context, arg *api.CursorReq) (res *api.CursorReply, err error) {
	res = new(api.CursorReply)
	// 获取鉴权mid
	au, ok := auth.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	network, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	var batchArg *arcApi.BatchPlayArg
	if arg.PlayerArgs != nil && s.config.Switch.PlayerArgs {
		// 公共秒开参数处理方法
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, network, au.Mid)
	} else {
		batchArg = buildBatchArg(arg.PlayerPreload, dev, network, au.Mid)
	}
	ctx := arcmid.NewContext(c, batchArg)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	return s.hisSvr.CursorGRPC(ctx, au.Mid, dev, plat, arg, network)
}

// CursorV2 for history
func (s *HistoryServer) CursorV2(c context.Context, arg *api.CursorV2Req) (res *api.CursorV2Reply, err error) {
	res = new(api.CursorV2Reply)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	network, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	var batchArg *arcApi.BatchPlayArg
	if arg.PlayerArgs != nil && s.config.Switch.PlayerArgs {
		// 公共秒开参数处理方法
		batchArg = arcmid.MossBatchPlayArgs(arg.PlayerArgs, dev, network, au.Mid)
	} else {
		batchArg = buildBatchArg(arg.PlayerPreload, dev, network, au.Mid)
	}
	ctx := arcmid.NewContext(c, batchArg)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	return s.hisSvr.CursorV2GRPC(ctx, au.Mid, dev, plat, arg, network)
}

// Delete is
func (s *HistoryServer) Delete(c context.Context, arg *api.DeleteReq) (res *api.NoReply, err error) {
	res = new(api.NoReply)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	var param []*hisApi.ModelHistory
	for _, v := range arg.HisInfo {
		param = append(param, &hisApi.ModelHistory{
			Kid:      v.Kid,
			Business: v.Business,
		})
	}
	err = s.hisSvr.Del(c, au.Mid, param, dev)
	return
}

func (s *HistoryServer) Search(c context.Context, arg *api.SearchReq) (res *api.SearchReply, err error) {
	res = new(api.SearchReply)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	network, ok := network.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	//兼容无business情况,设置默认为all
	if arg.Business == "" {
		arg.Business = "all"
	}
	return s.hisSvr.Search(c, au.Mid, plat, arg, dev, network)
}

func (s *HistoryServer) Clear(c context.Context, arg *api.ClearReq) (*api.NoReply, error) {
	res := new(api.NoReply)
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
	}
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	if err := s.hisSvr.ClearGRPC(c, arg, dev, plat, au.Mid); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *HistoryServer) LatestHistory(c context.Context, arg *api.LatestHistoryReq) (*api.LatestHistoryReply, error) {
	// 获取鉴权mid
	au, ok := auth.FromContext(c)
	if !ok {
		return nil, ecode.RequestErr
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
	batchArg := buildBatchArg(arg.PlayerPreload, dev, network, au.Mid)
	ctx := arcmid.NewContext(c, batchArg)
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	return s.hisSvr.LatestHistoryGRPC(ctx, au.Mid, dev, plat, arg)
}

func buildBatchArg(params *api.PlayerPreloadParams, dev device.Device, network network.Network, mid int64) *arcApi.BatchPlayArg {
	batchArg := &arcApi.BatchPlayArg{}
	if params != nil {
		batchArg.Fnval = params.Fnval
		batchArg.Fnver = params.Fnver
		batchArg.ForceHost = params.ForceHost
		batchArg.Fourk = params.Fourk
		batchArg.Qn = params.Qn
	}
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
