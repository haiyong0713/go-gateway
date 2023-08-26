package mission

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
)

const (
	_activityDetailKey = "activity:mission_activity:detail:actId:%d"
	_activityDetailTtl = 86400
	_activityTasksKey  = "activity:mission_activity:tasks:actId:%d"
	_activityTasksTtl  = 86400

	_activityUserCompleteTaskKey = "activity:mission_activity:user_complete:uid:%d:actId:%d:taskId:%d:period:%d:receivePeriod:%d"
	_activityUserCompleteTaskTtl = 86400

	_validActivityIdsKey = "activity:mission_activity:valid:activityIds"
	_validActivityIdsTtl = 86400

	_activityTaskByGroupIdKey = "activity:mission_activity:group:task:groupId:%d"
	_activityTaskByGroupIdTtl = 86400

	_activityTaskInfoKey = "activity:mission_activity:taskInfo:taskId:%d"
	_activityTaskInfoTtl = 86400

	_activityUserReceiveKey           = "activity:mission_activity:userReceiveInfo:uid:%d:actId:%d:receiveId:%d"
	_activityUserReceiveTtl           = 86400
	_activityUserReceiveNoReceivedTtl = 300

	_activityUserCompleteSerialNumCacheKey  = "activity:mission_activity:userReceiveInfo:uid:%d:serialNum:%s"
	_activityUserCompleteSerialNumCacheKTtl = 86400
)

func formatActivityTaskByGroupKey(groupId int64) string {
	return fmt.Sprintf(_activityTaskByGroupIdKey, groupId)
}

func formatActivityDetailKey(actId int64) string {
	return fmt.Sprintf(_activityDetailKey, actId)
}

func formatActivityTaskKey(actId int64) string {
	return fmt.Sprintf(_activityTasksKey, actId)
}

func formatUserReceiveInfoKey(mid int64, actId int64, receivedId int64) string {
	return fmt.Sprintf(_activityUserReceiveKey, mid, actId, receivedId)
}

func formatUserCompleteSerialNumKey(mid int64, serialNum string) string {
	return fmt.Sprintf(_activityUserCompleteSerialNumCacheKey, mid, serialNum)
}

