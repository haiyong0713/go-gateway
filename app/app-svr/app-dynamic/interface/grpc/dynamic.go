package grpc

import (
	"context"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	"strconv"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	abtest "go-common/component/tinker/middleware/grpc"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/http"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
	dynsvr "go-gateway/app/app-svr/app-dynamic/interface/service/dynamic"
	"go-gateway/app/app-svr/archive/middleware"
	middlewarev1 "go-gateway/app/app-svr/archive/middleware/v1"

	arcApi "go-gateway/app/app-svr/archive/service/api"

	metav1 "git.bilibili.co/bapis/bapis-go/bilibili/metadata"
	"github.com/gogo/protobuf/proto"
	xmetadata "google.golang.org/grpc/metadata"
)

const (
	// header
	_metadata = "x-bili-metadata-bin"
)

type Server struct {
	dynSvr *dynsvr.Service
	mauth  *mauth.Auth
	config *conf.Config
}

func New(wsvr *warden.Server, auth *mauth.Auth, svr *http.Server) (*Server, error) {
	s := &Server{
		dynSvr: svr.DynamicSvc,
		config: svr.Config,
		mauth:  auth,
	}
	api.RegisterDynamicServer(wsvr.Server(), s)

	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/SVideo", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC()) // guest: true/false
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynVideo", s.mauth.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynTab", s.mauth.UnaryServerInterceptor(false), abtest.UnaryServerInterceptor(), svr.FeatureSvc.BuildLimitGRPC()) // guest: true/false
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynOurCitySwitch", s.mauth.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())                        // guest: true/false
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynOurCity", s.mauth.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())                              // guest: true/false
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynDetails", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynVideoPersonal", s.mauth.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynUpdOffset", s.mauth.UnaryServerInterceptor(false), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynRed", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynMixUpListViewMore", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/DynMixUpListSearch", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/OurCityClickReport", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	wsvr.Add("/bilibili.app.dynamic.v1.Dynamic/GeoCoder", s.mauth.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	return s, nil
}

func (s *Server) DynVideo(c context.Context, req *api.DynVideoReq) (*api.DynVideoReqReply, error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	if req.RefreshType == 0 ||
		(req.RefreshType == dynmdl.RefreshTypeDown && (req.Offset == "" || req.Offset == "0")) {
		return nil, ecode.RequestErr
	}
	request := &dynmdl.DynVideoReq{
		Teenager:       int(req.TeenagersMode),
		UpdateBaseLine: req.UpdateBaseline,
		Offset:         req.Offset,
		Page:           int(req.Page),
		Refresh:        int(req.RefreshType),
		Mid:            au.Mid,
		VideoMate: &dynmdl.VideoMate{
			Qn:        int(req.Qn),
			Fnver:     int(req.Fnver),
			Fnval:     int(req.Fnval),
			ForceHost: int(req.ForceHost),
			Fourk:     int(req.Fourk),
		},
	}
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
	}
	ctx := s.buildPlayerArgs(c, nil, int64(req.Fnval), int64(req.Fnver), int64(req.ForceHost), int64(req.Qn), int64(req.Fourk))
	return s.dynSvr.DynVideo(ctx, header, request)
}

func (s *Server) DynDetails(c context.Context, req *api.DynDetailsReq) (*api.DynDetailsReply, error) {
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	dynIDs, err := xstr.SplitInts(req.DynamicIds)
	if err != nil {
		log.Errorc(c, "xstr.SplitInts() failed. error(%+v)", err)
		return nil, err
	}
	if len(dynIDs) == 0 {
		log.Errorc(c, "params check failed. len(dynIDs) == 0")
		return nil, ecode.RequestErr
	}
	request := &dynmdl.DynDetailsReq{
		Teenager: int(req.TeenagersMode),
		DynIDs:   dynIDs,
		Mid:      au.Mid,
		VideoMate: &dynmdl.VideoMate{
			Qn:        int(req.Qn),
			Fnver:     int(req.Fnver),
			Fnval:     int(req.Fnval),
			ForceHost: int(req.ForceHost),
			Fourk:     int(req.Fourk),
		},
	}
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
	}
	ctx := s.buildPlayerArgs(c, nil, int64(req.Fnval), int64(req.Fnver), int64(req.ForceHost), int64(req.Qn), int64(req.Fourk))
	return s.dynSvr.DynDetails(ctx, header, request)
}

