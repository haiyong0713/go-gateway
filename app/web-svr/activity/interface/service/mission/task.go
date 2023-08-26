package mission

import (
	"context"
	xsql "database/sql"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"time"
)

// getActivityTaskStatByTaskId 获取活动详情、任务信息以及当前库存使用信息
func (s *Service) getActivityTaskStatByTaskId(ctx context.Context, actId int64, taskId int64, period int64) (activity *v1.MissionActivityDetail, taskStat *mission.ActivityTaskStatInfo, err error) {
	activity, err = s.getMissionActivityInfo(ctx, actId, false, false)
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskStatByTaskId][getMissionActivityInfo][Error], err:%+v", err)
		return
	}
	taskDetail, err := s.GetMissionTaskDetail(ctx, &v1.GetMissionTaskDetailReq{
		TaskId: taskId,
	})
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskStatByTaskId][GetMissionTaskDetail][Error], err:%+v", err)
		return
	}
	taskStat, err = s.getActivityTaskStatWithStock(ctx, activity, taskDetail, period, false)
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskStatByTaskId][getActivityTaskStat][Error], err:%+v", err)
		return
	}
	return
}

// getUserReceiveInfo 获取用户的领取信息，如果不存在返回err, 否则返回其对应的领取记录详情
func (s *Service) getUserReceiveInfo(ctx context.Context, mid int64, actId int64, receiveId int64) (receiveInfo *mission.UserCompleteRecord, err error) {
	receiveInfo, err = s.dao.GetUserReceiveCache(ctx, mid, actId, receiveId)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[getUserReceiveInfo][GetUserReceiveCache][Error], err:%+v", err)
		return
	}
	if err == nil {
		return
	} else {
		err = nil
	}
	receiveInfo, err = s.dao.GetUserReceiveInfo(ctx, mid, actId, receiveId)
	if err != nil {
		log.Errorc(ctx, "[getUserReceiveInfo][GetUserReceiveCache][Error], err:%+v", err)
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "领取信息不存在")
			return
		}
		return
	}
	err = s.dao.SetUserReceiveCache(ctx, receiveInfo)
	if err != nil {
		log.Errorc(ctx, "[getUserReceiveInfo][SetUserReceiveCache][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) GetMissionTaskDetail(ctx context.Context, req *v1.GetMissionTaskDetailReq) (resp *v1.MissionTaskDetail, err error) {
	resp = new(v1.MissionTaskDetail)
	if req == nil || req.TaskId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	cacheValue, err := s.activityTaskInfoCache.Get(req.TaskId)
	if err != nil {
		log.Errorc(ctx, "[GetMissionTaskDetail][CacheGet][Error], err:%+v", err)
	}
	resp, ok := cacheValue.(*v1.MissionTaskDetail)
	if ok {
		return
	}
	// 获取任务详情信息
	resp, err = s.dao.GetTaskDetailByTaskId(ctx, req.TaskId)
	if err != nil {
		log.Errorc(ctx, "[GetMissionTaskDetail][GetTaskDetailByTaskId][Error], err:%+v", err)
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "任务不存在")
		}
		return
	}
	return
}

func (s *Service) GetActivityTaskStatByGroupId(ctx context.Context, groupId int64) (activity *v1.MissionActivityDetail, taskStat *mission.ActivityTaskStatInfo) {
	activity, task := s.getActivityTaskByGroupId(ctx, groupId)
	if activity == nil || activity.Id == 0 || task == nil || task.TaskId == 0 {
		return
	}
	taskStat, _ = s.getActivityTaskStatNoStock(ctx, activity, task, 0, true)
	return
}

func (s *Service) getActivityTaskStatNoStockByTime(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, timestamp int64) (taskStat *mission.ActivityTaskStatInfo, err error) {
	return
}

