package pay

import (
	"context"
	"encoding/json"
	xhttp "net/http"
	"strconv"
	"strings"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

const (
	_contentType = "Content-Type"
	_urlencoded  = "application/json"
)

// PayConfig .
type PayConfig struct {
	CustomerID   string
	MerchantCode string
	CoinType     string
	PayHost      string
	Token        string
	ActivityID   string
}

type Pay struct {
	payClient *bm.Client
}

func newPayClient(client *bm.Client) *Pay {
	return &Pay{payClient: client}
}

// requestInner new http request with method, uri, ip, values and headers.
func (d *Pay) requestInner(ctx context.Context, uri, token string, param Values) (*ResultInner, error) {
	params, err := param.Sign(token)
	if err != nil {
		log.Error("params params:%+v,sign err:%+v", params, err)
		return nil, err
	}
	bs, err := json.Marshal(params)
	if err != nil {
		log.Error("params Marshal:%+v,sign err:%+v", params, err)
		return nil, err
	}
	req, err := xhttp.NewRequest(xhttp.MethodPost, uri, strings.NewReader(string(bs)))
	if err != nil {
		log.Error("requestInner pay url:%s,body:%s,error:%+v", uri, string(bs), err)
		return nil, err
	}
	req.Header.Set(_contentType, _urlencoded)
	res := &ResultInner{}
	if bs, err = d.payClient.Raw(ctx, req); err != nil {
		log.Error("payClient requestInner pay url:%s,body:%s,error:%+v", uri, string(bs), err)
		return nil, err
	}
	log.Warn("requestInner url:%s,params:%+v,body:%s", uri, params, string(bs))
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("payClient request pay url:%s,body:%s,error:%+v", uri, string(bs), err)
	}
	return res, err
}

const _transferInner = "/payplatform/fund/acct/inner/profit/transfer"

// PayTransfer 活动红包转入 .
func (d *Pay) PayTransferInner(ctx context.Context, payConfig *PayConfig, pt *PayTransferInner) (res *ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                                  // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10))                   // 用户uid
	params.Set("customerId", payConfig.CustomerID)                     // 业务方id，由资产平台配置
	params.Set("orderNo", pt.OrderNo)                                  // 业务方转入红包的订单id（通过该字段保持幂等）
	params.Set("transBalance", strconv.FormatInt(pt.TransBalance, 10)) // 收益金额（单位人民币 元），小数点后两位
	params.Set("transDesc", pt.TransDesc)                              // 收益转入原因
	params.Set("activityId", payConfig.ActivityID)                     // 资产类型，由资产平台配置
	params.Set("startTme", strconv.FormatInt(pt.StartTme, 10))         // 红包解冻时间，毫秒值。若传入时间小于资产系统当前时间，则抛出异常，本次转入失败。
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10))       // 当前时间毫秒值
	params.Set("signType", "MD5")                                      // 签名校验类型，目前仅支持MD5

	return d.requestInner(ctx, payConfig.PayHost+_transferInner, payConfig.Token, params)
}

const _profitCancelInner = "/payplatform/fund/acct/inner/profit/cancel"

// ProfitCancelInner 活动红包撤回 .
func (d *Pay) ProfitCancelInner(ctx context.Context, payConfig *PayConfig, pt *ProfitCancelInner) (res *ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                            // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10))             // 用户uid
	params.Set("customerId", payConfig.CustomerID)               // 业务方id，由资产平台配置
	params.Set("activityId", payConfig.ActivityID)               // 资产类型，由资产平台配置
	params.Set("transOrderNo", pt.TransOrderNo)                  //
	params.Set("transDesc", pt.TransDesc)                        //
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10)) // 当前时间毫秒值
	params.Set("fullAmountCancel", "1")                          // 是否需要全量追回。不传则默认从原账户扣除。 0：从原活动账户扣除  1：若原活动账户过期或者余额不足，则从同一个提现的其他活动账户中扣除
	params.Set("signType", "MD5")                                // 签名校验类型，目前仅支持MD5

	return d.requestInner(ctx, payConfig.PayHost+_profitCancelInner, payConfig.Token, params)
}

const _balanceInner = "/payplatform/fund/acct/inner/current/balance/qry"

// Balance 查询用户余额 .
func (d *Pay) BalanceInner(ctx context.Context, payConfig *PayConfig, pt *BalanceInner) (res *ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                            // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10))             // 用户uid
	params.Set("customerId", payConfig.CustomerID)               // 业务方id，由资产平台配置
	params.Set("merchantCode", payConfig.MerchantCode)           // 业务方平台类型，由资产平台配置
	params.Set("coinType", payConfig.CoinType)                   // 资产类型，由资产平台配置
	params.Set("activityId", payConfig.ActivityID)               // 资产类型，由资产平台配置
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10)) // 当前时间毫秒值
	params.Set("signType", "MD5")                                // 签名校验类型，目前仅支持MD5

	return d.requestInner(ctx, payConfig.PayHost+_balanceInner, payConfig.Token, params)
}

const _incomeClear = "/payplatform/fund/acct/inner/income/clear"

// IncomeClear 用户收益清零 .
func (d *Pay) IncomeClearInner(ctx context.Context, payConfig *PayConfig, pt *IncomeClearInner) (res *ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10)) // 用户uid
	params.Set("orderNo", pt.OrderNo)
	params.Set("transDesc", pt.TransDesc)
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10)) // 当前时间毫秒值
	params.Set("signType", "MD5")                                // 签名校验类型，目前仅支持MD5

	params.Set("customerId", payConfig.CustomerID) // 业务方id，由资产平台配置
	params.Set("activityId", payConfig.ActivityID) // 资产类型，由资产平台配置

	return d.requestInner(ctx, payConfig.PayHost+_incomeClear, payConfig.Token, params)
}
