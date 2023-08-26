package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"

	"go-common/library/log"
)

type Configs struct {
	AwardMap map[int64] /*awardId*/ *api.RewardsAwardInfo
}

// GetAwardConfigById: 根据奖励Id查找奖励配置, err==nil时c一定不为nil, 调用方无需再次判断
func (s *service) GetAwardConfigById(ctx context.Context, awardId int64) (c *api.RewardsAwardInfo, err error) {
	if awardId == 0 { //used for debug
		return &api.RewardsAwardInfo{
			Id:      0,
			Type:    rewardTypeOther,
			Name:    "测试用奖品",
			JsonStr: "",
		}, nil
	}
	cs := s.awardsConfigs.Load().(*Configs)
	if cs == nil || cs.AwardMap == nil {
		err = fmt.Errorf("award config is emptySender")
		log.Errorc(ctx, "GetAwardConfigById error %v", err)
		return
	}
	c, ok := cs.AwardMap[awardId]
	if !ok || c == nil {
		err = fmt.Errorf("no such award id %v", awardId)
		log.Errorc(ctx, "GetAwardConfigById error %v", err)
		return
	}
	return
}

func deepCopyMap(old map[string]string) (new map[string]string) {
	new = make(map[string]string, 0)
	for k, v := range old {
		new[k] = v
	}
	return new
}

// GetAwardConfigById: 根据奖励Id获取奖励信息, err==nil时c一定不为nil, 调用方无需再次判断
// 使用场景: 用户中间后, 根据Id获取奖励信息返回给用户
// 与GetAwardConfigById的差别: GetAwardConfigById返回结果中不包含敏感信息
func (s *service) GetAwardSentInfoById(ctx context.Context, awardId, mid int64) (info *api.RewardsSendAwardReply, err error) {
	ac, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	info = &api.RewardsSendAwardReply{}
	info.Mid = mid
	info.AwardId = ac.Id
	info.Name = ac.Name
	info.ActivityId = ac.ActivityId
	info.ActivityName = ac.ActivityName
	info.Type = ac.Type
	info.Icon = ac.IconUrl
	info.ReceiveTime = time.Now().Unix()
	info.ExtraInfo = deepCopyMap(ac.ExtraInfo)
	return
}

// updateAwardConfigLoop: 定时更新内存中的奖励配置
func (s *service) updateAwardConfigLoop() {
	ctx := context.Background()
	for {
		cs, err := s.dao.GetAwardMap(ctx, 0)
		if err != nil {
			log.Errorc(ctx, "updateAwardConfigLoop error: %v", err)
		} else {
			s.awardsConfigs.Store(&Configs{AwardMap: cs})
		}

		time.Sleep(15 * time.Second)
	}
}

// AddAward: 增加奖励配置
func (s *service) AddAward(ctx context.Context, c *api.RewardsAddAwardReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "award AddAward error: %v", err)
		}
	}()
	afterCheckJsonStr := ""
	afterCheckJsonStr, err = s.validateJsonStr(c.Type, c.JsonStr)
	if err != nil {
		return
	}
	c.JsonStr = afterCheckJsonStr
	err = s.dao.AddAward(ctx, c)
	return
}

func (s *service) validateJsonStr(awardType, jsonStr string) (afterCheckJsonStr string, err error) {
	s.validateMu.Lock()
	defer s.validateMu.Unlock()
	var awardConfig interface{}

	awardConfig = awardsConfigMap[awardType]
	if awardConfig == nil {
		err = fmt.Errorf("no such award type: %v", awardType)
		return
	}
	err = json.Unmarshal([]byte(jsonStr), awardConfig)
	if err != nil {
		return
	}
	err = s.v.Struct(awardConfig)
	if err != nil {
		return
	}
	var afterCheckJsonBytes []byte
	afterCheckJsonBytes, err = json.Marshal(awardConfig)
	if err != nil {
		return
	}
	afterCheckJsonStr = string(afterCheckJsonBytes)
	return
}

func (s *service) DelAward(ctx context.Context, awardId int64) (err error) {
	return s.dao.DelAward(ctx, awardId)
}

func (s *service) GetAwards(ctx context.Context, c *api.RewardsListAwardReq) (res []*api.RewardsAwardInfo, err error) {
	res = make([]*api.RewardsAwardInfo, 0)
	return s.dao.GetAwardSlice(ctx, c.ActivityId, c.Keyword)
}

func (s *service) UpdateAward(ctx context.Context, c *api.RewardsAwardInfo) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "award UpdateAward error: %v", err)
		}
	}()
	afterCheckJsonStr := ""
	afterCheckJsonStr, err = s.validateJsonStr(c.Type, c.JsonStr)
	if err != nil {
		return
	}
	c.JsonStr = afterCheckJsonStr
	return s.dao.UpdateAward(ctx, c)
}

func (s *service) AddActivity(ctx context.Context, c *api.RewardsAddActivityReq) error {
	return s.dao.AddActivity(ctx, c)
}

func (s *service) DelActivity(ctx context.Context, c *api.RewardsDelActivityReq) error {
	return s.dao.DelActivity(ctx, c.ActivityId)
}

func (s *service) ListActivity(ctx context.Context, c *api.RewardsListActivityReq) (reply *api.RewardsListActivityReply, err error) {
	return s.dao.ListActivity(ctx, c, true)
}

func (s *service) UpdateActivity(ctx context.Context, c *api.RewardsUpdateActivityReq) error {
	return s.dao.UpdateActivity(ctx, c)
}

func (s *service) GetActivityDetail(ctx context.Context, c *api.RewardsGetActivityDetailReq) (*api.RewardsGetActivityDetailReply, error) {
	return s.dao.GetActivityDetail(ctx, c)
}
