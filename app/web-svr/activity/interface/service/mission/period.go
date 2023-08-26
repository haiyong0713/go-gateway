package mission

import (
	"context"
	"fmt"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"strconv"
	"time"
)

func (s *Service) calculateReceivePeriod(limitType int64) (receivePeriod int64, err error) {
	receivePeriod = 0
	if int64(v1.StockServerLimitType_StoreUpperLimit) == limitType {
		period := time.Now().Format(_timePeriodTemplate)
		receivePeriod, err = strconv.ParseInt(period, 10, 64)
		if err != nil {
			return
		}
	}
	return
}

func (s *Service) calculateStockPeriod(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, item *v1.CycleLimitStruct) (beginTime int64, endTime int64, err error) {
	period := time.Now().Format(_timePeriodTemplate)
	if item.LimitType == int32(v1.StockServerLimitType_StoreUpperLimit) {
		beginTimestamp, _ := time.ParseInLocation(_DailyPeriodTemplate, fmt.Sprintf("%s %s", period, item.CycleStartTime), time.Local)
		endTimestamp, _ := time.ParseInLocation(_DailyPeriodTemplate, fmt.Sprintf("%s %s", period, item.CycleEndTime), time.Local)
		beginTime = beginTimestamp.Unix()
		endTime = endTimestamp.Unix()
	} else {
		beginTime = activity.BeginTime.Time().Unix()
		endTime = activity.EndTime.Time().Unix()
	}
	return
}

func formatTaskPeriodStatByPeriod(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, period int64) (periodStat *mission.TaskPeriodStat, err error) {
	switch task.TaskPeriod {
	case mission.TaskPeriodNone:
		period = 0
		periodStat, err = formatDefaultPeriod(activity)
	case mission.TaskPeriodDaily:
		periodStat, err = formatTaskDailyPeriod(ctx, task, period)
	case mission.TaskPeriodWeekly:
		periodStat, err = formatTaskDefaultByPeriod(ctx, activity, task, period)
	case mission.TaskPeriodMonthly:
		periodStat, err = formatTaskDefaultByPeriod(ctx, activity, task, period)
	default:
		periodStat, err = formatDefaultPeriod(activity)
	}
	return
}

func (s *Service) calculateTaskPeriod(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, period int64, isNowPeriod bool) (periodStat *mission.TaskPeriodStat, err error) {
	periodStat = new(mission.TaskPeriodStat)
	if isNowPeriod {
		periodStat, err = formatPeriodStatByTime(ctx, activity, task, time.Now().Unix())
	} else {
		periodStat, err = formatTaskPeriodStatByPeriod(ctx, activity, task, period)
	}
	if err != nil {
		log.Errorc(ctx, "[calculatePeriod][Format][Error], err:%+v", err)
		return
	}
	return
}

func formatPeriodStatByTime(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, timestamp int64) (periodStat *mission.TaskPeriodStat, err error) {
	// 计算当前周期
	periodStat = new(mission.TaskPeriodStat)
	switch task.TaskPeriod {
	case mission.TaskPeriodNone:
		periodStat, err = formatDefaultPeriod(activity)
	case mission.TaskPeriodDaily:
		periodStr := time.Unix(timestamp, 0).Format(_timePeriodTemplate)
		// 目前天级别任务统一使用自然天处理活动周期
		//timeValue, errG := time.ParseInLocation(_DailyPeriodTemplate, fmt.Sprintf("%s %s", periodStr, _defaultTaskPeriodExtra), time.Local)
		//if errG != nil {
		//	err = errG
		//	return
		//}
		//period := int64(0)
		//if timestamp < timeValue.Unix() {
		//	// 昨天
		//	period, err = strconv.ParseInt(time.Unix(timestamp-86400, 0).Format(_timePeriodTemplate), 10, 64)
		//} else {
		//	period, err = strconv.ParseInt(time.Unix(timeValue.Unix(), 0).Format(_timePeriodTemplate), 10, 64)
		//}
		period, errG := strconv.ParseInt(periodStr, 10, 64)
		if errG != nil {
			err = errG
			return
		}
		periodStat, err = formatTaskDailyPeriod(ctx, task, period)
	case mission.TaskPeriodWeekly:
		periodYear, periodWeek := time.Unix(timestamp, 0).ISOWeek()
		period, errG := strconv.ParseInt(fmt.Sprintf("%d%d", periodYear, periodWeek), 10, 64)
		if errG != nil {
			err = errG
			return
		}
		periodStat, err = formatTaskDefaultByPeriod(ctx, activity, task, period)
	case mission.TaskPeriodMonthly:
		period, errG := strconv.ParseInt(time.Unix(timestamp, 0).Format(_MonthlyPeriodTemplate), 10, 64)
		if errG != nil {
			err = errG
			return
		}
		periodStat, err = formatTaskDefaultByPeriod(ctx, activity, task, period)
	default:
		periodStat.Period = 0
	}
	return
}

func formatTaskDefaultByPeriod(ctx context.Context, activity *v1.MissionActivityDetail, task *v1.MissionTaskDetail, period int64) (periodStat *mission.TaskPeriodStat, err error) {
	periodStat = &mission.TaskPeriodStat{
		Period:          period,
		PeriodBeginTime: activity.BeginTime.Time().Unix(),
		PeriodEndTime:   activity.EndTime.Time().Unix(),
	}
	return
}

func formatTaskDailyPeriod(ctx context.Context, task *v1.MissionTaskDetail, period int64) (periodStat *mission.TaskPeriodStat, err error) {
	periodStat = new(mission.TaskPeriodStat)
	beginTimestamp, err := time.ParseInLocation(_timePeriodTemplate, fmt.Sprintf("%d", period), time.Local)
	if err != nil {
		log.Errorc(ctx, "[formatTaskDailyPeriod][Parse][Erorr], err:%+v, period:%+v", err, period)
		return
	}
	periodStat.Period = period
	periodStat.PeriodBeginTime = beginTimestamp.Unix()
	periodStat.PeriodEndTime = periodStat.PeriodBeginTime + 86399
	return
}

func formatDefaultPeriod(activity *v1.MissionActivityDetail) (periodStat *mission.TaskPeriodStat, err error) {
	periodStat = new(mission.TaskPeriodStat)
	periodStat.Period = 0
	periodStat.PeriodBeginTime = activity.BeginTime.Time().Unix()
	periodStat.PeriodEndTime = activity.EndTime.Time().Unix()
	return
}
