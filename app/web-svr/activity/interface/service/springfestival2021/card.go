package springfestival2021

import (
	"context"
	"fmt"
	"go-common/library/database/sql"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	springfestival2021mdl "go-gateway/app/web-svr/activity/interface/model/springfestival2021"
	"strconv"

	"time"
)

const (
	midLimitMax = 1
	retry       = 3
	timeSleep   = 100 * time.Millisecond
)

// AddTimes 分享增加抽奖次数
func (s *Service) AddTimes(ctx context.Context, mid int64) (err error) {
	sid := s.c.SpringFestival2021.LotterySid
	orderNo := strconv.FormatInt(mid, 10) + strconv.Itoa(l.TimesShareType) + strconv.FormatInt(time.Now().Unix(), 10)
	return s.lotterySvr.AddLotteryTimes(ctx, sid, mid, 0, l.TimesShareType, 0, orderNo, true)
}

// Draw 抽卡
func (s *Service) Draw(ctx context.Context, mid int64, risk *riskmdl.Base, num int, ts int64) (res []*springfestival2021mdl.Card, err error) {
	sid := s.c.SpringFestival2021.LotterySid
	orderNo := fmt.Sprintf("%d_%d", mid, ts)
	gift, err := s.lotterySvr.DoLottery(ctx, sid, mid, risk, num, false, orderNo)
	res = make([]*springfestival2021mdl.Card, 0)
	if err != nil {
		log.Errorc(ctx, "springfestival do lottery err(%v)", err)
		return
	}
	if gift != nil {
		cardMapNum := make(map[string]int64)
		for _, v := range gift {
			giftID := v.GiftID
			giftName := v.GiftName
			imgURL := v.ImgURL
			ctime := v.Ctime
			if v.GiftID == 0 {
				res = append(res, &springfestival2021mdl.Card{GiftID: 0})
				continue
			}
			cardID := s.giftIDtoCardID(ctx, giftID)
			res = append(res, &springfestival2021mdl.Card{
				GiftID:   giftID,
				GiftName: giftName,
				ImgURL:   imgURL,
				CardID:   cardID,
				Ctime:    ctime,
			})
			// 获得的卡数+1
			if cardID > 0 {
				cardName := s.gardIDToCardDbName(ctx, cardID)
				if cardName != "" {
					if _, ok := cardMapNum[cardName]; !ok {
						cardMapNum[cardName] = 0
					}
					cardMapNum[cardName]++
				}
			}
		}
		if len(cardMapNum) > 0 {
			_, err = s.dao.UpdateCardNumsIncr(ctx, mid, cardMapNum)
			if err != nil {
				log.Errorc(ctx, "s.dao.UpdateCardNumsIncr err(%v)", err)
				return
			}
		}
	}
	err = s.dao.DeleteMidCardDetail(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteMidCardDetail mid(%d) err(%v)", mid, err)
		err = nil
	}
	return res, err
}

func (s *Service) gardIDToCardDbName(ctx context.Context, cardID int64) (cardName string) {
	if cardID == springfestival2021mdl.CardID1 {
		return springfestival2021mdl.Card1DB
	}
	if cardID == springfestival2021mdl.CardID2 {
		return springfestival2021mdl.Card2DB
	}
	if cardID == springfestival2021mdl.CardID3 {
		return springfestival2021mdl.Card3DB
	}
	if cardID == springfestival2021mdl.CardID4 {
		return springfestival2021mdl.Card4DB
	}
	if cardID == springfestival2021mdl.CardID5 {
		return springfestival2021mdl.Card5DB
	}
	return ""
}

// Times 剩余抽奖次数
func (s *Service) Times(ctx context.Context, mid int64) (res *l.TimesReply, err error) {
	sid := s.c.SpringFestival2021.LotterySid
	return s.lotterySvr.GetUnusedTimes(ctx, sid, mid)
}

