package rewards

//弹幕相关奖励发放
import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
)

const (
	//直播弹幕
	rewardTypeLiveDanmaku = "Danmuku"
)

func init() {
	awardsSendFuncMap[rewardTypeLiveDanmaku] = Client.liveDanmukuSender
	awardsConfigMap[rewardTypeLiveDanmaku] = &model.DanmukuConfig{}
}

func (s *service) liveDanmukuSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.liveDanmukuSender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	config := &model.DanmukuConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	rewards := make([]*model.BulletReward, 0, len(config.RoomIds))
	for _, v := range config.RoomIds {
		rewards = append(rewards, &model.BulletReward{
			RewardID:   6,
			ExpireTime: time.Now().Unix() + config.ExpireDays*24*60*60,
			Type:       11,
			ExtraData: &model.BulletExtraData{
				Type:   "color",
				Value:  config.Color,
				RoomID: int32(v),
			}})
	}
	res := &model.Bullet{
		Uids:    []int64{mid},
		MsgID:   uniqueID,
		Source:  1508,
		Rewards: rewards,
	}
	if err = s.liveDataBusPub.Send(ctx, fmt.Sprintf("%d", mid), res); err != nil {
		return
	}
	return
}
