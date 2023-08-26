package task

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/task"
)

// TaskList get task list.
func (s *Service) TaskList(c context.Context, businessID, foreignID, pn, ps int64) (list []*task.Item, count int64, err error) {
	var taskList []*task.Task
	source := s.dao.DB.Model(&task.Task{})
	if businessID > 0 {
		source = source.Where("business_id = ?", businessID)
	}
	if foreignID > 0 {
		source = source.Where("foreign_id = ?", foreignID)
	}
	if err = source.Count(&count).Error; err != nil {
		log.Error("TaskList count businessID(%d) foreignID(%d) pn(%d) ps(%d) error (%v)", businessID, foreignID, pn, ps, err)
		return
	}
	if count == 0 {
		return
	}
	if err = source.Order("rank ASC,id ASC").Offset((pn - 1) * ps).Limit(ps).Find(&taskList).Error; err != nil {
		log.Error("TaskList list businessID(%d) foreignID(%d) pn(%d) ps(%d) error (%v)", businessID, foreignID, pn, ps, err)
		return
	}
	if len(taskList) > 0 {
		var (
			taskIDs   []int64
			taskRules []*task.Rule
		)
		for _, v := range taskList {
			taskIDs = append(taskIDs, v.ID)
		}
		if err = s.dao.DB.Model(&task.Rule{}).Where("task_id IN (?)", taskIDs).Find(&taskRules).Error; err != nil {
			log.Error("TaskList rule ids(%v) error (%v)", taskIDs, err)
		}
		ruleRelate := make(map[int64]*task.Rule, len(taskRules))
		for _, rule := range taskRules {
			ruleRelate[rule.TaskID] = rule
		}
		for _, v := range taskList {
			item := &task.Item{Task: v}
			if rule, ok := ruleRelate[v.ID]; ok {
				item.Rule = rule
			}
			list = append(list, item)
		}
	}
	return
}

// AddTask add task.
func (s *Service) AddTask(c context.Context, arg *task.AddArg) (err error) {
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("AddTask s.dao.DB.Begin error(%v)", err)
		return
	}
	taskAdd := &task.Task{
		Name:          arg.Name,
		BusinessID:    arg.BusinessID,
		ForeignID:     arg.ForeignID,
		Rank:          arg.Rank,
		FinishCount:   arg.FinishCount,
		Attribute:     arg.Attribute,
		CycleDuration: arg.CycleDuration,
		AwardType:     arg.AwardType,
		AwardID:       arg.AwardID,
		AwardCount:    arg.AwardCount,
		State:         arg.State,
		Stime:         arg.Stime,
		Etime:         arg.Etime,
	}
	if err = tx.Model(&task.Task{}).Create(taskAdd).Error; err != nil {
		log.Error("AddTask s.dao.DB.Model Create(%+v) error(%v)", arg, err)
		err = tx.Rollback().Error
		return
	}
	if taskAdd.HasRule() {
		taskRule := &task.Rule{
			TaskID:  taskAdd.ID,
			PreTask: arg.PreTask,
			Level:   arg.Level,
		}
		if err = tx.Model(&task.Rule{}).Create(taskRule).Error; err != nil {
			log.Error("AddTask rule s.dao.DB.Model Create(%+v) error(%v)", taskRule, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// AddTaskV2 add task.
func (s *Service) AddTaskV2(c context.Context, arg *task.AddArgV2) (err error) {
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("AddTask s.dao.DB.Begin error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("AddTaskV2 %v", r)
		}
		if err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("commit(%v)", err)
			return
		}
	}()
	taskAdd := &task.Task{
		Name:          arg.Name,
		BusinessID:    arg.BusinessID,
		ForeignID:     arg.ForeignID,
		Rank:          arg.Rank,
		FinishCount:   arg.FinishCount,
		Attribute:     arg.Attribute,
		CycleDuration: arg.CycleDuration,
		AwardType:     arg.AwardType,
		AwardID:       arg.AwardID,
		AwardCount:    arg.AwardCount,
		State:         arg.State,
		Stime:         arg.Stime,
		Etime:         arg.Etime,
	}
	if err = tx.Model(&task.Task{}).Create(taskAdd).Error; err != nil {
		log.Error("AddTask s.dao.DB.Model Create(%+v) error(%v)", arg, err)
		return
	}
	if taskAdd.HasRule() {
		for _, v := range arg.TaskRule {
			taskRule := &task.Rule{
				TaskID:    taskAdd.ID,
				PreTask:   v.PreTask,
				Level:     v.Level,
				Count:     v.Count,
				CountType: v.CountType,
				Object:    v.Object,
			}
			if err = tx.Model(&task.Rule{}).Create(taskRule).Error; err != nil {
				log.Error("AddTask rule s.dao.DB.Model Create(%+v) error(%v)", taskRule, err)
				return
			}
		}
	}
	return
}

// SaveTask update task data.
func (s *Service) SaveTask(c context.Context, arg *task.SaveArg) (err error) {
	preData := &task.Item{
		Task: new(task.Task),
	}
	if err = s.dao.DB.Where("id = ?", arg.ID).First(&preData.Task).Error; err != nil {
		log.Error("SaveTask s.dao.DB.Where id(%d) error(%d)", arg.ID, err)
		return
	}
	taskRule := new(task.Rule)
	if err = s.dao.DB.Where("task_id = ?", arg.ID).First(&taskRule).Error; err != nil {
		if err != ecode.NothingFound {
			log.Error("SaveTask s.dao.DB.Where id(%d) error(%d)", arg.ID, err)
			return
		}
	}
	if taskRule.ID > 0 {
		preData.Rule = taskRule
	}
	return s.dao.SaveTask(c, arg, preData)
}

func (s *Service) AddAward(c context.Context, arg *task.AddAward) (err error) {
	if err = s.dao.AddAward(c, arg.TaskID, 72826, arg.Award); err != nil {
		log.Info("s.dao.AddAward() error(%v)", err)
	}
	return
}
