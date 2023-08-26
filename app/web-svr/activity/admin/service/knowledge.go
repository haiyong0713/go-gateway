package service

import (
	"context"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/client"
	"go-gateway/app/web-svr/activity/admin/model"
	pb "go-gateway/app/web-svr/activity/interface/api"
)

func (s *Service) UpdateKnowledgeHistoryBatch(ctx context.Context, param *model.ParamKnowledge) (err error) {
	var shareValue int64
	table, field, err := s.getKnowledgeConfig(ctx, param.ConfigId, param.TaskName)
	if err != nil {
		log.Errorc(ctx, "UpdateKnowledgeHistoryBatch param(%+v) error(%+v)", param, err)
		return
	}
	if table == "" {
		err = xecode.Errorf(xecode.RequestErr, "配制表不能为空")
		return
	}
	if field == "" {
		err = xecode.Errorf(xecode.RequestErr, "勋章名称不正确")
		return
	}
	if !param.IsBack {
		shareValue = 1
	}
	if err = s.dao.UpdateUserKnowTask(ctx, shareValue, table, field, param.UpdateMids); err != nil {
		log.Errorc(ctx, "UpdateKnowledgeHistoryBatch field(%s) error(%+v)", field, err)
		return
	}
	arg := &pb.DelKnowledgeCacheReq{
		TableName:  table,
		UpdateMids: param.UpdateMids,
	}
	if _, err = client.ActivityClient.DelKnowledgeCache(ctx, arg); err != nil {
		log.Errorc(ctx, "UpdateKnowledgeHistoryBatch client.ActivityClient.DelKnowledgeCache(%s) error(%+v)", field, err)
	}
	return
}

func (s *Service) getKnowledgeConfig(ctx context.Context, configID int64, taskName string) (table, fieldName string, err error) {
	knowConfig, err := s.dao.RawKnowledgeConfig(ctx, configID)
	if err != nil {
		log.Errorc(ctx, "getKnowledgeConfig s.dao.RawKnowledgeConfig() configID(%d) error(%+v)", configID, err)
		return
	}
	if knowConfig.ConfigDetails == nil || knowConfig.ConfigDetails.LevelTask == nil {
		log.Errorc(ctx, "getKnowledgeConfig knowConfig is nil knowConfig(%+v)", knowConfig)
		return
	}
	table = knowConfig.ConfigDetails.Table
	for _, levelTask := range knowConfig.ConfigDetails.LevelTask {
		fieldName = getTaskName(levelTask, taskName)
		if fieldName != "" {
			return
		}
	}
	return
}

func getTaskName(tasks interface{}, taskName string) (fieldName string) {
	levelTask, ok := tasks.(*model.LevelTask)
	if !ok {
		return
	}
	for _, task := range levelTask.Tasks {
		if task.TaskName == taskName {
			fieldName = task.TaskColumn
			return
		}
	}
	return
}
