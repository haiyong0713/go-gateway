package grpc

import (
	"context"
	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"

	"go-gateway/app/app-svr/app-card/interface/model/card"
	api "go-gateway/app/app-svr/app-show/interface/api/rank"
	"go-gateway/app/app-svr/app-show/interface/http"
	"go-gateway/app/app-svr/app-show/interface/model"
	rankSvr "go-gateway/app/app-svr/app-show/interface/service/rank"
)

type BRPCServer struct {
	rankSvc *rankSvr.Service
}

func rankGRPC(wsvr *warden.Server, svr *http.Server) {
	s := &BRPCServer{
		rankSvc: svr.RankSvc,
	}
	api.RegisterRankServer(wsvr.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	wsvr.Add("/bilibili.app.show.v1.Rank/RankAll", auther.UnaryServerInterceptor(true))
	wsvr.Add("/bilibili.app.show.v1.Rank/RankRegion", auther.UnaryServerInterceptor(true))
}

// nolint:gomnd
func (s *BRPCServer) RankAll(c context.Context, arg *api.RankAllResultReq) (res *api.RankListReply, err error) {
	var (
		mid  int64
		data []*api.Item
		ok   bool
	)
	if arg.Pn < 1 {
		arg.Pn = 1
	}
	if arg.Ps > 100 || arg.Ps <= 0 {
		arg.Ps = 100
	}
	if ((arg.Pn-1)*arg.Ps)+1 > 100 {
		res = &api.RankListReply{
			Items: []*api.Item{},
		}
		return
	}
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	if data, ok = s.rankSvc.Audit(c, dev.RawMobiApp, arg.Order, plat, int(dev.Build), 0, mid, dev.Device); !ok {
		data, err = s.rankSvc.RankShow(c, plat, 0, int(arg.Pn), int(arg.Ps), mid, arg.Order, dev.RawMobiApp, dev.Device)
	}
	if err != nil {
		return nil, err
	}
	res = &api.RankListReply{
		Items: data,
	}
	res = filterRankMidGreaterThanInt32(dev, res)
	return
}

// nolint:gomnd
func (s *BRPCServer) RankRegion(c context.Context, arg *api.RankRegionResultReq) (res *api.RankListReply, err error) {
	var (
		mid  int64
		data []*api.Item
		ok   bool
	)
	if arg.Pn < 1 {
		arg.Pn = 1
	}
	if arg.Ps > 100 || arg.Ps <= 0 {
		arg.Ps = 100
	}
	if ((arg.Pn-1)*arg.Ps)+1 > 100 {
		res = &api.RankListReply{
			Items: []*api.Item{},
		}
		return
	}
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	plat := model.Plat(dev.RawMobiApp, dev.Device)
	if data, ok = s.rankSvc.Audit(c, dev.RawMobiApp, "all", plat, int(dev.Build), int(arg.Rid), mid, dev.Device); !ok {
		data, err = s.rankSvc.RankShow(c, plat, int(arg.Rid), int(arg.Pn), int(arg.Ps), mid, "all", dev.RawMobiApp, dev.Device)
	}
	if err != nil {
		return nil, err
	}
	res = &api.RankListReply{
		Items: data,
	}
	res = filterRankMidGreaterThanInt32(dev, res)
	return
}

// 过滤mid大于int32
func filterRankMidGreaterThanInt32(dev device.Device, in *api.RankListReply) *api.RankListReply {
	if !card.CheckMidMaxInt32Version(dev) {
		return in
	}
	var tempItems []*api.Item
	for _, v := range in.Items {
		if card.CheckMidMaxInt32(v.Mid) {
			continue
		}
		tempItems = append(tempItems, v)
	}
	return &api.RankListReply{
		Items: tempItems,
	}
}
