package mission

import (
	"context"
	"fmt"
	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"strings"
	"time"
)

const (
	_timeDefaultTemplate          = "2006-01-02 15:04:05"
	_getActivityListByPage        = "select id, act_name, begin_time, end_time, uid_count, groups_act_id, bind_phone_check, status, mapping_type from act_mission_activity_base where is_deleted = 0 order by id desc limit ? offset ?"
	_getActivityListCount         = "select count(1) as num from act_mission_activity_base where is_deleted = 0"
	_getActivityTasks             = "select id as task_id, act_id, reward_id, stock_id, task_period, task_period_extra from act_mission_activity_tasks where act_id = ? and is_deleted = 0"
	_getActivity                  = "select id, act_name, begin_time, end_time, uid_count, groups_act_id, bind_phone_check, status, mapping_type from act_mission_activity_base where id = ? and is_deleted = 0"
	_getActivityTask              = "select id as task_id, act_id, reward_id, stock_id, task_period, task_period_extra from act_mission_activity_tasks where act_id = ? and id = ? and is_deleted = 0"
	_getActivityGroups            = "select id, act_id, task_id, group_id, complete_score from act_mission_activity_groups_mapping where task_id in (%s) and is_deleted = 0"
	_getValidActivityByTime       = "select id, act_name, begin_time, end_time, uid_count, groups_act_id, bind_phone_check, status, mapping_type from act_mission_activity_base where is_deleted = 0 and begin_time <= ? and end_time >= ? and status = ?"
	_getUserTaskComplete          = "select id, act_id, task_id, mid, complete_period, complete_time, failure_time, task_rewards_status, reason, receive_period, serial_num from act_user_task_record_%s_%s where mid = ? and task_id = ? and complete_period = ? and is_deleted = 0"
	_updateActStatus              = "update act_mission_activity_base set status = ? where id = ?"
	_insertAct                    = "insert into act_mission_activity_base (act_name, begin_time, end_time, uid_count, groups_act_id, bind_phone_check, status, mapping_type) values (?, ?, ?, ?, ?, ?, ?, ?)"
	_updateAct                    = "update act_mission_activity_base set act_name = ?, begin_time = ?, end_time = ?, groups_act_id = ?, bind_phone_check = ? , mapping_type = ? where id = ?"
	_insertActTask                = "insert act_mission_activity_tasks (act_id, reward_id, task_period, task_period_extra) values (?, ?, ?, ?)"
	_updateActTaskStockId         = "update act_mission_activity_tasks set stock_id = ? where act_id = ? and id = ?"
	_updateActTask                = "update act_mission_activity_tasks set reward_id = ?, task_period = ?, task_period_extra = ? where act_id = ? and id = ?"
	_removeActTask                = "update act_mission_activity_tasks set is_deleted = 1 where act_id = ? and id in (%s)"
	_removeTaskGroups             = "update act_mission_activity_groups_mapping set is_deleted = 1 where act_id = ? and task_id = ?"
	_batchInsertTaskGroups        = "insert into act_mission_activity_groups_mapping (act_id, task_id, group_id, complete_score) values %s"
	_getGroupMappingByGroupId     = "select id, act_id, task_id, group_id from act_mission_activity_groups_mapping where group_id = ? and is_deleted = 0"
	_getActivityTaskByTaskId      = "select id as task_id, act_id, reward_id, stock_id, task_period, task_period_extra from act_mission_activity_tasks where id = ? and is_deleted = 0"
	_getUserReceiveInfo           = "select id, act_id, task_id, mid, complete_period, complete_time, failure_time, task_rewards_status, reason, receive_period, serial_num from act_user_task_record_%s_%s where mid = ? and id = ? and is_deleted = 0"
	_insertUserCompleteTask       = "insert into act_user_task_record_%s_%s (act_id, task_id, mid, complete_period, complete_time, task_rewards_status) values (?, ?, ?, ?, ?, ?)"
	_getCompleteRecordBySerialNum = "select id, act_id, task_id, mid, complete_period, complete_time, failure_time, task_rewards_status, reason, serial_num from act_user_task_record_%s_%s where mid = ? and act_id = ? and task_id = ? and serial_num = ? and is_deleted = 0"
	_getGroupMappingByGroupIds    = "select id, act_id, task_id, group_id from act_mission_activity_groups_mapping where group_id in (%s) and is_deleted = 0"
	_removeTasksGroups            = "update act_mission_activity_groups_mapping set is_deleted = 1 where act_id = ? and task_id in (%s)"
	_getReceiveRecordByStatus     = "select id, act_id, task_id, mid, complete_period, complete_time, failure_time, task_rewards_status, reason, receive_period from act_user_task_record_%s_%s where act_id = ? and task_rewards_status = ? and is_deleted = 0 and mtime > ? and mtime < ? limit 100"
)

