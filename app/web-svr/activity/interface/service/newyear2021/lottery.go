package newyear2021

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/log"

	"go-common/library/log/infoc.v2"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	logid_008274 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	lotteryModel "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	rewardModel "go-gateway/app/web-svr/activity/interface/model/rewards"
	mdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"go-gateway/app/web-svr/activity/interface/service/lottery"
	"go-gateway/app/web-svr/activity/interface/tool"

	"github.com/gogo/protobuf/proto"
)

const (
	bizNameOfPubLiveLotteryReceive    = "live_lottery_receive"
	bizNameOfInvalidLiveLotteryReward = "invalid_live_lottery_reward"

	cacheKey4BackupOfLiveAwardRec = "bnj2021_live_award_rec_%02d"
)

func (s *Service) LiveLotteryDetail(ctx context.Context, mid int64) (list []*model.LiveRewardDetail, err error) {
	list = make([]*model.LiveRewardDetail, 0)
	list, err = newyear2021.FetchUserLiveLotteryList(ctx, mid)
	if err != nil {
		err = ecode.BNJTooManyUser
	}

	return
}

func (s *Service) LiveDrawReIssue(ctx context.Context, mid, sceneID int64) (list []*api.RewardsSendAwardReply, err error) {
	sc := s.GetConf()
	list = make([]*api.RewardsSendAwardReply, 0)
	reward := new(model.UserRewardInLiveRoom)
	reward, err = newyear2021.PopUserRewardBySceneID(ctx, mid, sceneID)
	if err != nil {
		return
	}

	if reward.MID > 0 && LiveLotteryProducer != nil {
		t := reward.Reward
		//预抽奖只配置了5个兜底, 实际兜底为60个.
		//判断老兜底的奖励ID, 从新从60个中随机选取一个
		if isValueInSlice(t.AwardId, sc.LiveOldGreetingAwardIds) {
			newId := getRandomValueFromSlice(sc.LiveNewUpGreetingAwardIds)
			tmpT, err := rewards.Client.GetAwardSentInfoById(ctx, newId, mid)
			if err == nil {
				t = tmpT
			}
		}

		if fidStr, ok := t.ExtraInfo["up_mid"]; ok {
			fid, err := strconv.ParseInt(fidStr, 10, 0)
			if err == nil {
				t.ExtraInfo["up_is_follow"] = s.dao.IsUserFollowedStr(ctx, mid, fid)
			}
		}
		list = append(list, t)
		key := fmt.Sprintf("%v-%v", mid, time.Now().Unix())
		bs, _ := json.Marshal(reward)
		if pubErr := LiveLotteryProducer.Send(ctx, key, bs); pubErr != nil {
			//err = ecode.BNJTooManyUser
			log.Errorc(ctx, "Live lottery rec pub err: %v, reward: ", pubErr, string(bs))
			tool.Metric4PubDatabus.WithLabelValues([]string{bizNameOfPubLiveLotteryReceive}...).Inc()
		}
	}

	return
}

func (s *Service) LiveLotteryExchange(ctx context.Context, mid, sceneID int64) (list []*api.RewardsSendAwardReply, err error) {
	sc := s.GetConf()
	list = make([]*api.RewardsSendAwardReply, 0)
	if currentUnix < BnjStrategyInfo.LiveStartTime {
		err = ecode.BNJLiveDrawNotStart

		return
	}

	if currentUnix >= BnjStrategyInfo.LiveEndTime {
		err = ecode.BNJLiveDrawEnd

		return
	}

	reward := new(model.UserRewardInLiveRoom)
	reward, err = newyear2021.PopUserRewardBySceneID(ctx, mid, sceneID)
	if err != nil {
		return
	}

	if reward.MID > 0 && LiveLotteryProducer != nil {
		t := reward.Reward
		//预抽奖只配置了5个兜底, 实际兜底为60个.
		//判断老兜底的奖励ID, 从新从60个中随机选取一个
		if isValueInSlice(t.AwardId, sc.LiveOldGreetingAwardIds) {
			newId := getRandomValueFromSlice(sc.LiveNewUpGreetingAwardIds)
			tmpT, err := rewards.Client.GetAwardSentInfoById(ctx, newId, mid)
			if err == nil {
				t = tmpT
			}
		}

		if fidStr, ok := t.ExtraInfo["up_mid"]; ok {
			fid, err := strconv.ParseInt(fidStr, 10, 0)
			if err == nil {
				t.ExtraInfo["up_is_follow"] = s.dao.IsUserFollowedStr(ctx, mid, fid)
			}
		}
		list = append(list, t)
		key := fmt.Sprintf("%v-%v", mid, time.Now().Unix())
		bs, _ := json.Marshal(reward)
		if pubErr := LiveLotteryProducer.Send(ctx, key, bs); pubErr != nil {
			//err = ecode.BNJTooManyUser
			log.Errorc(ctx, "Live lottery rec pub err: %v, reward: ", pubErr, string(bs))
			tool.Metric4PubDatabus.WithLabelValues([]string{bizNameOfPubLiveLotteryReceive}...).Inc()
		}
		_ = rewards.Client.AddTmpAwardSentInfoToCache(ctx, mid, t.AwardId)
	}

	return
}

