package newyear2021

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"time"
)

func (s *Service) getTaskAwardUniqueId(mid int64, taskId, awardId int64, isDailyTask bool) string {
	d := time.Now().Format("2006-01-02")
	if !isDailyTask {
		d = "0000-00-00"
	}
	return fmt.Sprintf("%v:%v:%v:%v", mid, taskId, awardId, d)
}

func (s *Service) ReceiveAward(ctx context.Context, mid int64, business string, taskConfig *model.Task, isDailyTask bool) (info *api.RewardsSendAwardReply, err error) {
	c := s.GetConf()
	if c == nil {
		log.Errorc(ctx, "get nil reward config")
		err = ecode.ActivityTaskAwardFailed
		return
	}
	//获取uniqueId, 供服务方幂等操作使用
	uniqueId := s.getTaskAwardUniqueId(mid, taskConfig.Id, taskConfig.AwardId, isDailyTask)

	//调用发奖客户端发放奖励
	info, err = rewards.Client.SendAwardByIdAsync(ctx, mid, uniqueId, business, taskConfig.AwardId, true, true)
	if err != nil {
		log.Errorc(ctx, "ReceiveAward SendAwardById error: %v", err)
	}
	return
}
