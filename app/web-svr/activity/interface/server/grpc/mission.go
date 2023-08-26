package grpc

import (
	"context"
	xecode "go-common/library/ecode"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/mission"
	"go-gateway/app/web-svr/activity/interface/service"
)

func (s *activityService) GetMissionActivityList(ctx context.Context, req *v1.GetMissionActivityListReq) (resp *v1.GetMissionActivityListResp, err error) {
	return service.MissionActivitySvr.GetMissionActivityList(ctx, req)
}

// 获取活动详情
func (s *activityService) GetMissionActivityInfo(ctx context.Context, req *v1.GetMissionActivityInfoReq) (resp *v1.MissionActivityDetail, err error) {
	return service.MissionActivitySvr.GetMissionActivityInfo(ctx, req)
}

// 更改活动状态
func (s *activityService) ChangeMissionActivityStatus(ctx context.Context, req *v1.ChangeMissionActivityStatusReq) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.ChangeMissionActivityStatus(ctx, req)
}

// 保存任务活动
func (s *activityService) SaveMissionActivity(ctx context.Context, req *v1.MissionActivityDetail) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.SaveMissionActivity(ctx, req)
}

// 获取活动的任务列表
func (s *activityService) GetMissionTasks(ctx context.Context, req *v1.GetMissionTasksReq) (resp *v1.GetMissionTasksResp, err error) {
	return service.MissionActivitySvr.GetMissionTasks(ctx, req)
}

// 活动下的任务全量保存
func (s *activityService) SaveMissionTasks(ctx context.Context, req *v1.SaveMissionTasksReq) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.SaveMissionTasks(ctx, req)
}

// 保存活动下的某个任务
func (s *activityService) SaveMissionTask(ctx context.Context, req *v1.MissionTaskDetail) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.SaveMissionTask(ctx, req)
}

// 获取活动下某个任务详情
func (s *activityService) GetMissionTaskInfo(ctx context.Context, req *v1.GetMissionTaskInfoReq) (resp *v1.MissionTaskDetail, err error) {
	return service.MissionActivitySvr.GetMissionTaskInfo(ctx, req)
}

// 任务活动下某个用户的完成状态
func (s *activityService) GetMissionTaskCompleteStatus(ctx context.Context, req *v1.GetMissionTaskCompleteStatusReq) (resp *v1.GetMissionTaskCompleteStatusResp, err error) {
	return service.MissionActivitySvr.GetMissionTaskCompleteStatus(ctx, req)
}

func (s *activityService) GetMissionTaskDetail(ctx context.Context, req *v1.GetMissionTaskDetailReq) (resp *v1.MissionTaskDetail, err error) {
	return service.MissionActivitySvr.GetMissionTaskDetail(ctx, req)
}

// 消费节点组进度消息构造任务的完成
func (s *activityService) GroupConsumerForTaskComplete(ctx context.Context, req *v1.GroupConsumerForTaskCompleteReq) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.GroupConsumerForTaskComplete(ctx, req)
}

// 获取有效的任务活动
func (s *activityService) GetValidMissionActivityIds(ctx context.Context, in *v1.NoReply) (resp *v1.GetValidMissionActivityIdsResp, err error) {
	resp = new(v1.GetValidMissionActivityIdsResp)
	validIds, err := service.MissionActivitySvr.GetValidActivityIds(ctx, true)
	if err != nil {
		return
	}
	resp.ActIds = validIds
	return
}

// 刷新任务活动的相关缓存
func (s *activityService) RefreshValidMissionActivityCache(ctx context.Context, in *v1.RefreshValidMissionActivityCacheReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	err = service.MissionActivitySvr.RefreshActivityCache(ctx, in.ActId)
	return
}

func (s *activityService) DelMissionTask(ctx context.Context, req *v1.DelMissionTaskReq) (resp *v1.NoReply, err error) {
	return service.MissionActivitySvr.DelMissionTask(ctx, req)
}

func (s *activityService) MissionCheckStock(ctx context.Context, in *v1.MissionCheckStockReq) (resp *v1.MissionCheckStockResp, err error) {
	return service.MissionActivitySvr.CheckStock(ctx, in)
}

func (s *activityService) GetMissionReceivingRecords(ctx context.Context, in *v1.GetMissionReceivingRecordsReq) (resp *v1.GetMissionReceivingRecordsResp, error error) {
	resp = new(v1.GetMissionReceivingRecordsResp)
	resp.List = make([]*v1.ReceivingRecord, 0)
	records, err := service.MissionActivitySvr.GetActivityReceiveRecordsByStatus(ctx, in.ActId, in.TableIndex, mission.TaskRewardStatusIn)
	if err != nil || records == nil || len(records) == 0 {
		return
	}
	for _, record := range records {
		resp.List = append(resp.List, &v1.ReceivingRecord{
			ReceiveId: record.ID,
			Mid:       record.Mid,
			ActId:     record.ActId,
		})
	}
	return
}

// 对任务活动领取中的记录进行领取重试
func (s *activityService) RetryMissionReceiveRecord(ctx context.Context, in *v1.RetryMissionReceiveRecordReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if in == nil || in.ReceiveId == 0 || in.Mid == 0 || in.ActId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	err = service.MissionActivitySvr.MakeUpRewards(ctx, in.ReceiveId, in.Mid, in.ActId)
	return
}
