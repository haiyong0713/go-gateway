package like

import (
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/job/model"
)

const (
	GiftTypeMoney               = 4
	OrderStatusPending          = 0
	OrderStatusFail             = 1 // TODO 轮训重试请求支付账户
	OrderStatusAccountNormal    = 2
	OrderStatusAccountFrozen    = 3
	OrderStatusAccountNotEnough = 4
	// pay status
	PayStatusNormal        = 0 // 默认状态
	PayStatusRisk          = 1 // 风控撤回中。。。
	PayStatusCancel        = 2 // 后台撤回中。。。
	PayStatusRiskSuccess   = 3 // 风控撤回成功
	PayStatusCancelSuccess = 4 // 后台撤回成功
)

type TransInfo struct {
	TransDesc       string
	WithdrawEndTime int64
	// WithdrawStartHour 多少小时后可以提现
	WithdrawStartHour int64
}

type WxLotteryLog struct {
	ID          int64         `json:"id"`
	Mid         int64         `json:"mid"`
	Buvid       string        `json:"buvid"`
	LotteryID   string        `json:"lottery_id"`
	GiftType    int64         `json:"gift_type"`
	GiftID      int64         `json:"gift_id"`
	GiftName    string        `json:"gift_name"`
	GiftMoney   int64         `json:"gift_money"`
	OrderID     string        `json:"order_id"`
	PayOrderID  string        `json:"pay_order_id"`
	OrderTime   int64         `json:"order_time"`
	OrderStatus int           `json:"order_status"`
	PayStatus   int           `json:"pay_status"`
	Ctime       model.StrTime `json:"ctime"`
	Mtime       model.StrTime `json:"mtime"`
}

// PayOrder .
type PayOrder struct {
	ID          int64      `json:"id"`
	Mid         int64      `json:"mid"`
	OrderDesc   string     `json:"order_desc"`
	Money       int64      `json:"money"`
	OrderID     string     `json:"order_id"`
	PayOrderID  string     `json:"pay_order_id"`
	OrderTime   int64      `json:"order_time"`
	PayTime     int64      `json:"pay_time"`
	OrderStatus int32      `json:"order_status"`
	PayStatus   int32      `json:"pay_status"`
	CTime       xtime.Time `json:"ctime"`
	MTime       xtime.Time `json:"mtime"`
}
