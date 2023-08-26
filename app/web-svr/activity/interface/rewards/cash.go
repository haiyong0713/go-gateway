package rewards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"net/http"
	"strconv"
	"time"
)

const (
	//现金
	rewardTypeCash    = "Cash"
	_transferInnerUrl = "/payplatform/fund/acct/inner/profit/transfer"
)

func init() {
	awardsSendFuncMap[rewardTypeCash] = Client.cashSender
	awardsConfigMap[rewardTypeCash] = &model.CashConfig{}
}

// cashSender: 发放现金
func (s *service) cashSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.comicsBnj2021CouponSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.CashConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	params := model.Values{}
	params.Set("traceId", uniqueID)                                            // 请求id 方便日志追踪
	params.Set("uid", strconv.FormatInt(mid, 10))                              // 用户uid
	params.Set("customerId", config.CustomerId)                                // 业务方id，由资产平台配置
	params.Set("orderNo", uniqueID)                                            // 业务方转入红包的订单id（通过该字段保持幂等）
	params.Set("transBalance", strconv.FormatInt(config.TransBalance, 10))     // 收益金额（单位人民币 元），小数点后两位
	params.Set("transDesc", business)                                          // 收益转入原因
	params.Set("activityId", config.ActivityID)                                // 资产类型，由资产平台配置
	params.Set("startTme", strconv.FormatInt(config.StartTme, 10))             // 红包解冻时间，毫秒值。若传入时间小于资产系统当前时间，则抛出异常，本次转入失败。
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/1000, 10)) // 当前时间毫秒值
	params.Set("signType", "MD5")                                              // 签名校验类型，目前仅支持MD5
	signedParams, err := params.Sign(s.c.Lottery.Pay.Token)
	if err != nil {
		return
	}

	paramJSON, err := json.Marshal(signedParams)
	if err != nil {
		return
	}
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		var req *http.Request
		if req, err = http.NewRequest("POST", s.c.Host.Pay+_transferInnerUrl, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if err = s.httpClient.Do(ctx, req, &res); err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
			continue
		}
		break
	}
	return
}
