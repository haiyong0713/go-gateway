package rewards

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"

	class "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"
)

const (
	//课堂优惠券
	rewardTypeClassCoupon = "ClassCoupon"
)

func init() {
	awardsSendFuncMap[rewardTypeClassCoupon] = Client.classCouponSender
	awardsConfigMap[rewardTypeClassCoupon] = &model.ClassCouponConfig{}
}

// classCouponSender: 领取课堂优惠券
func (s *service) classCouponSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.classCouponAsync mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.ClassCouponConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	senVc := 0
	if config.SendVc {
		senVc = 1
	}
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.classClient.AsynReceiveCoupon(ctx, &class.AsynReceiveCouponReq{ //异步发放
			Mid:        mid,
			BatchToken: config.BatchToken,
			SendVc:     int32(senVc),
		})
		if err == nil {
			break
		}
	}

	return
}
