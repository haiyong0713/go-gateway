package s10

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"
	s10dao "go-gateway/app/web-svr/activity/interface/dao/s10"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
)

func UserCostRecord(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	res, err := s10dao.PointDetailCache(ctx, mid)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}
	if subTable {
		res, err = s10dao.UserCostRecordSub(ctx, mid)
	} else {
		res, err = s10dao.UserCostRecord(ctx, mid)
	}

	if err != nil {
		return nil, err
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime > res[j].Ctime
	})
	if err = cache.Do(context.Background(), func(ctx context.Context) {
		tmpres := res
		if tmpres == nil {
			tmpres = make([]*s10.CostRecord, 0, 1)
		}
		s10dao.AddPointDetailCache(ctx, mid, tmpres)
	}); err != nil {
		log.Errorc(ctx, "cache.Do() error(%v)", err)
	}
	return res, nil
}

func (s *Service) OtherActivity(ctx context.Context) ([]*s10.Other, error) {
	matchCfg := conf.LoadS10MatchesCfg()
	return matchCfg.OtherActivity, nil
}

func (s *Service) ActGoods(ctx context.Context, mid int64) (*s10.ActGoods, error) {
	if !s.whiteCheck(mid) {
		return new(s10.ActGoods), nil
	}
	rep := new(s10.ActGoods)
	rep.Currtime = time.Now().Unix()
	matchCfg := conf.LoadS10MatchesCfg()
	rep.Other = matchCfg.OtherActivity
	bonuses := s.bonuses.Load().(map[int32][]*s10.Bonus)
	rep.Bonuses = bonuses[0]
	if mid <= 0 {
		return rep, nil
	}
	staticMap, err := s.dao.ExchangeStaticCache(ctx, mid)
	if err != nil {
		return rep, nil
	}
	currdate, err := s.timeToDate(rep.Currtime)
	if err != nil {
		return rep, nil
	}
	roundStaticMap, err := s.dao.ExchangeRoundStaticCache(ctx, mid, xtime.Time(currdate))
	if err != nil {
		return rep, nil
	}
	if staticMap == nil || roundStaticMap == nil {
		staticMap, roundStaticMap, err = s.userCostStatic(ctx, mid, xtime.Time(currdate))
		if err != nil {
			return rep, nil
		}
		cache.Do(context.Background(), func(ctx context.Context) {
			s.dao.AddExchangeRoundStaticCache(ctx, mid, xtime.Time(currdate), roundStaticMap)
			s.dao.AddExchangeStaticCache(ctx, mid, staticMap)
		})
	}
	res := make([]*s10.Bonus, 0, len(rep.Bonuses))
	for _, v := range rep.Bonuses {
		tmp := new(s10.Bonus)
		*tmp = *v
		res = append(res, tmp)
		if tmp.IsHaust != 0 {
			continue
		}
		if !tmp.IsRoundInfinite && currdate == tmp.CurrDate {
			if tmp.RoundSend >= tmp.RoundStock {
				tmp.IsHaust = 2
				continue
			}
			if tmp.LeftTimes > tmp.RoundStock-tmp.RoundSend {
				tmp.LeftTimes = tmp.RoundStock - tmp.RoundSend
			}
		}
		value := staticMap[tmp.ID]
		if tmp.ExchangeTimes != 0 {
			if value >= tmp.ExchangeTimes {
				tmp.IsHaust = 3
				continue
			}
			if tmp.LeftTimes > tmp.ExchangeTimes-value {
				tmp.LeftTimes = tmp.ExchangeTimes - value
			}
		}
		value = roundStaticMap[tmp.ID]
		if tmp.RoundExchangeTimes != 0 {
			if value >= tmp.RoundExchangeTimes {
				tmp.IsHaust = 4
				continue
			}
			if tmp.LeftTimes > tmp.RoundExchangeTimes-value {
				tmp.LeftTimes = tmp.RoundExchangeTimes - value
			}
		}
	}
	rep.Bonuses = res
	return rep, nil
}

