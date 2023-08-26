package mission

import (
	"context"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"time"
)

const (
	_timePeriodTemplate       = "20060102"
	_DailyPeriodTemplate      = "20060102 15:04:05"
	_MonthlyPeriodTemplate    = "200601"
	_DailyPeriodExtraTemplate = "15:04:05"
)

func (s *Service) getMissionActivityInfo(ctx context.Context, actId int64, skipLocal bool, skipCache bool) (activityDetail *v1.MissionActivityDetail, err error) {
	if !skipLocal {
		activityDetail, err = s.getMissionActivityFromLocal(ctx, actId)
		if err == nil {
			return
		}
	}
	if !skipCache {
		activityDetail, err = s.dao.GetActivityDetailCache(ctx, actId)
		if err != nil && err != redis.ErrNil {
			log.Errorc(ctx, "[Service][getMissionActivityInfo][GetActivityDetailCache] err:%+v", err)
			return
		}
		if err == nil {
			return
		}
	}
	activityDetail, err = s.dao.GetActivityInfo(ctx, actId)
	if err != nil {
		log.Errorc(ctx, "[Service][getMissionActivityInfo][GetActivityInfo] err:%+v", err)
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "活动不存在")
		}
		return
	}
	_ = s.dao.SetActivityDetailCache(ctx, activityDetail)
	return
}

func (s *Service) GetValidActivityIds(ctx context.Context, skipCache bool) (validIds []int64, err error) {
	if !skipCache {
		validIds, err = s.dao.GetValidActivityIdsCache(ctx)
		if err != nil && err != redis.ErrNil {
			log.Errorc(ctx, "[getValidActivityIds][GetValidActivityIdsCache][Error], err:%+v", err)
			return
		}
		if err == nil {
			return
		}
	}
	// 回源
	now := time.Now().Unix()
	endCompare := now - int64(time.Duration(s.conf.MissionActivityConf.CacheRule.ValidBeforeTime).Seconds())
	beginCompare := now + int64(time.Duration(s.conf.MissionActivityConf.CacheRule.ValidAfterTime).Seconds())

	list, err := s.dao.GetValidActivityListByTime(ctx, endCompare, beginCompare, mission.ActivityNormalStatus)
	if err != nil {
		log.Errorc(ctx, "[getValidActivityIds][GetValidActivityListByTime][Error], err:%+v", err)
		return
	}
	validIds = make([]int64, 0)
	for _, v := range list {
		validIds = append(validIds, v.Id)
	}
	_ = s.dao.SetValidActivityIdsCache(ctx, validIds)
	return
}

// getActivityTasksWithStats 通过活动id获取其任务列表，包括各个任务的奖品库存，各个任务的当前周期信息
func (s *Service) getActivityTasksWithStats(ctx context.Context, actId int64) (activity *v1.MissionActivityDetail, tasks []*mission.ActivityTaskStatInfo, err error) {
	activity, err = s.getMissionActivityInfo(ctx, actId, false, false)
	if err != nil {
		log.Errorc(ctx, "[getActivityTasksWithStats][getMissionActivityInfo][Error], err:%+v", err)
		return
	}
	list, err := s.getActivityTasks(ctx, actId, false, false)
	if err != nil {
		log.Errorc(ctx, "[getActivityTasksWithStats][getActivityTasks][Error], err:%+v", err)
		return
	}
	tasks = make([]*mission.ActivityTaskStatInfo, 0)
	// 计算任务的当前周期
	stockIds := make([]int64, 0)
	for _, v := range list {
		task := new(mission.ActivityTaskStatInfo)
		task.TaskDetail = v
		task.PeriodStat, err = s.calculateTaskPeriod(ctx, activity, v, 0, true)
		if err != nil {
			log.Errorc(ctx, "[getActivityTasksWithStats][calculatePeriod][Error], err:%+v", err)
			return
		}
		stockIds = append(stockIds, v.StockId)
		tasks = append(tasks, task)
	}
	if len(tasks) == 0 || len(stockIds) == 0 {
		return
	}
	// 获取任务的当前库存信息
	resp, err := s.stockSvr.GetStocksByIds(ctx, &v1.GetStocksReq{
		StockIds: stockIds,
	})
	if err != nil {
		log.Errorc(ctx, "[getActivityTasksWithStats][GetStocksByIds][Error], err:%+v", err)
		return
	}
	stockMap := resp.StockMap
	if len(stockMap) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "库存信息错误")
		log.Errorc(ctx, "[getActivityTasksWithStats][GetStocksByIds][Error], stockEmpty, stock:%+v", stockMap)
		return
	}
	for _, task := range tasks {
		stockItem, ok := stockMap[task.TaskDetail.StockId]
		if !ok {
			err = xecode.Errorf(xecode.RequestErr, "库存信息错误")
			log.Errorc(ctx, "[getActivityTasksWithStats][GetStocksByIds][Error], stockEmpty, stockId:%+v, stock:%+v", task.TaskDetail.StockId, stockMap)
			return
		}
		stockDetail := stockItem.List
		if stockDetail == nil || len(stockDetail) != 1 {
			err = xecode.Errorf(xecode.RequestErr, "库存信息错误")
			log.Errorc(ctx, "[getActivityTasksWithStats][GetStocksByIds][Error], stockError, stockId:%+v, stock:%+v", task.TaskDetail.StockId, stockMap)
			return
		}
		task.StockStat = new(mission.TaskStockStat)
		task.StockStat.LimitType = int64(stockDetail[0].CycleLimitObj.LimitType)
		task.StockStat.Total = int64(stockDetail[0].LimitNum)
		task.StockStat.Consumed = int64(stockDetail[0].LimitNum - stockDetail[0].StockNum)
		task.StockStat.StockBeginTime, task.StockStat.StockEndTime, err = s.calculateStockPeriod(ctx, activity, task.TaskDetail, stockDetail[0].CycleLimitObj)
		task.StockStat.StockPeriod, err = s.calculateReceivePeriod(int64(stockDetail[0].CycleLimitObj.LimitType))
		if err != nil {
			err = xecode.Errorf(xecode.RequestErr, "库存信息错误")
			log.Errorc(ctx, "[getActivityTasksWithStats][StockPeriod][Error], stockError, stockId:%+v, stock:%+v, err:%+v", task.TaskDetail.StockId, stockMap, err)
			return
		}
	}
	return
}

func (s *Service) getActivityTasks(ctx context.Context, actId int64, skipLocal bool, skipCache bool) (tasks []*v1.MissionTaskDetail, err error) {
	if !skipLocal {
		tasks, err = s.getActivityTasksFromLocal(ctx, actId)
		if err == nil {
			return
		} else {
			err = nil
		}
	}
	if !skipCache {
		tasks, err = s.dao.GetActivityTaskCache(ctx, actId)
		if err != nil && err != redis.ErrNil {
			log.Errorc(ctx, "[Service][getActivityTasks][GetActivityTaskCache] err:%+v", err)
			return
		}
		if err == nil {
			return
		}
	}
	tasks, err = s.dao.GetActivityTasks(ctx, actId)
	if err != nil {
		log.Errorc(ctx, "[Service][getActivityTasks][GetActivityTasks] err:%+v", err)
		return
	}
	_ = s.dao.SetActivityTaskCache(ctx, actId, tasks)
	return
}
