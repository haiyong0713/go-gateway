package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/job/model/match"
	"go-gateway/app/web-svr/activity/job/model/s10"

	"go-common/library/log"
	"go-common/library/stat/prom"
	xtime "go-common/library/time"
)

func (s *Service) userLotteryRecord(m *match.Message) {
	var (
		err        error
		newLottery *s10.UserLotteryRecord
		ctx        = context.Background()
	)
	if newLottery, err = s.parseNewUserLotteryRecord(m); err != nil {
		return
	}
	if newLottery.State <= 1 {
		return
	}
	tmp := &s10.MatchUser{IsLottery: true}
	tmp.IsRecieve = newLottery.State > 0
	tmp.Lucky = &s10.Lucky{Gid: newLottery.Gid, State: newLottery.State, Extra: newLottery.Extra}
	err = s.s10Dao.AddLotteryCache(ctx, newLottery.Mid, newLottery.Robin, tmp)
	if err != nil {
		str := fmt.Sprintf("S10Lottery mid:%d,robin:%d", newLottery.Mid, newLottery.Robin)
		log.Errorc(ctx, str)
		prom.BusinessErrCount.Incr("S10Lottery")
	}
}

func (s *Service) userCostRecord(m *match.Message) {
	var (
		err     error
		cost    int32
		newCost *s10.UserCostRecord
		res     []*s10.CostRecord
		ctx     = context.Background()
	)
	if newCost, err = s.parseNewUserCostRecord(m); err != nil {
		return
	}
	if m.Action == "insert" {
		if newCost.Gid < 10 {
			s.s10Dao.AddLotteryUser(ctx, newCost.Mid, newCost.Gid)
			s.s10Dao.AddLotteryCache(ctx, newCost.Mid, newCost.Gid, &s10.MatchUser{IsLottery: true})
		}
	}
	currDate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		return
	}
	curr := xtime.Time(currDate)
	if s.s10General != nil && s.s10General.Switch {
		res, err = s.s10Dao.UserCostRecordSubTab(ctx, newCost.Mid)
	} else {
		res, err = s.s10Dao.UserCostRecord(ctx, newCost.Mid)
	}

	if err != nil {
		return
	}
	totalGidMap := make(map[int32]int32, 10)
	roundGidMap := make(map[int32]int32, 10)
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
	s.s10Dao.AddExchangeRoundStaticCache(ctx, newCost.Mid, curr, roundGidMap)
	s.s10Dao.AddExchangeStaticCache(ctx, newCost.Mid, totalGidMap)
	s.s10Dao.AddPointsCache(ctx, newCost.Mid, 1, cost)
	s.s10Dao.AddPointDetailCache(ctx, newCost.Mid, res)
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
