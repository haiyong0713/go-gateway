package cards

import (
	"context"
	"go-common/library/log"

	cards "go-gateway/app/web-svr/activity/interface/model/cards"
)

// CardsConfig 用户已经获得的卡
func (d *Dao) CardsConfig(c context.Context, activity string) (res *cards.Cards, err error) {
	cards, err := d.RawCardsConfig(c, activity)
	if err != nil {
		log.Errorc(c, "d.RawCardsConfig err(%v)", err)
	}
	if cards != nil && err == nil {
		return cards, nil
	}
	cache, err := d.GetCardsConfig(c, activity)
	if err != nil {
		log.Errorc(c, "d.GetCardsConfig(c, %s) err(%v)", activity, err)
		return nil, err
	}
	err = d.CacheCardsConfig(c, activity, cache)
	if err != nil {
		log.Errorc(c, " d.CacheCardsConfig err(%v)", err)
	}
	return res, nil
}

func (d *Dao) GetMidsCardsNew(c context.Context, mid, activityID int64) (res []*cards.CardMid, err error) {
	cards, err := d.MidCardDetailNew(c, mid, activityID)
	if err != nil {
		log.Errorc(c, "d.MidCardDetailNew err(%v)", err)
	}
	if cards != nil && err == nil {
		return cards, nil
	}
	giftMid, err := d.MidNumsNew(c, mid, activityID)
	if err != nil {
		log.Errorc(c, "d.MidNums(c, %d) err(%v)", mid, err)
		return nil, err
	}
	err = d.AddMidCardDetailNew(c, mid, activityID, giftMid)
	if err != nil {
		log.Errorc(c, " d.AddMidCardDetail err(%v)", err)
	}
	return giftMid, nil
}

// GetMidCards 用户已经获得的卡
func (d *Dao) GetMidCards(c context.Context, mid int64, activity string) (res *cards.MidNums, err error) {
	cards, err := d.MidCardDetail(c, mid, activity)
	if err != nil {
		log.Errorc(c, "d.MidCardDetail err(%v)", err)
	}
	if cards != nil && err == nil {
		return cards, nil
	}
	giftMid, err := d.MidNums(c, mid)
	if err != nil {
		log.Errorc(c, "d.MidNums(c, %d) err(%v)", mid, err)
		return nil, err
	}
	midUsed, err := d.MidComposeUsed(c, mid)
	if err != nil {
		log.Errorc(c, "d.MidComposeUsed(c, %d) err(%v)", mid, err)
		return nil, err
	}
	if midUsed != nil {
		giftMid.Compose = giftMid.Compose - midUsed.ComposeUsed
		if giftMid.Compose < 0 {
			giftMid.Compose = 0
		}
	}
	err = d.AddMidCardDetail(c, mid, activity, giftMid)
	if err != nil {
		log.Errorc(c, " d.AddMidCardDetail err(%v)", err)
	}
	return giftMid, nil
}
