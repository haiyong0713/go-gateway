package rewards

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
)

const (
	//测试用, 直接返回成功
	rewardTypeOther = "Other"
	//Up主祝福
	rewardTypeUpGreeting = "UpGreetings"
)

func init() {
	//Up主祝福, 不实际发放奖励. 配合extraInfo返回祝福语. 并保留中奖记录
	awardsSendFuncMap[rewardTypeUpGreeting] = Client.emptySender
	awardsConfigMap[rewardTypeUpGreeting] = &model.EmptyConfig{}
	//测试用
	awardsSendFuncMap[rewardTypeOther] = Client.emptySender
	awardsConfigMap[rewardTypeOther] = &model.EmptyConfig{}
}

// 不实际发放奖励. 可配合extraInfo返回自定义信息
func (s *service) emptySender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	return
}
