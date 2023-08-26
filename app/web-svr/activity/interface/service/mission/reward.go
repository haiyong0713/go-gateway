package mission

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	model "go-gateway/app/web-svr/activity/interface/model/mission"

	oauth2 "git.bilibili.co/bapis/bapis-go/account/service/oauth2"
)

func (s *Service) stockCheck(a *model.TaskStockStat) (err error) {
	now := time.Now()
	if a.StockBeginTime > now.Unix() {
		err = ecode.MissionActivityReceiveTimeError
		return
	}

	if a.StockEndTime < now.Unix() {
		err = ecode.MissionActivityReceiveTimeError
		return
	}

	if int64(v1.StockServerLimitType_StoreUpperLimit) == a.LimitType && a.Consumed >= a.Total {
		err = ecode.MissionActivityNoStockError
	}
	return
}

func (s *Service) activityTimeCheck(a *v1.MissionActivityDetail) (err error) {
	now := time.Now()

	if a.BeginTime.Time().Unix() > now.Unix() {
		err = ecode.MissionActivityNotStartError
		return
	}
	if a.EndTime.Time().Unix() < now.Unix() {
		err = ecode.MissionActivityEndError
		return
	}
	if a.Status != model.ActivityNormalStatus {
		err = ecode.MissionActivityEndError
		return
	}
	return

}

func (s *Service) generateTaskUniqueId(mid, actId, taskId int64, taskStat *model.ActivityTaskStatInfo) string {
	// 当前不考虑taskStat.StockStat.StockPeriod, 所以默认为0
	return fmt.Sprintf("mis-%v-%v-%v-%v-%v", mid, actId, taskId, taskStat.PeriodStat.Period, 0)
}

func (s *Service) parseTaskUniqueId(uniqueId string) (mid, actId, taskId, taskPeriod, stockPeriod int64, err error) {
	defer func() {
		err = errors.Wrap(err, "parseTaskUniqueId")
	}()
	fields := strings.Split(uniqueId, "-")
	if len(fields) < 6 {
		err = fmt.Errorf("unknown uniqueId")
		return
	}
	mid, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return
	}
	actId, err = strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return
	}
	taskId, err = strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return
	}
	taskPeriod, err = strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return
	}
	stockPeriod, err = strconv.ParseInt(fields[5], 10, 64)
	if err != nil {
		return
	}
	return
}