// Cards 用户已经获得的卡及合成情况
func (s *Service) Cards(ctx context.Context, mid int64) (res *springfestival2021mdl.CardsReply, err error) {
	var (
		cards *springfestival2021mdl.MidNums
	)
	res = &springfestival2021mdl.CardsReply{}
	res.Cards = &springfestival2021mdl.MidCard{}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if cards, err = s.dao.GetMidCards(ctx, mid); err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards err(%v)", err)
		}
		return
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}
	if cards != nil {
		var card1, card2, card3, card4, card5 int64
		if cards.Card1-cards.Card1Used > 0 {
			card1 = cards.Card1 - cards.Card1Used
		}
		if cards.Card2-cards.Card2Used > 0 {
			card2 = cards.Card2 - cards.Card2Used
		}
		if cards.Card3-cards.Card3Used > 0 {
			card3 = cards.Card3 - cards.Card3Used
		}
		if cards.Card4-cards.Card4Used > 0 {
			card4 = cards.Card4 - cards.Card4Used
		}
		if cards.Card5-cards.Card5Used > 0 {
			card5 = cards.Card5 - cards.Card5Used
		}
		res.Cards = &springfestival2021mdl.MidCard{
			Card1:   card1,
			Card2:   card2,
			Card3:   card3,
			Card4:   card4,
			Card5:   card5,
			Compose: cards.Compose,
		}
		if card1 > 0 && card2 > 0 && card3 > 0 && card4 > 0 && card5 > 0 {
			res.CanCompose = true
		}
	}

	return res, nil
}

// giftIDtoCardID ...
func (s *Service) giftIDtoCardID(ctx context.Context, giftID int64) int64 {
	for _, cardGiftID := range s.c.SpringFestival2021.Card1GiftID {
		if giftID == cardGiftID {
			return springfestival2021mdl.CardID1
		}
	}
	for _, cardGiftID := range s.c.SpringFestival2021.Card2GiftID {

		if giftID == cardGiftID {
			return springfestival2021mdl.CardID2
		}
	}
	for _, cardGiftID := range s.c.SpringFestival2021.Card3GiftID {

		if giftID == cardGiftID {
			return springfestival2021mdl.CardID3
		}
	}
	for _, cardGiftID := range s.c.SpringFestival2021.Card4GiftID {

		if giftID == cardGiftID {
			return springfestival2021mdl.CardID4
		}
	}
	for _, cardGiftID := range s.c.SpringFestival2021.Card5GiftID {

		if giftID == cardGiftID {
			return springfestival2021mdl.CardID5
		}
	}
	return 0
}

func (s *Service) checkCompose(c context.Context, cards *springfestival2021mdl.MidNums) (err error) {

	if cards.Card1-cards.Card1Used <= 0 {
		err = ecode.SpringFestivalComposeCardStoreErr
		return
	}
	if cards.Card2-cards.Card2Used <= 0 {
		err = ecode.SpringFestivalComposeCardStoreErr
		return
	}
	if cards.Card3-cards.Card3Used <= 0 {
		err = ecode.SpringFestivalComposeCardStoreErr
		return
	}
	if cards.Card4-cards.Card4Used <= 0 {
		err = ecode.SpringFestivalComposeCardStoreErr
		return
	}
	if cards.Card5-cards.Card5Used <= 0 {
		err = ecode.SpringFestivalComposeCardStoreErr
		return
	}
	return nil
}

