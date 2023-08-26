package s10

import (
	"context"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/admin/model/s10"

	"go-common/library/log"
	xtime "go-common/library/time"
)

func (s *Service) UserCostCacheFlush(ctx context.Context, mid int64) error {
	currDate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		return err
	}
	curr := xtime.Time(currDate)
	err = s.dao.DelExchangeStaticCache(ctx, mid)
	if err != nil {
		return err
	}
	err = s.dao.DelExchangeRoundStaticCache(ctx, mid, curr)
	if err != nil {
		return err
	}
	var res []*s10.CostRecord
	if s.subTabSwitch {
		res, err = s.dao.UserCostRecordSub(ctx, mid)
	} else {
		res, err = s.dao.UserCostRecord(ctx, mid)
	}

	if err != nil {
		return err
	}
	totalGidMap := make(map[int32]int32, 10)
	roundGidMap := make(map[int32]int32, 10)
	var cost int32
	for _, v := range res {
		totalGidMap[v.Gid] += 1
		cost += v.Cost
		if v.Ctime > curr {
			roundGidMap[v.Gid] += 1
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime > res[j].Ctime
	})
	if res == nil {
		res = make([]*s10.CostRecord, 0, 1)
	}
	err = s.dao.AddExchangeRoundStaticCache(ctx, mid, curr, roundGidMap)
	if err != nil {
		return err
	}
	err = s.dao.AddExchangeStaticCache(ctx, mid, totalGidMap)
	if err != nil {
		return err
	}
	err = s.dao.AddPointsCache(ctx, mid, 1, cost)
	if err != nil {
		return err
	}
	err = s.dao.AddPointDetailCache(ctx, mid, res)
	return err
}

func (s *Service) timeToDate(currentTime int64) (int64, error) {
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Error(" time.LoadLocation error(%v)", err)
		return 0, err
	}
	pubDateStr := time.Unix(currentTime, 0).Format("2006/01/02")
	curDate, err := time.ParseInLocation("2006/01/02", pubDateStr, location)
	if err != nil {
		log.Error("time.ParseInLocation error(%v)", err)
		return 0, err
	}
	return curDate.Unix(), nil
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

func (s *Service) lotteryGoods(ctx context.Context, mid int64) (map[int32]*s10.MatchUser, error) {
	robins := s.robins
	if len(robins) == 0 {
		return nil, nil
	}
	var lotteryInfo []int32
	var err error
	if s.subTabSwitch {
		lotteryInfo, err = s.dao.UserLotterySubInfo(ctx, mid, robins)
	} else {
		lotteryInfo, err = s.dao.UserLotteryInfo(ctx, mid, robins)
	}
	if err != nil {
		return nil, err
	}
	lotteryMap := make(map[int32]struct{}, len(lotteryInfo))
	for _, v := range lotteryInfo {
		lotteryMap[v] = struct{}{}
	}
	resMap, err := s.dao.UserLottery(ctx, mid)
	if err != nil {
		return nil, err
	}
	res := make(map[int32]*s10.MatchUser, len(lotteryInfo))
	for _, robin := range robins {
		tmpRobin := int32(robin)
		tmp, _ := resMap[tmpRobin]
		_, ok := lotteryMap[tmpRobin]
		robinLottery := &s10.MatchUser{IsLottery: ok, Lucky: tmp, IsRecieve: tmp != nil && tmp.State > 0}
		res[tmpRobin] = robinLottery
	}
	return res, nil
}

func (s *Service) DelGoodsStockByGid(ctx context.Context, gid int32) error {
	return s.dao.DelRestCountByGoodsCache(ctx, gid)
}
