package pay

import (
	"context"
	"strconv"
)

const _statchange = "/payplatform/fund/user/account/statchange"

// Statchange 用户账户状态变更 .
func (d *Pay) Statchange(ctx context.Context, payConfig *PayConfig, pt *Statchange) (res *ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                                       // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10))                        // 用户uid
	params.Set("operateType", strconv.FormatInt(int64(pt.OperateType), 10)) // 操作类型，0解冻，1冻结
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10))            // 当前时间毫秒值
	params.Set("signType", "MD5")                                           // 签名校验类型，目前仅支持MD5
	params.Set("reason", pt.Reason)

	params.Set("customerId", payConfig.CustomerID)     // 业务方id，由资产平台配置
	params.Set("merchantCode", payConfig.MerchantCode) // 业务方平台类型，由资产平台配置
	params.Set("coinType", payConfig.CoinType)         // 资产类型，由资产平台配置

	return d.requestInner(ctx, payConfig.PayHost+_statchange, payConfig.Token, params)
}