// Compose 合成卡
func (s *Service) Compose(c context.Context, mid int64, risk *riskmdl.Base, mobiApp string) (err error) {
	cards, err := s.dao.GetMidCards(c, mid)
	if err != nil {
		log.Errorc(c, "s.dao.GetMidCards err(%v)", err)
		return err
	}
	err = s.checkCompose(c, cards)
	if err != nil {
		log.Errorc(c, "s.checkCompose (%v)", err)
		return err
	}

	// 风控
	spRisk := &riskmdl.Sf21Compose{
		Base:        *risk,
		Mid:         mid,
		ActivityUID: ActivityUID,
		MobiApp:     mobiApp,
	}
	_, err = s.risk(c, mid, riskmdl.ActionSf21Compose, spRisk, spRisk.EsTime)
	if err != nil {
		log.Errorc(c, "s.risk mid(%d) compose err(%v)", mid, err)
	}
	var (
		tx  *sql.Tx
		res *springfestival2021mdl.MidNums
	)
	if tx, err = s.dao.BeginTran(c); err != nil {
		log.Errorc(c, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "Compose %v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
			return
		}
	}()
	if res, err = s.dao.MidNumsForUpdateTx(c, tx, mid); err != nil {
		log.Errorc(c, "Add s.dao.MidNumsForUpdateTx(%d) failed. error(%v)", mid, err)
		return
	}

	// 使用卡
	err = s.composeUsedCard(c, res)
	if err != nil {
		return err
	}
	_, err = s.dao.UpdateCardNums(c, tx, mid, res.Card1, res.Card1Used, res.Card2, res.Card2Used, res.Card3, res.Card3Used, res.Card4, res.Card4Used, res.Card5, res.Card5Used, res.Compose)
	if err != nil {
		log.Errorc(c, "s.dao.UpdateCardNums err(%v)", err)
		return ecode.SpringFestivalComposeCardErr
	}
	_, err = s.dao.InsertComposeLogTx(c, tx, mid)
	if err != nil {
		log.Errorc(c, "s.dao.InsertComposeLogTx err(%v)", err)
		return ecode.SpringFestivalComposeCardErr
	}
	err = s.dao.DeleteMidCardDetail(c, mid)
	if err != nil {
		log.Errorc(c, "s.dao.DeleteMidCardDetail mid(%d) err(%v)", mid, err)
		err = nil
	}
	return
}

// composeUsedCard 合成用卡
func (s *Service) composeUsedCard(c context.Context, cards *springfestival2021mdl.MidNums) (err error) {
	err = s.checkCompose(c, cards)
	if err != nil {
		log.Errorc(c, "s.checkCompose (%v)", err)
		return err
	}
	cards.Card1Used++
	cards.Card2Used++
	cards.Card3Used++
	cards.Card4Used++
	cards.Card5Used++
	cards.Compose++
	return nil
}

func (s *Service) checkCardStore(ctx context.Context, cardID int64, cards *springfestival2021mdl.MidNums) (err error) {
	if cardID == springfestival2021mdl.CardID1 {
		if cards.Card1-cards.Card1Used <= 1 {
			return ecode.SpringFestivalCardStoreErr
		}
	}
	if cardID == springfestival2021mdl.CardID2 {
		if cards.Card2-cards.Card2Used <= 1 {
			return ecode.SpringFestivalCardStoreErr
		}
	}
	if cardID == springfestival2021mdl.CardID3 {
		if cards.Card3-cards.Card3Used <= 1 {
			return ecode.SpringFestivalCardStoreErr
		}
	}
	if cardID == springfestival2021mdl.CardID4 {
		if cards.Card4-cards.Card4Used <= 1 {
			return ecode.SpringFestivalCardStoreErr
		}
	}
	if cardID == springfestival2021mdl.CardID5 {
		if cards.Card5-cards.Card5Used <= 1 {
			return ecode.SpringFestivalCardStoreErr
		}
	}
	return nil
}

// CardShare ...
func (s *Service) CardShare(ctx context.Context, mid int64, cardID int64) (res *springfestival2021mdl.CardTokenReply, err error) {

	res = &springfestival2021mdl.CardTokenReply{}
	if err = s.midLimit(ctx, mid, midLimitMax); err != nil {
		log.Errorc(ctx, "mid(%d) limit error error(%v)", mid, err)
		return res, err
	}
	cards, err := s.dao.GetMidCards(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetMidCards err(%v)", err)
		return res, err
	}
	// 检查库存
	err = s.checkCardStore(ctx, cardID, cards)
	if err != nil {
		return res, err
	}

	token := s.createToken(ctx, mid, fmt.Sprintf("%d", cardID), time.Now().Unix())
	cardToken := &springfestival2021mdl.CardTokenMid{
		Mid:    mid,
		CardID: cardID,
	}
	err = s.dao.AddShareCardToken(ctx, token, cardToken)
	if err != nil {
		log.Errorc(ctx, " s.dao.AddShareCardToken err(%v)", err)
		return res, ecode.SpringFestivalSendCardErr
	}
	res.Token = token
	return
}

