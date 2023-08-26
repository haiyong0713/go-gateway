package service

import (
	"context"
	"go-common/library/log"
	rm "go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) loadLimitFreeOnline() {
	reply, err := s.resolutionDao.FetchAllLimitFreeOnline(context.Background())
	if err != nil {
		log.Error("loadLimitFreeOnline failed error(%+v)", err)
		return
	}
	s.limitFreeOnline = reply
	log.Info("load limit free success %+v", reply)
}

func (s *Service) FetchLimitFreeOnline() (*rm.LimitFreeReply, error) {
	reply := &rm.LimitFreeReply{
		LimitFreeWithAid: make(map[int64]*rm.LimitFreeInfo, len(s.limitFreeOnline)),
	}
	for _, v := range s.limitFreeOnline {
		if v == nil {
			continue
		}
		reply.LimitFreeWithAid[v.Aid] = v
	}
	return reply, nil
}
