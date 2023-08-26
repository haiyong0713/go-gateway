package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
)

func (s *Service) CardFollow(c context.Context, arg *api.NoArgRequest) (res *api.CardFollowReply, err error) {
	list, err := s.cardDao.Follow(c)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	res = &api.CardFollowReply{List: list}
	return
}

func (s *Service) CardPosRecs(c context.Context, arg *api.CardPosRecReplyRequest) (*api.CardPosRecReply, error) {
	res := &api.CardPosRecReply{}
	if arg == nil || len(arg.CardIds) == 0 {
		return nil, ecode.RequestErr
	}
	card := map[int64]*api.CardPosRec{}
	for _, id := range arg.CardIds {
		if p, ok := s.feedPosRecCache[id]; ok {
			card[p.Id] = p
		}
	}
	res.Card = card
	return res, nil
}

func (s *Service) loadPosRec() {
	tmp, err := s.cardDao.PosRec(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.feedPosRecCache = tmp
}
