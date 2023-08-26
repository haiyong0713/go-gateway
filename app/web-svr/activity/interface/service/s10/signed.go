package s10

import (
	"context"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/s10"
)

func (s *Service) SignIn(ctx context.Context, mid int64) error {
	timestamp := time.Now().Unix()
	if err := s.s10PointsTimePeriod(timestamp); err != nil {
		return err
	}
	err := s.accountCheck(ctx, mid)
	if err != nil {
		return err
	}
	sign, err := s.Signed(ctx, mid, timestamp)
	if err != nil {
		return ecode.ActivitySignNotOpen
	}
	if sign > 0 {
		return ecode.ActivityAlreadySigned
	}
	if err = s.dao.TaskPubDataBus(ctx, mid, timestamp, "sign"); err != nil {
		return ecode.ActivitySignNotOpen
	}
	if err = s.dao.AddSignedCache(ctx, mid); err != nil {
		return ecode.ActivitySignNotOpen
	}
	return nil
}

func (s *Service) Signed(ctx context.Context, mid, timestamp int64) (int32, error) {
	sign, err := s.dao.SignedCache(ctx, mid)
	if err != nil || sign > 0 {
		return sign, err
	}
	return s.dao.GetCounterRes(ctx, mid, timestamp, s10.S10ActSign, s.s10Act)
}
