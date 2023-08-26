package mission

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
)

func (d *Dao) GetTaskDetailFromDBByTaskId(ctx context.Context, taskId int64) (taskInfo *v1.MissionTaskDetail, err error) {
	taskInfo = new(v1.MissionTaskDetail)
	if err = d.db.QueryRow(ctx, _getActivityTaskByTaskId, taskId).Scan(&taskInfo.TaskId, &taskInfo.ActId, &taskInfo.RewardId, &taskInfo.StockId, &taskInfo.TaskPeriod, &taskInfo.TaskPeriodExtra); err != nil {
		log.Errorc(ctx, "[GetTaskDetailByTaskId][Scan][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetTaskDetailByTaskId(ctx context.Context, taskId int64) (taskInfo *v1.MissionTaskDetail, err error) {
	taskInfo, err = d.GetTaskDetailByTaskIdFromCache(ctx, taskId)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[GetTaskDetailByTaskId][GetTaskDetailByTaskIdFromCache][Error], err:%+v", err)
		return
	}
	if err == nil {
		return
	}
	taskInfo, err = d.GetTaskDetailFromDBByTaskId(ctx, taskId)
	if err != nil {
		log.Errorc(ctx, "[GetTaskDetailByTaskId][GetTaskDetailFromDBByTaskId][Error], err:%+v", err)
		return
	}
	_ = d.SetTaskDetailCacheByTaskId(ctx, taskInfo)
	return
}

func (d *Dao) GetUserReceiveInfo(ctx context.Context, mid int64, actId int64, receiveId int64) (receiveInfo *mission.UserCompleteRecord, err error) {
	receiveInfo = new(mission.UserCompleteRecord)
	tablePrefixFirst, tablePrefixSecond := getUserTable(actId, mid)
	if err = d.db.QueryRow(ctx,
		fmt.Sprintf(_getUserReceiveInfo, tablePrefixFirst, tablePrefixSecond),
		mid,
		receiveId,
	).Scan(
		&receiveInfo.ID,
		&receiveInfo.ActId,
		&receiveInfo.TaskId,
		&receiveInfo.Mid,
		&receiveInfo.CompletePeriod,
		&receiveInfo.CompleteTime,
		&receiveInfo.FailureTime,
		&receiveInfo.TaskRewardsStatus,
		&receiveInfo.Reason,
		&receiveInfo.ReceivePeriod,
		&receiveInfo.SerialNum,
	); err != nil {
		log.Errorc(ctx, "[GetUserReceiveInfo][Scan][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetUserReceiveInfoBySerialNum(ctx context.Context, actId int64, taskId int64, mid int64, serialNum string) (receiveInfo *mission.UserCompleteRecord, err error) {
	receiveInfo = new(mission.UserCompleteRecord)
	tablePrefixFirst, tablePrefixSecond := getUserTable(actId, mid)
	err = d.db.QueryRow(ctx,
		fmt.Sprintf(_getCompleteRecordBySerialNum, tablePrefixFirst, tablePrefixSecond),
		mid,
		actId,
		taskId,
		serialNum,
	).Scan(
		&receiveInfo.ID,
		&receiveInfo.ActId,
		&receiveInfo.TaskId,
		&receiveInfo.Mid,
		&receiveInfo.CompletePeriod,
		&receiveInfo.CompleteTime,
		&receiveInfo.FailureTime,
		&receiveInfo.TaskRewardsStatus,
		&receiveInfo.Reason,
		&receiveInfo.SerialNum,
	)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "[GetUserReceiveInfo][Scan][Error], err:%+v", err)
		return
	}
	return
}
