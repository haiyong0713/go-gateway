package newyear2021

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	rewardModel "go-gateway/app/web-svr/activity/interface/model/rewards"
)

func (s *Service) isTaskAwardAlreadyReceived(ctx context.Context, mid, taskId, awardId int64, isDailyTask bool) (bool, error) {
	uniqueId := s.getTaskAwardUniqueId(mid, taskId, awardId, isDailyTask)
	return rewards.Client.IsAwardAlreadySend(ctx, mid, awardId, uniqueId)
}

func (s *Service) GetDailyTaskStatus(ctx context.Context, mid int64) (res *model.PersonalTaskResult, err error) {
	eg := errgroup.Group{}

	resMu := sync.Mutex{}
	taskConfigs := s.GetConf().TaskConfig
	res = &model.PersonalTaskResult{
		DailyTasks: make(map[string]*model.TaskResult, 0),
	}
	for _, taskConfig := range taskConfigs.DailyTasks.Tasks {
		tmpConfig := taskConfig.DeepCopy()
		eg.Go(func(ctx context.Context) (err error) {
			var count int64
			isFinished := false
			isReceived := false
			if mid != 0 {
				isDailyTask := true
				count, err = client.GetActPlatformCounterRes(ctx, mid, time.Now().Unix(), tmpConfig.ActPlatCounterId, tmpConfig.ActPlatId)
				if tmpConfig.HideOnFinish {
					//隐藏任务不是每日任务. 周期内只能完成一次, 所以读取Total
					count, err = client.GetActPlatformCounterTotal(ctx, mid, tmpConfig.ActPlatCounterId, tmpConfig.ActPlatId)
					isDailyTask = false
				}
				if err != nil {
					return err
				}
				isFinished = count >= tmpConfig.RequireCount
				isReceived = false
				if isFinished { //only check isReceived when task finished
					isReceived, err = s.isTaskAwardAlreadyReceived(ctx, mid, tmpConfig.Id, tmpConfig.AwardId, isDailyTask)
					if err != nil {
						return err
					}
				}
			}

			tmp := &model.TaskResult{
				Id:            tmpConfig.Id,
				Name:          tmpConfig.DisplayName,
				PcUrl:         tmpConfig.PcUrl,
				H5Url:         tmpConfig.H5Url,
				FinishCount:   count,
				RequiredCount: tmpConfig.RequireCount,
				IsFinish:      isFinished,
				IsReceived:    isReceived,
			}
			if tmpConfig.ActPlatCounterId == "ogv" { //大会员片单每日更新
				season, ok := taskConfigs.DailyTaskOgvSeasons[time.Now().Format("20060102")]
				if !ok {
					season = taskConfigs.DailyTaskOgvDefaultSeason
				}
				tmp.PcUrl = strings.Replace(tmp.PcUrl, "{{SEASON_ID}}", season, -1)
				tmp.H5Url = strings.Replace(tmp.H5Url, "{{SEASON_ID}}", season, -1)
			}

			a, err := s.getAwardInfoById(ctx, tmpConfig.AwardId)
			if err != nil {
				return
			}
			tmp.Award = a
			if tmp.FinishCount > tmpConfig.RequireCount {
				tmp.FinishCount = tmpConfig.RequireCount
			}
			resMu.Lock()
			res.DailyTasks[tmpConfig.ActPlatCounterId] = tmp
			resMu.Unlock()
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return res, err
	}
	return
}

func (s *Service) ReceivePersonalDailyReward(ctx context.Context, mid int64, taskId int64, debugSkipAllCheck bool) (info *api.RewardsSendAwardReply, err error) {
	var taskConfig *model.Task
	config := s.GetConf()
	taskConfigs := config.TaskConfig
	periodConfig := config.TimePeriod
	//0.检查当前是否在活动时间内
	nowUnix := time.Now().Unix()
	if nowUnix < periodConfig.Start.Time().Unix() {
		err = ecode.ActivityTaskNotStart
		return
	}

	if nowUnix > periodConfig.End.Time().Unix() {
		err = ecode.ActivityTaskOverEnd
		return
	}

	defer func() {
		if err != nil {
			log.Errorc(ctx, "ReceivePersonalDailyReward error: %v", err)
		}
	}()

	//1.检查任务是否完成
	{
		for _, tc := range taskConfigs.DailyTasks.Tasks {
			if tc.Id == taskId {
				taskConfig = tc
				break
			}
		}
		if taskConfig == nil {
			err = ecode.ActivityTaskNotExist
			return
		}

		var count int64
		count, err = client.GetActPlatformCounterRes(ctx, mid, time.Now().Unix(), taskConfig.ActPlatCounterId, taskConfig.ActPlatId)
		if err != nil {
			return
		}
		if count < taskConfig.RequireCount && !debugSkipAllCheck {
			err = ecode.ActivityTaskNotFinish
			return
		}
	}

	//2.检查奖励是否已经领取过
	isDailyTask := true
	if taskConfig.HideOnFinish {
		isDailyTask = false
	}
	received, err := s.isTaskAwardAlreadyReceived(ctx, mid, taskConfig.Id, taskConfig.AwardId, isDailyTask)
	if err != nil {
		return
	}
	if received && !debugSkipAllCheck {
		err = ecode.ActivityTaskHadAward
		return
	}

	if taskConfig.VipHidden { //付费隐藏任务
		paid := false
		paid, err = s.dao.IsUserPaid(ctx, mid, taskConfig.VipSuitID)
		if err != nil {
			err = ecode.BNJTooManyUser
			return
		}
		if !paid { //用户未付费
			err = ecode.BNJUserNotPaid
			return
		}
	}

	//3.发放奖励
	info, err = s.ReceiveAward(ctx, mid, "bnj2021Task2", taskConfig, isDailyTask)
	return

}

func (s *Service) GetLevelTaskStatus(ctx context.Context, mid int64) (res *model.LevelTaskStatus, err error) {
	config := s.GetConf()
	c := config.TaskConfig.LevelTask
	var count int64
	if mid != 0 {
		count, err = client.GetActPlatformCounterTotal(ctx, mid, c.ActPlatCounterId, c.ActPlatId)
		if err != nil {
			return
		}
	}

	res = &model.LevelTaskStatus{
		FinishCount: count,
		Tasks:       make([]*model.LevelTaskResult, 0, len(c.Stages)),
	}
	for _, stage := range c.Stages {
		stageRes := &model.LevelTaskResult{
			Id:            stage.Id,
			Name:          stage.DisplayName,
			RequiredCount: stage.RequireCount,
			PcUrl:         stage.PcUrl,
			H5Url:         stage.H5Url,
			IsFinish:      false,
			IsReceived:    false,
		}
		var userPaid bool
		if stage.VipHidden {
			userPaid, _ = s.dao.IsUserPaid(ctx, mid, stage.VipSuitID)
		}
		if count >= stage.RequireCount {
			if stage.VipHidden { //VIP隐藏任务, 付费才能完成
				stageRes.IsFinish = userPaid
			} else { //普通任务,count满足后即为完成
				stageRes.IsFinish = true
			}
			//完成条件满足后才验证是否领取
			if stageRes.IsFinish {
				isReceived, err := s.isTaskAwardAlreadyReceived(ctx, mid, stage.Id, stage.AwardId, false)
				if err != nil {
					return res, err
				}
				stageRes.IsReceived = isReceived
			}

		}
		var a *model.AwardInfo
		a, err = s.getAwardInfoById(ctx, stage.AwardId)
		if err != nil {
			return
		}
		stageRes.Award = a
		if stage.VipHidden { //VIP隐藏任务, 赋值到上一个任务的HiddenTask
			if userPaid { //只有在已付费用户可以看到此任务
				res.Tasks[len(res.Tasks)-1].HiddenTask = stageRes
			}
		} else {
			res.Tasks = append(res.Tasks, stageRes)
		}
	}
	return
}

func (s *Service) ReceivePersonalLevelReward(ctx context.Context, mid int64, taskId int64, debugSkipAllCheck bool) (info *api.RewardsSendAwardReply, err error) {
	var taskConfig *model.Task
	config := s.GetConf()
	taskConfigs := config.TaskConfig
	periodConfig := config.TimePeriod
	//0.检查当前是否在活动时间内
	nowUnix := time.Now().Unix()
	if nowUnix < periodConfig.Start.Time().Unix() {
		err = ecode.ActivityTaskNotStart
		return
	}

	if nowUnix > periodConfig.End.Time().Unix() {
		err = ecode.ActivityTaskOverEnd
		return
	}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "ReceivePersonalLevelReward error: %v", err)
		}
	}()

	//1.检查任务是否完成
	for _, st := range taskConfigs.LevelTask.Stages {
		if st.Id == taskId {
			taskConfig = st
			taskConfig.ActPlatCounterId = taskConfigs.LevelTask.ActPlatCounterId
			taskConfig.ActPlatId = taskConfigs.LevelTask.ActPlatId
		}
	}
	if taskConfig == nil {
		err = ecode.ActivityTaskNotExist
		return
	}

	if taskConfig.VipHidden { //付费隐藏任务
		paid := false
		paid, err = s.dao.IsUserPaid(ctx, mid, taskConfig.VipSuitID)
		if err != nil {
			err = ecode.BNJTooManyUser
			return
		}
		if !paid { //用户未付费
			err = ecode.BNJUserNotPaid
			return
		}
	}

	count, err := client.GetActPlatformCounterTotal(ctx, mid, taskConfig.ActPlatCounterId, taskConfig.ActPlatId)
	if err != nil {
		err = ecode.BNJTooManyUser
		return
	}
	if count < taskConfig.RequireCount && !debugSkipAllCheck {
		err = ecode.ActivityTaskNotFinish
		return
	}

	//2.检查奖励是否已经领取过
	received, err := s.isTaskAwardAlreadyReceived(ctx, mid, taskConfig.Id, taskConfig.AwardId, false)
	if err != nil {
		err = ecode.BNJTooManyUser
		return
	}
	if received && !debugSkipAllCheck {
		err = ecode.ActivityTaskHadAward
		return
	}

	//3.发放奖励
	info, err = s.ReceiveAward(ctx, mid, "bnj2021Task3", taskConfig, false)
	return
}

func (s *Service) getAwardInfoById(ctx context.Context, awardId int64) (res *model.AwardInfo, err error) {
	c, err := rewards.Client.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return res, err
	}
	return &model.AwardInfo{
		AwardName: c.Name,
		Type:      c.Type,
		Icon:      c.IconUrl,
		ExtraInfo: c.ExtraInfo,
	}, nil
}

func (s *Service) PubMallVisit(ctx context.Context, mid int64) (err error) {
	sc := s.GetConf()
	msg := &rewardModel.ActPlatActivityPoints{
		Points:    1,
		Timestamp: time.Now().Unix(),
		Mid:       mid,
		Source:    408933983,
		Activity:  sc.ActPlatActId,
		Business:  sc.ActPlatMallCounterName,
		Extra:     "",
	}
	err = s.actPlatDatabus.Send(ctx, fmt.Sprintf("%v-%v", mid, time.Now().Unix()), msg)
	if err != nil { //do not return error here
		log.Errorc(ctx, "s.DoLottery send actPlatDatabus error: %v", err)
	}
	return
}