func (s *Service) ReceiveTaskAward(ctx context.Context, mid, actId, taskId, receiveId int64) (rewardInfo *v1.RewardsSendAwardReply, err error) {
	// 1.校验活动id,receiveId
	info, err := s.getUserReceiveInfo(ctx, mid, actId, receiveId)
	if err != nil {
		log.Errorc(ctx, "getUserReceiveInfo error: %v", err)
		return
	}
	if info.TaskRewardsStatus == model.TaskRewardStatusSuccess {
		err = ecode.ActivityTaskHadAward
		return
	}
	timeNow := time.Now().Unix()
	activity, taskStat, err := s.getActivityTaskStatByTaskId(ctx, actId, taskId, info.CompletePeriod)
	if err != nil {
		log.Errorc(ctx, "getActivityTaskStatByTaskId error: %v", err)
		return
	}
	// 2.时间校验、当前库存校验
	{
		err = s.activityTimeCheck(activity)
		if err != nil {
			log.Errorc(ctx, "activityTimeCheck error: %v", err)
			return
		}
		if info.CompletePeriod != taskStat.PeriodStat.Period {
			err = ecode.MissionActivityReceiveTimeError
			return
		}
		err = s.stockCheck(taskStat.StockStat)
		if err != nil {
			log.Errorc(ctx, "periodTimeCheck error: %v", err)
			return
		}
	}

	// 3.校验是否需要用户绑定手机号
	{
		bindPhone := int32(0)
		if activity.BindPhoneCheck != 0 {
			// 校验用户是否绑定了手机
			bindPhone, err = s.dao.GetUserPhoneBind(ctx, mid)
			if err != nil {
				err = xecode.Errorf(xecode.RequestErr, "用户账号信息获取失败")
				return
			}
			if bindPhone == 0 {
				err = xecode.Errorf(xecode.RequestErr, "绑定手机号后才能参与活动")
				return
			}
		}
	}

	// 4.查询发奖状态
	uniqueId := s.generateTaskUniqueId(mid, actId, taskId, taskStat)
	{
		//4.1 校验是否已领取
		var send *v1.RewardsCheckSentStatusResp
		send, err = rewards.Client.RewardsCheckSentStatusReq(ctx, &v1.RewardsCheckSentStatusReq{
			Mid:      mid,
			UniqueId: uniqueId,
			AwardId:  taskStat.TaskDetail.RewardId,
		})
		if err != nil {
			log.Errorc(ctx, "rewards.IsAwardAlreadySend error: %v", err)
			return
		}
		if send.Result {
			err = ecode.ActivityTaskHadAward
			return
		}
		// 4.2 校验用户是否满足发放条件
		err = rewards.Client.AwardSendPreCheck(ctx, mid, uniqueId, "mission", taskStat.TaskDetail.RewardId)
		if err != nil {
			return
		}
	}

	var stockNo string
	//5.进行发放
	{

		//5.1 扣减库存
		stockNo, err = s.stockSvr.ConsumerSingleStockById(ctx, &v1.ConsumerSingleStockReq{
			StockId: taskStat.TaskDetail.StockId,
			RetryId: uniqueId,
			Ts:      timeNow,
			Mid:     mid,
		})
		if err != nil {
			return
		}

		//5.2 标记任务领取中
		err = s.dao.UpdateUserReceiveRecordToStarting(ctx, info, taskStat.PeriodStat.Period, uniqueId, stockNo)
		if err != nil {
			log.Errorc(ctx, "dao.UpdateUserReceiveRecordToStarting error: %v", err)
			return
		}
		info.TaskRewardsStatus = model.TaskRewardStatusIn
		_ = s.dao.SetUserReceiveCache(ctx, info)
		// 5.3.调用奖励发放
		rewardInfo, err = rewards.Client.SendAwardByIdAsync(ctx, mid, uniqueId, "mission", taskStat.TaskDetail.RewardId, true, true)
		if err != nil {
			return
		}
		info.SerialNum = uniqueId
		err = s.dao.UpdateUserReceiveRecordToFinish(ctx, info, taskStat.PeriodStat.Period)
		if err != nil {
			log.Errorc(ctx, "dao.UpdateUserReceiveRecordToFinish error: %v", err)
			err = nil
			return
		}
		_ = s.dao.SetUserReceiveCache(ctx, info)
		_ = s.dao.SetUserCompleteRecordCacheBySerialNum(ctx, info)
	}
	//库存异步提交
	{
		if component.StockServerSyncProducer != nil {
			msg := &v1.StockServerSyncStruct{
				StockId:     taskStat.TaskDetail.StockId,
				Ts:          timeNow,
				StockOrders: []string{stockNo},
			}
			err1 := component.StockServerSyncProducer.Send(ctx, uniqueId, msg)
			if err1 != nil {
				log.Errorc(ctx, "commit StockServerSyncProducer message error: %v", err1)
			}
		}
	}
	return
}

