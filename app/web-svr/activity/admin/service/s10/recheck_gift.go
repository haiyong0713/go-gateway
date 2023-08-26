package s10

import (
	"context"
	"time"

	"go-gateway/app/web-svr/activity/admin/model/s10"

	xtime "go-common/library/time"
)

func (s *Service) AckCostInfo(ctx context.Context, id, mid int64) error {
	var err error
	if s.subTabSwitch {
		_, err = s.dao.AckUserCostActSub(ctx, mid, id)

	} else {
		_, err = s.dao.AckUserCostAct(ctx, mid, id)
	}
	return err
}

func (s *Service) UpdateUserCostState(ctx context.Context, id, mid int64) error {
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
	if s.subTabSwitch {
		_, err = s.dao.UpdateUserCostStateSub(ctx, mid, id)

	} else {
		_, err = s.dao.UpdateUserCostState(ctx, mid, id)
	}
	return err
}

func (s *Service) AckGiftInfo(ctx context.Context, id, mid int64) error {
	_, err := s.dao.AckUserGiftAct(ctx, id)
	return err
}
func (s *Service) RedeliveryCostInfo(ctx context.Context, id, mid int64) error {
	var (
		err      error
		userInfo *s10.UserCostRecord
	)
	if s.subTabSwitch {
		userInfo, err = s.dao.UserCostRecordByIDSub(ctx, id, mid)
	} else {
		userInfo, err = s.dao.UserCostRecordByID(ctx, id, mid)
	}

	if err != nil {
		return err
	}
	if userInfo.Ack > 0 || userInfo.State == 1 {
		return nil
	}
	err = s.dao.Redelivery(ctx, id, mid, int64(userInfo.Gid), 0)
	if err != nil {
		return err
	}
	if s.subTabSwitch {
		_, err = s.dao.AckUserCostActSub(ctx, mid, id)

	} else {
		_, err = s.dao.AckUserCostAct(ctx, mid, id)
	}
	return err
}

func (s *Service) RedeliveryGiftInfo(ctx context.Context, id, mid int64) error {
	userInfo, err := s.dao.UserGiftByID(ctx, id)
	if err != nil {
		return err
	}
	if userInfo.Ack > 0 || userInfo.State <= 1 {
		return nil
	}
	err = s.dao.Redelivery(ctx, id, mid, int64(userInfo.Gid), 1)
	if err != nil {
		return err
	}
	_, err = s.dao.AckUserGiftAct(ctx, id)
	return err
}
