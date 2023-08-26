package s10

import (
	"context"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	s10dao "go-gateway/app/web-svr/activity/interface/dao/s10"
	"go-gateway/app/web-svr/activity/interface/model/s10"
	model "go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
	xtime "go-common/library/time"
)

func (s *Service) RedeliveryGift(ctx context.Context, gid, act int32, id, mid int64) error {
	gift := s.goodsInfo.Load().(map[int32]*s10.Bonus)[gid]
	if gift == nil {
		return ecode.ActivityGoodsNoExist
	}
	return s.exchangeAboutOtherBusiness(ctx, mid, id, act, gift, 0)
}

func (s *Service) PointFlush(ctx context.Context, mid int64) error {
	var (
		err  error
		cost int32
	)
	if s.splitTab {
		cost, err = s.dao.UserCountCostSub(ctx, mid)
	} else {
		cost, err = s.dao.UserCountCost(ctx, mid)
	}
	if err != nil {
		return err
	}
	return s.dao.AddPointsCache(ctx, mid, model.S10PointsCost, cost)
}

func (s *Service) UserLotteryInfoFlush(ctx context.Context, mid int64) error {
	userLottery, err := s.lotteryGoods(ctx, mid)
	if err != nil {
		return err
	}
	if len(userLottery) == 0 {
		userLottery = make(map[int32]*s10.MatchUser, 1)
		userLottery[s10.S10LotterySentinels] = new(s10.MatchUser)
	}
	return s.dao.AddLotteryCache(ctx, mid, userLottery)
}

func (s *Service) UserCostPointsDetailFlush(ctx context.Context, mid int64) error {
	var (
		err error
		res []*s10.CostRecord
	)
	if s.splitTab {
		res, err = s10dao.UserCostRecordSub(ctx, mid)
	} else {
		res, err = s10dao.UserCostRecord(ctx, mid)
	}
	if err != nil {
		return err
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime > res[j].Ctime
	})
	if res == nil {
		res = make([]*s10.CostRecord, 0, 1)
	}
	return s10dao.AddPointDetailCache(ctx, mid, res)
}

func (s *Service) UserCostStaticFlush(ctx context.Context, mid int64) error {
	currDate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		return err
	}
	return s.exchangeRebuild(ctx, mid, xtime.Time(currDate))
}

func (s *Service) DelGoodsStockByGid(ctx context.Context, gid int32) error {
	return s.dao.DelRestCountByGoodsCache(ctx, gid)
}

func (s *Service) DelRoundGoodsStockByGid(ctx context.Context, gid int32) error {
	currdate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s10 timeToDate error:%v", err)
	}
	return s.dao.DelRoundRestCountByGoodsCache(ctx, gid, xtime.Time(currdate))
}

func (s *Service) DelUserStatic(ctx context.Context, mid int64) error {
	return s.dao.DelExchangeStaticCache(ctx, mid)
}

func (s *Service) DelRoundUserStatic(ctx context.Context, mid int64) error {
	currdate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s10 timeToDate error:%v", err)
	}
	return s.dao.DelRoundExchangeStaticCache(ctx, mid, xtime.Time(currdate))
}