func pubAwardRecIntoBackup(ctx context.Context, info string) (err error) {
	cacheKey := fmt.Sprintf(cacheKey4BackupOfLiveAwardRec, genBackupRandSuffix())
	_, err = component.BackUpMQ.Do(ctx, "LPUSH", cacheKey, info)
	if err != nil {
		log.Errorc(ctx, "BNJ_Live pubAwardRecIntoBackup failed, err: %v", err)
	}

	return
}

func (s *Service) getLotteryAwardUniqueId(mid int64, detail *lotteryModel.RecordDetail, idx int) string {
	ts := time.Now().UnixNano()/1e6 - 1600000000000
	return fmt.Sprintf("%v-%v-%v-%v-%v", detail.GiftID, detail.Type, mid, idx, ts)
}

func PubUserDrawLog(ctx context.Context, mid, opType int64) {
	if BnjStrategyInfo.DWLogID4Draw == "" {
		return
	}

	info := new(logid_008274.BusinessMessage)
	{
		info.Mid = mid
		info.LogTime = time.Now().Unix()
		info.OpType = int32(opType)
	}

	if bs, err := proto.Marshal(info); err == nil {
		payload := infoc.NewPbPayload(BnjStrategyInfo.DWLogID4Draw, bs)
		if err := component.DWInfo.Info(ctx, payload); err != nil {
			log.Errorc(ctx, "PubUserDrawLog msg: %v, failed: %v", string(bs), err)
			tool.IncrCommonBizStatus("pub_bnj_draw", tool.StatusOfFailed)
		}
	} else {
		tool.IncrCommonBizStatus("pub_bnj_draw", tool.StatusOfFailed)
	}
}

func getRandomValueFromSlice(s []int64) int64 {
	idx := int(rand.NewSource(time.Now().UnixNano()).Int63()) % len(s)
	return s[idx]
}

func isValueInSlice(v int64, s []int64) bool {
	for _, i := range s {
		if v == i {
			return true
		}
	}
	return false
}