func (s *Service) CheckTencentGameAward(ctx context.Context, taskId int64, openId, serialNum string) (ok bool, err error) {
	// get mid by taskId and openId
	task, err := s.GetMissionTaskDetail(ctx, &v1.GetMissionTaskDetailReq{
		TaskId: taskId,
	})
	if err != nil {
		err = errors.Wrap(err, "s.GetMissionTaskDetail error")
		return
	}

	accountInfoId, err := rewards.Client.GetTencentAwardAccountId(ctx, task.RewardId)
	if err != nil {
		err = errors.Wrap(err, "rewards.Client.GetTencentAwardAccountId")
		return
	}

	_, gameConfig, err := s.bindSvr.GetBindAndGameConfig(ctx, accountInfoId)
	if err != nil {
		err = errors.Wrap(err, "s.bindSvr.GetBindAndGameConfig")
		return
	}

	midReply, err := client.BiliOAuth2Client.MidByOpenID(ctx, &oauth2.MidByOpenIDReq{
		Oauth2Appkey: gameConfig.ClientId,
		Openid:       openId,
	})
	if err != nil {
		err = errors.Wrap(err, "client.BiliOAuth2Client.MidByOpenID")
		return
	}

	state, err := rewards.Client.RewardsCheckSentStatusReq(ctx, &v1.RewardsCheckSentStatusReq{
		Mid:      midReply.Mid,
		UniqueId: serialNum,
		AwardId:  task.RewardId,
	})
	if err != nil {
		err = errors.Wrap(err, "rewards.Client.RewardsCheckSentStatusReq")
		return
	}
	if !state.Result {
		ok = false
		return
	}
	uniqueId, err := rewards.Client.GetParentUniqueId(ctx, task.RewardId, serialNum)
	if err != nil {
		err = errors.Wrap(err, "rewards.Client.GetParentUniqueId")
		return
	}
	// check task status by mid and uniqueId
	completeRecord, err := s.getUserCompleteRecordBySerialNum(ctx, midReply.Mid, task.ActId, taskId, uniqueId)
	if err != nil {
		err = errors.Wrap(err, "s.getUserCompleteRecordBySerialNum")
	}
	if err != nil || completeRecord == nil {
		ok = false
		return
	}
	ok = true
	return

}

func (s *Service) CheckStock(ctx context.Context, in *v1.MissionCheckStockReq) (resp *v1.MissionCheckStockResp, err error) {
	resp = &v1.MissionCheckStockResp{}
	mid, actId, taskId, _, _, err := s.parseTaskUniqueId(in.UniqueId)
	if err != nil {
		err = xecode.Error(xecode.RequestErr, err.Error())
		return
	}
	resp.Status, err = s.dao.CheckStock(ctx, mid, actId, taskId, in.UniqueId, in.StockNo)
	return
}

func (s *Service) MakeUpRewards(ctx context.Context, receiveId int64, mid int64, actId int64) (err error) {
	info, err := s.dao.GetUserReceiveInfo(ctx, mid, actId, receiveId)
	if err != nil {
		log.Errorc(ctx, "getUserReceiveInfo error: %v", err)
		return
	}
	if info.TaskRewardsStatus != model.TaskRewardStatusIn || info.SerialNum == "" {
		log.Warnc(ctx, "[MakeUpRewards][StatusNotMatch] info:%+v", info)
		return
	}
	_, taskStat, err := s.getActivityTaskStatByTaskId(ctx, actId, info.TaskId, info.CompletePeriod)
	if err != nil {
		log.Errorc(ctx, "getActivityTaskStatByTaskId error: %v", err)
		return
	}
	var send *v1.RewardsCheckSentStatusResp
	send, err = rewards.Client.RewardsCheckSentStatusReq(ctx, &v1.RewardsCheckSentStatusReq{
		Mid:      mid,
		UniqueId: info.SerialNum,
		AwardId:  taskStat.TaskDetail.RewardId,
	})
	if err != nil {
		log.Errorc(ctx, "rewards.IsAwardAlreadySend error: %v", err)
		return
	}
	if send.Result {
		// 已发放
		err = s.updateReceiveFinish(ctx, info, taskStat)
		return
	}
	// 调用奖励发放
	_, err = rewards.Client.SendAwardByIdAsync(ctx, mid, info.SerialNum, "mission", taskStat.TaskDetail.RewardId, true, true)
	if err != nil {
		return
	}
	err = s.updateReceiveFinish(ctx, info, taskStat)
	return
}

func (s *Service) updateReceiveFinish(ctx context.Context, info *model.UserCompleteRecord, taskStat *model.ActivityTaskStatInfo) (err error) {
	err = s.dao.UpdateUserReceiveRecordToFinish(ctx, info, taskStat.PeriodStat.Period)
	if err != nil {
		log.Errorc(ctx, "dao.UpdateUserReceiveRecordToFinish error: %v", err)
		err = nil
		return
	}
	_ = s.dao.SetUserReceiveCache(ctx, info)
	_ = s.dao.SetUserCompleteRecordCacheBySerialNum(ctx, info)
	return
}
