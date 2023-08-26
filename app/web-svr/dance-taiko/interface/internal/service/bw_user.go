package service

import (
	"context"

	"go-common/library/log"

	account "git.bilibili.co/bapis/bapis-go/account/service"
)

func (s *Service) user(c context.Context, mid int64) (resPro *account.Card, err error) {
	resp, err := s.accClient.Card3(c, &account.MidReq{Mid: mid})
	if err != nil {
		log.Error("accClient.Card3 err(%v)", err)
		return nil, err
	}
	return resp.GetCard(), nil
}

func (s *Service) users(c context.Context, mids []int64) (resPro map[int64]*account.Card, err error) {
	resp, err := s.accClient.Cards3(c, &account.MidsReq{Mids: mids})
	if err != nil {
		log.Error("accClient.Cards3 err(%v)", err)
		return nil, err
	}
	return resp.Cards, nil
}
