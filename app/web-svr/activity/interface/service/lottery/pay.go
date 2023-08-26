package lottery

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-gateway/app/web-svr/activity/interface/model/pay"
)

const (
	retryTimes = 3
	timeSleep  = 100 * time.Millisecond
)

// PayTransferInner 活动红包转入 .
func (s *Service) payTransferInner(c context.Context, mid, money int64, orderID string, transDesc string, t time.Time, withdrawStartHour int64, activityID string) (reply *pay.ResultInner, err error) {
	millisecond := t.UnixNano() / 1000 / 1000
	pt := &pay.TransferInner{
		TraceID:      strconv.FormatInt(millisecond, 10),
		UID:          mid,
		OrderNo:      orderID,                                    // 业务方转入红包的订单id（通过该字段保持幂等）
		TransBalance: money,                                      // 分转成元
		TransDesc:    transDesc,                                  // 红包名称
		StartTme:     millisecond + withdrawStartHour*60*60*1000, // 红包解冻时间
		Timestamp:    millisecond,
		CustomerID:   s.c.Lottery.Pay.CustomerID,
		ActivityID:   activityID,
		Token:        s.c.Lottery.Pay.Token,
	}
	reply, err = s.pay.PayTransferInner(c, pt)
	if pay.OrderStatusFail == reply.OrderStatus() {
		err = fmt.Errorf("retry activityUID:%d,payTransferInner:%+v ", mid, pt)
	}
	return
}