// getActivityTaskStatNoStock 返回活动周期对应的周期时间 ， 库存周期及时间，不包含库存的动态信息
func (s *Service) getActivityTaskStatWithStock(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, period int64, isNowPeriod bool) (taskStat *mission.ActivityTaskStatInfo, err error) {
	if task == nil {
		return
	}
	periodStat, err := s.calculateTaskPeriod(ctx, activity, task, period, isNowPeriod)
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskStatNoStock][calculatePeriod][Error], err:%+v", err)
		return
	}
	resp, err := s.stockSvr.GetStocksByIds(ctx, &v1.GetStocksReq{
		StockIds:  []int64{task.StockId},
		SkipCache: false,
	})
	if err != nil || resp == nil {
		log.Errorc(ctx, "[getActivityTaskStatWithStock][GetStocksByIds][Error], err:%+v, resp:%+v", err, resp)
		err = xecode.Errorf(xecode.ServerErr, "库存信息获取失败")
		return
	}
	itemList := resp.StockMap[task.StockId]
	if itemList == nil || itemList.List == nil || len(itemList.List) != 1 {
		log.Errorc(ctx, "[getActivityTaskStatNoStock][GetStocksByIds][Error][NoMatch], err:%+v, resp:%+v", err, resp)
		err = xecode.Errorf(xecode.ServerErr, "库存信息获取失败")
		return
	}
	item := itemList.List[0]
	stockStat := &mission.TaskStockStat{
		CycleType: int64(item.CycleLimitObj.CycleType),
		LimitType: int64(item.CycleLimitObj.LimitType),
		Total:     int64(item.LimitNum),
		Consumed:  int64(item.LimitNum - item.StockNum),
	}
	stockStat.StockPeriod, err = s.calculateReceivePeriod(stockStat.CycleType)
	stockStat.StockBeginTime, stockStat.StockEndTime, err = s.calculateStockPeriod(ctx, activity, task, item.CycleLimitObj)

	taskStat = &mission.ActivityTaskStatInfo{
		TaskDetail: task,
		PeriodStat: periodStat,
		StockStat:  stockStat,
	}
	return
}

// getActivityTaskStatNoStock 返回活动周期对应的周期时间 ， 库存周期及时间，不包含库存的动态信息
func (s *Service) getActivityTaskStatNoStock(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, period int64, isNowPeriod bool) (taskStat *mission.ActivityTaskStatInfo, err error) {
	if task == nil {
		return
	}
	periodStat, err := s.calculateTaskPeriod(ctx, activity, task, period, isNowPeriod)
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskStatNoStock][calculatePeriod][Error], err:%+v", err)
		return
	}
	resp, err := s.stockSvr.QueryStockRecord(ctx, task.StockId, false)
	if err != nil || resp == nil {
		log.Errorc(ctx, "[getActivityTaskStatNoStock][GetStocksByIds][Error], err:%+v, resp:%+v", err, resp)
		err = xecode.Errorf(xecode.ServerErr, "库存信息获取失败")
		return
	}
	itemList := resp.CycleLimitObj
	if itemList == nil || len(itemList) != 1 {
		log.Errorc(ctx, "[getActivityTaskStatNoStock][GetStocksByIds][Error][NoMatch], err:%+v, resp:%+v", err, resp)
		err = xecode.Errorf(xecode.ServerErr, "库存信息获取失败")
		return
	}
	item := itemList[0]
	stockStat := &mission.TaskStockStat{
		CycleType: int64(item.CycleType),
		LimitType: int64(item.LimitType),
	}
	stockStat.StockBeginTime, stockStat.StockEndTime, err = s.calculateStockPeriod(ctx, activity, task, item)

	taskStat = &mission.ActivityTaskStatInfo{
		TaskDetail: task,
		PeriodStat: periodStat,
		StockStat:  stockStat,
	}
	return
}

// getActivityTaskByGroupId 通过groupId获取其对应的活动详情，任务详情
func (s *Service) getActivityTaskByGroupId(ctx context.Context, groupId int64) (activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail) {
	cacheValue, err := s.groupTaskMappingCache.Get(groupId)
	if err != nil {
		log.Warnc(ctx, "[getActivityTaskByGroupId][Error], err:%+v", err)
		err = nil
	}
	task, ok := cacheValue.(*v1.MissionTaskDetail)
	if !ok {
		log.Warnc(ctx, "[getActivityTaskByGroupId][Assert][False], err:%+v", err)
	}
	actId, err := s.dao.GetActivityTaskIdCacheByGroupId(ctx, groupId)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[getActivityTaskByGroupId][GetActivityTaskIdCacheByGroupId][Error], err:%+v", err)
		return
	}
	if err != nil {
		err = nil
		// 回源
		actId, err = s.dao.GetMappingByGroupId(ctx, groupId)
		if err != nil {
			log.Errorc(ctx, "[getActivityTaskByGroupId][GetMappingByGroupId][Error], err:%+v", err)
			return
		}
		_ = s.dao.SetActivityTaskIdCacheByGroupId(ctx, groupId, actId)
	}
	if actId == 0 {
		return
	}
	activity, err = s.getMissionActivityInfo(ctx, actId, false, false)
	if err != nil {
		return
	}
	task, err = s.getTaskByActIdGroupId(ctx, actId, groupId)
	if err != nil {
		return
	}
	return
}

