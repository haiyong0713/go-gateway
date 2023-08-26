package rewards

//会员购奖励
//docs: https://info.bilibili.co/pages/viewpage.action?pageId=111058327
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"net/http"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"

	user "git.bilibili.co/bapis/bapis-go/passport/service/user"
)

const (
	//会员购满减卷
	rewardTypeMallCoupon = "MallCoupon"
	//会员购优惠券(新)
	rewardTypeMallCouponV2 = "MallCouponV2"
	//会员购商品
	rewardTypeMallPrize = "MallPrize"
	//会员购实物(新)
	rewardTypeMallPrizeV2 = "MallPrizeV2"
	//会员购红包
	rewardTypeMallPay = "MallPay"

	//优惠券, 魔晶请求地址
	_mallCouponURI = "/mall-marketing/coupon_code/createV2"

	//奖品
	_mallPrizeURI = "/mall-marketing-core/game/receive_game_prize"

	//优惠券, 魔晶, 红包, 实物发放地址(新接口)
	_mallCouponURIV2 = "discovery://open.mall.mall-asset-watchdog/asset/release_asset"

	_mallChannelTypeVipMall = 1
	_mallPrizeSourceId      = 10
)

func init() {
	awardsSendFuncMap[rewardTypeMallCoupon] = Client.mallCouponSender
	awardsSendFuncMap[rewardTypeMallPrize] = Client.mallPrizeSender
	awardsSendFuncMap[rewardTypeMallCouponV2] = Client.mallCouponV2Sender
	awardsSendFuncMap[rewardTypeMallPay] = Client.mallPayV2Sender
	awardsSendFuncMap[rewardTypeMallPrizeV2] = Client.mallPrizeV2Sender

	awardsConfigMap[rewardTypeMallCoupon] = &model.MallCouponConfig{}
	awardsConfigMap[rewardTypeMallPrize] = &model.MallPrizeConfig{}
	awardsConfigMap[rewardTypeMallCouponV2] = &model.MallCouponConfigV2{}
	awardsConfigMap[rewardTypeMallPay] = &model.MallCouponPayConfigV2{}
	awardsConfigMap[rewardTypeMallPrizeV2] = &model.MallPrizeConfigV2{}
}

// mallCouponSender: 优惠券/魔晶发放.
func (s *service) mallCouponSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.mallCouponSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.MallCouponConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	userDetail, err := s.userAccClient.UserDetail(ctx, &user.UserDetailReq{
		Mid: mid,
	})
	if err != nil {
		return
	}
	param := &struct {
		MID              int64  `json:"mid"`
		CouponID         string `json:"couponId"`
		SourceId         int64  `json:"sourceId"`
		SourceActivityId string `json:"sourceActivityId"`
		SourceBizId      string `json:"sourceBizId"`
		UserName         string `json:"uname"`
	}{
		MID:              mid,
		CouponID:         config.CouponId,
		SourceId:         10,
		SourceActivityId: "masteractivity",
		SourceBizId:      uniqueID,
		UserName:         userDetail.UserID,
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		var req *http.Request
		if req, err = http.NewRequest("POST", s.mallCouponURL, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"message"`
		}
		err = s.httpClient.Do(ctx, req, &res)
		if err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
			if res.Code == 83110020 {
				log.Error("日志告警 会员购优惠券到期 mid: %v, uniqueId: %v, couponId:%v", mid, uniqueID, config.CouponId)
				break
			}
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.mallCouponSender fail, wait next retry. error: %v", err)
	}
	return
}

// mallPrizeSender: 奖品发放.
func (s *service) mallPrizeSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.mallPrizeSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.MallPrizeConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	param := &struct {
		BizInfoDTO struct {
			BizId    string `json:"bizId"`
			SourceId int64  `json:"sourceId"`
		} `json:"bizInfoDTO"`
		MID         int64  `json:"mid"`
		PrizeNo     int64  `json:"prizeNo"`
		PrizePoolId int64  `json:"prizePoolId"`
		GameId      string `json:"gameId"`
	}{
		BizInfoDTO: struct {
			BizId    string `json:"bizId"`
			SourceId int64  `json:"sourceId"`
		}{
			BizId:    uniqueID,
			SourceId: 10,
		},
		MID:         mid,
		PrizeNo:     config.PrizeNo,
		PrizePoolId: config.PrizePoolId,
		GameId:      config.GameId,
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		var req *http.Request
		if req, err = http.NewRequest("POST", s.mallPrizeURL, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"message"`
		}
		err = s.httpClient.Do(ctx, req, &res)
		if err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.mallPrizeSender fail, wait next retry. error: %v", err)
	}
	return
}

