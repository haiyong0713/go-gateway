package bwsonline

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"time"
)

func (s *Service) CurrencyFind(ctx context.Context, mid, bid int64) error {
	if check, err := s.likeDao.RsSetNX(ctx, fmt.Sprintf("currency_find_%d", mid), 1); err != nil || !check {
		log.Warnc(ctx, "CurrencyFind mid:%d to fast err:%v", mid, err)
		return ecode.ActivityRapid
	}
	today := todayDate()
	usedTimes, err := s.dao.UsedTimes(ctx, mid, today)
	if err != nil {
		log.Errorc(ctx, "CurrencyFind RawUsedTimes mid:%d date:%d error:%v", mid, today, err)
		return err
	}
	if _, ok := usedTimes[bwsonline.UsedTimeTypeAd]; ok {
		return ecode.BwsOnlineTimeUsed
	}
	if _, err = s.dao.AddUsedTimes(ctx, mid, bwsonline.UsedTimeTypeAd, today); err != nil {
		log.Errorc(ctx, "CurrencyFind AddUsedTimes mid:%d date:%d error:%v", mid, today, err)
		return err
	}
	if err = s.upUserCurrency(ctx, mid, bwsonline.CurrTypeEnergy, bwsonline.CurrAddTypeNormal, s.c.BwsOnline.FreeEnergy, bid); err != nil {
		log.Errorc(ctx, "CurrencyFind upUserCurrency mid:%d date:%d error:%v", mid, today, err)
		return err
	}
	if err = s.upUserCurrency(ctx, mid, bwsonline.CurrTypeCoin, bwsonline.CurrAddTypeNormal, s.c.BwsOnline.FreeCoin, bid); err != nil {
		log.Errorc(ctx, "CurrencyFind upUserCurrency mid:%d date:%d error:%v", mid, today, err)
		return err
	}
	return nil
}

func (s *Service) userCurrency(ctx context.Context, mid int64, bid int64) (map[int64]int64, error) {
	userCurrency, err := s.dao.UserCurrency(ctx, mid, bid)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int64)
	for i, v := range userCurrency {
		res[i] = v
	}
	nowEnergy := res[bwsonline.CurrTypeEnergy]
	if nowEnergy >= s.c.BwsOnline.MaxEnergy {
		return res, nil
	}
	func() {
		lastAutoTs, err := s.dao.LastAutoEnergy(ctx, mid, bid)
		if err != nil {
			log.Errorc(ctx, "userCurrency LastAutoEnergy mid:%d error:%v", mid, err)
			return
		}
		autoEnergy := func() int64 {
			var addEnergy int64
			if lastAutoTs <= 0 {
				return s.c.BwsOnline.FirstEnergy
			}
			if lastAutoTs > 0 {
				nowHour := hourInt(time.Now().Unix())
				lastHour := hourInt(lastAutoTs)
				if nowHour-lastHour-1 > 0 {
					addEnergy = nowHour - lastHour - 1
				}
			}
			return addEnergy
		}()
		if nowEnergy+autoEnergy > s.c.BwsOnline.MaxEnergy {
			autoEnergy = s.c.BwsOnline.MaxEnergy - nowEnergy
		}
		if res == nil {
			res = make(map[int64]int64)
		}
		res[bwsonline.CurrTypeEnergy] = nowEnergy + autoEnergy
		if lastAutoTs == 0 || autoEnergy > 0 {
			s.cache.Do(ctx, func(ctx context.Context) {
				if err = s.upUserCurrency(ctx, mid, bwsonline.CurrTypeEnergy, bwsonline.CurrAddTypeAuto, autoEnergy, bid); err != nil {
					log.Errorc(ctx, "userCurrency upUserCurrency auto mid:%d error:%v", mid, err)
					return
				}
				s.dao.DelCacheLastAutoEnergy(ctx, mid, bid)
			})
		}
	}()
	return res, nil
}

func (s *Service) upUserCurrency(ctx context.Context, mid, typ, addType, amount, bid int64) error {
	if _, err := s.dao.AddUserCurrencyLog(ctx, mid, typ, addType, amount, bid); err != nil {
		return err
	}
	if amount != 0 {
		affected, err := s.dao.UpUserCurrency(ctx, mid, typ, amount, bid)
		if err != nil {
			return err
		}
		if amount < 0 && affected == 0 {
			return ecode.BwsOnlineCoinLow
		}
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserCurrency(ctx, mid, bid)
	})
	return nil
}
