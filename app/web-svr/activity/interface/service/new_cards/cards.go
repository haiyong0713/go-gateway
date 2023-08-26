package cards

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/interface/component"

	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"strconv"

	"time"
)

const (
	midLimitMax = 1
	retry       = 3
	timeSleep   = 100 * time.Millisecond
	isInternal  = true
)

// AddTimes 分享增加抽奖次数
func (s *Service) AddTimes(ctx context.Context, mid int64, activity string) (err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	sid := cardsConfig.SID
	orderNo := strconv.FormatInt(mid, 10) + strconv.Itoa(l.TimesShareType) + strconv.FormatInt(time.Now().Unix(), 10)
	return s.lotterySvr.AddLotteryTimes(ctx, sid, mid, 0, l.TimesShareType, 0, orderNo, true)
}

// Draw 抽卡
func (s *Service) Draw(ctx context.Context, mid int64, risk *riskmdl.Base, num int, ts int64, activity string) (res []*cardsmdl.Card, err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	sid := cardsConfig.SID
	orderNo := fmt.Sprintf("%d_%d", mid, ts)
	gift, err := s.lotterySvr.DoLottery(ctx, sid, mid, risk, num, false, orderNo)
	res = make([]*cardsmdl.Card, 0)
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
				res = append(res, &cardsmdl.Card{GiftID: 0})
				continue
			}
			cardID := s.giftIDtoCardID(ctx, giftID, cardsConfig.Cards)
			res = append(res, &cardsmdl.Card{
				GiftID:   giftID,
				GiftName: giftName,
				ImgURL:   imgURL,
				CardID:   cardID,
				Ctime:    ctime,
			})
			// 获得的卡数+1
			if cardID > 0 {
				cardName := fmt.Sprintf("%d", cardID)
				if cardName != "" {
					if _, ok := cardMapNum[cardName]; !ok {
						cardMapNum[cardName] = 0
					}
					cardMapNum[cardName]++
				}
			}
		}
		if len(cardMapNum) > 0 {
			_, err = s.dao.UpdateCardNumsIncrNew(ctx, mid, cardsConfig.ID, cardMapNum)
			if err != nil {
				log.Errorc(ctx, "s.dao.UpdateCardNumsIncrNew err(%v)", err)
				return
			}
		}
	}
	err = s.dao.DeleteMidCardDetailNew(ctx, mid, cardsConfig.ID)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteMidCardDetailNew mid(%d) err(%v)", mid, err)
		err = nil
	}
	return res, err
}

// Times 剩余抽奖次数
func (s *Service) Times(ctx context.Context, mid int64, activity string) (res *l.TimesReply, err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	sid := cardsConfig.SID
	return s.lotterySvr.GetUnusedTimes(ctx, sid, mid)
}

// Cards 用户已经获得的卡及合成情况
func (s *Service) Cards(ctx context.Context, mid int64, activity string) (res *cardsmdl.CardsReplyNew, err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	var (
		cards []*cardsmdl.CardMid
	)
	res = &cardsmdl.CardsReplyNew{}
	res.Cards = make([]*cardsmdl.CardMidRes, 0)

	if cards, err = s.dao.GetMidsCardsNew(ctx, mid, cardsConfig.ID); err != nil {
		log.Errorc(ctx, "s.dao.GetMidCards err(%v)", err)
	}

	if len(cards) > 0 {
		var canCompose = true
		for _, v := range cards {
			remain := v.Nums - v.Used
			if remain <= 0 {
				remain = 0
				if v.CardID != cardsmdl.ComposeCardID {
					canCompose = false
				}
			}
			res.Cards = append(res.Cards, &cardsmdl.CardMidRes{
				CardID: v.CardID,
				Nums:   remain,
			})

		}
		res.CanCompose = canCompose
	}
	return res, nil
}

// giftIDtoCardID ...
func (s *Service) giftIDtoCardID(ctx context.Context, giftID int64, cards string) int64 {
	c := make(map[string]int64)
	if err := json.Unmarshal([]byte(cards), &c); err != nil {
		log.Errorc(ctx, "Task json.Unmarshal(%s) error(%v)", cards, err)
		return 0
	}
	for k, v := range c {
		if k == strconv.FormatInt(giftID, 10) {
			return v
		}
	}

	return 0
}

func (s *Service) checkCompose(c context.Context, cards []*cardsmdl.CardMid) (err error) {
	for _, v := range cards {
		if v.CardID != cardsmdl.ComposeCardID {
			if v.Nums-v.Used <= 0 {
				err = ecode.SpringFestivalComposeCardStoreErr
				return
			}

		}
	}
	return nil
}

