package s10

import (
	"context"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/s10"

	xtime "go-common/library/time"
)

func (s *Service) pointCheck(ctx context.Context, mid int64, score int32) (bool, error) {
	total, err := s.dao.TotalPoints(ctx, mid, s.s10Act)
	if err != nil {
		return false, err
	}
	points, err := s.dao.PointsCache(ctx, mid)
	if err != nil {
		return false, err
	}
	cost, ok := points[model.S10PointsCost]
	if ok {
		if total-cost < score {
			return false, nil
		}
	}
	if s.splitTab {
		cost, err = s.dao.UserCountCostMasterSub(ctx, mid)
	} else {
		cost, err = s.dao.UserCountCostMaster(ctx, mid)
	}
	if err != nil {
		return false, err
	}
	if total-cost >= score {
		return true, nil
	}
	return false, nil
}

func (s *Service) lotteryTimesAboutGoods(ctx context.Context, mid int64, gid, count, roundCount int32, currTime xtime.Time) (bool, error) {
	countCache, exist1, err := s.dao.ExchangeFieldStaticCache(ctx, mid, gid)
	if err != nil {
		return false, err
	}
	roundcountCache, exist2, err := s.dao.ExchangeFieldRoundStaticCache(ctx, mid, gid, currTime)
	if err != nil {
		return false, err
	}
	if !exist1 && count != 0 || !exist2 && roundCount != 0 {
		cache.Do(context.Background(), func(ctx context.Context) {
			s.exchangeRebuild(ctx, mid, currTime)
		})
	}
	var res int32
	if count > 0 {
		if exist1 && countCache >= count {
			return false, nil
		}
		if exist2 && roundcountCache >= roundCount && roundCount > 0 {
			return false, nil
		}
		if s.splitTab {
			res, err = s.dao.UserCostRecordCountByGidSub(ctx, mid, gid, 0)
		} else {
			res, err = s.dao.UserCostRecordCountByGid(ctx, mid, gid, 0)
		}
		if err != nil {
			return false, err
		}
		if res >= count {
			return false, nil
		}
	}
	if roundCount > 0 {
		if exist2 && roundcountCache >= roundCount {
			return false, nil
		}
		if s.splitTab {
			res, err = s.dao.UserCostRecordCountByGidSub(ctx, mid, gid, currTime)
		} else {
			res, err = s.dao.UserCostRecordCountByGid(ctx, mid, gid, currTime)
		}
		if err != nil {
			return false, err
		}
		if res >= roundCount {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) checkPointAndTimes(ctx context.Context, mid int64, gid, score, count, roundCount int32, currTime xtime.Time) error {
	pass, err := s.pointCheck(ctx, mid, score)
	if err != nil {
		return ecode.ActivityExchangePointFail
	}
	if !pass {
		return ecode.ActivityInsufficient
	}
	pass, err = s.lotteryTimesAboutGoods(ctx, mid, gid, count, roundCount, currTime)
	if err != nil {
		return ecode.ActivityExchangePointFail
	}
	if !pass {
		if gid <= 10 {
			return ecode.ActivityMatchExchangedPoint
		}
		return ecode.ActivityGoodsOverTimes
	}
	return nil
}

func (s *Service) accountCheck(ctx context.Context, mid int64) error {
	accInfo, err := s.dao.Profile(ctx, mid)
	if err != nil {
		return nil
	}
	if accInfo.GetTelStatus() != 1 {
		return ecode.ActivityVogueTelValid
	}
	ok, err := s.dao.InBackList(ctx, mid, []string{"s10point"})
	if err != nil {
		return nil
	}
	if ok {
		return ecode.ActivityUerInBlackList
	}
	return nil
}

func (s *Service) s10GoodsTimePeriod(timestamp int64) error {
	s10TimePeriod := conf.LoadS10TimePeriodCfg()
	if s10TimePeriod.Goods == nil {
		return ecode.ActivityAppstoreEnd
	}
	if s10TimePeriod.Goods.Start > timestamp {
		return ecode.ActivityAppstoreNotStart
	}
	if s10TimePeriod.Goods.End < timestamp {
		return ecode.ActivityAppstoreEnd
	}
	return nil
}

func (s *Service) whiteCheck(mid int64) bool {
	if s.whiteSwitch {
		_, ok := s.whiteMap[mid]
		return ok
	}
	return true
}