func (s *Service) UpdateUserLotteryState(ctx context.Context, robin int32, mid int64, number, name, addr string) error {
	if !s.whiteCheck(mid) {
		return nil
	}
	currTime := time.Now().Unix()
	if err := s.s10GoodsTimePeriod(currTime); err != nil {
		return err
	}
	matchCfg := conf.LoadS10MatchesCfg()
	flag := false
	for _, v := range matchCfg.Matches {
		if v.Robin == robin && v.LotteryExpire > currTime {
			if currTime <= v.Lottery {
				return ecode.ActivityLotteryNotStart
			}
			flag = true
			break
		}
	}
	if !flag {
		return ecode.ActivityGoodsExpired
	}
	res, err := s.dao.LotteryCache(ctx, mid)
	if err != nil {
		return ecode.ActivityLotteryGiftGetFail
	}
	if res != nil {
		if v, ok := res[robin]; ok {
			if v.Lucky == nil {
				return ecode.ActivityLotteryNotLucky
			}
			if v.Lucky.State > 0 {
				return ecode.ActivityLotteryGiftReceived
			}
		}
	}
	luckyGoods, err := s.dao.UserLotteryByRobin(ctx, mid, robin)
	if err != nil {
		return ecode.ActivityLotteryGiftGetFail
	}
	if luckyGoods == nil {
		return ecode.ActivityLotteryNotLucky
	}
	if luckyGoods.State > 0 {
		return ecode.ActivityLotteryGiftReceived
	}
	bonuses := s.goodsInfo.Load().(map[int32]*s10.Bonus)
	gift := bonuses[luckyGoods.Gid]
	if gift == nil {
		log.Errorc(ctx, "s10 UpdateUserLooteryState gift nil")
		return nil
	}
	_, err = s.dao.UpdateUserLotteryState(ctx, robin, gift.Type, mid, number, name, addr)
	if err != nil {
		return ecode.ActivityLotteryGiftGetFail
	}
	if gift.Type == 0 || gift.Type == 1 {
		return nil
	}
	err = s.exchangeAboutOtherBusiness(ctx, mid, int64(robin), 1, gift, 0)
	if err == nil {
		s.dao.AckUserCostGift(ctx, mid, robin)
	}
	if err != nil {
		err = ecode.ActivityCallServiceFail
	}
	return err
}

func (s *Service) StageLottery(ctx context.Context, robin int32, mid int64) error {
	if !s.whiteCheck(mid) {
		return nil
	}
	err := s.accountCheck(ctx, mid)
	if err != nil {
		return err
	}
	matchCfg := conf.LoadS10MatchesCfg()
	var match *s10.Match
	for _, v := range matchCfg.Matches {
		if v.Robin == robin {
			match = v
			break
		}
	}
	if match == nil {
		return xecode.RequestErr
	}
	currentTime := time.Now()
	timeInt64 := currentTime.Unix()
	if err = s.s10GoodsTimePeriod(timeInt64); err != nil {
		return err
	}
	if match.Start >= timeInt64 {
		return ecode.ActivityMatchStageNotStart
	}
	if match.End <= timeInt64 {
		return ecode.ActivityMatchStageEnd
	}
	err = s.dao.UserLock(ctx, mid)
	if err != nil {
		return ecode.ActivityExchangePointFail
	}
	defer s.dao.UserUnlock(ctx, mid)
	err = s.checkPointAndTimes(ctx, mid, robin, match.Points, 1, 0, 0)
	if err != nil {
		return err
	}
	if err = s.dao.DelPointsFieldCache(ctx, mid, 1); err != nil {
		return ecode.ActivityExchangePointFail
	}
	if err = s.dao.DelLotteryFieldCache(ctx, mid, robin); err != nil {
		return ecode.ActivityExchangePointFail
	}
	if s.splitTab {
		if _, err = s.dao.AddUserCostRecordToSubTab(ctx, mid, robin, match.Points, fmt.Sprintf("%s抽奖", match.Title), xtime.Time(timeInt64)); err != nil {
			err = ecode.ActivityExchangePointFail
		}
	} else {
		if _, err = s.dao.AddUserCostRecord(ctx, mid, robin, match.Points, fmt.Sprintf("%s抽奖", match.Title), xtime.Time(timeInt64)); err != nil {
			err = ecode.ActivityExchangePointFail
		}
	}
	s10dao.DelPointDetailCache(ctx, mid)
	return err
}

