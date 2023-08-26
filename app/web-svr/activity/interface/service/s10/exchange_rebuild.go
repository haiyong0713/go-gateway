package s10

import (
	"context"
	"time"

	s10dao "go-gateway/app/web-svr/activity/interface/dao/s10"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
	xtime "go-common/library/time"
)

func (s *Service) timeToDate(currentTime int64) (int64, error) {
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Error("s10 time.LoadLocation error(%v)", err)
		return 0, err
	}
	pubDateStr := time.Unix(currentTime, 0).Format("2006/01/02")
	curDate, err := time.ParseInLocation("2006/01/02", pubDateStr, location)
	if err != nil {
		log.Error("s10 time.ParseInLocation error(%v)", err)
		return 0, err
	}
	return curDate.Unix(), nil
}

func (s *Service) exchangeRebuild(ctx context.Context, mid int64, curDate xtime.Time) error {
	totalGidMap, roundGidMap, err := s.userCostStatic(ctx, mid, curDate)
	if err != nil {
		return err
	}
	s.dao.AddExchangeRoundStaticCache(ctx, mid, curDate, roundGidMap)
	s.dao.AddExchangeStaticCache(ctx, mid, totalGidMap)
	return nil
}

func (s *Service) userCostStatic(ctx context.Context, mid int64, curDate xtime.Time) (totalGidMap, roundGidMap map[int32]int32, err error) {
	var res []*s10.CostRecord
	if s.splitTab {
		res, err = s10dao.UserCostRecordSub(ctx, mid)
	} else {
		res, err = s10dao.UserCostRecord(ctx, mid)
	}
	if err != nil {
		return nil, nil, err
	}
	totalGidMap = make(map[int32]int32, 10)
	roundGidMap = make(map[int32]int32, 10)
	for _, v := range res {
		totalGidMap[v.Gid] += 1
		if v.Ctime > curDate {
			roundGidMap[v.Gid] += 1
		}
	}
	return
}
