package s10

import (
	"context"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
)

func (s *Service) s10PointsTimePeriod(timestamp int64) error {
	s10TimePeriod := conf.LoadS10TimePeriodCfg()
	if s10TimePeriod.Points == nil {
		return ecode.ActivityAppstoreEnd
	}
	if s10TimePeriod.Points.Start > timestamp {
		return ecode.ActivityAppstoreNotStart
	}
	if s10TimePeriod.Points.End < timestamp {
		return ecode.ActivityAppstoreEnd
	}
	return nil
}

func (s *Service) Points(ctx context.Context, mid int64) (res model.Points, err error) {
	res.Total, res.Rest, err = s.RestPoint(ctx, mid)
	return
}

func (s *Service) RestPoint(ctx context.Context, mid int64) (int32, int32, error) {
	pointsMap, err := s.dao.PointsCache(ctx, mid)
	if err != nil {
		return 0, 0, ecode.ActivityPointGetFail
	}
	total, err1 := s.dao.TotalPoints(ctx, mid, s.s10Act)
	cost, ok := pointsMap[model.S10PointsCost]
	if !ok {
		if s.splitTab {
			if cost, err = s.dao.UserCountCostSub(ctx, mid); err != nil {
				return 0, 0, ecode.ActivityPointGetFail
			}
		} else {
			if cost, err = s.dao.UserCountCost(ctx, mid); err != nil {
				return 0, 0, ecode.ActivityPointGetFail
			}
		}

	}
	if !ok || (err1 == nil && pointsMap[model.S10PointsTotal] != total) {
		if err1 != nil {
			total = pointsMap[model.S10PointsTotal]
		}
		if err = cache.Do(context.Background(), func(ctx context.Context) {
			s.dao.AddAllPointsCache(ctx, mid, total, cost)
		}); err != nil {
			log.Errorc(ctx, "s10 s.cache.Do() error(%v)", err)
		}
	}
	if total == 0 {
		total = pointsMap[model.S10PointsTotal]
	}
	res := total - cost
	if res < 0 {
		res = 0
	}
	return total, res, nil
}
