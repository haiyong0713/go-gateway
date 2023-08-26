package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"time"

	vip "git.bilibili.co/bapis/bapis-go/vip/resource/service"

	account "git.bilibili.co/bapis/bapis-go/account/service/coupon"
)

const (
	//大会员代金券
	rewardTypeVipCoupon = "VipCoupon"
	//大会员
	rewardTypeVip = "Vip"
)

var (
	businessMap map[string]string
)

func init() {
	businessMap = make(map[string]string, 0)
	businessMap["bnj2021Lottery1"] = "拜年纪活动奖励"

	awardsSendFuncMap[rewardTypeVipCoupon] = Client.vipCouponGrpc
	awardsSendFuncMap[rewardTypeVip] = Client.vipGrpc

	awardsConfigMap[rewardTypeVipCoupon] = &model.VipCouponConfig{}
	awardsConfigMap[rewardTypeVip] = &model.VipConfig{}
}

// vipCouponHttp: 发放大会员代金券(GRPC方式)
// https://info.bilibili.co/pages/viewpage.action?pageId=102963040
func (s *service) vipCouponGrpc(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.vipCouponGrpc mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.VipCouponConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		err = fmt.Errorf("vipCouponGrpc : config illegal")
		return
	}
	req := &account.AllowanceReceiveReq{
		Mid:        mid,
		BatchToken: config.BatchToken,
		OrderNo:    uniqueID,
		Appkey:     config.AppKey,
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.accountCouponClient.AllowanceReceive(ctx, req)
		if ecode.Cause(err).Code() == 69090 {
			err = nil
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.vipCouponGrpc fail, wait next retry. error: %v", err)
	}

	return
}

func (s *service) vipGrpc(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.vipGrpc mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	tmpBusiness, ok := businessMap[business]
	if ok {
		business = tmpBusiness
	}
	config := &model.VipConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		err = fmt.Errorf("vipGrpc : config illegal")
		return
	}

	req := &vip.ResourceUseReq{
		BatchToken: config.BatchToken,
		Mid:        mid,
		OrderNo:    uniqueID,
		Remark:     business,
		WithDetail: false,
		Appkey:     config.AppKey,
		Ts:         time.Now().Unix(),
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.vipClient.ResourceUse(ctx, req)
		if ecode.Cause(err).Code() == 69090 {
			err = nil
		}
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.vipGrpc fail, wait next retry. error: %v", err)
	}
	return
}
