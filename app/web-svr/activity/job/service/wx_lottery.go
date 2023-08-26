package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/dao/pay"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
)

const (
	_wxLotteryID = "wx_lottery_01"
	_timeSleep   = 100 * time.Millisecond
	_retry       = 3
	_pageLimit   = 50
)

// wxLotteryLogHandle .
func (s *Service) wxLotteryLogHandle(msg *match.Message) error {
	return nil
}

func (s *Service) wxLotteryLogStatus(ctx context.Context, t time.Time, v *like.WxLotteryLog) error {
	var (
		err   error
		reply *pay.ResultInner
	)
	orderID := s.md5(fmt.Sprintf("%d_%s_%s_%d", v.Mid, _wxLotteryID, v.LotteryID, v.ID))
	var payOrder *like.PayOrder
	if payOrder, err = s.pay.PayOrderByOrderID(ctx, orderID); err != nil {
		return err
	}
	if payOrder == nil || payOrder.OrderID != orderID {
		s.retry(context.Background(), func() error {
			_, err = s.pay.InsertPayOrder(ctx, v.Mid, s.c.WxLottery.TransDesc, v.GiftMoney, orderID)
			return err
		})
	}
	if (payOrder != nil) && (payOrder.OrderStatus > like.OrderStatusFail) {
		log.Warn("wxLotteryLogStatus payOrder%+v", payOrder)
		return nil
	}
	reply, err = s.payTransferInner(ctx, v.Mid, v.GiftMoney, orderID, t, s.wxLottery)
	if err != nil {
		return err
	}
	var orderStatus = reply.OrderStatus()
	s.pay.UpdateWxLotteryLogPayOrderID(ctx, orderID, reply.Data.TransOrderNo, t.Unix(), orderStatus, v.Mid, v.ID)
	s.pay.UpdatePayOrder(ctx, reply.Data.TransOrderNo, t.Unix(), orderStatus, v.Mid, orderID)
	return nil
}

func (s *Service) wxLotteryLogStatusCancel(ctx context.Context, t time.Time, v *like.WxLotteryLog) (bool, error) {
	var (
		err             error
		payStatus       int32
		payStatusCancel bool
		reply           *pay.ResultInner
	)
	switch v.PayStatus {
	case like.PayStatusRisk:
		payStatus = like.PayStatusRiskSuccess
		payStatusCancel = true
	case like.PayStatusCancel:
		payStatus = like.PayStatusCancelSuccess
		payStatusCancel = true
	case like.PayStatusRiskSuccess:
		return true, nil
	case like.PayStatusCancelSuccess:
		return true, nil
	}
	if v.PayOrderID == "" {
		log.Warn("wxLotteryLogStatusCancel 无需撤回 upLottery:%+v", v)
		return payStatusCancel, nil
	}
	if payStatusCancel {
		reply, err = s.profitCancelInner(ctx, v.Mid, v.PayOrderID, t, s.wxLottery)
		if err != nil {
			return payStatusCancel, err
		}
		if reply.Code != 0 {
			return payStatusCancel, err
		}
		s.pay.UpdateWxLotteryLogPayStatus(ctx, payStatus, v.Mid, v.ID)
		s.pay.UpdatePayOrderPayStatus(ctx, payStatus, v.Mid, v.OrderID, v.PayOrderID)
		return payStatusCancel, err
	}
	return payStatusCancel, nil
}

// wxLotteryLogPage  logical complement
func (s *Service) wxLotteryLogPage() {
	return
}

func (s *Service) md5(source string) string {
	md5Str := md5.New()
	md5Str.Write([]byte(source))
	return hex.EncodeToString(md5Str.Sum(nil))
}

func (s *Service) retry(ctx context.Context, f func() error) {
	for i := 0; i < _retry; i++ {
		err := f()
		if err == nil {
			return
		}
		log.Error("retry info:%+v", err)
		time.Sleep(_timeSleep)
	}
}

// PayTransferInner 活动红包转入 .
func (s *Service) payTransferInner(ctx context.Context, mid, money int64, orderID string, t time.Time, info *like.TransInfo) (reply *pay.ResultInner, err error) {
	millisecond := t.UnixNano() / 1000 / 1000
	pt := &pay.PayTransferInner{
		TraceID:      strconv.FormatInt(millisecond, 10),
		UID:          mid,
		OrderNo:      orderID,                                         // 业务方转入红包的订单id（通过该字段保持幂等）
		TransBalance: money,                                           // 分转成元
		TransDesc:    info.TransDesc,                                  // 红包名称
		StartTme:     millisecond + info.WithdrawStartHour*60*60*1000, // 红包解冻时间
		Timestamp:    millisecond,                                     // 当前时间毫秒值
	}
	s.retry(ctx, func() error {
		reply, err = s.pay.PayTransferInner(ctx, pt)
		if pay.OrderStatusFail == reply.OrderStatus() {
			err = fmt.Errorf("retry activityUID:%d,payTransferInner:%+v ", mid, pt)
		}
		return err
	})
	return
}

// ProfitCancel 活动红包撤回 .
func (s *Service) profitCancelInner(ctx context.Context, mid int64, transOrderNo string, t time.Time, info *like.TransInfo) (reply *pay.ResultInner, err error) {
	millisecond := t.UnixNano() / 1000 / 1000
	pc := &pay.ProfitCancelInner{
		TraceID:          strconv.FormatInt(millisecond, 10),
		UID:              mid,
		TransOrderNo:     transOrderNo,
		TransDesc:        info.TransDesc, //  红包名称
		Timestamp:        millisecond,
		FullAmountCancel: 0, // 是否需要全量追回。不传则默认从原账户扣除。 0：从原活动账户扣除  1：若原活动账户过期或者余额不足，则从同一个提现的其他活动账户中扣除
	}
	s.retry(ctx, func() error {
		reply, err = s.pay.ProfitCancelInner(ctx, pc)
		if pay.OrderStatusFail == reply.OrderStatus() {
			err = fmt.Errorf("retry profitCancelInner:%+v ", pc)
		}
		return err
	})
	return
}
