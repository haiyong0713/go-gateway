package rewards

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"strconv"
	"strings"
	"time"
)

const (
	//CdKey奖励
	rewardTypeCdKey = "CdKey"
)

func init() {
	//CdKey奖励
	awardsSendFuncMap[rewardTypeCdKey] = Client.cdKeySender
	awardsConfigMap[rewardTypeCdKey] = &model.CdKeyConfig{}
}

// cdKeySender: 兑换码
func (s *service) cdKeySender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	var cdKeyId int64
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.cdKeySender mids(%v) uniqueId(%v) config(%v) error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		uri := strings.Replace(c.NotifyJumpUri2, "{{CDKEY_ID}}", strconv.FormatInt(cdKeyId, 10), -1)
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, uri)
	}()
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		cdKeyId, err = s.dao.SendCdKey(ctx, mid, c.ActivityId, c.Id, uniqueID)
		if err == nil {
			break
		}
		log.Errorc(ctx, "rewards.cdKeySender fail, wait next retry. error: %v", err)
	}

	return
}
