package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/collection-splash/api"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) AddSplash(ctx context.Context, param *pb.AddSplashReq) (*pb.SetSplashReply, error) {
	id, err := s.dao.AddSplash(ctx, param)
	if err != nil {
		log.Error("Failed to AddSplash: %+v, %+v", param, err)
		return nil, err
	}
	if id == 0 {
		return nil, ecode.Error(ecode.NothingFound, "添加失败，返回的自增id为0")
	}
	return &pb.SetSplashReply{Id: id}, nil
}

func (s *Service) UpdateSplash(ctx context.Context, param *pb.UpdateSplashReq) (*pb.SetSplashReply, error) {
	if err := s.checkSplash(ctx, param.Id); err != nil {
		return nil, err
	}
	row, err := s.dao.UpdateSplash(ctx, param)
	if err != nil {
		log.Error("Failed to UpdateSplash: %+v, %+v", param, err)
		return nil, err
	}
	return &pb.SetSplashReply{
		Id: row,
	}, nil
}

func (s *Service) DeleteSplash(ctx context.Context, param *pb.SplashReq) (*pb.SetSplashReply, error) {
	if err := s.checkSplash(ctx, param.Id); err != nil {
		return nil, err
	}
	row, err := s.dao.DeleteSplash(ctx, param)
	if err != nil {
		log.Error("Failed to DeleteSplash: %+v, %+v", param, err)
		return nil, err
	}
	if row == 0 {
		return nil, ecode.Error(ecode.NothingFound, "删除失败，返回的影响行数为0")
	}
	return &pb.SetSplashReply{
		Id: row,
	}, nil
}

func (s *Service) checkSplash(ctx context.Context, id int64) error {
	_, err := s.Splash(ctx, &pb.SplashReq{Id: id})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Splash(ctx context.Context, param *pb.SplashReq) (*pb.SplashReply, error) {
	splash, err := s.dao.Splash(ctx, param)
	if err != nil {
		log.Error("Failed to Splash: %+v, %+v", param, err)
		return nil, err
	}
	if splash.GetId() == 0 {
		return nil, ecode.Error(ecode.NothingFound, "找不到对应的闪屏配置")
	}
	if splash.GetIsDeleted() {
		return nil, ecode.Error(ecode.NothingFound, "查询的闪屏配置已被删除")
	}
	return &pb.SplashReply{
		Splash: splash,
	}, nil
}

func (s *Service) SplashList(ctx context.Context, _ *empty.Empty) (*pb.SplashListReply, error) {
	splashList, err := s.dao.SplashList(ctx)
	if err != nil {
		log.Error("Failed to SplashList: %+v", err)
		return nil, err
	}
	return &pb.SplashListReply{
		List: splashList,
	}, nil
}
