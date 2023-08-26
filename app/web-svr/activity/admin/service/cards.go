package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/cards"
)

func (s *Service) AddCards(ctx context.Context, card *cards.Cards) (err error) {
	id, err := s.dao.AddCards(ctx, card)
	if err != nil {
		log.Errorc(ctx, "s.dao.AddCards err(%v)", err)
		return
	}
	err = s.dao.CreateMidCards(ctx, id)
	if err != nil {
		log.Errorc(ctx, "s.dao.AddCards err(%v)", err)
		return
	}
	return
}