// qpsLimit ...
func (s *Service) midLimit(c context.Context, mid int64, maxLimit int64) error {
	limit, err := s.dao.MidLimit(c, mid)
	if err != nil {
		log.Errorc(c, "s.dao.MidLimit(%d) error(%v)", mid, err)
		return err
	}
	if limit > maxLimit {
		return ecode.SpringFestivalTooFastErr
	}
	return nil
}

// GetCard ...
func (s *Service) GetCard(ctx context.Context, mid int64, token string, risk *riskmdl.Base, mobiApp string) (err error) {
	card, err := s.dao.ShareCardToken(ctx, token)
	if err != nil {
		log.Errorc(ctx, "s.dao.ShareCardToken err(%v)", err)
		return ecode.SpringFestivalCardAlreadyErr
	}
	if card == nil {
		return ecode.SpringFestivalCardAlreadyErr
	}
	if card.Mid <= 0 {
		return ecode.SpringFestivalCardAlreadyErr
	}
	if card.IsReceived == springfestival2021mdl.IsReceived {
		return ecode.SpringFestivalCardAlreadyErr
	}
	var (
		tx          *sql.Tx
		senderCards *springfestival2021mdl.MidNums
		midCards    *springfestival2021mdl.MidNums
	)

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		senderCards, err = s.dao.GetMidCards(ctx, card.Mid)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards mid(%d) err(%v)", card.Mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		midCards, err = s.dao.GetMidCards(ctx, mid)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards mid(%d) err(%v)", mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		err = s.midInsertSpringNums(ctx, mid)
		if err != nil {
			log.Errorc(ctx, "s.midInsertSpringNums mid(%d)", mid)
			return err
		}
		return err
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}
	if card.Mid == mid {
		return ecode.SpringFestivalCantGetCardErr
	}

	spRisk := &riskmdl.Sf21SendCard{
		Base:        *risk,
		Mid:         mid,
		InvitedMid:  card.Mid,
		ActivityUID: ActivityUID,
		MobiApp:     mobiApp,
		CardType:    card.CardID,
	}
	riskReply, err := s.risk(ctx, mid, riskmdl.ActionSf21SendCard, spRisk, spRisk.EsTime)
	if err != nil {
		log.Errorc(ctx, "s.risk mid(%d) send err(%v)", mid, err)
	} else {
		err = s.riskCheck(ctx, mid, riskReply)
		if err != nil {
			return err
		}
	}

	// 送卡验证
	err = s.sendCard(ctx, senderCards, midCards, card.CardID)
	if err != nil {
		return err
	}
	cardName := s.gardIDToCardDbName(ctx, card.CardID)
	cardMapNum := make(map[string]int64)
	if cardName != "" {
		cardMapNum[cardName] = 1
	}
	if len(cardMapNum) == 0 {
		return ecode.SpringFestivalGetCardErr
	}
	if tx, err = s.dao.BeginTran(ctx); err != nil {
		log.Errorc(ctx, "s.lotDao.BeginTran() failed. error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(ctx, "Compose %v", r)
			return
		}
		if err != nil {
			log.Errorc(ctx, "GetCard err(%v)", err)
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
			return
		}
	}()

	// 更新送卡人库存
	update, err := s.dao.UpdateCardNumsUsedIncrTx(ctx, tx, card.Mid, cardMapNum)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpdateCardNums mid(%d) err(%v)", card.Mid, err)
		return ecode.SpringFestivalGetCardErr
	}
	if update == 0 {
		return ecode.SpringFestivalAlreadyDonatedErr
	}
	// 更新被送卡人库存
	_, err = s.dao.UpdateCardNumsIncrTx(ctx, tx, mid, cardMapNum)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpdateCardNums mid(%d) err(%v)", card.Mid, err)
		return ecode.SpringFestivalGetCardErr
	}
	_, err = s.dao.InsertSendCardLogTx(ctx, tx, card.Mid, mid, card.CardID)
	if err != nil {
		log.Errorc(ctx, "s.dao.InsertSendCardLogTx err(%v)", err)
		return ecode.SpringFestivalGetCardErr
	}
	err = s.dao.DeleteMidCardDetail(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteMidCardDetail mid(%d) err(%v)", mid, err)
		err = nil
	}
	card.IsReceived = springfestival2021mdl.IsReceived
	card.ReceiveMid = mid
	err = s.dao.SetShareCardToken(ctx, token, card)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteShareCardToken token(%s) err(%v) ", token, err)
	}

	err = s.actSend(ctx, card.Mid, card.CardID, donateBusiness)
	if err != nil {
		log.Errorc(ctx, "s.actSend err(%v)", err)
	}

	return nil
}