func (s *Service) GroupConsumerForTaskComplete(ctx context.Context, req *v1.GroupConsumerForTaskCompleteReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req == nil || req.GroupId == 0 || req.Mid == 0 {
		log.Warnc(ctx, "[GroupConsumerForTaskComplete][Params][Warns], req:%+v", err)
		return
	}
	activity, task := s.getActivityTaskByGroupId(ctx, req.GroupId)
	if activity == nil || task == nil || activity.Id == 0 || task.TaskId == 0 {
		// 节点组无绑定的任务
		return
	}
	if check := s.activityTimeCheck(activity); check != nil {
		return
	}

	groupIds := make([]int64, 0)
	for _, v := range task.Groups {
		if v.GroupId == req.GroupId && req.Total < v.CompleteScore {
			return
		}
		groupIds = append(groupIds, v.GroupId)
	}

	res, err := s.reserveSvr.ActivityProgress(ctx, &v1.ActivityProgressReq{
		Sid:  activity.GroupsActId,
		Gids: groupIds,
		Type: 2,
		Mid:  req.Mid,
		Time: req.Timestamp,
	})
	if err != nil {
		log.Errorc(ctx, "[GroupConsumerForTaskComplete][ActivityProgress][Error], err:%+v", err)
		return
	}

	groupSchedule := make([]*mission.GroupSchedule, 0)
	for _, group := range task.Groups {
		groupInfo, ok := res.Groups[group.GroupId]
		if !ok {
			return
		}
		if groupInfo.Total < group.CompleteScore {
			return
		}
		groupSchedule = append(groupSchedule, &mission.GroupSchedule{
			TaskId:           task.TaskId,
			ActId:            activity.Id,
			GroupId:          group.GroupId,
			GroupBaseNum:     group.CompleteScore,
			GroupCompleteNum: group.CompleteScore,
		})
	}
	periodStat, err := formatPeriodStatByTime(ctx, activity, task, req.Timestamp)
	if err != nil {
		log.Errorc(ctx, "[GroupConsumerForTaskComplete][formatPeriodStatByTime][Error], err:%+v, req:%+v", err, req)
		return
	}
	taskStat, err := s.getActivityTaskStatNoStock(ctx, activity, task, periodStat.Period, true)
	if err != nil {
		log.Errorc(ctx, "[singleTaskCompleteHandler][getActivityTaskStatNoStock][Error], err:%+v", err)
		return
	}
	receivePeriod, err := s.calculateReceivePeriod(taskStat.StockStat.LimitType)
	if err != nil {
		log.Errorc(ctx, "[singleTaskCompleteHandler][calculateReceivePeriod][Error], err:%+v", err)
		return
	}
	userTask := &mission.UserTaskDetail{
		ID:            task.TaskId,
		ActId:         activity.Id,
		GroupList:     groupSchedule,
		TaskPeriod:    periodStat.Period,
		RewardId:      task.RewardId,
		RewardType:    task.TaskPeriod,
		ReceivePeriod: receivePeriod,
	}
	err = s.singleTaskCompleteHandler(ctx, req.Mid, userTask)
	if err != nil {
		return
	}
	return
}

func (s *Service) getTaskByActIdGroupId(ctx context.Context, actId int64, groupId int64) (task *v1.MissionTaskDetail, err error) {
	tasks, err := s.getActivityTasks(ctx, actId, false, false)
	if err != nil {
		log.Errorc(ctx, "[getActivityTaskByGroupId][GetActivityTaskIdCacheByGroupId][Error], err:%+v", err)
		return
	}
	for _, item := range tasks {
		for _, group := range item.Groups {
			if group.GroupId == groupId {
				task = item
				return
			}
		}
	}
	return
}

