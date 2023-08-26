package pay

// PayTransferInner .
type PayTransferInner struct {
	TraceID      string
	UID          int64
	OrderNo      string
	TransBalance int64
	TransDesc    string
	StartTme     int64
	Timestamp    int64
}

// ProfitCancelInner .
type ProfitCancelInner struct {
	TraceID          string
	UID              int64
	TransOrderNo     string
	TransDesc        string
	Timestamp        int64
	FullAmountCancel int32 // 是否需要全量追回。不传则默认从原账户扣除。 0：从原活动账户扣除  1：若原活动账户过期或者余额不足，则从同一个提现的其他活动账户中扣除
}

// BalanceInner .
type BalanceInner struct {
	TraceID   string
	UID       int64
	Timestamp int64
	SignType  string
}

// IncomeClearInner .
type IncomeClearInner struct {
	TraceID   string
	UID       int64
	OrderNo   string
	TransDesc string
	Timestamp int64
}

// BalanceInner .
type Statchange struct {
	TraceID     string
	UID         int64
	OperateType int32
	Timestamp   int64
	Reason      string
}

// const .
const (
	OrderStatusPending          = 0
	OrderStatusFail             = 1 // TODO 轮训重试请求支付账户
	OrderStatusAccountNormal    = 2
	OrderStatusAccountFrozen    = 3
	OrderStatusAccountNotEnough = 4
	OrderStatusAccountNotFund   = 5

	// 8004070003 // 资产平台交易订单号不能为空

	PayOrderAccountNotEnough = 8004070903 // 8004070903 // 账户余额不足
	PayOrderAccountNotFund   = 8004070401
	// 8004070401 : account not fund 账户信息未找到，请稍后重试
)

// ResultInner
// NOTE: 一般情况下 Code 和 ErrNO 等价的.
type ResultInner struct {
	Code    int64  `json:"code"`
	ErrNO   int64  `json:"errno,omitempty"`
	MSG     string `json:"msg,omitempty"`
	ShowMsg string `json:"showMsg"`
	ErrTag  int64  `json:"errtag,omitempty"`
	Data    struct {
		UID             string `json:"uid"`
		TransOrderNo    string `json:"transOrderNo"`
		OrderNo         string `json:"orderNo"`
		Status          int64  `json:"status"`
		TransBalance    int64  `json:"transBalance"`
		LimitBalance    int64  `json:"limitBalance"`
		EarnedBalance   int64  `json:"earnedBalance,omitempty"`
		ActivityID      string `json:"activityId,omitempty"`
		CanceledBalance int64  `json:"canceledBalance,omitempty"`
	} `json:"data"`
}

// OrderStatus .
func (r *ResultInner) OrderStatus() int64 {
	switch {
	case r == nil:
		return OrderStatusFail
	case r.Code == PayOrderAccountNotEnough:
		return OrderStatusAccountNotEnough
	case r.Code == PayOrderAccountNotFund:
		return OrderStatusAccountNotFund
	case r.Code != 0:
		return OrderStatusFail
	// case r.Data == nil:
	// 	return OrderStatusFail
	case r.Data.Status == 0:
		return OrderStatusAccountNormal
	case r.Data.Status == 1:
		return OrderStatusAccountFrozen
	}
	return OrderStatusFail
}

// 8004070401 : account not fund 账户信息未找到，请稍后重试
// EARN_TRANSFER_IN_ERROR(8004070907L, "用户红包转入失败，请稍后重试"),
// FUND_ORDER_REVOKE_ERROR(8004070902L, "red packet revoke failed", "用户红包撤回失败，请稍后重试"),
// USER_SUBACCT_TRANSFER_ERROR(8004070900L, "user account balance transfer failed","账户变更异常，请稍后重试")
// REQUEST_FREQUENTLY_ERROR(8004070015L, "request frequently", "请求过于频繁，请稍后重试"),

//
// BALANCE_NOT_ENOUGH_ERROR(8004070903L, "user account money not enough", "账户余额不足")
// REVOKE_ORDER_NOT_FOUNT(8004070904L, "can not fund this order", "未找到该订单！"),
// ACCOUNT_UNNORMAL(8004070906L, "account has been frozen", "该账户已经被冻结！"),
// ORDER_ALREADY_EXISTS(8004070908L, "order already exists","订单已经存在，请不要重复转入"),
// SIGN_ERROR(8004070001L, "SIGN_ERROR","签名异常"),
// INTERNAL_ERROR(8004070004L, "INTERNAL_ERROR","内部错误")