func (s *Server) headerGet(c context.Context) (md metav1.Metadata, err error) {
	gmd, ok := xmetadata.FromIncomingContext(c)
	if !ok {
		log.Error("headerGet xmetadata.FromIncomingContext ok(%v) gmd(%v)", ok, gmd)
		return
	}
	tmp := gmd.Get(_metadata)
	if len(tmp) > 0 && tmp[0] != "" {
		if err = proto.Unmarshal([]byte(tmp[0]), &md); err != nil {
			log.Error("headerGet proto.Unmarshal error(%v)", err)
			return
		}
	}
	return
}

func (s *Server) DynVideoPersonal(c context.Context, req *api.DynVideoPersonalReq) (*api.DynVideoPersonalReply, error) {
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	if req.HostUid <= 0 {
		log.Errorc(c, "params check failed. host_uid(%v) <= 0", req.HostUid)
		return nil, ecode.RequestErr
	}
	request := &dynmdl.DynVideoPersonalReq{
		Teenager:  int(req.TeenagersMode),
		Offset:    req.Offset,
		Page:      int(req.Page),
		IsPreload: int(req.IsPreload),
		VideoMate: &dynmdl.VideoMate{
			Qn:        int(req.Qn),
			Fnver:     int(req.Fnver),
			Fnval:     int(req.Fnval),
			ForceHost: int(req.ForceHost),
			Fourk:     int(req.Fourk),
		},
		Mid:     au.Mid,
		HostUID: req.HostUid,
	}
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
		IP:       metadata.String(c, metadata.RemoteIP),
	}
	ctx := s.buildPlayerArgs(c, nil, int64(req.Fnval), int64(req.Fnver), int64(req.ForceHost), int64(req.Qn), int64(req.Fourk))
	return s.dynSvr.DynVideoPersonal(ctx, header, request)
}

func (s *Server) DynUpdOffset(c context.Context, req *api.DynUpdOffsetReq) (*api.NoReply, error) {
	au, _ := auth.FromContext(c)
	if req.HostUid <= 0 || req.ReadOffset == "" {
		log.Errorc(c, "params check failed. required params is empty. host_uid(%v), read_offset(%v)", req.HostUid, req.ReadOffset)
		return nil, ecode.RequestErr
	}
	request := &dynmdl.DynUpdOffsetReq{
		HostUID:    req.HostUid,
		ReadOffset: req.ReadOffset,
		Mid:        au.Mid,
	}
	return s.dynSvr.DynUpdOffset(c, request)
}

func (s *Server) SVideo(c context.Context, req *api.SVideoReq) (*api.SVideoReply, error) {
	// 获取鉴权mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
	}
	vm := &dynmdl.VideoMate{}
	vm.FromSVideo(req)
	var idx int64
	if dynmdl.IsPopularSv(req) && req.Offset != "" {
		var err error
		if idx, err = strconv.ParseInt(req.Offset, 10, 64); err != nil {
			return nil, ecode.RequestErr
		}
	}
	// 秒开信息处理
	ctx := s.buildPlayerArgs(c, req.PlayerArgs, int64(req.Fnval), int64(req.Fnver), int64(req.ForceHost), int64(req.Qn), int64(req.Fourk))
	return s.dynSvr.SVideo(ctx, req, au.Mid, header, vm, idx)
}

