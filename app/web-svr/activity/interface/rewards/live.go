package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"net/url"
	"strconv"
	"time"
)

const (
	//直播金瓜子
	rewardTypeLiveGold       = "LiveGold"
	_liveGoldUri             = "/xlive/internal/revenue/v1/apcenter/operateGoldPoolAward"
	_liveGoldTimeStampFormat = "2006-01-02 15:04:05"
)

func init() {
	awardsSendFuncMap[rewardTypeLiveGold] = Client.liveGoldSender
	awardsConfigMap[rewardTypeLiveGold] = &model.LiveGoldConfig{}
}

// liveGoldSender: 直播金瓜子发放.
func (s *service) liveGoldSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.liveGoldSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.LiveGoldConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	params := url.Values{}
	params.Set("order_id", uniqueID)
	params.Set("order_source", strconv.FormatInt(config.OrderSource, 10))
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("coin_type", strconv.FormatInt(config.Type, 10))
	params.Set("coin", strconv.FormatInt(config.Count, 10))
	params.Set("remark", config.Remark)
	params.Set("pool_id", strconv.FormatInt(config.PoolId, 10))
	params.Set("timestamp", time.Now().Format(_liveGoldTimeStampFormat))

	var res struct {
		Code int `json:"code"`
		Msg  struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Status int `json:"status"`
			} `json:"data"`
		} `json:"data"`
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		err = s.httpClient.Get(ctx, s.liveGoldURL, metadata.String(ctx, metadata.RemoteIP), params, &res)
		if err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("outer code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
		}
		if res.Msg.Data.Status != 0 {
			err = fmt.Errorf("inner code: %v, message: %v", ecode.Int(res.Msg.Data.Status).Error(), res.Msg)
			if res.Msg.Data.Status != 4 {
				//1: 预算池余额不足, 2: 预算池已下线, 3: 预算池不存在, 4: 系统错误, 5: 参数错误
				//只有code=4时需要重试
				return
			}
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.liveGoldSender fail, wait next retry. error: %v", err)
	}
	return
}