func (d *Dao) GetActivityDetailCache(ctx context.Context, actId int64) (activityInfo *v1.MissionActivityDetail, err error) {
	reply, err := redis.Bytes(d.redis.Do(ctx, "get", formatActivityDetailKey(actId)))
	if err != nil {
		log.Errorc(ctx, "[GetActivityDetailCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &activityInfo); err != nil {
		log.Errorc(ctx, "[GetActivityDetailCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetActivityDetailCache(ctx context.Context, activityInfo *v1.MissionActivityDetail) (err error) {
	if activityInfo == nil || activityInfo.Id == 0 {
		return
	}
	cacheValue, err := json.Marshal(activityInfo)
	if err != nil {
		log.Errorc(ctx, "[SetActivityDetailCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", formatActivityDetailKey(activityInfo.Id), _activityDetailTtl, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[SetActivityDetailCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetActivityTaskCache(ctx context.Context, actId int64) (tasks []*v1.MissionTaskDetail, err error) {
	tasks = make([]*v1.MissionTaskDetail, 0)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get", formatActivityTaskKey(actId)))
	if err != nil {
		log.Errorc(ctx, "[GetActivityTaskCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &tasks); err != nil {
		log.Errorc(ctx, "[GetActivityTaskCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetActivityTaskCache(ctx context.Context, actId int64, tasks []*v1.MissionTaskDetail) (err error) {
	cacheValue, err := json.Marshal(tasks)
	if err != nil {
		log.Errorc(ctx, "[SetActivityDetailCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx, "setEx", formatActivityTaskKey(actId), _activityTasksTtl, cacheValue)
	if err != nil {
		log.Errorc(ctx, "[SetActivityDetailCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func formatUserCompleteTaskKey(mid int64, actId int64, taskId int64, period int64, receivePeriod int64) string {
	return fmt.Sprintf(_activityUserCompleteTaskKey, mid, actId, taskId, period, receivePeriod)
}

func (d *Dao) SetUserCompleteTaskCache(ctx context.Context, completeRecord *mission.UserCompleteRecord) (err error) {
	cacheValue, err := json.Marshal(completeRecord)
	if err != nil {
		log.Errorc(ctx, "[SetUserCompleteTaskCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx,
		"setEx",
		formatUserCompleteTaskKey(completeRecord.Mid, completeRecord.ActId, completeRecord.TaskId, completeRecord.CompletePeriod, 0),
		_activityUserCompleteTaskTtl,
		cacheValue,
	)
	if err != nil {
		log.Errorc(ctx, "[SetUserCompleteTaskCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetUserCompleteTaskCache(ctx context.Context, mid int64, actId int64, taskId int64, period int64, receivePeriod int64) (completeRecord *mission.UserCompleteRecord, err error) {
	completeRecord = new(mission.UserCompleteRecord)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get",
		formatUserCompleteTaskKey(mid, actId, taskId, period, receivePeriod),
	))
	if err != nil {
		log.Errorc(ctx, "[GetUserCompleteTaskCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &completeRecord); err != nil {
		log.Errorc(ctx, "[GetUserCompleteTaskCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetValidActivityIdsCache(ctx context.Context) (actIds []int64, err error) {
	actIds = make([]int64, 0)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get", _validActivityIdsKey))
	if err != nil {
		log.Errorc(ctx, "[GetValidActivityIdsCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &actIds); err != nil {
		log.Errorc(ctx, "[GetValidActivityIdsCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetValidActivityIdsCache(ctx context.Context, actIds []int64) (err error) {
	cacheValue, err := json.Marshal(actIds)
	if err != nil {
		log.Errorc(ctx, "[SetValidActivityIdsCache][Marshal][Error], err:%+v", err)
		return
	}
	_, err = d.redis.Do(ctx,
		"setEx",
		_validActivityIdsKey,
		_validActivityIdsTtl,
		cacheValue,
	)
	if err != nil {
		log.Errorc(ctx, "[SetValidActivityIdsCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetActivityTaskIdCacheByGroupId(ctx context.Context, groupId int64) (actId int64, err error) {
	actId, err = redis.Int64(d.redis.Do(ctx, "get", formatActivityTaskByGroupKey(groupId)))
	if err != nil {
		log.Errorc(ctx, "[GetActivityTaskIdCacheByGroupId][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetActivityTaskIdCacheByGroupId(ctx context.Context, groupId int64, actId int64) (err error) {
	_, err = d.redis.Do(ctx,
		"setEx",
		formatActivityTaskByGroupKey(groupId),
		_activityTaskByGroupIdTtl,
		actId,
	)
	if err != nil {
		log.Errorc(ctx, "[SetActivityTaskIdCacheByGroupId][Redis][Error], err:%+v", err)
		return
	}
	return
}

func formatTaskInfoCacheKey(taskId int64) string {
	return fmt.Sprintf(_activityTaskInfoKey, taskId)
}

func (d *Dao) GetTaskDetailByTaskIdFromCache(ctx context.Context, taskId int64) (task *v1.MissionTaskDetail, err error) {
	task = new(v1.MissionTaskDetail)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get",
		formatTaskInfoCacheKey(taskId),
	))
	if err != nil {
		log.Errorc(ctx, "[GetTaskDetailByTaskIdFromCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &task); err != nil {
		log.Errorc(ctx, "[GetTaskDetailByTaskIdFromCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetTaskDetailCacheByTaskId(ctx context.Context, task *v1.MissionTaskDetail) (err error) {
	cacheValue, err := json.Marshal(task)
	_, err = d.redis.Do(ctx,
		"setEx",
		formatTaskInfoCacheKey(task.TaskId),
		_activityTaskInfoTtl,
		cacheValue,
	)
	if err != nil {
		log.Errorc(ctx, "[SetTaskDetailByTaskIdFromCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetUserReceiveCache(ctx context.Context, mid int64, actId int64, receiveId int64) (receiveInfo *mission.UserCompleteRecord, err error) {
	receiveInfo = new(mission.UserCompleteRecord)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get",
		formatUserReceiveInfoKey(mid, actId, receiveId),
	))
	if err != nil {
		log.Errorc(ctx, "[GetUserReceiveCache][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &receiveInfo); err != nil {
		log.Errorc(ctx, "[GetUserReceiveCache][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetUserReceiveCache(ctx context.Context, receiveInfo *mission.UserCompleteRecord) (err error) {
	cacheValue, err := json.Marshal(receiveInfo)
	_, err = d.redis.Do(ctx,
		"setEx",
		formatUserReceiveInfoKey(receiveInfo.Mid, receiveInfo.ActId, receiveInfo.ID),
		_activityUserReceiveTtl,
		cacheValue,
	)
	if err != nil {
		log.Errorc(ctx, "[SetTaskDetailByTaskIdFromCache][Redis][Error], err:%+v", err)
		return
	}
	err = d.SetUserCompleteTaskCache(ctx, receiveInfo)
	return
}

func (d *Dao) GetUserCompleteRecordCacheBySerialNum(ctx context.Context, mid int64, serialNum string) (record *mission.UserCompleteRecord, err error) {
	record = new(mission.UserCompleteRecord)
	reply, err := redis.Bytes(d.redis.Do(ctx, "get",
		formatUserCompleteSerialNumKey(mid, serialNum),
	))
	if err != nil {
		log.Errorc(ctx, "[GetUserCompleteRecordCacheBySerialNum][Redis][Error], err:%+v", err)
		return
	}
	if err = json.Unmarshal(reply, &record); err != nil {
		log.Errorc(ctx, "[GetUserCompleteRecordCacheBySerialNum][Unmarshal][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) SetUserCompleteRecordCacheBySerialNum(ctx context.Context, record *mission.UserCompleteRecord) (err error) {
	cacheValue, err := json.Marshal(record)
	_, err = d.redis.Do(ctx,
		"setEx",
		formatUserCompleteSerialNumKey(record.Mid, record.SerialNum),
		_activityUserCompleteSerialNumCacheKTtl,
		cacheValue,
	)
	if err != nil {
		log.Errorc(ctx, "[SetUserCompleteRecordCacheBySerialNum][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) DelActivityCache(ctx context.Context, actId int64) (err error) {
	_, err = d.redis.Do(ctx, "del", formatActivityDetailKey(actId))
	if err != nil {
		log.Errorc(ctx, "[DelActivityCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) DelActivityTasksCache(ctx context.Context, actId int64) (err error) {
	_, err = d.redis.Do(ctx, "del", formatActivityTaskKey(actId))
	if err != nil {
		log.Errorc(ctx, "[DelActivityTasksCache][Redis][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) DelActivityTaskCache(ctx context.Context, taskId int64) (err error) {
	_, err = d.redis.Do(ctx, "del", formatTaskInfoCacheKey(taskId))
	if err != nil {
		log.Errorc(ctx, "[DelActivityTaskCache][Redis][Error], err:%+v", err)
		return
	}
	return
}