// SceneID=1: 扭蛋抽奖
// SceneID=2: 直播间抽奖
// SceneID=3: 直播间预约抽奖
func (s *Service) DoLottery(ctx context.Context, mid, sceneId, lotteryCount int64, lotteryService *lottery.Service, lotteryParams *mdl.Base, lotteryActivityId int64,
	shouldSend, debugNoCost, updateCache, updateDB bool) (list []*api.RewardsSendAwardReply, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "s.DoLottery error: %v", err)
		}
	}()
	if lotteryParams == nil {
		lotteryParams = &mdl.Base{}
	}
	lotteryParams.ActivityID = lotteryActivityId
	lotteryParams.MID = mid
	lotteryParams.Action = mdl.ActionLottery
	list = make([]*api.RewardsSendAwardReply, 0)

	lotteryCost := lotteryCount
	if lotteryCost == 10 {
		lotteryCost = 9
	}
	//测试插桩, 不消耗奖券.
	if !debugNoCost {
		err = newyear2021.DecreaseARCouponByCAS(ctx, mid, lotteryCost)
		if err != nil {
			return
		}
	}
	conf := s.GetConf()
	var activityId, userAlreadyWinTimes int64
	var upGreetingAwardIds []int64
	var lotterySid string
	switch sceneId {
	case 1:
		activityId = conf.NiuDanRewardsActivityId
		lotterySid = conf.NiuDanLotterySid
		upGreetingAwardIds = conf.NiuDanUpGreetingAwardIds
	case 2:
		activityId = conf.LiveRewardsActivityId1
		lotterySid = conf.LiveLotterySid1
		upGreetingAwardIds = conf.UpGreetingAwardIds1
	case 3:
		activityId = conf.LiveRewardsActivityId2
		lotterySid = conf.LiveLotterySid2
		upGreetingAwardIds = conf.UpGreetingAwardIds2
	case 4:
		activityId = conf.LiveRewardsActivityId3
		lotterySid = conf.LiveLotterySid3
		upGreetingAwardIds = conf.UpGreetingAwardIds3
	default:
		err = fmt.Errorf("no such sceneId: %v", sceneId)
		return
	}

	lotteryList := make([]*lotteryModel.RecordDetail, 0)
	userAlreadyWinTimes, err = rewards.Client.GetAwardCountByMidAndActivityId(ctx, mid, activityId)
	if err != nil {
		return
	}
	log.Infoc(ctx, "s.DoLottery call SimpleLottery with params, mid: %v,  risk: %+v, cost: %v, alreadyWin: %v", mid, lotteryParams, mid, lotteryCount, userAlreadyWinTimes)
	lotteryList, err = lotteryService.SimpleLottery(ctx, lotterySid, mid, lotteryParams, int(lotteryCount), int(userAlreadyWinTimes), false)
	if err != nil {
		switch err {
		case ecode.ActivityLotteryRiskInfo:
			err = ecode.BNJDrawNothing
		case ecode.ActivityWriteHandBlocked:
			err = nil
			lotteryList = make([]*lotteryModel.RecordDetail, 0)
		}

		if err != nil {
			return
		}
	}

	for idx, lotteryItem := range lotteryList {
		rewardIdStr, ok := lotteryItem.Extra["award_id"]
		if !ok {
			log.Errorc(ctx, "s.DoLottery lotteryItem got empty award_id, detail: %+v, Extra: %+v", lotteryItem, lotteryItem.Extra)
			continue
		}
		rewardId := 0
		rewardId, err = strconv.Atoi(rewardIdStr)
		if err != nil {
			log.Errorc(ctx, "s.DoLottery lottery id %v parse award_id fail: %v", lotteryItem.ID, err)
			continue
		}
		var info *api.RewardsSendAwardReply
		var sendErr error
		uniqueId := s.getLotteryAwardUniqueId(mid, lotteryItem, idx)
		if shouldSend {
			info, sendErr = rewards.Client.SendAwardByIdAsync(ctx, mid, uniqueId, "bnj2021Lottery1", int64(rewardId), updateCache, updateDB)
		} else {
			info, sendErr = rewards.Client.GetAwardSentInfoById(ctx, int64(rewardId), mid)
		}
		if sendErr != nil {
			log.Errorc(ctx, "s.DoLottery SendAwardByIdAsync error: mid: %v, uniqueId: %v, err: %v", mid, uniqueId, sendErr)
			err = sendErr
			continue
		}
		list = append(list, info)
	}

	//如果未中奖, 则补充Up主祝福
	if len(lotteryList) < int(lotteryCount) {
		missing := int(lotteryCount) - len(lotteryList)
		ts := time.Now().UnixNano()/1e6 - 1600000000000
		for i := 0; i < missing; i++ {
			var info *api.RewardsSendAwardReply
			var sendErr error
			uniqueId := fmt.Sprintf("niudan-append-%v-%v-%v-%v", mid, lotteryCount, ts, i)
			if shouldSend {
				info, sendErr = rewards.Client.SendAwardByIdAsync(ctx, mid, uniqueId, "bnj2021Lottery1", getRandomValueFromSlice(upGreetingAwardIds), updateCache, updateDB)
			} else {
				info, sendErr = rewards.Client.GetAwardSentInfoById(ctx, getRandomValueFromSlice(upGreetingAwardIds), mid)
			}
			if sendErr != nil {
				log.Errorc(ctx, "s.DoLottery SendAwardByIdAsync error: mid: %v, uniqueId: %v", mid, uniqueId)
				err = sendErr
				continue
			}
			if fidStr, ok := info.ExtraInfo["up_mid"]; ok {
				fid, err := strconv.ParseInt(fidStr, 10, 0)
				if err == nil {
					info.ExtraInfo["up_is_follow"] = s.dao.IsUserFollowedStr(ctx, mid, fid)
				}
			}
			list = append(list, info)
		}
	}

	// reset last draw award
	resetUserShareAward(ctx, mid, list)

	sc := s.GetConf()
	msg := &rewardModel.ActPlatActivityPoints{
		Points:    1,
		Timestamp: time.Now().Unix(),
		Mid:       mid,
		Source:    408933983,
		Activity:  sc.ActPlatActId,
		Business:  sc.ActPlatLotteryCounterName,
		Extra:     "",
	}
	errDataBus := s.actPlatDatabus.Send(ctx, fmt.Sprintf("%v-%v", mid, time.Now().Unix()), msg)
	if errDataBus != nil { //do not return error here
		log.Errorc(ctx, "s.DoLottery send actPlatDatabus error: %v", err)
	}
	return
}