// mallCouponV2Sender: 优惠券/魔晶发放(新接口).
func (s *service) mallCouponV2Sender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.mallCouponV2Sender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.MallCouponConfigV2{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	param := &struct {
		SourceId          string `json:"sourceId"`
		SourceAuthorityId string `json:"sourceAuthorityId"`
		AssetRequest      struct {
			Channel     int64  `json:"channel"`
			SourceBizId string `json:"sourceBizId"`
			MID         int64  `json:"uid"`
		} `json:"assetRequest"`
	}{
		SourceId:          conf.Conf.Rewards.VipMallSourceId,
		SourceAuthorityId: config.SourceAuthorityId,
		AssetRequest: struct {
			Channel     int64  `json:"channel"`
			SourceBizId string `json:"sourceBizId"`
			MID         int64  `json:"uid"`
		}{
			Channel:     _mallChannelTypeVipMall,
			SourceBizId: uniqueID,
			MID:         mid},
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		var req *http.Request
		if req, err = http.NewRequest("POST", _mallCouponURIV2, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"message"`
		}
		err = s.discoveryHttpClient.Do(ctx, req, &res)
		if err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
			if res.Code >= 83003001 && res.Code <= 83003004 {
				log.Errorc(ctx, "rewards.mallCouponV2Sender fail, wait next retry. error: %v", err)
				log.Errorc(ctx, "日志告警 会员购优惠券权限非法 mid: %v, uniqueId: %v, SourceAuthorityId:%v, Code:%v", mid, uniqueID, config.SourceAuthorityId, res.Code)
				break
			}
		}
		if err == nil {
			break
		}
	}
	return
}

// mallPayV2Sender: 会员购红包发放(新接口).
func (s *service) mallPayV2Sender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.mallPayV2Sender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.MallCouponPayConfigV2{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	param := &struct {
		SourceId          string `json:"sourceId"`
		SourceAuthorityId string `json:"sourceAuthorityId"`
		AssetRequest      struct {
			ReferenceId    string `json:"referenceId"`
			SourcePlatform string `json:"sourcePlatform"`
			SourceName     string `json:"sourceName"`
			MID            int64  `json:"mid"`
			IsCheckRisk    bool   `json:"isCheckRisk"`
		} `json:"assetRequest"`
	}{
		SourceId:          conf.Conf.Rewards.VipMallSourceId,
		SourceAuthorityId: config.SourceAuthorityId,
		AssetRequest: struct {
			ReferenceId    string `json:"referenceId"`
			SourcePlatform string `json:"sourcePlatform"`
			SourceName     string `json:"sourceName"`
			MID            int64  `json:"mid"`
			IsCheckRisk    bool   `json:"isCheckRisk"`
		}{
			ReferenceId:    uniqueID,
			SourceName:     c.Name,
			SourcePlatform: conf.Conf.Rewards.VipMallSourcePlatform,
			MID:            mid,
			IsCheckRisk:    false},
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		var req *http.Request
		if req, err = http.NewRequest("POST", _mallCouponURIV2, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"message"`
		}
		err = s.discoveryHttpClient.Do(ctx, req, &res)
		if err != nil {
			continue
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
			if res.Code >= 83003001 && res.Code <= 83003004 {
				log.Errorc(ctx, "日志告警 会员购红包参数无效 mid: %v, uniqueId: %v, SourceAuthorityId:%v, Code:%v", mid, uniqueID, config.SourceAuthorityId, res.Code)
				break
			}
		}
		if err == nil {
			break
		}

		log.Errorc(ctx, "rewards.mallPayV2Sender fail, wait next retry. error: %v", err)
	}
	return
}

// mallPrizeV2Sender: 会员购实物发放(新接口).
func (s *service) mallPrizeV2Sender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.mallPrizeV2Sender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.MallPrizeConfigV2{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	param := &struct {
		SourceId          string `json:"sourceId"`
		SourceAuthorityId string `json:"sourceAuthorityId"`
		AssetRequest      struct {
			BizInfoDTO struct {
				BizId    string `json:"bizId"`
				SourceId int64  `json:"sourceId"`
			} `json:"bizInfoDTO"`
			Channel int64 `json:"channel"`
			MID     int64 `json:"mid"`
		} `json:"assetRequest"`
	}{
		SourceId:          conf.Conf.Rewards.VipMallSourceId,
		SourceAuthorityId: config.SourceAuthorityId,
		AssetRequest: struct {
			BizInfoDTO struct {
				BizId    string `json:"bizId"`
				SourceId int64  `json:"sourceId"`
			} `json:"bizInfoDTO"`
			Channel int64 `json:"channel"`
			MID     int64 `json:"mid"`
		}{
			BizInfoDTO: struct {
				BizId    string `json:"bizId"`
				SourceId int64  `json:"sourceId"`
			}{BizId: uniqueID, SourceId: _mallPrizeSourceId},
			Channel: _mallChannelTypeVipMall,
			MID:     mid,
		},
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}

	err = retry.WithAttempts(ctx, "rewards.mallPrizeV2Sender", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		var req *http.Request
		if req, err = http.NewRequest("POST", _mallCouponURIV2, bytes.NewReader(paramJSON)); err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		var res struct {
			Code int    `json:"code"`
			Msg  string `json:"message"`
		}
		err = s.discoveryHttpClient.Do(ctx, req, &res)
		if err != nil {
			return
		}
		if res.Code != 0 {
			err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
			if res.Code >= 83003001 && res.Code <= 83003004 {
				log.Errorc(ctx, "rewards.mallPrizeV2Sender fail, wait next retry. error: %v", err)
				log.Errorc(ctx, "日志告警 会员购奖品权限非法 mid: %v, uniqueId: %v, SourceAuthorityId:%v, Code:%v", mid, uniqueID, config.SourceAuthorityId, res.Code)
				return
			}
		}
		return
	})

	return
}
