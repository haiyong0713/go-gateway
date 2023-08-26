package service

import (
	"context"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) MenuExtVer(c context.Context, arg *pb.MenuExtVerReq) (*pb.MenuExtVerReply, error) {
	click, err := s.cacheDao.CacheMenuVer(c, arg.Id, arg.Buvid, arg.Ver)
	if err != nil {
		return nil, err
	}
	return &pb.MenuExtVerReply{Click: int32(click)}, nil
}

func (s *Service) AddMenuExtVer(c context.Context, arg *pb.AddMenuExtVerReq) (*pb.AddMenuExtVerReply, error) {
	err := s.cacheDao.AddMenuVer(c, arg.Id, arg.Buvid, arg.Ver)
	if err != nil {
		return nil, err
	}
	return &pb.AddMenuExtVerReply{}, nil
}