// Compose 合成卡
func (s *Service) Compose(c context.Context, mid int64, risk *riskmdl.Base, activity string) (err error) {
	timestamp := time.Now().Unix()
	cardsConfig, err := s.cardsConfig(c, activity)
	if err != nil {
		log.Errorc(c, "s.dao.CardsConfig err(%v)", err)
		return
	}
	cards, err := s.dao.GetMidsCardsNew(c, mid, cardsConfig.ID)
	if err != nil {
		log.Errorc(c, "s.dao.GetMidsCardsNew err(%v)", err)
		return err
	}
	err = s.checkCompose(c, cards)
	if err != nil {
		log.Errorc(c, "s.checkCompose (%v)", err)
		return err
	}

	// 风控
	spRisk := &riskmdl.Task{
		Base:        *risk,
		Mid:         mid,
		ActivityUID: activity,
		Subscene:    riskmdl.ActionCompose,
	}
	_, err = s.risk(c, mid, spRisk, spRisk.EsTime)
	if err != nil {
		log.Errorc(c, "s.risk mid(%d) compose err(%v)", mid, err)
	}
	var (
		tx  *sql.Tx
		res []*cardsmdl.CardMid
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
	if res, err = s.dao.MidNumsForUpdateTxNew(c, tx, mid, cardsConfig.ID); err != nil {
		log.Errorc(c, "Add s.dao.MidNumsForUpdateTxNew(%d) failed. error(%v)", mid, err)
		return
	}
	// 使用卡
	err = s.composeUsedCard(c, res)
	if err != nil {
		return err
	}
	var affect int64
	affect, err = s.dao.UpdateCardNumsNew(c, tx, mid, cardsConfig.ID, res)
	if err != nil || affect == 0 {
		log.Errorc(c, "s.dao.UpdateCardNumsNew err(%v)", err)
		return ecode.SpringFestivalComposeCardErr
	}

	_, err = s.dao.InsertComposeLogTx(c, tx, mid, activity)
	if err != nil {
		log.Errorc(c, "s.dao.InsertComposeLogTx err(%v)", err)
		return ecode.SpringFestivalComposeCardErr
	}
	err = s.dao.DeleteMidCardDetailNew(c, mid, cardsConfig.ID)
	if err != nil {
		log.Errorc(c, "s.dao.DeleteMidCardDetailNew mid(%d) err(%v)", mid, err)
		err = nil
	}
	_ = s.cache.SyncDo(c, func(c context.Context) {
		err = s.composeFinishSend(c, mid, cardsConfig.ID, timestamp, activity)
		if err != nil {
			log.Errorc(c, "s.composeFinishSend(%d,%d,%s) err(%v)", cardsConfig.ID, timestamp, activity, err)
		}
		return
	})
	return
}

// composeFinishSend
func (s *Service) composeFinishSend(ctx context.Context, mid, activityID, timestamp int64, activity string) (err error) {
	midStr := fmt.Sprintf("%d", mid)
	data := cardsmdl.CardsComposeMessage{
		MID:        mid,
		ActivityID: activityID,
		Timestamp:  timestamp,
		Activity:   activity,
		Nums:       1,
	}
	b, _ := json.Marshal(data)
	err = component.CardsComposeProducer.Send(ctx, fmt.Sprintf(midStr), b)
	if err != nil {
		log.Errorc(ctx, "composeFinishSend sync failed:%v", err)
		return
	}
	log.Infoc(ctx, "composeFinishSend: component.CardsComposeProducer.Send(%d,%+v)", mid, data)
	return
}

// composeUsedCard 合成用卡
func (s *Service) composeUsedCard(c context.Context, cards []*cardsmdl.CardMid) (err error) {
	err = s.checkCompose(c, cards)
	if err != nil {
		log.Errorc(c, "s.checkCompose (%v)", err)
		return err
	}
	for i, v := range cards {
		index := i
		if v.CardID != cardsmdl.ComposeCardID {
			cards[index].Used++
		} else {
			cards[index].Nums++
		}
	}
	return nil
}

// GetCard ...
func (s *Service) GetCard(ctx context.Context, mid int64, token string, risk *riskmdl.Base, activity string) (err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	card, err := s.dao.ShareCardToken(ctx, token, activity)
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
	if card.IsReceived == cardsmdl.IsReceived {
		return ecode.SpringFestivalCardAlreadyErr
	}
	if card.Mid == mid {
		return ecode.SpringFestivalCantGetCardErr
	}
	var (
		tx          *sql.Tx
		senderCards []*cardsmdl.CardMid
		midCards    []*cardsmdl.CardMid
	)

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		senderCards, err = s.dao.GetMidsCardsNew(ctx, card.Mid, cardsConfig.ID)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards mid(%d) err(%v)", card.Mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		midCards, err = s.dao.GetMidsCardsNew(ctx, mid, cardsConfig.ID)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidCards mid(%d) err(%v)", mid, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		err = s.midInsertInit(ctx, mid, cardsConfig.ID, cardsConfig.CardsNum)
		if err != nil {
			log.Errorc(ctx, "s.midInsertInit mid(%d)", mid)
			return err
		}
		return err
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return
	}

	// 风控
	spRisk := &riskmdl.Task{
		Base:        *risk,
		Mid:         mid,
		TargetMid:   card.Mid,
		ActivityUID: activity,
		Subscene:    riskmdl.ActionSendCard,
	}
	riskReply, err := s.risk(ctx, mid, spRisk, spRisk.EsTime)
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
	if card.CardID > cardsConfig.CardsNum || card.CardID <= 0 {
		return ecode.SpringFestivalGetCardErr
	}
	cardMapNum := make(map[string]int64)
	cardName := fmt.Sprintf("%d", card.CardID)
	if cardName != "" {
		cardMapNum[cardName] = 1
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
	update, err := s.dao.UpdateCardNumsDescNewTx(ctx, tx, card.Mid, cardsConfig.ID, cardMapNum)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpdateCardNums mid(%d) err(%v)", card.Mid, err)
		return ecode.SpringFestivalGetCardErr
	}
	if update == 0 {
		return ecode.SpringFestivalAlreadyDonatedErr
	}
	// 更新被送卡人库存
	_, err = s.dao.UpdateCardNumsIncrNewTx(ctx, tx, mid, cardsConfig.ID, cardMapNum)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpdateCardNums mid(%d) err(%v)", card.Mid, err)
		return ecode.SpringFestivalGetCardErr
	}
	_, err = s.dao.InsertSendCardLogTx(ctx, tx, card.Mid, mid, card.CardID, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.InsertSendCardLogTx err(%v)", err)
		return ecode.SpringFestivalGetCardErr
	}
	err = s.dao.DeleteMidCardDetailNew(ctx, mid, cardsConfig.ID)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteMidCardDetailNew mid(%d) err(%v)", mid, err)
		err = nil
	}
	card.IsReceived = cardsmdl.IsReceived
	card.ReceiveMid = mid
	err = s.dao.SetShareCardToken(ctx, token, card, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.DeleteShareCardToken token(%s) err(%v) ", token, err)
	}
	err = s.taskSvr.CardsActSend(ctx, card.Mid, donateBusiness, activity, time.Now().Unix(), nil, isInternal)
	if err != nil {
		log.Errorc(ctx, "s.CardsActSend err(%v)", err)
	}
	return nil
}