func (s *Service) GetUserTasks(ctx context.Context, actId int64, mid int64) (userTasks []*mission.UserTaskDetail, err error) {
	activity, tasks, err := s.getActivityTasksWithStats(ctx, actId)
	if err != nil {
		log.Errorc(ctx, "[GetUserTasks][getActivityTasksWithStats][Error], err:%+v", err)
		return
	}
	groupIds := make([]int64, 0)
	for _, task := range tasks {
		for _, group := range task.TaskDetail.Groups {
			groupIds = append(groupIds, group.GroupId)
		}
	}
	resp := new(v1.ActivityProgressReply)
	if mid != 0 && len(groupIds) > 0 {
		resp, err = s.reserveSvr.ActivityProgress(ctx, &v1.ActivityProgressReq{
			Sid:  activity.GroupsActId,
			Gids: groupIds,
			Type: 2,
			Mid:  mid,
			Time: time.Now().Unix(),
		})
		if err != nil {
			log.Errorc(ctx, "[GetUserTasks][ActivityProgress][Error], err:%+v", err)
			return
		}
	}

	userTasks = make([]*mission.UserTaskDetail, 0)
	tasksMapping := make(map[int64]*v1.MissionTaskDetail)
	for _, v := range tasks {
		rewardResp, errG := rewards.Client.GetAwardConfigById(ctx, v.TaskDetail.RewardId)
		if errG != nil {
			err = errG
			log.Errorc(ctx, "[GetUserTasks][GetAwardConfigById][Error], err:%+v", errG)
			return
		}
		rewardInfo := &mission.RewardInfo{
			RewardId:    rewardResp.Id,
			RewardName:  rewardResp.Name,
			RewardIcon:  rewardResp.IconUrl,
			RewardActId: rewardResp.ActivityId,
			Type:        rewardResp.Type,
			Awards: []*mission.AwardInfo{
				{
					AwardId:   rewardResp.Id,
					AwardName: rewardResp.Name,
					AwardIcon: rewardResp.IconUrl,
					AwardType: rewardResp.Type,
				},
			},
		}
		userTask := &mission.UserTaskDetail{
			ID:                      v.TaskDetail.TaskId,
			ActId:                   actId,
			GroupList:               s.formatGroupsSchedule(v.TaskDetail, resp.Groups),
			TaskPeriod:              v.PeriodStat.Period,
			RewardId:                v.TaskDetail.RewardId,
			RewardType:              v.StockStat.LimitType,
			RewardInfo:              rewardInfo,
			RewardPeriodPoolNum:     v.StockStat.Total,
			RewardPeriodReceivedNum: v.StockStat.Consumed,
			RewardPeriodStockNum:    v.StockStat.Total - v.StockStat.Consumed,
			RewardReceiveBeginTime:  v.StockStat.StockBeginTime,
			RewardReceiveEndTime:    v.StockStat.StockEndTime,
			ReceiveId:               0,
			ReceiveStatus:           0,
			ReceivePeriod:           v.StockStat.StockPeriod,
		}
		userTasks = append(userTasks, userTask)
		tasksMapping[v.TaskDetail.TaskId] = v.TaskDetail
	}
	if check := s.activityTimeCheck(activity); check != nil {
		return
	}
	err = s.completeTasksHandler(ctx, mid, userTasks)
	if err != nil {
		return
	}
	return
}

func (s *Service) completeTasksHandler(ctx context.Context, mid int64, userTasks []*mission.UserTaskDetail) (err error) {
	if mid == 0 {
		return
	}
	errgroup2 := errgroup.WithContext(ctx)
	for _, v := range userTasks {
		userTask := v
		errgroup2.Go(func(ctx context.Context) (err error) {
			err = s.singleTaskCompleteHandler(ctx, mid, userTask)
			if err != nil {
				log.Errorc(ctx, "[singleTaskCompleteHandler][Error] userTask:%+v, err:%+v", v, err)
				return
			}
			return
		})
	}
	err = errgroup2.Wait()
	if err != nil {
		log.Errorc(ctx, "[completeTasksHandler][errgroup2][Error] err:%+v", err)
		return
	}
	return
}

