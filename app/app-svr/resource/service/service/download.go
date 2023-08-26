package service

import (
	"context"

	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) DownLoad(c context.Context, arg *api.NoArgRequest) (res *api.DownLoadCardReply, err error) {
	list, err := s.cardDao.DownLoad(c)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	res = &api.DownLoadCardReply{List: list}
	return
}