// sendCard 合成用卡
func (s *Service) sendCard(c context.Context, sendCards []*cardsmdl.CardMid, midCards []*cardsmdl.CardMid, cardID int64) (err error) {
	// 检查库存
	err = s.checkCardStore(c, cardID, sendCards)
	if err != nil {
		return ecode.SpringFestivalAlreadyDonatedErr
	}
	for i, v := range sendCards {
		if v.CardID == cardID {
			sendCards[i].Used++
		}

	}
	for i, v := range midCards {
		if v.CardID == cardID {
			midCards[i].Nums++
		}
	}
	return nil
}

// CardTokenToMid 分享token转mid
func (s *Service) CardTokenToMid(ctx context.Context, token, activity string) (res *cardsmdl.CardTokenToMidReply, err error) {
	cardsConfig, err := s.cardsConfig(ctx, activity)
	if err != nil {
		log.Errorc(ctx, "s.dao.CardsConfig err(%v)", err)
		return
	}
	res = &cardsmdl.CardTokenToMidReply{}
	res.Account = &cardsmdl.Account{}
	res.Card = &cardsmdl.CardIsReceived{}
	var (
		cards []*cardsmdl.CardMid
	)
	card, err := s.dao.ShareCardToken(ctx, token, activity)
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
		cards, err = s.dao.GetMidsCardsNew(ctx, card.Mid, cardsConfig.ID)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetMidsCardsNew err(%v)", err)
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
		res.Card.IsInStock = cardsmdl.IsInStock
	}
	if err != nil {
		err = nil
	}

	res.Card.CardID = card.CardID
	res.Card.IsReceived = card.IsReceived
	res.Card.Mid = card.ReceiveMid
	return
}
