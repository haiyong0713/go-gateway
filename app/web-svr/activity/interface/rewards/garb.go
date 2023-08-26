package rewards

//装扮相关奖励发放
//docs: https://info.bilibili.co/pages/viewpage.action?pageId=36650764
import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"time"

	garbDiy "git.bilibili.co/bapis/bapis-go/garb/diy/service"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	garbCoupon "git.bilibili.co/bapis/bapis-go/vas/coupon/service"
)

const (
	//装扮套装
	rewardTypeGarbSuit = "GarbSuit"
	//装扮, 非套装
	rewardTypeGarbDressUp = "GarbDressUp"
	//2021拜年祭装扮
	rewardType2021BnjDressUp = "2021BnjDressUp"
	//装扮折扣券
	rewardTypeGarbCoupon = "GarbCoupon"
	//装扮Diy工具
	rewardTypeGarbDiyTool = "GarbDiyTool"
)

func init() {
	awardsSendFuncMap[rewardTypeGarbSuit] = Client.garbSuitSender
	awardsSendFuncMap[rewardTypeGarbDressUp] = Client.garbDressUpSender
	awardsSendFuncMap[rewardType2021BnjDressUp] = Client.garbBnj2021DressUpSender
	awardsSendFuncMap[rewardTypeGarbCoupon] = Client.garbCouponSender
	awardsSendFuncMap[rewardTypeGarbDiyTool] = Client.garbDiyToolSender

	awardsConfigMap[rewardTypeGarbSuit] = &model.SuitConfig{}
	awardsConfigMap[rewardTypeGarbDressUp] = &model.DressUpConfig{}
	awardsConfigMap[rewardType2021BnjDressUp] = &model.DressUpConfig{}
	awardsConfigMap[rewardTypeGarbCoupon] = &model.GarbCouponConfig{}
	awardsConfigMap[rewardTypeGarbDiyTool] = &model.GarbDiyToolConfig{}
}

//garbSuitSender: 发放套装
//business: 业务名
/*
message GrantSuitReq{
	repeated int64 mids = 1; // 必填 多个mid 最多100个
	int64 suitID = 2; // 必填 套装id
	int64 addSecond = 3; // 必填 发放时长 单位秒 大于0
	string token = 4; // 非必填 同一个token 只发一次
	string business = 5; // 业务方名发放目的 最长16个字符
}
*/
func (s *service) garbSuitSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	req := &garb.GrantSuitReq{}
	config := &model.SuitConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.garbSuitSender mids(%v) uniqueId(%v) config: %+v, error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	addSecond := config.ExpireDays * 24 * 60 * 60
	if config.ExpireDays < 0 {
		addSecond = -1
	}
	req.Mids = []int64{mid}
	req.SuitID = config.Id
	req.Token = uniqueID
	req.Business = business
	req.AddSecond = addSecond

	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.garbClient.GrantSuit(ctx, req)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.garbSuitSender fail, wait next retry. error: %v", err)
	}
	return
}

//garbDressUpSender: 发放装扮(非套装)
//docs: https://info.bilibili.co/pages/viewpage.action?pageId=129813057
/*
message GrantByBizReq{
	repeated int64 mids = 1; // 多个mid 最多100个
	repeated int64 ids = 2; // 装扮id（非套装）
	int64 addSecond = 3; // 必填 发放时长 单位秒 大于0
	string token = 4; // 不必填 同一个token 只发一次
}
*/
func (s *service) garbDressUpSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	req := &garb.GrantByBizReq{}
	config := &model.DressUpConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.garbDressUpSender mids(%v) uniqueId(%v) config: %+v, error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()

	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	addSecond := config.ExpireDays * 24 * 60 * 60
	if config.ExpireDays < 0 {
		addSecond = -1
	}
	req.Mids = []int64{mid}
	req.Ids = []int64{config.Id}
	req.AddSecond = addSecond
	req.Token = uniqueID
	req.Business = business
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.garbClient.GrantByBiz(ctx, req)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.garbDressUpSender fail, wait next retry. error: %v", err)
	}
	return
}

// garbBnj2021DressUpSender: 发放2021拜年祭装扮(非套装)
func (s *service) garbBnj2021DressUpSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	req := &garb.UserAssetUnlockReq{}
	config := &model.DressUpConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.garbBnj2021DressUpSender mids(%v) uniqueId(%v), config: %+v, req: %+v, error:%v", mid, uniqueID, c.JsonStr, req, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	addSecond := config.ExpireDays * 24 * 60 * 60
	if config.ExpireDays < 0 {
		addSecond = -1
	}
	req.Mid = mid
	req.ItemId = config.Id
	req.Business = business
	req.AddSecond = addSecond
	req.Token = uniqueID
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.garbClient.UserAssetUnlock(ctx, req)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.garbBnj2021DressUpSender fail, wait next retry. error: %v", err)
	}
	return
}

// garbCouponSender: 发放装扮优惠券
func (s *service) garbCouponSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	req := &garbCoupon.GrantCouponReq{}
	config := &model.GarbCouponConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.garbCouponSender mids(%v) uniqueId(%v), config: %+v, req: %+v, error:%v", mid, uniqueID, c.JsonStr, req, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	req.Mid = mid
	req.IdempotentNo = uniqueID
	req.Source = business
	req.BatchToken = config.BatchToken
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.garbCouponClient.GrantCoupon(ctx, req)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.garbBnj2021DressUpSender fail, wait next retry. error: %v", err)
	}
	return
}

// garbCouponSender: 发放装扮优惠券
func (s *service) garbDiyToolSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	req := &garbDiy.GrantDiyPendantByBizReq{}
	config := &model.GarbDiyToolConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.garbDiyToolSender mids(%v) uniqueId(%v), config: %+v, req: %+v, error:%v", mid, uniqueID, c.JsonStr, req, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	addSecond := config.ExpireDays * 24 * 60 * 60
	if config.ExpireDays < 0 {
		addSecond = -1
	}
	req.Mid = mid
	req.Token = uniqueID
	req.ActivityID = config.ActivityId
	req.AddSecond = addSecond
	req.Business = business
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		_, err = s.garbDiyClient.GrantDiyPendantByBiz(ctx, req)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.garbBnj2021DressUpSender fail, wait next retry. error: %v", err)
	}
	return
}
