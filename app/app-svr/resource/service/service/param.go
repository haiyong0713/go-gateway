package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) loadParamCache() {
	params, err := s.show.ParamList(context.Background())
	if err != nil {
		log.Error("loadParamCache %+v", err)
		return
	}
	s.paramsCache = params
	log.Info("loadParamCache success")
}

// ParamList .
func (s *Service) ParamList(_ context.Context, req *v1.ParamReq) (*v1.ParamReply, error) {
	if len(req.GetPlats()) == 0 {
		return &v1.ParamReply{List: s.paramsCache}, nil
	}
	platMap := make(map[int64]struct{}, len(req.GetPlats()))
	for _, v := range req.GetPlats() {
		platMap[v] = struct{}{}
	}
	reply := new(v1.ParamReply)
	for _, v := range s.paramsCache {
		if _, ok := platMap[v.Plat]; ok {
			reply.List = append(reply.List, v)
		}
	}
	return reply, nil
}
