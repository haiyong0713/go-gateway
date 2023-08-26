package pay

import (
	"context"
	"encoding/json"
	xhttp "net/http"
	"strconv"
	"strings"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/model/pay"
)

const (
	_contentType = "Content-Type"
	_urlencoded  = "application/json"
)

// Config ...
type Config struct {
	CustomerID   string
	MerchantCode string
	CoinType     string
	PayHost      string
	Token        string
	ActivityID   string
}

// Pay ...
type Pay struct {
	payClient *bm.Client
}

func newPayClient(client *bm.Client) *Pay {
	return &Pay{payClient: client}
}

// requestInner new http request with method, uri, ip, values and headers.
func (d *Pay) requestInner(ctx context.Context, uri, token string, param Values) (*pay.ResultInner, error) {
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
	res := &pay.ResultInner{}
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

// PayTransferInner 活动红包转入 .
func (d *dao) PayTransferInner(ctx context.Context, pt *pay.TransferInner) (res *pay.ResultInner, err error) {
	params := Values{}
	params.Set("traceId", pt.TraceID)                                  // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(pt.UID, 10))                   // 用户uid
	params.Set("customerId", pt.CustomerID)                            // 业务方id，由资产平台配置
	params.Set("orderNo", pt.OrderNo)                                  // 业务方转入红包的订单id（通过该字段保持幂等）
	params.Set("transBalance", strconv.FormatInt(pt.TransBalance, 10)) // 收益金额（单位人民币 元），小数点后两位
	params.Set("transDesc", pt.TransDesc)                              // 收益转入原因
	params.Set("activityId", pt.ActivityID)                            // 资产类型，由资产平台配置
	params.Set("startTme", strconv.FormatInt(pt.StartTme, 10))         // 红包解冻时间，毫秒值。若传入时间小于资产系统当前时间，则抛出异常，本次转入失败。
	params.Set("timestamp", strconv.FormatInt(pt.Timestamp, 10))       // 当前时间毫秒值
	params.Set("signType", "MD5")                                      // 签名校验类型，目前仅支持MD5
	res, err = d.pay.requestInner(ctx, d.payHost+_transferInner, pt.Token, params)
	log.Infoc(ctx, "payTransferInner res(%v) err(%v)", res, err)
	return res, err
}