func (s *Server) DynTab(c context.Context, req *api.DynTabReq) (res *api.DynTabReply, err error) {
	res = new(api.DynTabReply)
	au, ok := auth.FromContext(c)
	if !ok {
		return res, ecode.RequestErr
	}
	md, err := s.headerGet(c)
	if err != nil || req == nil {
		return res, ecode.RequestErr
	}
	header := &dynmdl.Header{
		MobiApp:  md.MobiApp,
		Device:   md.Device,
		Buvid:    md.Buvid,
		Platform: md.Platform,
		Build:    int(md.Build),
	}
	if req.TeenagersMode != model.TeenagersClose && req.TeenagersMode != model.TeenagersOpen {
		return res, ecode.RequestErr
	}
	return s.dynSvr.DynTab(c, au.Mid, header, req)
}

func (s *Server) DynOurCitySwitch(c context.Context, req *api.DynOurCitySwitchReq) (res *api.NoReply, err error) {
	return nil, ecode.MethodNotAllowed
}

func (s *Server) DynOurCity(c context.Context, req *api.DynOurCityReq) (res *api.DynOurCityReply, err error) {
	return nil, ecode.MethodNotAllowed
}

func (s *Server) DynRed(c context.Context, req *api.DynRedReq) (res *api.DynRedReply, err error) {
	res = new(api.DynRedReply)
	au, _ := auth.FromContext(c)
	md, err := s.headerGet(c)
	if err != nil || req == nil {
		return res, ecode.RequestErr
	}
	header := &dynmdl.Header{
		MobiApp:  md.MobiApp,
		Device:   md.Device,
		Buvid:    md.Buvid,
		Platform: md.Platform,
		Build:    int(md.Build),
	}
	return s.dynSvr.DynRed(c, header, req, au.Mid)
}

func (s *Server) DynMixUpListSearch(c context.Context, req *api.DynMixUpListSearchReq) (res *api.DynMixUpListSearchReply, err error) {
	// 获取设备信息
	dev, _ := device.FromContext(c)
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
	}
	return s.dynSvr.DynMixUpListSearch(c, header, req)
}

func (s *Server) DynMixUpListViewMore(c context.Context, req *api.NoReq) (res *api.DynMixUpListViewMoreReply, err error) {
	// 获取设备信息
	dev, _ := device.FromContext(c)
	header := &dynmdl.Header{
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Buvid:    dev.Buvid,
		Platform: dev.RawPlatform,
		Build:    int(dev.Build),
	}
	return s.dynSvr.DynMixUpListViewMore(c, header, req)
}

func (s *Server) buildPlayerArgs(c context.Context, playArg *middlewarev1.PlayerArgs, fnval, fnver, forcehost, qn, fourk int64) context.Context {
	// 获取鉴权 mid
	au, _ := auth.FromContext(c)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	// 获取网络信息
	net, _ := network.FromContext(c)
	//高版本秒开参数
	var batchArg *arcApi.BatchPlayArg
	if s.config.Ctrl.PlayerArgs && playArg != nil {
		batchArg = middleware.MossBatchPlayArgs(playArg, dev, net, au.Mid)
	} else {
		batchArg = &arcApi.BatchPlayArg{
			Ip:        net.RemoteIP,
			Build:     dev.Build,
			Device:    dev.Device,
			NetType:   arcApi.NetworkType(net.Type),
			Qn:        qn,
			MobiApp:   dev.RawMobiApp,
			Fnver:     fnver,
			Fnval:     fnval,
			ForceHost: forcehost,
			Buvid:     dev.Buvid,
			Mid:       au.Mid,
			Fourk:     fourk,
			TfType:    arcApi.TFType(net.TF),
		}
	}
	return middleware.NewContext(c, batchArg)
}

func (s *Server) OurCityClickReport(c context.Context, req *api.OurCityClickReportReq) (res *api.OurCityClickReportReply, err error) {
	return nil, ecode.MethodNotAllowed
}

func (s *Server) GeoCoder(c context.Context, req *api.GeoCoderReq) (res *api.GeoCoderReply, err error) {
	return s.dynSvr.GeoCoder(c, req.Lat, req.Lng, req.From)
}
