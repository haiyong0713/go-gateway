package service

import (
	"context"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

// SearchOgv .
func (s *Service) SearchOgv(ctx context.Context, req *pb.SearchOgvReq) (res *pb.SearchOgvReply, err error) {
	res = &pb.SearchOgvReply{Sids: []int64{}}
	tmp := s.searchOgvCache
	if ids, ok := tmp[req.Id]; ok {
		res.Sids = ids
	}
	return
}