func (s *Service) ExchangeGoods(ctx context.Context, mid int64, gid int32) error {
	if !s.whiteCheck(mid) {
		return nil
	}
	currentTime := time.Now().Unix()
	if err := s.s10GoodsTimePeriod(currentTime); err != nil {
		return err
	}
	err := s.accountCheck(ctx, mid)
	if err != nil {
		return err
	}
	goods := s.goodsInfo.Load().(map[int32]*s10.Bonus)
	gift := goods[gid]
	if gift == nil || gift.Robin != 0 {
		return xecode.RequestErr
	}

	if int64(gift.Start) >= currentTime {
		return ecode.ActivityGoodsNotStart
	}
	if int64(gift.End) <= currentTime {
		return ecode.ActivityGoodsEnd
	}
	currDate, err := s.timeToDate(currentTime)
	if err != nil {
		return err
	}
	if gift.IsHaust != 0 {
		return ecode.ActivityGoodsNoStoreErr
	}
	if !gift.IsRoundInfinite && currDate == gift.CurrDate {
		if gift.RoundSend >= gift.RoundStock {
			return ecode.ActivityGoodsNoStoreErr
		}
	}
	err = s.dao.UserLock(ctx, mid)
	if err != nil {
		return ecode.ActivityExchangePointFail
	}
	defer s.dao.UserUnlock(ctx, mid)
	err = s.checkPointAndTimes(ctx, mid, gid, gift.Score, gift.ExchangeTimes, gift.RoundExchangeTimes, xtime.Time(currDate))
	if err != nil {
		return err
	}
	if err = s.dao.DelPointsFieldCache(ctx, mid, 1); err != nil {
		return ecode.ActivityExchangePointFail
	}
	if !gift.IsInfinite {
		if gift.IsHaust != 0 {
			return ecode.ActivityGoodsNoStoreErr
		}
		exist, err := s.dao.IncrRestCountByGoodsCache(ctx, gid)
		if err != nil {
			return ecode.ActivityExchangePointFail
		}
		if !exist {
			if err = s.goodsRestCountRebuild(ctx, gid); err != nil {
				return ecode.ActivityExchangePointFail
			}
		}
	}

	if gift.IsRound && !gift.IsRoundInfinite {
		exist, err := s.dao.IncrRoundRestCountByGoodsCache(ctx, gid, xtime.Time(currDate))
		if err != nil {
			//s.correctAllGoodsStock(ctx, gift, 0, mid, "decr_goods", xtime.Time(currDate))
			return ecode.ActivityExchangePointFail
		}
		if !exist {
			if err = s.goodsRoundRestCountRebuild(ctx, gid, xtime.Time(currDate)); err != nil {
				//s.correctAllGoodsStock(ctx, gift, 0, mid, "decr_goods", xtime.Time(currDate))
				return ecode.ActivityExchangePointFail
			}
		}
	}
	err = s.exchangeGoodsWrite(ctx, mid, currentTime, gift)
	return err
}

func (s *Service) exchangeGoodsWrite(ctx context.Context, mid int64, currTime int64, gift *s10.Bonus) error {
	var effect int64
	currDate, err := s.timeToDate(currTime)
	if err != nil {
		return err
	}
	if !gift.IsInfinite {
		effect, err = s.dao.UpdateGoodsSendCount(ctx, gift.ID)
		if err != nil {
			//s.correctAllGoodsStock(ctx, gift, 0, mid, "decr_round_goods", xtime.Time(currDate))
			return ecode.ActivityExchangePointFail
		}
		if effect == 0 {
			return ecode.ActivityGoodsNoStoreErr
		}
	}
	if gift.IsRound && !gift.IsRoundInfinite {
		effect, err = s.dao.UpdateGoodsRoundSendCount(ctx, gift.ID, xtime.Time(currDate))
		if err != nil {
			//s.correctAllGoodsStock(ctx, gift, 0, mid, "act_goods", xtime.Time(currDate))
			return ecode.ActivityExchangePointFail
		}
		if effect == 0 {
			return ecode.ActivityGoodsNoStoreErr
		}
	}
	if s.splitTab {
		effect, err = s.dao.AddUserCostRecordToSubTab(ctx, mid, gift.ID, gift.Score, gift.Name, xtime.Time(currTime))
	} else {
		effect, err = s.dao.AddUserCostRecord(ctx, mid, gift.ID, gift.Score, gift.Name, xtime.Time(currTime))

	}
	if err != nil {
		//s.correctAllGoodsStock(ctx, gift, 0, mid, "act_round_goods", xtime.Time(currDate))
		return ecode.ActivityExchangePointFail
	}
	err = s.exchangeAboutOtherBusiness(ctx, mid, effect, 0, gift, xtime.Time(currDate))
	if err == nil {
		if s.splitTab {
			s.dao.AckUserCostRecordSub(ctx, effect, mid)
		} else {
			s.dao.AckUserCostRecord(ctx, effect)
		}
	}
	if err != nil {
		err = ecode.ActivityCallServiceFail
	}
	return err
}

func (s *Service) exchangeAboutOtherBusiness(ctx context.Context, mid, id int64, act int32, gift *s10.Bonus, currdate xtime.Time) (err error) {
	uniqueID := fmt.Sprintf("s10:%d:%d:%d", act, mid, id)
	switch gift.Type {
	case 2:
		// 漫画
		return s.dao.CartoonDiscount(ctx, mid, gift.Extra, uniqueID)
	case 3:
		// 会员购
		return s.dao.MallCoupon(ctx, mid, gift.Extra, uniqueID)
	case 4:
		// 大会员
		return s.dao.MemberCoupon(ctx, mid, gift.Extra, uniqueID)
	case 5:
		// 直播用户头衔
		err = s.dao.LiveUserTitlePub(ctx, mid, uniqueID, gift.Extra)
	case 6:
		// 直播弹幕颜色
		err = s.dao.LiverBulletPub(ctx, mid, uniqueID, gift.Extra)
	case 7:
		// 战队头像框
		return s.dao.GrantByBiz(ctx, mid, uniqueID, gift.Extra)
	case 8:
		// 战队装扮
		return s.dao.GrantSuit(ctx, mid, uniqueID, gift.Extra)
	}
	return
}
