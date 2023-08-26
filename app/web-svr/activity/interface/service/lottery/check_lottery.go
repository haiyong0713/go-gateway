package lottery

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"

	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/pay"
)

// PreCheck 前置校验接口
type PreCheck interface {
	check(c context.Context, s *Service, mid int64) error
	checkError(c context.Context, err error) error
}

// PayLottery ...
type PayLottery struct {
}

func (p *PayLottery) check(c context.Context, s *Service, mid int64) error {
	res, err := s.cheeseClient.HistoryPaid(c, &cheeseapi.HistoryPaidReq{Mid: mid})
	if err != nil || res == nil {
		log.Errorc(c, "s.cheeseClient.HistoryPaid err(%v) res(%v)", err, res)
		return ecode.ActivityLotteryNetWorkError
	}
	if res.IsPaid {
		return nil
	}
	return ecode.ActivityLotteryNoPayError
}

func (p *PayLottery) checkError(c context.Context, err error) error {
	if err != nil && xecode.EqualError(ecode.ActivityNoTimes, err) {
		return ecode.ActivityLotteryPayJoinedError
	}
	return err
}

// PreCheck ...
func (s *Service) PreCheck(c context.Context, sid string) (PreCheck, error) {
	switch sid {
	case s.c.Lottery.PayOneYear: // 付费一周年
		return &PayLottery{}, nil
	}
	return nil, nil

}
