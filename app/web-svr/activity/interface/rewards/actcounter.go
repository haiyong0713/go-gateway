package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/api"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
)

const (
	//活动平台Counter
	rewardTypeActCounter = "ActCounter"
)

func init() {
	awardsSendFuncMap[rewardTypeActCounter] = Client.actCounterSender
	awardsConfigMap[rewardTypeActCounter] = &model.ActCounterConfig{}
}

// actCounterSender: 活动平台Counter奖励发放
func (s *service) actCounterSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.actCounterSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.ActCounterConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	var ts int64
	ts, err = s.dao.GetActCounterUnusedTimestampByMid(ctx, mid)
	if err != nil {
		return
	}
	midStr := fmt.Sprintf("%d", mid)
	data := &model.ActPlatActivityPoints{
		Points:    config.Points,
		Timestamp: ts,
		Mid:       mid,
		Source:    408933983,
		Activity:  config.Activity,
		Business:  config.Business,
		Extra:     config.Extra,
	}
	for i := 0; i < 3; i++ {
		err = s.actPlatPub.Send(ctx, midStr, data)
		if err == nil {
			break
		}
	}
	return
}
