package service

import (
	"context"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) GetS10PopEntranceAids(ctx context.Context, req *pb.GetS10PopEntranceAidsReq) (reply *pb.GetS10PopEntranceAidsReply, err error) {
	aids, err := s.cacheDao.CacheAIChannelRes(ctx, s.c.PopEntranceS10Id)
	if err != nil {
		return nil, err
	}
	reply = &pb.GetS10PopEntranceAidsReply{Aids: aids}
	return reply, nil
}
