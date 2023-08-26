package grpc

import (
	"context"

	xecode "go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"

	pb "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/interface/conf"
	service "go-gateway/app/web-svr/native-page/interface/service/like"
)

// New new a grpc server.
func New(c *warden.ServerConfig, svc *service.Service, cfg *conf.Config) *warden.Server {
	ws := warden.NewServer(c)
	limiter := quota.New(cfg.QuotaConf)
	ws.Use(limiter.Limit())
	pb.RegisterNaPageServer(ws.Server(), &naPageService{svr: svc})
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}

type naPageService struct {
	svr *service.Service
}

var _ pb.NaPageServer = &naPageService{}

func (s *naPageService) ModuleConfig(c context.Context, r *pb.ModuleConfigReq) (rs *pb.ModuleConfigReply, err error) {
	if rs, err = s.svr.ModuleConfig(c, r); err != nil {
		return
	}
	if rs == nil {
		err = xecode.NothingFound
	}
	return
}

// NatInfoFromForeign .
func (s *naPageService) NatInfoFromForeign(c context.Context, r *pb.NatInfoFromForeignReq) (rs *pb.NatInfoFromForeignReply, err error) {
	var (
		pages map[int64]*pb.NativePage
	)
	rs = &pb.NatInfoFromForeignReply{}
	if pages, err = s.svr.NatInfoFromForeign(c, r.Fids, r.PageType, r.Content); err != nil {
		return
	}
	rs.List = pages
	return
}

// NatConfig .
func (s *naPageService) NatConfig(c context.Context, r *pb.NatConfigReq) (rs *pb.NatConfigReply, err error) {
	if r.Ps <= 0 {
		r.Ps = 20
	}
	if rs, err = s.svr.NatConfig(c, r); err != nil {
		return
	}
	if rs == nil {
		err = xecode.NothingFound
	}
	return
}

// NatConfig .
func (s *naPageService) BaseConfig(c context.Context, r *pb.BaseConfigReq) (rs *pb.BaseConfigReply, err error) {
	if r.Ps == 0 {
		// 兼容老版本获取全部组件
		r.Ps = -1
	}
	if rs, err = s.svr.BaseConfig(c, r); err != nil {
		return
	}
	if rs == nil {
		err = xecode.NothingFound
	}
	return
}

// ModuleMixExt .
func (s *naPageService) ModuleMixExt(c context.Context, r *pb.ModuleMixExtReq) (rs *pb.ModuleMixExtReply, err error) {
	if rs, err = s.svr.ModuleMixExt(c, r.ModuleID, r.Offset, r.Ps, r.MType); err != nil {
		return
	}
	if rs == nil {
		err = xecode.NothingFound
	}
	return
}

// ModuleMixExt .
func (s *naPageService) ModuleMixExts(c context.Context, r *pb.ModuleMixExtsReq) (rs *pb.ModuleMixExtsReply, err error) {
	if rs, err = s.svr.ModuleMixExts(c, r.ModuleID, r.Offset, r.Ps); err != nil {
		return
	}
	if rs == nil {
		err = xecode.NothingFound
	}
	return
}

func (s *naPageService) NativePages(c context.Context, r *pb.NativePagesReq) (*pb.NativePagesReply, error) {
	rly, err := s.svr.NativePages(c, r.Pids)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativePagesReply{}, nil
	}
	return &pb.NativePagesReply{List: rly}, nil
}

func (s *naPageService) NativePageCards(c context.Context, r *pb.NativePageCardsReq) (*pb.NativePageCardsReply, error) {
	rly, err := s.svr.NativePageCards(c, r.Pids, r.Build, r.MobiApp, r.Platform)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativePageCardsReply{}, nil
	}
	return &pb.NativePageCardsReply{List: rly}, nil
}

func (s *naPageService) NativePagesExt(c context.Context, r *pb.NativePagesExtReq) (*pb.NativePagesExtReply, error) {
	rly, err := s.svr.NativePagesExt(c, r.Pids)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativePagesExtReply{}, nil
	}
	return &pb.NativePagesExtReply{List: rly}, nil
}

