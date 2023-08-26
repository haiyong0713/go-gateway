package knowledge

import (
	"context"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/knowledge"
	"go-gateway/app/web-svr/activity/interface/tool"
)

func (s *Service) BadgeProgress(ctx context.Context, mid, configID int64) (res *model.KnowConfigDetail, err error) {
	var finishBadge float64
	res = &model.KnowConfigDetail{}
	globalKnowConfig, ok := goingKnowledgeConfigMap[configID]
	if !ok {
		err = xecode.Errorf(xecode.RequestErr, "配制ID不存在")
		return
	}
	knowConfigDetail, err := deepCopyKnowledgeConfig(globalKnowConfig)
	if err != nil {
		log.Errorc(ctx, "BadgeProgress deepCopyKnowledgeConfig() mid(%d) config(%d) error(%+v)", mid, configID, err)
		return
	}
	// 用户未登录直接返回
	if mid == 0 {
		res = globalKnowConfig.ConfigDetails
		return
	}
	//获取用户进度
	userProgressMap, err := s.dao.UserKnowledgeTask(ctx, mid, knowConfigDetail.Table)
	if err != nil {
		log.Errorc(ctx, "BadgeProgress s.dao.UserKnowledgeTask() mid(%d) config(%d) error(%+v)", mid, configID, err)
		return
	}
	if len(userProgressMap) > 0 && userProgressMap["id"] > 0 {
		for _, levelTask := range knowConfigDetail.LevelTask {
			for _, task := range levelTask.Tasks {
				task.IsFinish = isTaskFinish(task, userProgressMap)
				if task.IsFinish {
					finishBadge++
				}
			}
		}
		knowConfigDetail.TotalProgress = getTotalProgress(finishBadge, knowConfigDetail.TotalBadge)
	}
	res = knowConfigDetail
	return
}

func getTotalProgress(finishBadge, totalBadge float64) float64 {
	if totalBadge > 0 {
		tmpTotal := (finishBadge / totalBadge) * 100
		return tool.Decimal(tmpTotal, 0)
	}
	return 0
}

func isTaskFinish(task *model.KnowTask, userProgressMap map[string]int64) bool {
	if task == nil {
		return false
	}
	progress, ok := userProgressMap[task.TaskColumn]
	if !ok {
		return false
	}
	if progress >= task.TaskFinish {
		return true
	}
	return false
}

func rebuildLevelTask(origin *model.LevelTask) (res *model.LevelTask) {
	res = new(model.LevelTask)
	res.Priority = origin.Priority
	res.ParentName = origin.ParentName
	res.ParentDescription = origin.ParentDescription
	res.Tasks = make([]*model.KnowTask, 0)
	return
}

func deepCopyKnowledgeConfig(knowledgeConfig *model.KnowConfig) (res *model.KnowConfigDetail, err error) {
	if knowledgeConfig == nil || knowledgeConfig.ConfigDetails == nil {
		err = xecode.Errorf(xecode.RequestErr, "配制不存在")
		return
	}
	if knowledgeConfig.ConfigDetails.Table == "" {
		err = xecode.Errorf(xecode.RequestErr, "配制表不能为空")
		return
	}
	configDetail := knowledgeConfig.ConfigDetails
	res = new(model.KnowConfigDetail)
	tmpDetail := new(model.KnowConfigDetail)
	{
		tmpDetail.Name = configDetail.Name
		tmpDetail.TotalBadge = configDetail.TotalBadge
		tmpDetail.Table = configDetail.Table
		tmpDetail.TableCount = configDetail.TableCount
		tmpDetail.ShareFields = configDetail.ShareFields
		tmpDetail.LevelTask = make(map[string]*model.LevelTask, 0)
		for taskName, tasks := range configDetail.LevelTask {
			tmpDetail.LevelTask[taskName] = rebuildLevelTask(tasks)
			tmpDetail.LevelTask[taskName].Tasks = deepCopyKnowTasks(tasks.Tasks)
		}
	}
	res = tmpDetail
	return
}

func deepCopyKnowTasks(tasks []*model.KnowTask) (taskList []*model.KnowTask) {
	if len(tasks) > 0 {
		for _, task := range tasks {
			tmpTask := new(model.KnowTask)
			*tmpTask = *task
			taskList = append(taskList, tmpTask)
		}
	}
	return
}

func (s *Service) BadgeShare(ctx context.Context, mid, configID int64, shareName string) (err error) {
	table, err := canShare(configID, shareName)
	if err != nil {
		return
	}
	if err = s.dao.UpdateInsertUserKnowledgeTask(ctx, shareName, mid); err != nil {
		log.Errorc(ctx, "BadgeShare s.dao.UpdateUserKnowledgeTask() mid(%d) error(%+v)", mid, err)
		return
	}
	if err = s.dao.DelCacheUserKnowledgeTask(ctx, mid, table); err != nil {
		log.Errorc(ctx, "BadgeShare s.dao.DelCacheUserKnowledgeTask() mid(%d) error(%+v)", mid, err)
		return
	}
	return
}

func canShare(configID int64, shareName string) (table string, err error) {
	var haveShareName bool
	globalKnowConfig, ok := goingKnowledgeConfigMap[configID]
	if !ok {
		err = xecode.Errorf(xecode.RequestErr, "配制ID不存在")
		return
	}
	if globalKnowConfig.ConfigDetails == nil {
		err = xecode.Errorf(xecode.RequestErr, "配制不存在")
		return
	}
	table = globalKnowConfig.ConfigDetails.Table
	if table == "" {
		err = xecode.Errorf(xecode.RequestErr, "配制表不存在")
		return
	}
	if globalKnowConfig.ConfigDetails.ShareFields == "" {
		return
	}
	fields := strings.Split(globalKnowConfig.ConfigDetails.ShareFields, ",")
	for _, field := range fields {
		if field == shareName {
			haveShareName = true
			return
		}
	}
	if !haveShareName {
		err = xecode.Errorf(xecode.RequestErr, "share_name参数值不存在")
		return
	}
	return
}

func (s *Service) BadgeConfig(ctx context.Context, jsonConfig string, configId int64) (err error) {
	if err = s.dao.UpdateKnowledgeConfig(ctx, jsonConfig, configId); err != nil {
		log.Errorc(ctx, "BadgeConfig s.dao.UpdateKnowledgeConfig() configId(%d) error(%+v)", configId, err)
		return
	}
	return
}

func (s *Service) DelKnowledgeCache(ctx context.Context, req *pb.DelKnowledgeCacheReq) (res *pb.NoReply, err error) {
	res = &pb.NoReply{}
	s.cache.Do(ctx, func(c context.Context) {
		for _, mid := range req.UpdateMids {
			if err = s.dao.DelCacheUserKnowledgeTask(ctx, mid, req.TableName); err != nil {
				log.Errorc(ctx, "DelKnowledgeCache s.dao.DelCacheUserKnowledgeTask() mid(%d) error(%+v)", mid, err)
				return
			}
		}
	})
	return
}