// sendCard 合成用卡
func (s *Service) sendCard(c context.Context, sendCards *springfestival2021mdl.MidNums, midCards *springfestival2021mdl.MidNums, cardID int64) (err error) {
	// 检查库存
	err = s.checkCardStore(c, cardID, sendCards)
	if err != nil {
		return ecode.SpringFestivalAlreadyDonatedErr
	}
	if cardID == springfestival2021mdl.CardID1 {
		sendCards.Card1Used++
		midCards.Card1++
		if midCards.Card1-midCards.Card1Used > s.c.SpringFestival2021.MaxCard {
			return ecode.SpringFestivalGetCardMaxErr
		}
	}
	if cardID == springfestival2021mdl.CardID2 {
		sendCards.Card2Used++
		midCards.Card2++
		if midCards.Card2-midCards.Card2Used > s.c.SpringFestival2021.MaxCard {
			return ecode.SpringFestivalGetCardMaxErr
		}
	}
	if cardID == springfestival2021mdl.CardID3 {
		sendCards.Card3Used++
		midCards.Card3++
		if midCards.Card3-midCards.Card3Used > s.c.SpringFestival2021.MaxCard {
			return ecode.SpringFestivalGetCardMaxErr
		}
	}
	if cardID == springfestival2021mdl.CardID4 {
		sendCards.Card4Used++
		midCards.Card4++
		if midCards.Card4-midCards.Card4Used > s.c.SpringFestival2021.MaxCard {
			return ecode.SpringFestivalGetCardMaxErr
		}
	}
	if cardID == springfestival2021mdl.CardID5 {
		sendCards.Card5Used++
		midCards.Card5++
		if midCards.Card5-midCards.Card5Used > s.c.SpringFestival2021.MaxCard {
			return ecode.SpringFestivalGetCardMaxErr
		}
	}

	return nil
}

// CardTokenToMid 分享token转mid
func (s *Service) CardTokenToMid(ctx context.Context, token string) (res *springfestival2021mdl.CardTokenToMidReply, err error) {
	res = &springfestival2021mdl.CardTokenToMidReply{}
	res.Account = &springfestival2021mdl.Account{}
	res.Card = &springfestival2021mdl.CardIsReceived{}
	var (
		cards *springfestival2021mdl.MidNums
	)
	card, err := s.dao.ShareCardToken(ctx, token)
	if err != nil {
		log.Errorc(ctx, "s.dao.ShareCardToken err(%v)", err)
		return res, ecode.SpringFestivalCardAlreadyErr
	}
	if card == nil {
		return res, ecode.SpringFestivalCardAlreadyErr
	}
	if card.Mid <= 0 {
		return res, ecode.SpringFestivalCardAlreadyErr
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		res.Account, err = s.midToAccount(ctx, card.Mid)
		if err != nil {
			log.Errorc(ctx, "s.midToAccount mid(%d) err(%v)", card.Mid, err)
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		cards, err = s.dao.GetMidCards(ctx, card.Mid)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards err(%v)", err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return res, err
	}
	// 检查库存
	err = s.checkCardStore(ctx, card.CardID, cards)
	if err == nil {
		res.Card.IsInStock = springfestival2021mdl.IsInStock
	}
	if err != nil {
		err = nil
	}

	res.Card.CardID = card.CardID
	res.Card.IsReceived = card.IsReceived
	res.Card.Mid = card.ReceiveMid
	return
}