func (s *naPageService) NativeValidPagesExt(c context.Context, r *pb.NativeValidPagesExtReq) (*pb.NativeValidPagesExtReply, error) {
	rly, err := s.svr.NativeValidPagesExt(c, r)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativeValidPagesExtReply{}, nil
	}
	return &pb.NativeValidPagesExtReply{List: rly}, nil
}

func (s *naPageService) NativePage(c context.Context, r *pb.NativePageReq) (*pb.NativePageReply, error) {
	rly, err := s.svr.NativePage(c, r.Pid)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativePageReply{}, nil
	}
	return &pb.NativePageReply{Item: rly}, nil
}

// NatTabModules .
func (s *naPageService) NatTabModules(c context.Context, r *pb.NatTabModulesReq) (*pb.NatTabModulesReply, error) {
	rly, err := s.svr.NatTabModules(c, r)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NatTabModulesReply{}, nil
	}
	return rly, nil
}

func (s *naPageService) NativePagesTab(c context.Context, r *pb.NativePagesTabReq) (*pb.NativePagesTabReply, error) {
	rly, err := s.svr.NativePagesTab(c, r.Pids, r.Category)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativePagesTabReply{}, nil
	}
	return rly, nil
}

// up主发起活动白名单接口 .
func (s *naPageService) IsUpActUid(c context.Context, r *pb.IsUpActUidReq) (*pb.IsUpActUidReply, error) {
	match, err := s.svr.IsUpActUid(c, r.Mid)
	if err != nil {
		return nil, err
	}
	return &pb.IsUpActUidReply{Match: match}, nil
}

// up主发起有效活动列表接口 .
func (s *naPageService) UpActNativePages(c context.Context, r *pb.UpActNativePagesReq) (*pb.UpActNativePagesReply, error) {
	rly, err := s.svr.UpActNativePages(c, r.Mid, r.Offset, r.Ps)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.UpActNativePagesReply{}, nil
	}
	return rly, nil
}

// up主发起活动-进审核态
func (s *naPageService) UpActNativePageBind(c context.Context, r *pb.UpActNativePageBindReq) (*pb.UpActNativePageBindReply, error) {
	rly, err := s.svr.UpActNativePageBind(c, r)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.UpActNativePageBindReply{}, nil
	}
	return rly, nil
}

func (s *naPageService) NativeForbidList(c context.Context, r *pb.NativeForbidListReq) (*pb.NoReply, error) {
	if err := s.svr.NativeForbidList(c, r); err != nil {
		return nil, err
	}
	return &pb.NoReply{}, nil
}

func (s *naPageService) SponsorNativePages(ctx context.Context, req *pb.SponsorNativePagesReq) (*pb.SponsorNativePagesReply, error) {
	return &pb.SponsorNativePagesReply{}, nil
}

// GetReserveProgress 获取预约数据
func (s *naPageService) GetNatProgressParams(ctx context.Context, req *pb.GetNatProgressParamsReq) (*pb.GetNatProgressParamsReply, error) {
	return s.svr.GetNatProgressParams(ctx, req)
}

func (s *naPageService) NativeAllPages(c context.Context, r *pb.NativeAllPagesReq) (*pb.NativeAllPagesReply, error) {
	rly, err := s.svr.NativeAllPages(c, r.Pids)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &pb.NativeAllPagesReply{}, nil
	}
	return &pb.NativeAllPagesReply{List: rly}, nil
}

func (s *naPageService) SpaceSyncSetting(c context.Context, r *pb.SpaceSyncSettingReq) (*pb.SpaceSyncSettingReply, error) {
	return s.svr.SpaceSyncSetting(c, r)
}

func (s *naPageService) NativeAllPageCards(c context.Context, r *pb.NativeAllPageCardsReq) (*pb.NativeAllPageCardsReply, error) {
	rly, err := s.svr.NativeAllPageCards(c, r.Pids)
	if err != nil {
		return nil, err
	}
	return &pb.NativeAllPageCardsReply{List: rly}, nil
}