func (d *Dao) GetActivityListByPage(ctx context.Context, page int64, pageSize int64) (acts []*v1.MissionActivityDetail, total int64, err error) {
	acts = make([]*v1.MissionActivityDetail, 0)
	total = 0
	offset := (page - 1) * pageSize
	rows, err := d.db.Query(ctx, _getActivityListByPage, pageSize, offset)
	if err != nil {
		log.Errorc(ctx, "[GetActivityListByPage][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		act := new(v1.MissionActivityDetail)
		if err = rows.Scan(&act.Id, &act.ActName, &act.BeginTime, &act.EndTime, &act.UidCount, &act.GroupsActId, &act.BindPhoneCheck, &act.Status, &act.MappingType); err != nil {
			log.Errorc(ctx, "[GetActivityListByPage][Query][Scan][Error], err:%+v", err)
			return
		}
		acts = append(acts, act)
	}
	// 获取总数
	countStruct := new(struct {
		Num int64
	})

	if err = d.db.QueryRow(ctx, _getActivityListCount).Scan(&countStruct.Num); err != nil {
		log.Errorc(ctx, "[GetActivityListByPage][QueryRow][Error], err:%+v", err)
		return
	}
	total = countStruct.Num
	return
}

func (d *Dao) GetActivityTasks(ctx context.Context, actId int64) (tasks []*v1.MissionTaskDetail, err error) {
	tasks = make([]*v1.MissionTaskDetail, 0)
	rows, err := d.db.Query(ctx, _getActivityTasks, actId)
	if err != nil {
		log.Errorc(ctx, "[GetActivityTasks][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	taskIds := make([]int64, 0)
	for rows.Next() {
		taskInfo := new(v1.MissionTaskDetail)
		if err = rows.Scan(&taskInfo.TaskId, &taskInfo.ActId, &taskInfo.RewardId, &taskInfo.StockId, &taskInfo.TaskPeriod, &taskInfo.TaskPeriodExtra); err != nil {
			log.Errorc(ctx, "[GetActivityTasks][Scan][Error], err:%+v", err)
			return
		}
		taskIds = append(taskIds, taskInfo.TaskId)
		tasks = append(tasks, taskInfo)
	}
	if len(taskIds) == 0 {
		return
	}
	groupsRows, err := d.db.Query(ctx, fmt.Sprintf(_getActivityGroups, xstr.JoinInts(taskIds)))
	if err != nil {
		log.Errorc(ctx, "[GetActivityTasks][Groups][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = groupsRows.Close()
	}()
	taskGroups := make(map[int64][]*v1.TaskGroups)
	for groupsRows.Next() {
		taskGroup := new(mission.TaskGroupsMapping)
		if err = groupsRows.Scan(&taskGroup.ID, &taskGroup.ActId, &taskGroup.TaskId, &taskGroup.GroupId, &taskGroup.CompleteScore); err != nil {
			log.Errorc(ctx, "[GetActivityTasks][Groups][Scan][Error], err:%+v", err)
			return
		}
		if taskGroups[taskGroup.TaskId] == nil {
			taskGroups[taskGroup.TaskId] = make([]*v1.TaskGroups, 0)
		}
		taskGroups[taskGroup.TaskId] = append(taskGroups[taskGroup.TaskId], &v1.TaskGroups{
			GroupId:       taskGroup.GroupId,
			CompleteScore: taskGroup.CompleteScore,
		})
	}
	for _, v := range tasks {
		if taskGroups[v.TaskId] != nil {
			v.Groups = taskGroups[v.TaskId]
		}
	}
	return
}

// GetActivityInfo 获取活动详情
func (d *Dao) GetActivityInfo(ctx context.Context, actId int64) (act *v1.MissionActivityDetail, err error) {
	act = new(v1.MissionActivityDetail)
	if err = d.db.QueryRow(ctx, _getActivity, actId).Scan(&act.Id, &act.ActName, &act.BeginTime, &act.EndTime, &act.UidCount, &act.GroupsActId, &act.BindPhoneCheck, &act.Status, &act.MappingType); err != nil {
		log.Errorc(ctx, "[GetActivityInfo][QueryRow][Scan][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetActivityTaskInfo(ctx context.Context, actId int64, taskId int64) (taskInfo *v1.MissionTaskDetail, err error) {
	taskInfo = new(v1.MissionTaskDetail)
	if err = d.db.QueryRow(ctx, _getActivityTask, actId, taskId).Scan(&taskInfo.TaskId, &taskInfo.ActId, &taskInfo.RewardId, &taskInfo.StockId, &taskInfo.TaskPeriod, &taskInfo.TaskPeriodExtra); err != nil {
		log.Errorc(ctx, "[GetActivityTaskInfo][QueryRow][Error], err:%+v", err)
		return
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_getActivityGroups, xstr.JoinInts([]int64{taskId})))
	if err != nil {
		log.Errorc(ctx, "[GetActivityTaskInfo][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	groups := make([]*v1.TaskGroups, 0)
	for rows.Next() {
		taskGroup := new(mission.TaskGroupsMapping)
		if err = rows.Scan(&taskGroup.ID, &taskGroup.ActId, &taskGroup.TaskId, &taskGroup.GroupId, &taskGroup.CompleteScore); err != nil {
			log.Errorc(ctx, "[GetActivityTaskInfo][Query][Error], err:%+v", err)
			return
		}
		groups = append(groups, &v1.TaskGroups{
			GroupId:       taskGroup.GroupId,
			CompleteScore: taskGroup.CompleteScore,
		})
	}
	taskInfo.Groups = groups
	return
}

func (d *Dao) GetValidActivityListByTime(ctx context.Context, endCompare int64, beginCompare int64, status int64) (list []*v1.MissionActivityDetail, err error) {
	list = make([]*v1.MissionActivityDetail, 0)
	rows, err := d.db.Query(ctx,
		_getValidActivityByTime,
		time.Unix(beginCompare, 0).Format(_timeDefaultTemplate),
		time.Unix(endCompare, 0).Format(_timeDefaultTemplate),
		status)
	if err != nil {
		log.Errorc(ctx, "[GetValidActivityListByTime][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		act := new(v1.MissionActivityDetail)
		if err = rows.Scan(&act.Id, &act.ActName, &act.BeginTime, &act.EndTime, &act.UidCount, &act.GroupsActId, &act.BindPhoneCheck, &act.Status, &act.MappingType); err != nil {
			log.Errorc(ctx, "[GetValidActivityListByTime][Query][Scan][Error], err:%+v", err)
			return
		}
		list = append(list, act)
	}
	return
}

func (d *Dao) ActivityStatusUpdate(ctx context.Context, actId int64, status int64) (err error) {
	_, err = d.db.Exec(ctx, _updateActStatus, status, actId)
	if err != nil {
		log.Errorc(ctx, "[Dao][ActivityStatusUpdate][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) AddActivity(ctx context.Context, activity *v1.MissionActivityDetail) (err error) {
	if _, err = d.db.Exec(ctx,
		_insertAct,
		activity.ActName,
		activity.BeginTime,
		activity.EndTime,
		activity.UidCount,
		activity.GroupsActId,
		activity.BindPhoneCheck,
		mission.ActivityNormalStatus,
		activity.MappingType,
	); err != nil {
		log.Errorc(ctx, "[Dao][AddActivity][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) UpdateActivity(ctx context.Context, activity *v1.MissionActivityDetail) (err error) {
	if _, err = d.db.Exec(ctx, _updateAct,
		activity.ActName,
		activity.BeginTime,
		activity.EndTime,
		activity.GroupsActId,
		activity.BindPhoneCheck,
		activity.MappingType,
		activity.Id,
	); err != nil {
		log.Errorc(ctx, "[Dao][AddActivity][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) AddActivityTask(ctx context.Context, actId int64, task *v1.MissionTaskDetail) (err error) {
	if err = d.checkGroupsMappingValid(ctx, task); err != nil {
		return
	}
	res, err := d.db.Exec(ctx, _insertActTask,
		actId,
		task.RewardId,
		task.TaskPeriod,
		task.TaskPeriodExtra,
	)
	if err != nil {
		log.Errorc(ctx, "[Dao][AddActivityTask][Exec][Error], err:%+v", err)
		return
	}
	task.TaskId, err = res.LastInsertId()
	if err != nil {
		log.Errorc(ctx, "[Dao][AddActivityTask][LastInsertId][Error], err:%+v", err)
		return
	}
	if err = d.batchAddTaskGroupMapping(ctx, actId, task); err != nil {
		log.Errorc(ctx, "[Dao][batchAddTaskGroupMapping][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) UpdateActivityTask(ctx context.Context, actId int64, task *v1.MissionTaskDetail) (err error) {
	// 更新任务和节点组的关系映射
	if _, err = d.db.Exec(ctx, _removeTaskGroups,
		actId, task.TaskId); err != nil {
		log.Errorc(ctx, "[Dao][UpdateActivityTask][RemoveTaskGroups][Exec][Error], err:%+v", err)
		return
	}
	if err = d.checkGroupsMappingValid(ctx, task); err != nil {
		return
	}
	if err = d.batchAddTaskGroupMapping(ctx, actId, task); err != nil {
		log.Errorc(ctx, "[Dao][batchAddTaskGroupMapping][Error], err:%+v", err)
		return
	}
	if _, err = d.db.Exec(ctx, _updateActTask,
		task.RewardId,
		task.TaskPeriod,
		task.TaskPeriodExtra,
		actId,
		task.TaskId,
	); err != nil {
		log.Errorc(ctx, "[Dao][AddActivityTask][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) checkGroupsMappingValid(ctx context.Context, task *v1.MissionTaskDetail) (err error) {
	groupIds := make([]int64, 0)
	for _, group := range task.Groups {
		groupIds = append(groupIds, group.GroupId)
	}
	mappingList, err := d.getGroupsMappingByGroupIds(ctx, groupIds)
	if err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[Dao][getGroupsMappingByGroupIds][Error], err:%+v", err)
		return
	}
	for _, mapping := range mappingList {
		if mapping.TaskId != task.TaskId {
			err = xecode.Errorf(xecode.RequestErr, fmt.Sprintf("节点组：%d 已被其他任务使用", mapping.GroupId))
			return
		}
	}
	return
}

func (d *Dao) getGroupsMappingByGroupIds(ctx context.Context, groupIds []int64) (list []*mission.TaskGroupsMapping, err error) {
	list = make([]*mission.TaskGroupsMapping, 0)
	if len(groupIds) == 0 {
		return
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_getGroupMappingByGroupIds, xstr.JoinInts(groupIds)))
	if err != nil {
		log.Errorc(ctx, "[Dao][getGroupsMappingByGroupIds][Exec][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		taskGroup := new(mission.TaskGroupsMapping)
		if err = rows.Scan(&taskGroup.ID, &taskGroup.ActId, &taskGroup.TaskId, &taskGroup.GroupId); err != nil {
			log.Errorc(ctx, "[Dao][GetMappingByGroupId][QueryRow][Error], err:%+v", err)
			return
		}
		list = append(list, taskGroup)
	}
	return
}

func (d *Dao) batchAddTaskGroupMapping(ctx context.Context, actId int64, task *v1.MissionTaskDetail) (err error) {
	values := make([]string, 0)
	items := make([]interface{}, 0)
	for _, v := range task.Groups {
		values = append(values, "(?, ?, ?, ?)")
		items = append(items, actId, task.TaskId, v.GroupId, v.CompleteScore)
	}
	if _, err = d.db.Exec(ctx,
		fmt.Sprintf(_batchInsertTaskGroups, strings.Join(values, ",")),
		items...,
	); err != nil {
		log.Errorc(ctx, "[Dao][UpdateActivityTask][RemoveTaskGroups][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) RemoveTasks(ctx context.Context, actId int64, taskIds []int64) (err error) {
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_removeActTask, xstr.JoinInts(taskIds)), actId); err != nil {
		log.Errorc(ctx, "[Dao][RemoveTask][Exec][Error], err:%+v", err)
		return
	}
	// 移除节点组映射
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_removeTasksGroups, xstr.JoinInts(taskIds)), actId); err != nil {
		log.Errorc(ctx, "[Dao][RemoveTasksGroups][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) UpdateTaskStockId(ctx context.Context, actId int64, taskId int64, stockId int64) (err error) {
	if _, err = d.db.Exec(ctx, _updateActTaskStockId, stockId, actId, taskId); err != nil {
		log.Errorc(ctx, "[Dao][UpdateTaskStockId][Exec][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetMappingByGroupId(ctx context.Context, groupId int64) (actId int64, err error) {
	taskGroup := new(mission.TaskGroupsMapping)
	if err = d.db.QueryRow(ctx, _getGroupMappingByGroupId, groupId).Scan(&taskGroup.ID, &taskGroup.ActId, &taskGroup.TaskId, &taskGroup.GroupId); err != nil {
		log.Errorc(ctx, "[Dao][GetMappingByGroupId][QueryRow][Error], err:%+v", err)
		return
	}
	actId = taskGroup.ActId
	return
}

func (d *Dao) GetUserCompleteTaskFromDB(ctx context.Context, mid int64, actId int64, taskId int64, period int64, receivePeriod int64) (completeRecord *mission.UserCompleteRecord, err error) {
	completeRecord = new(mission.UserCompleteRecord)
	tablePrefixFirst, tablePrefixSecond := getUserTable(actId, mid)
	if err = d.db.QueryRow(ctx,
		fmt.Sprintf(_getUserTaskComplete, tablePrefixFirst, tablePrefixSecond),
		mid,
		taskId,
		period,
	).Scan(
		&completeRecord.ID,
		&completeRecord.ActId,
		&completeRecord.TaskId,
		&completeRecord.Mid,
		&completeRecord.CompletePeriod,
		&completeRecord.CompleteTime,
		&completeRecord.FailureTime,
		&completeRecord.TaskRewardsStatus,
		&completeRecord.Reason,
		&completeRecord.ReceivePeriod,
		&completeRecord.SerialNum,
	); err != nil {
		log.Errorc(ctx, "[GetUserCompleteTaskFromDB][Scan][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) InsertUserCompleteTaskRecord(ctx context.Context, mid int64, userTask *mission.UserTaskDetail) (recordId int64, err error) {
	tablePrefixFirst, tablePrefixSecond := getUserTable(userTask.ActId, mid)
	res, err := d.db.Exec(ctx,
		fmt.Sprintf(_insertUserCompleteTask, tablePrefixFirst, tablePrefixSecond),
		userTask.ActId,
		userTask.ID,
		mid,
		userTask.TaskPeriod,
		time.Now().Unix(),
		mission.TaskRewardStatusIdle,
	)
	if err != nil {
		log.Errorc(ctx, "[InsertUserCompleteTaskRecord][Insert][Error], err:%+v", err)
		return
	}
	recordId, err = res.LastInsertId()
	if err != nil {
		log.Errorc(ctx, "[InsertUserCompleteTaskRecord][LastInsertId][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetActivityReceiveRecordsByStatus(ctx context.Context, index int64, actId int64, status int64, beginMTime int64, endMTime int64) (records []*mission.UserCompleteRecord, err error) {
	records = make([]*mission.UserCompleteRecord, 0)
	tablePrefixFirst, tablePrefixSecond := getUserTable(actId, index)

	rows, err := d.db.Query(ctx,
		fmt.Sprintf(_getReceiveRecordByStatus, tablePrefixFirst, tablePrefixSecond),
		actId,
		status,
		time.Unix(beginMTime, 0).Format(_timeDefaultTemplate),
		time.Unix(endMTime, 0).Format(_timeDefaultTemplate),
	)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		completeRecord := new(mission.UserCompleteRecord)
		if err = rows.Scan(
			&completeRecord.ID,
			&completeRecord.ActId,
			&completeRecord.TaskId,
			&completeRecord.Mid,
			&completeRecord.CompletePeriod,
			&completeRecord.CompleteTime,
			&completeRecord.FailureTime,
			&completeRecord.TaskRewardsStatus,
			&completeRecord.Reason,
			&completeRecord.ReceivePeriod,
		); err != nil {
			log.Errorc(ctx, "[GetActivityReceiveRecordsByStatus][Scan][Error], err:%+v", err)
			return
		}
		records = append(records, completeRecord)
	}
	return
}

func getUserTable(actId int64, mid int64) (tablePrefixFirst string, tablePrefixSecond string) {
	tablePrefixFirst = fmt.Sprintf("%02d", actId%5)
	tablePrefixSecond = fmt.Sprintf("%03d", mid%100)
	return
}