func (s *Service) singleTaskCompleteHandler(ctx context.Context, mid int64, userTask *mission.UserTaskDetail) (err error) {
	for _, schedule := range userTask.GroupList {
		if schedule.GroupCompleteNum < schedule.GroupBaseNum {
			return
		}
	}

	complete, err := s.dao.GetUserCompleteTaskCache(ctx, mid, userTask.ActId, userTask.ID, userTask.TaskPeriod, 0)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[singleTaskCompleteHandler][GetUserCompleteTaskCache][Error], err:%+v", err)
		return
	}
	if err == nil {
		userTask.ReceiveId = complete.ID
		userTask.ReceiveStatus = complete.TaskRewardsStatus
		return
	}
	complete, err = s.dao.GetUserCompleteTaskFromDB(ctx, mid, userTask.ActId, userTask.ID, userTask.TaskPeriod, userTask.ReceivePeriod)
	if err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[singleTaskCompleteHandler][GetUserCompleteTaskFromDB][Error], err:%+v", err)
		return
	}
	if err == nil && complete != nil {
		userTask.ReceiveId = complete.ID
		userTask.ReceiveStatus = complete.TaskRewardsStatus
		err = s.dao.SetUserReceiveCache(ctx, complete)
		if err != nil {
			log.Errorc(ctx, "[singleTaskCompleteHandler][SetUserCompleteTaskCache][Error], err:%+v", err)
			return
		}
		return
	}
	// 插入完成记录
	receiveId, err := s.dao.InsertUserCompleteTaskRecord(ctx, mid, userTask)
	if err != nil {
		log.Errorc(ctx, "[singleTaskCompleteHandler][InsertUserCompleteTaskRecord][Error], err:%+v", err)
		return
	}
	userTask.ReceiveId = receiveId
	userTask.ReceiveStatus = mission.TaskRewardStatusIdle
	err = s.dao.SetUserReceiveCache(ctx, complete)
	if err != nil {
		log.Errorc(ctx, "[singleTaskCompleteHandler][SetUserCompleteTaskCache][AfterInsert][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) formatGroupsSchedule(task *v1.MissionTaskDetail, groupProgressMapping map[int64]*v1.ActivityProgressGroup) (groupSchedules []*mission.GroupSchedule) {
	groupSchedules = make([]*mission.GroupSchedule, 0)
	groups := task.Groups
	if len(groups) == 0 {
		return
	}
	for _, group := range groups {
		completeNum := int64(0)
		_, ok := groupProgressMapping[group.GroupId]
		if ok {
			completeNum = groupProgressMapping[group.GroupId].Total
			if completeNum > group.CompleteScore {
				completeNum = group.CompleteScore
			}
		}
		if s.conf.MissionActivityConf != nil && s.conf.MissionActivityConf.SpecialActivityId != 0 && task.ActId == s.conf.MissionActivityConf.SpecialActivityId {
			completeNum = group.CompleteScore
		}
		groupSchedule := &mission.GroupSchedule{
			TaskId:           task.TaskId,
			ActId:            task.ActId,
			GroupId:          group.GroupId,
			GroupBaseNum:     group.CompleteScore,
			GroupCompleteNum: completeNum,
		}
		groupSchedules = append(groupSchedules, groupSchedule)
	}
	return
}

func (s *Service) getUserCompleteRecordBySerialNum(ctx context.Context, mid int64, actId int64, taskId int64, serialNum string) (completeRecord *mission.UserCompleteRecord, err error) {
	completeRecord, err = s.dao.GetUserCompleteRecordCacheBySerialNum(ctx, mid, serialNum)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[getUserCompleteRecordBySerialNum][GetUserCompleteRecordCacheBySerialNum][AfterInsert][Error], err:%+v", err)
		return
	}
	if err == nil {
		return
	}
	completeRecord, err = s.dao.GetUserReceiveInfoBySerialNum(ctx, actId, taskId, mid, serialNum)
	if err != nil {
		log.Errorc(ctx, "[getUserCompleteRecordBySerialNum][GetUserCompleteRecordCacheBySerialNum][AfterInsert][Error], err:%+v", err)
		return
	}
	_ = s.dao.SetUserCompleteRecordCacheBySerialNum(ctx, completeRecord)
	return
}

func (s *Service) GetActivityReceiveRecordsByStatus(ctx context.Context, actId int64, tableIndex int32, status int64) (records []*mission.UserCompleteRecord, err error) {
	records, err = s.dao.GetActivityReceiveRecordsByStatus(ctx, int64(tableIndex), actId, status, time.Now().Unix()-3600, time.Now().Unix()-1)
	if err != nil {
		log.Errorc(ctx, "[GetActivityReceiveRecordsByStatus][GetActivityReceiveRecordsByStatus][Error], err:%+v", err)
		return
	}
	return
}
