package service

import (
	"context"

	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) Converge(c context.Context, arg *api.NoArgRequest) (res *api.ConvergeCardReply, err error) {
	list, err := s.cardDao.ConvergeCards(c)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	res = &api.ConvergeCardReply{List: list}
	return
}
