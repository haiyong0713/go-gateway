package cards

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	"time"
)

// InviteShare 分享 返回token
func (s *Service) InviteShare(ctx context.Context, mid int64, activity string) (res *cardsmdl.InviteTokenReply, err error) {
	res = &cardsmdl.InviteTokenReply{}
	token, err := s.dao.GetInviteMidToToken(ctx, mid, activity)
	if err != nil {
		log.Errorc(ctx, " s.dao.GetInviteMidToToken err(%v)", err)
		return
	}
	if token == "" {
		return res, ecode.SpringFestivalInviterTokenErr
	}
	res.Token = token
	return
}

// CardShare ...
func (s *Service) CardShare(ctx context.Context, mid int64, cardID int64, activity string) (res *cardsmdl.CardTokenReply, err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	if cardID > cardsConfig.CardsNum {
		err = ecode.SpringFestivalCardsErr
		return
	}
	res = &cardsmdl.CardTokenReply{}
	if err = s.midLimit(ctx, mid, midLimitMax, activity); err != nil {
		log.Errorc(ctx, "mid(%d) limit error error(%v)", mid, err)
		return res, err
	}
	cards, err := s.dao.GetMidsCardsNew(ctx, mid, cardsConfig.ID)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetMidCards err(%v)", err)
		return res, err
	}
	// 检查库存
	err = s.checkCardStore(ctx, cardID, cards)
	if err != nil {
		return res, err
	}

	token := s.createToken(ctx, mid, fmt.Sprintf("%s%d", activity, cardID), time.Now().Unix())
	cardToken := &cardsmdl.CardTokenMid{
		Mid:    mid,
		CardID: cardID,
	}
	err = s.dao.AddShareCardToken(ctx, token, cardToken, activity)
	if err != nil {
		log.Errorc(ctx, " s.dao.AddShareCardToken err(%v)", err)
		return res, ecode.SpringFestivalSendCardErr
	}
	res.Token = token
	return
}

// qpsLimit ...
func (s *Service) midLimit(c context.Context, mid int64, maxLimit int64, activity string) error {
	limit, err := s.dao.MidLimit(c, mid, activity)
	if err != nil {
		log.Errorc(c, "s.dao.MidLimit(%d) error(%v)", mid, err)
		return err
	}
	if limit > maxLimit {
		return ecode.SpringFestivalTooFastErr
	}
	return nil
}

func (s *Service) checkCardStore(ctx context.Context, cardID int64, cards []*cardsmdl.CardMid) (err error) {
	for _, v := range cards {
		if cardID == v.CardID {
			if v.Nums-v.Used <= 1 {
				return ecode.SpringFestivalCardStoreErr
			}
		}
	}
	return nil
}