func resetUserShareAward(ctx context.Context, mid int64, list []*api.RewardsSendAwardReply) {
	var (
		awardLevel int64
		awardName  string
	)

	for _, v := range list {
		if d, ok := v.ExtraInfo["level"]; ok {
			if level, err := strconv.ParseInt(d, 10, 64); err != nil {
				if awardLevel == 0 || level > awardLevel {
					awardLevel = level
					awardName = v.Name

					continue
				}
			}
		}
	}

	if awardLevel > 0 && awardName != "" {
		_ = newyear2021.ResetLastDrawAward(ctx, mid, awardName)
	}
}

// SceneID=1: 扭蛋抽奖
// SceneID=2: 直播间抽奖
// SceneID=3: 直播间预约抽奖
func (s *Service) GetAwardRecordByMid(ctx context.Context, mid, sceneId int64) (res []*rewardModel.AwardSentInfo, couponLeft int64, err error) {
	res = make([]*rewardModel.AwardSentInfo, 0)
	if currentUnix < BnjStrategyInfo.LiveStartTime && sceneId != 1 {
		return
	}

	activityIds := make([]int64, 0)
	conf := s.GetConf()
	switch sceneId {
	case 1:
		activityIds = append(activityIds, conf.NiuDanRewardsActivityId)
	case 2:
		activityIds = append(activityIds, conf.LiveRewardsActivityId1)
		activityIds = append(activityIds, conf.LiveRewardsActivityId2)
	case 3:
		activityIds = append(activityIds, conf.LiveRewardsActivityId1)
		activityIds = append(activityIds, conf.LiveRewardsActivityId2)
	default:
		err = fmt.Errorf("no such sceneId: %v", sceneId)
		return
	}
	couponLeft, err = s.FetchUserCoupon(ctx, mid, sceneId)
	if err != nil {
		return
	}
	if mid != 0 && len(activityIds) != 0 {
		res, err = rewards.Client.GetAwardRecordByMidAndActivityIdWithCache(ctx, mid, activityIds, 100)
	}
	return
}

// SceneID=1: 扭蛋抽奖
// SceneID=2: 直播间抽奖
// SceneID=3: 直播间预约抽奖
func (s *Service) FetchUserCoupon(ctx context.Context, mid int64, sceneId int64) (userCouponLeft int64, err error) {
	if mid == 0 {
		return
	}

	switch sceneId {
	case 1:
		u, fErr := newyear2021.FetchUserCoupon(ctx, mid)
		if fErr != nil {
			err = fErr
			return
		}
		userCouponLeft = u.ND
	case 2:
		userCouponLeft, err = newyear2021.FetchLiveLotteryQuota(ctx, mid, newyear2021.SceneID4LiveView)
	case 3:
		userCouponLeft, err = newyear2021.FetchLiveLotteryQuota(ctx, mid, newyear2021.SceneID4Reserve)
	}
	return

}
